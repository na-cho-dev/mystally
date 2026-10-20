package ledger

import (
	"fmt"
	"strings"
	"time"
)

type Kobo int64
type EventType int

type Event struct {
	ID      int       `json:"id"`
	Creator string    `json:"creator"`
	Type    EventType `json:"type"`
	At      time.Time `json:"at"`
}

const (
	Unknown EventType = iota
	View
	Like
	Save
	Comment
	Share
)

// Naira converts kobo to naira for display.
// Value receiver: it only reads kb, and Kobo is a small number.
func (kb Kobo) Naira() float64 {
	naira := float64(kb) / 100
	return naira
}

// String returns the amount in naira, example ₦12.50.
// Value receiver: it only reads kb, and Kobo is a small number.
func (kb Kobo) String() string {
	return fmt.Sprintf("₦%.2f", kb.Naira())
}

// String returns the name a user types, example "like".
// Value receiver: it only reads ev.
func (ev EventType) String() string {
	switch ev {
	case View:
		return ("view")
	case Like:
		return ("like")
	case Save:
		return ("save")
	case Comment:
		return ("comment")
	case Share:
		return ("share")
	}

	return fmt.Sprintf("EventType(%d)", int(ev))
}

// MarshalText returns the name, like "like", so JSON stores a word, not a number.
// Value reciever: it only reads ev
func (ev EventType) MarshalText() ([]byte, error) {
	return []byte(ev.String()), nil
}

// UnmarshalText reads a name such as "view" from JSON and stores the matching EventType.
// Pointer receiver: it must change the caller's variable.
func (ev *EventType) UnmarshalText(b []byte) error {
	parsed, err := ParseEventType(string(b))
	if err != nil {
		return err
	}

	*ev = parsed
	return nil
}

// ParseEventType turns a word like "share" into the constant Share, or returns an error.
func ParseEventType(s string) (EventType, error) {
	events := []string{}

	for val := View; val <= Share; val++ {
		if strings.EqualFold(val.String(), s) {
			return EventType(val), nil
		}
		events = append(events, val.String())
	}

	return 0, fmt.Errorf("unknown event type: %q; choose one of: %s", s, strings.Join(events, ", "))
}
