package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

const eventsFile = "events.json"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return errors.New(
			"usage: mystally add <creator> <event>\nusage: mystally list",
		)
	}

	switch args[0] {
	case "add":
		if len(args) != 3 {
			return errors.New("usage: mystally add <creator> <event>")
		}

		if strings.Contains(args[1], ",") {
			return errors.New("creator name must not contain a comma")
		}

		err := addCmd(args)
		if err != nil {
			return err
		}
	case "list":
		if len(args) != 1 {
			return errors.New("usage: mystally list")
		}

		err := listCmd()
		if err != nil {
			return err
		}
	default:
		return errors.New(
			"usage: mystally add <creator> <event>\nusage: mystally list",
		)
	}

	return nil
}

func addCmd(args []string) error {
	parsedType, err := ParseEventType(args[2])
	if err != nil {
		return err
	}

	events, err := loadEvents(eventsFile)
	if err != nil {
		return err
	}
	var eventId int
	if len(events) == 0 {
		eventId = 1
	} else {
		last := events[len(events)-1]
		eventId = last.ID + 1
	}

	newEvent := Event{ID: eventId, Creator: args[1], Type: parsedType, At: time.Now()}
	events = append(events, newEvent)

	return saveEvents(eventsFile, events)
}

func listCmd() error {
	events, err := loadEvents(eventsFile)
	if err != nil {
		return err
	}

	fmt.Printf("%-15s %-15s %-15s %-15s\n", "ID", "CREATOR", "TYPE", "TIME")
	for idx := range events {
		fmt.Printf("%-15d %-15s %-15s %-15s\n", events[idx].ID, events[idx].Creator, events[idx].Type, events[idx].At.Format(time.DateTime))
	}

	return nil
}
