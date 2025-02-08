package store

import (
	"database/sql"
	"fmt"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin"
)

type TicketStore interface {
	CreateTicket(ticket *Ticket) error
	GetTicketByID(id string) (*Ticket, error)
	UpdateTicket(ticket *Ticket) error
	DeleteTicket(id string) error
}

type SQLStore struct {
	api plugin.API
	db  *sql.DB
}

func NewSQLStore(api plugin.API) (*SQLStore, error) {
	db, err := sql.Open("sqlite3", "tickets.db")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// Initialize schema
	if err := initSchema(db); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %v", err)
	}

	return &SQLStore{
		api: api,
		db:  db,
	}, nil
}

func initSchema(db *sql.DB) error {
	// Create tables
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS tickets (
			id TEXT PRIMARY KEY,
			subject TEXT NOT NULL,
			description TEXT,
			creator_id TEXT NOT NULL,
			status TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		);
	`)
	return err
}
