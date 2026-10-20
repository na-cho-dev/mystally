package ledger

type Ledger struct {
	events []Event
	rates  map[EventType]Kobo
	store  Store
}

func NewLedger(s Store) *Ledger {
	return &Ledger{
		rates: map[EventType]Kobo{
			View:    1,
			Like:    5,
			Save:    10,
			Comment: 20,
			Share:   50,
		},
		store: s,
	}
}

func (l *Ledger) Load() error {
	events, err := l.store.Load()
	if err != nil {
		return err
	}
	l.events = events

	return nil
}

func (l *Ledger) Add(ev Event) error {
	var eventId int
	if len(l.events) == 0 {
		eventId = 1
	} else {
		lastEvent := l.events[len(l.events)-1]
		eventId = lastEvent.ID + 1
	}
	ev.ID = eventId

	l.events = append(l.events, ev)
	err := l.store.Save(l.events)

	return err
}

func (l *Ledger) Events() []Event {
	return l.events
}

func (l *Ledger) BalanceFor(creator string) Kobo {
	var amount Kobo
	for _, ev := range l.events {
		if ev.Creator == creator {
			amount += l.rates[ev.Type]
		}
	}

	return amount
}

func (l *Ledger) Balances() map[string]Kobo {
	balances := make(map[string]Kobo)

	for _, event := range l.events {
		balances[event.Creator] += l.rates[event.Type]
	}

	return balances
}
