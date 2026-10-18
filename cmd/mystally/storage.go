package main

import (
	"encoding/json"
	"errors"
	"os"
)

// loadEvents returns the list of Events from the json file
func loadEvents(path string) ([]Event, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var event []Event
	err = json.Unmarshal(data, &event)

	return event, err
}

// saveEvents replaces the json file with a list of Events
func saveEvents(path string, event []Event) error {
	data, err := json.MarshalIndent(event, "", " ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}
