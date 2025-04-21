package store

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type Workstream struct {
    ID          int
    Name        string
    Code        string
    Location    string
    Description string
    Quote       string
}

type WorkstreamStore struct {
    DB *sql.DB
}

func NewWorkstreamStore(dbPath string) (*WorkstreamStore, error) {
	
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return nil, err
    }

	schema := `
    CREATE TABLE IF NOT EXISTS workstreams (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        code TEXT,
        location TEXT,
        description TEXT,
        quote TEXT
    );`
    _, err = db.Exec(schema)
    if err != nil {
        return nil, err
    }

    return &WorkstreamStore{DB: db}, nil
}

func (s *WorkstreamStore) CreateWorkstream(ws Workstream) error {
    _, err := s.DB.Exec(
        `INSERT INTO workstreams (name, code, location, description, quote) VALUES (?, ?, ?, ?, ?)`,
        ws.Name, ws.Code, ws.Location, ws.Description, ws.Quote,
    )
    return err
}

func (s *WorkstreamStore) DeleteWorkstream(id int) error {
    _, err := s.DB.Exec(`DELETE FROM workstreams WHERE id = ?`, id)
    return err
}

func (s *WorkstreamStore) GetAllWorkstreams() ([]Workstream, error) {
    rows, err := s.DB.Query(`SELECT id, name, code, location, description, quote FROM workstreams`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var workstreams []Workstream
    for rows.Next() {
        var ws Workstream
        err := rows.Scan(&ws.ID, &ws.Name, &ws.Code, &ws.Location, &ws.Description, &ws.Quote)
        if err != nil {
            return nil, err
        }
        workstreams = append(workstreams, ws)
    }
    return workstreams, nil
}