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
			"usage: mystally add <creator> <event>\nusage: mystally list\nusage: mystally balance <creator>\nusage: mystally balances",
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

		if err := listCmd(); err != nil {
			return err
		}
	case "balance":
		if len(args) != 2 {
			return errors.New("usage: mystally balance <creator>")
		}

		if err := balanceCmd(args[1]); err != nil {
			return err
		}
	case "balances":
		if len(args) != 1 {
			return errors.New("usage: mystally balances")
		}

		if err := allBalanceCmd(); err != nil {
			return err
		}
	default:
		return errors.New(
			"usage: mystally add <creator> <event>\nusage: mystally list\nusage: mystally balance <creator>\nusage: mystally balances",
		)
	}

	return nil
}

func createLedger() (*Ledger, error) {
	events, err := loadEvents(eventsFile)
	if err != nil {
		return nil, err
	}
	return newLedger(events), nil
}

func addCmd(args []string) error {
	ledger, err := createLedger()
	if err != nil {
		return err
	}

	parsedType, err := ParseEventType(args[2])
	if err != nil {
		return err
	}

	newEvent := Event{Creator: args[1], Type: parsedType, At: time.Now()}
	ledger.Add(newEvent)

	return saveEvents(eventsFile, ledger.Events())
}

func listCmd() error {
	ledger, err := createLedger()
	if err != nil {
		return err
	}

	fmt.Printf("%-15s %-15s %-15s %-15s\n", "ID", "CREATOR", "TYPE", "TIME")
	for _, ev := range ledger.Events() {
		fmt.Printf("%-15d %-15s %-15s %-15s\n", ev.ID, ev.Creator, ev.Type, ev.At.Format(time.DateTime))
	}

	return nil
}

func balanceCmd(creator string) error {
	ledger, err := createLedger()
	if err != nil {
		return err
	}

	fmt.Printf("%-15s %-15s\n", creator, ledger.BalanceFor(creator))

	return nil
}

func allBalanceCmd() error {
	ledger, err := createLedger()
	if err != nil {
		return err
	}

	b := ledger.Balances()
	for _, name := range slices.Sorted(maps.Keys(b)) {
		fmt.Printf("%-15s %-15s\n", name, b[name])
	}

	return nil
}
