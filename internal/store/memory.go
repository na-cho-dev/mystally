package store

import (
	"slices"

	"github.com/na-cho-dev/mystally/internal/ledger"
)

type MemoryStore struct {
	events []ledger.Event
}

var _ ledger.Store = (*MemoryStore)(nil)

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{}
}

func (ms *MemoryStore) Load() ([]ledger.Event, error) {
	return slices.Clone(ms.events), nil
}
func (ms *MemoryStore) Save(events []ledger.Event) error {
	ms.events = slices.Clone(events)
	return nil
}
