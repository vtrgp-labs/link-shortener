package main

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"math/big"
	"strings"

	_ "modernc.org/sqlite"
)

const (
	codeLength   = 6
	codeAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	maxAttempts  = 5
)

// ErrNotFound is returned when a short code does not exist.
var ErrNotFound = errors.New("link not found")

// Link is a stored short link.
type Link struct {
	Code   string
	URL    string
	Clicks int64
}

// Store persists links in SQLite.
type Store struct {
	db *sql.DB
}

// NewStore opens (or creates) the SQLite database at path and ensures the schema exists.
func NewStore(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	// SQLite allows a single writer; serializing connections keeps
	// concurrent click increments from failing with SQLITE_BUSY.
	db.SetMaxOpenConns(1)

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS links (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		code       TEXT NOT NULL UNIQUE,
		url        TEXT NOT NULL,
		clicks     INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

// Close closes the underlying database.
func (s *Store) Close() error {
	return s.db.Close()
}

// Create stores url under a new random short code and returns the code.
func (s *Store) Create(url string) (string, error) {
	for i := 0; i < maxAttempts; i++ {
		code, err := generateCode()
		if err != nil {
			return "", err
		}
		_, err = s.db.Exec(`INSERT INTO links (code, url) VALUES (?, ?)`, code, url)
		if err == nil {
			return code, nil
		}
		if !strings.Contains(err.Error(), "UNIQUE") {
			return "", err
		}
	}
	return "", errors.New("could not generate a unique code")
}

// Get returns the original URL for code.
func (s *Store) Get(code string) (string, error) {
	var url string
	err := s.db.QueryRow(`SELECT url FROM links WHERE code = ?`, code).Scan(&url)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	return url, err
}

// Visit atomically increments the click counter for code and returns its URL.
func (s *Store) Visit(code string) (string, error) {
	var url string
	err := s.db.QueryRow(`UPDATE links SET clicks = clicks + 1 WHERE code = ? RETURNING url`, code).Scan(&url)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	return url, err
}

// List returns all links, newest first.
func (s *Store) List() ([]Link, error) {
	rows, err := s.db.Query(`SELECT code, url, clicks FROM links ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []Link
	for rows.Next() {
		var l Link
		if err := rows.Scan(&l.Code, &l.URL, &l.Clicks); err != nil {
			return nil, err
		}
		links = append(links, l)
	}
	return links, rows.Err()
}

func generateCode() (string, error) {
	b := make([]byte, codeLength)
	max := big.NewInt(int64(len(codeAlphabet)))
	for i := range b {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		b[i] = codeAlphabet[n.Int64()]
	}
	return string(b), nil
}
