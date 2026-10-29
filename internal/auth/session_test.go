package auth

import (
	"testing"
	"time"
)

func TestSessions(t *testing.T) {
	s := NewSessions()
	a, err := s.Create("user")
	if err != nil {
		t.Fatal(err)
	}
	b, err := s.Create("user")
	if err != nil {
		t.Fatal(err)
	}
	if a == b || len(a) < 40 {
		t.Fatal("tokens must be unpredictable and unique")
	}
	if id, ok := s.Get(a); !ok || id != "user" {
		t.Fatal("valid session rejected")
	}
	if _, ok := s.Get("user"); ok {
		t.Fatal("forged session accepted")
	}
	s.Delete(a)
	if _, ok := s.Get(a); ok {
		t.Fatal("revoked session accepted")
	}
	s.mu.Lock()
	entry := s.entries[b]
	entry.expires = time.Now().Add(-time.Second)
	s.entries[b] = entry
	s.mu.Unlock()
	if _, ok := s.Get(b); ok {
		t.Fatal("expired session accepted")
	}
}
