package main

import (
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"
	"time"
)

func main() {
	events := map[string]int{
		"view":    1,
		"like":    5,
		"save":    10,
		"comment": 20,
		"share":   50,
	}
	eventKeys := slices.Sorted(maps.Keys(events))
	args := os.Args
	fileName := "events.txt"

	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: mystally add <creator> <event>")
		fmt.Fprintln(os.Stderr, "usage: mystally list")
		os.Exit(1)
	}

	switch args[1] {
	case "add":
		if len(args) != 4 {
			fmt.Fprintln(os.Stderr, "usage: mystally add <creator> <event>")
			os.Exit(1)
		}

		if strings.Contains(args[2], ",") {
			fmt.Fprintln(os.Stderr, "creator name must not contain a comma")
			os.Exit(1)
		}

		_, ok := events[args[3]]
		if !ok {
			fmt.Fprintln(os.Stderr, "usage: mystally add <creator> <event>")
			fmt.Fprintf(os.Stderr, "unknown event %q; choose one of: %s\n", args[3], strings.Join(eventKeys, ", "))
			os.Exit(1)
		}

		err := add(args, fileName)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "list":
		if len(args) != 2 {
			fmt.Fprintln(os.Stderr, "usage: mystally list")
			os.Exit(1)
		}

		content, err := list(fileName)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		for _, line := range content {
			if line != "" {
				fmt.Println(line)
			}
		}
	default:
		fmt.Fprintln(os.Stderr, "usage: mystally add <creator> <event>")
		fmt.Fprintln(os.Stderr, "usage: mystally list")
		os.Exit(1)
	}
}

func add(args []string, fileName string) error {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	line := strings.Join([]string{args[2], args[3], time.Now().Format(time.RFC3339)}, ",")
	_, err = file.WriteString(line + "\n")

	return err
}

func list(fileName string) ([]string, error) {
	content, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(content), "\n")

	return lines, nil
}
