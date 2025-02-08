package email

import (
    "fmt"
    "io"
    "time"
    "github.com/emersion/go-imap"
    "github.com/emersion/go-imap/client"
    "github.com/emersion/go-message/mail"
    "github.com/mattermost/mattermost-server/v6/model"
)

type EmailService struct {
    ticker    *time.Ticker
    done      chan bool
    client    *client.Client
    store     TicketStore
    api       *model.API
    config    *imapConfig
}

type imapConfig struct {
    server     string
    port       int
    username   string
    password   string
    interval   time.Duration
    SupportChannelID string
}

func NewEmailService(store TicketStore, api *model.API, config *imapConfig) *EmailService {
    return &EmailService{
        store:    store,
        api:      api,
        config:   config,
    }
}

func (es *EmailService) StartEmailPolling() error {
    es.ticker = time.NewTicker(es.config.interval)
    es.done = make(chan bool)

    if err := es.connect(); err != nil {
        return err
    }

    go es.pollLoop()
    return nil
}

func (es *EmailService) connect() error {
    c, err := client.DialTLS(
        fmt.Sprintf("%s:%d", es.config.server, es.config.port), 
        nil,
    )
    if err != nil {
        return err
    }
    es.client = c
    return es.client.Login(es.config.username, es.config.password)
}

func (es *EmailService) pollLoop() {
    for {
        select {
        case <-es.done:
            es.client.Logout()
            return
        case <-es.ticker.C:
            es.pollEmails()
        }
    }
}

func (es *EmailService) pollEmails() {
    // Select INBOX
    _, err := es.client.Select("INBOX", false)
    if err != nil {
        es.api.LogError("Failed to select mailbox", "error", err.Error())
        return
    }

    // Search for unseen messages
    criteria := imap.NewSearchCriteria()
    criteria.WithoutFlags = []string{imap.SeenFlag}
    ids, err := es.client.Search(criteria)
    if err != nil {
        es.api.LogError("Email search failed", "error", err.Error())
        return
    }

    if len(ids) == 0 {
        return
    }

    seqset := new(imap.SeqSet)
    seqset.AddNum(ids...)

    messages := make(chan *imap.Message, 10)
    go func() {
        if err := es.client.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope, imap.FetchBody}, messages);
            es.api.LogError("Message fetch failed", "error", err.Error())
        }
    }()

    for msg := range messages {
        body, err := io.ReadAll(msg.GetBody(&imap.BodySectionName{}))
        if err != nil {
            es.api.LogError("Failed to read email body", "error", err.Error())
            continue
        }

        ticket := &Ticket{
            ID:          extractTicketID(msg.Envelope.Subject),
            Subject:     msg.Envelope.Subject,
            Description: string(body),
            FromEmail:   msg.Envelope.From[0].Address(),
            Status:      "open",
            CreatedAt:   time.Now().Unix(),
        }

        if err := es.store.CreateTicket(ticket); err != nil {
            es.api.LogError("Failed to save ticket", "error", err.Error())
            continue
        }

        post := &model.Post{
            ChannelId: es.config.SupportChannelID,
            Message:   fmt.Sprintf("New ticket from %s: **%s**\n%s", ticket.FromEmail, ticket.Subject, ticket.Description),
        }
        if _, err := es.api.CreatePost(post); err != nil {
            es.api.LogError("Failed to create post", "error", err.Error())
        }
    }
}

func (es *EmailService) StopEmailPolling() {
    es.ticker.Stop()
    es.done <- true
}
