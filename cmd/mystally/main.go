package main

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/na-cho-dev/mystally/internal/ledger"
	"github.com/na-cho-dev/mystally/internal/store"
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

func createLedger() (*ledger.Ledger, error) {
	store := store.NewFileStore(eventsFile)
	ldg := ledger.NewLedger(store)
	if err := ldg.Load(); err != nil {
		return nil, err
	}

	return ldg, nil
}

func addCmd(args []string) error {
	ldg, err := createLedger()
	if err != nil {
		return err
	}

	parsedType, err := ledger.ParseEventType(args[2])
	if err != nil {
		return err
	}

	newEvent := ledger.Event{Creator: args[1], Type: parsedType, At: time.Now()}
	err = ldg.Add(newEvent)

	return err
}

func listCmd() error {
	ldg, err := createLedger()
	if err != nil {
		return err
	}

	fmt.Printf("%-15s %-15s %-15s %-15s\n", "ID", "CREATOR", "TYPE", "TIME")
	for _, ev := range ldg.Events() {
		fmt.Printf("%-15d %-15s %-15s %-15s\n", ev.ID, ev.Creator, ev.Type, ev.At.Format(time.DateTime))
	}

	return nil
}

func balanceCmd(creator string) error {
	ldg, err := createLedger()
	if err != nil {
		return err
	}

	fmt.Printf("%-15s %-15s\n", creator, ldg.BalanceFor(creator))

	return nil
}

func allBalanceCmd() error {
	ldg, err := createLedger()
	if err != nil {
		return err
	}

	b := ldg.Balances()
	for _, name := range slices.Sorted(maps.Keys(b)) {
		fmt.Printf("%-15s %-15s\n", name, b[name])
	}

	return nil
}
