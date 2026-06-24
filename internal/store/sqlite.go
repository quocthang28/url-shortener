package store

import (
	"database/sql"
	"errors"
	"fmt"

	sqlite "modernc.org/sqlite"     // pure-Go SQLite driver (no cgo); registers driver name "sqlite".
	sqlite3 "modernc.org/sqlite/lib" // extended result-code constants.
)

// schema is the canonical DDL, applied by NewSQLite on startup.
const schema = `CREATE TABLE IF NOT EXISTS urls (
    short_code   TEXT PRIMARY KEY,
    original_url TEXT NOT NULL UNIQUE
);`

// SQLiteStore is the production Store backed by SQLite.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLite opens (or creates) the database at path and applies the schema.
//
// busy_timeout makes a writer wait for a held lock instead of immediately
// failing with "database is locked" (which would surface as a 500); WAL lets
// readers run concurrently with the single writer.
func NewSQLite(path string) (*SQLiteStore, error) {
	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite %q: %w", path, err)
	}
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}
	return &SQLiteStore{db: db}, nil
}

func (s *SQLiteStore) Save(shortCode, originalURL string) error {
	_, err := s.db.Exec(
		`INSERT INTO urls (short_code, original_url) VALUES (?, ?)`,
		shortCode, originalURL,
	)
	if err == nil {
		return nil
	}

	// map constraint violations to error codes
	if sqErr, ok := errors.AsType[*sqlite.Error](err); ok {
		switch sqErr.Code() {
		case sqlite3.SQLITE_CONSTRAINT_PRIMARYKEY:
			return ErrCodeTaken
		case sqlite3.SQLITE_CONSTRAINT_UNIQUE:
			return ErrURLExists
		}
	}
	return fmt.Errorf("insert url: %w", err)
}

func (s *SQLiteStore) FindByCode(shortCode string) (string, error) {
	var originalURL string

	err := s.db.QueryRow(
		`SELECT original_url FROM urls WHERE short_code = ?`, shortCode,
	).Scan(&originalURL)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	
	if err != nil {
		return "", fmt.Errorf("find by code: %w", err)
	}
	
	return originalURL, nil
}

func (s *SQLiteStore) FindByURL(originalURL string) (string, error) {
	var shortCode string
	
	err := s.db.QueryRow(
		`SELECT short_code FROM urls WHERE original_url = ?`, originalURL,
	).Scan(&shortCode)
	
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	
	if err != nil {
		return "", fmt.Errorf("find by url: %w", err)
	}
	
	return shortCode, nil
}

func (s *SQLiteStore) Close() error {
	if s.db == nil {
		return nil
	}
	
	return s.db.Close()
}
