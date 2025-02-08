package main

import (
	"fmt"
	"strings"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin"
)

func createTicketCommand() *model.Command {
	return &model.Command{
		Trigger:          "ticket",
		DisplayName:      "Support Ticket",
		Description:      "Create and manage support tickets",
		AutoComplete:     true,
		AutoCompleteHint: "[create|status|close]",
		AutoCompleteDesc: "Manage support tickets: create new tickets, check status, or close resolved issues",
	}
}

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	split := strings.Fields(args.Command)
	if len(split) < 2 {
		return &model.CommandResponse{
			Text: "Usage: /ticket [create|status|close]",
		}, nil
	}

	command := split[1]
	parameters := split[2:]

	switch command {
	case "create":
		return p.handleCreateTicket(parameters, args)
	case "status":
		return p.handleTicketStatus(parameters, args)
	case "close":
		return p.handleCloseTicket(parameters, args)
	default:
		return &model.CommandResponse{
			Text: "Unknown command. Available commands: create, status, close",
		}, nil
	}
}

func (p *Plugin) handleCreateTicket(parameters []string, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	if len(parameters) < 2 {
		return &model.CommandResponse{
			Text: "Usage: /ticket create [subject] [description]",
		}, nil
	}

	ticket := &Ticket{
		Subject:     parameters[0],
		Description: strings.Join(parameters[1:], " "),
		CreatorID:   args.UserId,
		Status:      "open",
	}

	if err := p.ticketService.CreateTicket(ticket); err != nil {
		p.API.LogError("Failed to create ticket", "error", err.Error())
		return &model.CommandResponse{
			Text: "Failed to create ticket. Please try again.",
		}, nil
	}

	return &model.CommandResponse{
		Text: fmt.Sprintf("Ticket created: %s\nStatus: %s", ticket.Subject, ticket.Status),
	}, nil
}
