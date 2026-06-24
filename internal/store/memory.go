package store

import "sync"

// MemoryStore is an in-memory Store used by tests (no disk I/O).
type MemoryStore struct {
	byCode sync.Map // short_code (string) -> original_url (string)
	byURL  sync.Map // original_url (string) -> short_code (string)
}

// NewMemory returns an empty MemoryStore. A zero sync.Map is ready to use,
// so there are no maps to pre-allocate.
func NewMemory() *MemoryStore {
	return &MemoryStore{}
}

func (m *MemoryStore) Save(shortCode, originalURL string) error {
	if _, loaded := m.byCode.LoadOrStore(shortCode, originalURL); loaded {
		return ErrCodeTaken
	}

	if _, loaded := m.byURL.LoadOrStore(originalURL, shortCode); loaded {
		m.byCode.Delete(shortCode) // release the code we just claimed
		return ErrURLExists
	}

	return nil
}

func (m *MemoryStore) FindByCode(shortCode string) (string, error) {
	if v, ok := m.byCode.Load(shortCode); ok {
		return v.(string), nil
	}
	return "", ErrNotFound
}

func (m *MemoryStore) FindByURL(originalURL string) (string, error) {
	if v, ok := m.byURL.Load(originalURL); ok {
		return v.(string), nil
	}
	
	return "", ErrNotFound
}

func (m *MemoryStore) Close() error { return nil }
