package main

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"
	"time"
)

const eventsFile = "events.txt"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	events := map[string]int{
		"view":    1,
		"like":    5,
		"save":    10,
		"comment": 20,
		"share":   50,
	}
	eventKeys := slices.Sorted(maps.Keys(events))

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

		_, ok := events[args[2]]
		if !ok {
			return fmt.Errorf("usage: mystally add <creator> <event>\nunknown event %q; choose one of: %s", args[2], strings.Join(eventKeys, ", "))
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
	file, err := os.OpenFile(eventsFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	line := strings.Join([]string{args[1], args[2], time.Now().Format(time.RFC3339)}, ",")
	_, err = file.WriteString(line + "\n")

	return err
}

func listCmd() error {
	content, err := os.ReadFile(eventsFile)
	if err != nil {
		return err
	}

	lines := strings.SplitSeq(string(content), "\n")
	for line := range lines {
		if line != "" {
			fmt.Println(line)
		}
	}

	return nil
}
