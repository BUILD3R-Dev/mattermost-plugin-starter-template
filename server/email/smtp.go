package email

import (
	"fmt"
	"net/smtp"
	"github.com/mattermost/mattermost-server/v6/model"
)

type SMTPService struct {
	config    *smtpConfig
	api       *model.API
}

type smtpConfig struct {
	host        string
	port        int
	username    string
	password    string
	fromAddress string
}

func NewSMTPService(api *model.API, config *smtpConfig) *SMTPService {
	return &SMTPService{
		config: config,
		api:    api,
	}
}

func (s *SMTPService) SendTicketUpdate(ticket *Ticket) error {
	auth := smtp.PlainAuth("", s.config.username, s.config.password, s.config.host)
	
	msg := fmt.Sprintf(
		"From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: [Ticket:%s] Update: %s\r\n"+
		"Content-Type: text/plain\r\n"+
		"\r\n"+
		"Your ticket '%s' has been updated:\n"+
		"Status: %s\n"+
		"Last Updated: %s\n",
		s.config.fromAddress,
		ticket.FromEmail,
		ticket.ID,
		ticket.Subject,
		ticket.Subject,
		ticket.Status,
		time.Now().Format(time.RFC1123),
	)

	err := smtp.SendMail(
		fmt.Sprintf("%s:%d", s.config.host, s.config.port),
		auth,
		s.config.fromAddress,
		[]string{ticket.FromEmail},
		[]byte(msg),
	)
	
	if err != nil {
		s.api.LogError("Failed to send ticket update email", "error", err.Error())
	}
	return err
}
