package ledger

type Store interface {
	Load() ([]Event, error)
	Save([]Event) error
}
