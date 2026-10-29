package auth

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"
)

const SessionTTL = 12 * time.Hour

type session struct {
	userID  string
	expires time.Time
}

// Sessions holds opaque random tokens server-side. Restarting the app logs everyone out.
// A multi-instance deployment needs a shared session store instead.
type Sessions struct {
	mu      sync.Mutex
	entries map[string]session
}

func NewSessions() *Sessions { return &Sessions{entries: make(map[string]session)} }
func (s *Sessions) Create(userID string) (string, error) {
	var bytes [32]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}
	token := base64.RawURLEncoding.EncodeToString(bytes[:])
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for key, value := range s.entries {
		if !now.Before(value.expires) {
			delete(s.entries, key)
		}
	}
	s.entries[token] = session{userID: userID, expires: now.Add(SessionTTL)}
	return token, nil
}
func (s *Sessions) Get(token string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	value, ok := s.entries[token]
	if !ok || !time.Now().Before(value.expires) {
		delete(s.entries, token)
		return "", false
	}
	return value.userID, true
}
func (s *Sessions) Delete(token string) { s.mu.Lock(); defer s.mu.Unlock(); delete(s.entries, token) }
