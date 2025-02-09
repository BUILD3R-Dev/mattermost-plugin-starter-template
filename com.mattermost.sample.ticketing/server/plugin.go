// Updated import block with additional dependencies.
package main

import (
	"net/http"
	"sync"
	"time"
	"encoding/json"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost-plugin-starter-template/server/command"
	"github.com/mattermost/mattermost-plugin-starter-template/server/store/kvstore"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/mattermost/mattermost/server/public/pluginapi"
	"github.com/mattermost/mattermost/server/public/pluginapi/cluster"
	"github.com/pkg/errors"
)

// Ticket represents a support ticket.
type Ticket struct {
	ID          string   `json:"id"`
	Subject     string   `json:"subject"`
	Description string   `json:"description"`
	Comments    []string `json:"comments,omitempty"`
	CreatedAt   int64    `json:"created_at"`
}

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin

	// kvstore is the client used to read/write KV records for this plugin.
	kvstore kvstore.KVStore

	// client is the Mattermost server API client.
	client *pluginapi.Client

	// commandClient is the client used to register and execute slash commands.
	commandClient command.Command

	backgroundJob *cluster.Job

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration
}

// OnActivate is invoked when the plugin is activated. If an error is returned, the plugin will be deactivated.
func (p *Plugin) OnActivate() error {
	p.client = pluginapi.NewClient(p.API, p.Driver)

	p.kvstore = kvstore.NewKVStore(p.client)

	p.commandClient = command.NewCommandHandler(p.client)

	job, err := cluster.Schedule(
		p.API,
		"BackgroundJob",
		cluster.MakeWaitForRoundedInterval(1*time.Hour),
		p.runJob,
	)
	if err != nil {
		return errors.Wrap(err, "failed to schedule background job")
	}

	p.backgroundJob = job

	return nil
}

// OnDeactivate is invoked when the plugin is deactivated.
func (p *Plugin) OnDeactivate() error {
	if p.backgroundJob != nil {
		if err := p.backgroundJob.Close(); err != nil {
			p.API.LogError("Failed to close background job", "err", err)
		}
	}
	return nil
}

// This will execute the commands that were registered in the NewCommandHandler function.
func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	response, err := p.commandClient.Handle(args)
	if err != nil {
		return nil, model.NewAppError("ExecuteCommand", "plugin.command.execute_command.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	return response, nil
}

// See https://developers.mattermost.com/extend/plugins/server/reference/
// Ticket management endpoints

// CreateTicket handles ticket creation request.
func (p *Plugin) CreateTicket(w http.ResponseWriter, r *http.Request) {
	var ticket Ticket
	if err := json.NewDecoder(r.Body).Decode(&ticket); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	ticket.ID = strconv.FormatInt(time.Now().UnixNano(), 10)
	ticket.CreatedAt = time.Now().Unix()
	var tickets []Ticket
	data, appErr := p.kvstore.Get("tickets")
	if appErr == nil && data != nil {
		json.Unmarshal([]byte(data.(string)), &tickets)
	}
	tickets = append(tickets, ticket)
	b, err := json.Marshal(tickets)
	if err != nil {
		http.Error(w, "Error saving ticket", http.StatusInternalServerError)
		return
	}
	if err := p.kvstore.Set("tickets", string(b)); err != nil {
		http.Error(w, "Error saving ticket", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ticket)
}

// UpdateTicket handles ticket update request.
func (p *Plugin) UpdateTicket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ticketId := vars["ticketId"]

	var updatedData Ticket
	if err := json.NewDecoder(r.Body).Decode(&updatedData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Retrieve existing tickets from the KV store.
	var tickets []Ticket
	data, appErr := p.kvstore.Get("tickets")
	if appErr == nil && data != nil {
		json.Unmarshal([]byte(data.(string)), &tickets)
	}

	updated := false
	var updatedTicket Ticket
	for i, t := range tickets {
		if t.ID == ticketId {
			// Update fields if provided.
			if updatedData.Subject != "" {
				tickets[i].Subject = updatedData.Subject
			}
			if updatedData.Description != "" {
				tickets[i].Description = updatedData.Description
			}
			updated = true
			updatedTicket = tickets[i]
			break
		}
	}

	if !updated {
		http.Error(w, "Ticket not found", http.StatusNotFound)
		return
	}

	// Save the updated tickets list back to KV store.
	b, err := json.Marshal(tickets)
	if err != nil {
		http.Error(w, "Error updating ticket", http.StatusInternalServerError)
		return
	}
	if err := p.kvstore.Set("tickets", string(b)); err != nil {
		http.Error(w, "Error updating ticket", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedTicket)
}

// ListTickets handles ticket listing request.
func (p *Plugin) ListTickets(w http.ResponseWriter, r *http.Request) {
	var tickets []Ticket
	data, appErr := p.kvstore.Get("tickets")
	if appErr == nil && data != nil {
		json.Unmarshal([]byte(data.(string)), &tickets)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tickets)
}

// DeleteTicket handles ticket deletion request.
func (p *Plugin) DeleteTicket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ticketId := vars["ticketId"]

	var tickets []Ticket
	data, appErr := p.kvstore.Get("tickets")
	if appErr == nil && data != nil {
		json.Unmarshal([]byte(data.(string)), &tickets)
	}

	newTickets := make([]Ticket, 0)
	found := false
	for _, t := range tickets {
		if t.ID == ticketId {
			found = true
		} else {
			newTickets = append(newTickets, t)
		}
	}
	if !found {
		http.Error(w, "Ticket not found", http.StatusNotFound)
		return
	}

	b, err := json.Marshal(newTickets)
	if err != nil {
		http.Error(w, "Error deleting ticket", http.StatusInternalServerError)
		return
	}
	if err := p.kvstore.Set("tickets", string(b)); err != nil {
		http.Error(w, "Error deleting ticket", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetTicket handles retrieving a single ticket by its ID.
func (p *Plugin) GetTicket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ticketId := vars["ticketId"]

	var tickets []Ticket
	data, appErr := p.kvstore.Get("tickets")
	if appErr == nil && data != nil {
		json.Unmarshal([]byte(data.(string)), &tickets)
	}
	for _, t := range tickets {
		if t.ID == ticketId {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(t)
			return
		}
	}
	http.Error(w, "Ticket not found", http.StatusNotFound)
}

// AddComment handles adding a comment to an existing ticket.
func (p *Plugin) AddComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ticketId := vars["ticketId"]

	type CommentRequest struct {
		Comment string `json:"comment"`
	}
	var req CommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var tickets []Ticket
	data, appErr := p.kvstore.Get("tickets")
	if appErr == nil && data != nil {
		json.Unmarshal([]byte(data.(string)), &tickets)
	}

	updated := false
	var updatedTicket Ticket
	for i, t := range tickets {
		if t.ID == ticketId {
			tickets[i].Comments = append(tickets[i].Comments, req.Comment)
			updated = true
			updatedTicket = tickets[i]
			break
		}
	}

	if !updated {
		http.Error(w, "Ticket not found", http.StatusNotFound)
		return
	}

	b, err := json.Marshal(tickets)
	if err != nil {
		http.Error(w, "Error adding comment", http.StatusInternalServerError)
		return
	}
	if err := p.kvstore.Set("tickets", string(b)); err != nil {
		http.Error(w, "Error adding comment", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedTicket)
}

// ProcessEmails handles processing of incoming emails.
func (p *Plugin) ProcessEmails(w http.ResponseWriter, r *http.Request) {
	// Define a structure for incoming email messages.
	type EmailMessage struct {
		MessageID string `json:"message_id"`
		From      string `json:"from"`
		Subject   string `json:"subject"`
		Body      string `json:"body"`
		CreatedAt int64  `json:"created_at"`
	}

	var email EmailMessage
	if err := json.NewDecoder(r.Body).Decode(&email); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Retrieve existing tickets.
	var tickets []Ticket
	data, appErr := p.kvstore.Get("tickets")
	if appErr == nil && data != nil {
		json.Unmarshal([]byte(data.(string)), &tickets)
	}

	// Check if a ticket with the email's subject already exists.
	var existingTicket *Ticket
	for i, t := range tickets {
		if t.Subject == email.Subject {
			existingTicket = &tickets[i]
			break
		}
	}

	if existingTicket != nil {
		// Append email content to the ticket's description.
		existingTicket.Description += "\n\nEmail from " + email.From + ": " + email.Body
	} else {
		// Create a new ticket from the email.
		newTicket := Ticket{
			Subject:     email.Subject,
			Description: "Email from " + email.From + ":\n" + email.Body,
			ID:          strconv.FormatInt(time.Now().UnixNano(), 10),
			CreatedAt:   time.Now().Unix(),
		}
		tickets = append(tickets, newTicket)
		existingTicket = &newTicket
	}

	// Save the updated tickets list back to KV store.
	b, err := json.Marshal(tickets)
	if err != nil {
		http.Error(w, "Error processing email", http.StatusInternalServerError)
		return
	}
	if err := p.kvstore.Set("tickets", string(b)); err != nil {
		http.Error(w, "Error processing email", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existingTicket)
}
