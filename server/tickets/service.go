package main

import (
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-plugin-starter-template/server/email"
)

type TicketService struct {
	store       TicketStore
	smtpService *email.SMTPService
}

func NewTicketService(store TicketStore, smtpService *email.SMTPService) *TicketService {
	return &TicketService{
		store:       store,
		smtpService: smtpService,
	}
}

func (s *TicketService) UpdateTicketStatus(id string, status string) error {
	ticket, err := s.store.GetTicketByID(id)
	if err != nil {
		return err
	}

	ticket.Status = status
	if err := s.store.UpdateTicketStatus(id, status); err != nil {
		return err
	}

	// Send email update
	if err := s.smtpService.SendTicketUpdate(ticket); err != nil {
		model.LogError("Failed to send ticket update email", "ticket_id", id, "error", err.Error())
	}
	return nil
}
