package main

import (
	"net/http"
	"sync"
	"time"

	"github.com/mattermost/mattermost-plugin-starter-template/server/command"
	"github.com/mattermost/mattermost-plugin-starter-template/server/store/kvstore"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/mattermost/mattermost/server/public/pluginapi"
	"github.com/mattermost/mattermost/server/public/pluginapi/cluster"
	"github.com/pkg/errors"
)

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

	// Add our services
	ticketStore  TicketStore
	emailService EmailService
	smtpService  SMTPService
	ticketService TicketService
}

// TicketStore handles database operations
type TicketStore interface {
	CreateTicket(ticket *Ticket) error
	GetTicketByID(id string) (*Ticket, error)
	UpdateTicketStatus(id string, status string) error
}

// EmailService handles email synchronization
type EmailService interface {
	StartEmailPolling() error
	StopEmailPolling()
}

// SMTPService handles SMTP operations
type SMTPService interface {
	SendEmail(to string, subject string, body string) error
}

// TicketService handles ticket operations
type TicketService interface {
	// Add methods for ticket service
}

// OnActivate is invoked when the plugin is activated. If an error is returned, the plugin will be deactivated.
func (p *Plugin) OnActivate() error {
	p.client = pluginapi.NewClient(p.API, p.Driver)

	p.kvstore = kvstore.NewKVStore(p.client)

	p.commandClient = command.NewCommandHandler(p.client)

	// Initialize store
	store, err := store.NewSQLStore(p.API)
	if err != nil {
		return fmt.Errorf("failed to initialize store: %v", err)
	}
	p.ticketStore = store

	// Initialize email services
	emailConfig := &email.imapConfig{
		server:    p.configuration.EmailServer,
		port:      p.configuration.EmailPort,
		username:  p.configuration.EmailUsername,
		password:  p.configuration.EmailPassword,
		interval:  5 * time.Minute,
	}
	p.emailService = email.NewEmailService(p.ticketStore, p.API, emailConfig)

	smtpConfig := &email.smtpConfig{
		host:        p.configuration.SMTPHost,
		port:        p.configuration.SMTPPort,
		username:    p.configuration.SMTPUsername,
		password:    p.configuration.SMTPPassword,
		fromAddress: p.configuration.FromAddress,
	}
	p.smtpService = email.NewSMTPService(p.API, smtpConfig)

	// Initialize ticket service
	p.ticketService = tickets.NewTicketService(
		p.ticketStore,
		p.smtpService,
		p.API,
	)

	// Register commands
	if err := p.API.RegisterCommand(createTicketCommand()); err != nil {
		return fmt.Errorf("failed to register command: %v", err)
	}

	// Start email polling
	if err := p.emailService.StartEmailPolling(); err != nil {
		return fmt.Errorf("failed to start email service: %v", err)
	}

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
