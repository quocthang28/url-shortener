package store

import "errors"

// Sentinel errors returned by Store implementations.
var (
	// ErrNotFound is returned by the Find* methods when no row matches.
	ErrNotFound = errors.New("not found")
	// ErrCodeTaken is returned by Save when the short_code already exists
	// (a generation collision — the caller should retry with a new code).
	ErrCodeTaken = errors.New("short code already exists")
	// ErrURLExists is returned by Save when the original_url already exists
	// (a concurrent encode of the same URL won — the caller should re-read by URL).
	ErrURLExists = errors.New("original url already exists")
)

// Store persists the mapping between short codes and original URLs.
// Implementations must be safe for concurrent use.
type Store interface {
	// Save inserts a new mapping. Returns ErrCodeTaken or ErrURLExists on conflict.
	Save(shortCode, originalURL string) error
	// FindByCode returns the original URL for a short code, or ErrNotFound.
	FindByCode(shortCode string) (string, error)
	// FindByURL returns the existing short code for a URL, or ErrNotFound.
	FindByURL(originalURL string) (string, error)
	// Close releases underlying resources.
	Close() error
}
