package audit

import (
	"sync"
	"time"
)

type Event struct {
	ID        string    `json:"id"`
	Actor     string    `json:"actor"`
	Action    string    `json:"action"`
	Target    string    `json:"target"`
	CreatedAt time.Time `json:"createdAt"`
}

type Store struct {
	mu     sync.RWMutex
	events []Event
}

func NewStore() *Store { return &Store{events: make([]Event, 0, 256)} }

func (s *Store) Add(e Event) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append([]Event{e}, s.events...)
	if len(s.events) > 1000 {
		s.events = s.events[:1000]
	}
}

func (s *Store) List() []Event {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Event, len(s.events))
	copy(out, s.events)
	return out
}
