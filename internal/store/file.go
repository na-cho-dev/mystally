// Package store holds the places events can be kept.
package store

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/na-cho-dev/mystally/internal/ledger"
)

type FileStore struct {
	path string
}

var _ ledger.Store = (*FileStore)(nil)

func NewFileStore(path string) *FileStore {
	return &FileStore{path: path}
}

// Load returns the list of Events from the json file
func (fs *FileStore) Load() ([]ledger.Event, error) {
	data, err := os.ReadFile(fs.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var event []ledger.Event
	err = json.Unmarshal(data, &event)

	return event, err
}

// Save replaces the json file with a list of Events
func (fs *FileStore) Save(event []ledger.Event) error {
	data, err := json.MarshalIndent(event, "", " ")
	if err != nil {
		return err
	}

	return os.WriteFile(fs.path, data, 0o644)
}
