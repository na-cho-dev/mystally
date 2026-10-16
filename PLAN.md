# Mystally: a creator earnings ledger in Go

A command-line tool that records engagement events for creators, works out what each creator has earned in kobo, and reports balances. Later it grows a concurrent importer, an HTTP API, and a Postgres store.

The product is not the point. The point is that every stage forces one or two Go fundamentals, in the order real Go code uses them. Finish a stage completely before starting the next.

Module path: `github.com/<your-github-user>/mystally`
Binary name: `mystally`

---

## How to work on this

- One stage per one or two evenings. Do not skip ahead.
- No AI-written code. You may ask for a concept to be explained, or ask "why does this error happen" with the error pasted. Never ask for the code.
- No third-party packages until stage 9. The standard library covers every stage before that.
- Look things up in this order: compiler error text, `go doc <pkg>` or the editor dropdown, pkg.go.dev/std, gobyexample.com. See "Finding the right package" at the bottom.
- A stage is done only when every box in its "Done when" list is ticked and `go vet ./...` and `go test ./...` are clean.
- Commit at the end of every stage with a one-line message, for example `stage 3: ledger balances`.

## Toolchain (the whole thing)

```bash
go run ./cmd/mystally ...   # run from source
go build ./cmd/mystally     # produce the binary ./mystally
go test ./...            # run all tests
go test -race ./...      # run tests with the race detector (stage 7+)
go vet ./...             # static checks
gofmt -w .               # format everything
go doc os.ReadFile       # read docs for one function
go doc -all strings      # read a whole package
```

Install the Go extension in VS Code. Turn on format-on-save. It will show compile errors inline and give you the `os.` dropdown.

## Domain

Event types and the rate each one pays, in kobo (100 kobo = 1 naira):

| Event   | Rate (kobo) |
|---------|-------------|
| view    | 1           |
| like    | 5           |
| save    | 10          |
| comment | 20          |
| share   | 50          |

An **event** is: an id, a creator username, an event type, and a timestamp.
A **ledger** holds events and can answer: what has creator X earned, and what has everyone earned.

## Target layout (you arrive here by stage 4, not before)

```
mystally/
  go.mod
  PLAN.md
  cmd/mystally/main.go        the CLI: parse args, call the ledger, print
  internal/ledger/         Event, EventType, Kobo, Ledger, rates, Store interface
  internal/store/          FileStore, MemoryStore, later PostgresStore
  internal/httpapi/        stage 8 handlers and middleware
  testdata/                sample JSON files for tests and imports
```

`internal/` is a Go convention: packages under it cannot be imported from outside this module.

---

## Stage 0: setup

**Goal:** a running hello program, a git repo, and the doc tools in your hands.

**Build**
- `go mod init github.com/<you>/mystally`
- `cmd/mystally/main.go` that prints `mystally 0.1`
- `git init`, a `.gitignore` containing the binary name `mystally`
- Run `go doc os`, `go doc strings`, `go doc fmt.Println` and read what comes back

**Learn:** modules, `package main`, `func main`, how `go run` finds the program.

**Done when**
- [ ] `go run ./cmd/mystally` prints the version line
- [ ] `go build ./cmd/mystally && ./mystally` does the same
- [ ] first commit exists

---

## Stage 1: a CLI that reads and writes a text file

**Goal:** the smallest useful program. Arguments in, file on disk, lines out.

**Build**
- `mystally add <creator> <event>` appends one line `creator,event,timestamp` to `events.txt`
- `mystally list` prints every line, one per row
- Any bad input prints a message to stderr and exits with code 1
- Running with no command prints usage

**Learn:** `os.Args`, `os.ReadFile`, `os.OpenFile` with append, `strings.Split`, `strings.Join`, `time.Now`, `for ... range`, slices, the `if err != nil { return err }` shape, `os.Exit`.

**Look in:** `os`, `strings`, `fmt`, `time`.

**Done when**
- [ ] three `add` calls then `list` shows three rows
- [ ] `mystally add` with a missing argument exits 1 with a clear message
- [ ] `mystally list` on a missing file does not crash
- [ ] `go vet ./...` is clean

---

## Stage 2: real types and JSON

**Goal:** replace strings with a model that the compiler can check.

**Build**
- `type Kobo int64` with a `Naira() float64` method and a `String()` method that prints `₦12.50`
- `type EventType int` with `iota` constants for the five event types, a `String()` method, and a `ParseEventType(s string) (EventType, error)` function
- `type Event struct` with `ID`, `Creator`, `Type`, `At`, with JSON tags
- Storage moves from `events.txt` to `events.json`: a JSON array of events
- `mystally list` prints an aligned table: id, creator, type, time

**Learn:** named types, methods and receivers, `iota`, struct tags, `encoding/json` Marshal and Unmarshal, `fmt.Sprintf` verbs like `%-10s` and `%6d`, `time.Time` and `time.RFC3339`.

**Look in:** `encoding/json`, `fmt`, `time`, `strconv`.

**Done when**
- [ ] `events.json` is readable by eye and round-trips through load and save
- [ ] `mystally add nacho like` works; `mystally add nacho dance` fails with "unknown event type"
- [ ] `fmt.Println(Kobo(1250))` prints `₦12.50`
- [ ] every method receiver choice (value or pointer) has a one-line comment saying why

---

## Stage 3: the ledger

**Goal:** put the logic in one struct with methods, away from the CLI.

**Build**
- `type Ledger struct` holding a slice of events
- `func (l *Ledger) Add(e Event)`
- `func (l *Ledger) BalanceFor(creator string) Kobo`
- `func (l *Ledger) Balances() map[string]Kobo`
- rates live in a `map[EventType]Kobo`
- `mystally balances` prints every creator and their balance, sorted by creator name
- `mystally balance <creator>` prints one

**Learn:** pointer receivers, maps, `slices.Sort`, `maps.Keys`, the zero value of a map (nil, cannot be written to, must be made).

**Look in:** `maps`, `slices`, `sort`.

**Done when**
- [ ] `main.go` no longer contains any arithmetic; it only parses args and prints
- [ ] balances output is deterministic (sorted) on every run
- [ ] an unknown creator reports a zero balance, not a crash

---

## Stage 4: packages and the Store interface

**Goal:** split the program into packages and introduce the first interface.

**Build**
- Move model and logic into `internal/ledger`
- Declare in that package: `type Store interface { Load() ([]Event, error); Save([]Event) error }`
- Implement `FileStore` in `internal/store` backed by `events.json`
- Implement `MemoryStore` in `internal/store` backed by a slice
- `Ledger` gets a `NewLedger(s Store) *Ledger` constructor and uses the store to load and save
- `main` wires `FileStore` into `NewLedger` by hand
- Add `var _ ledger.Store = (*FileStore)(nil)` next to each implementation

**Learn:** packages and imports, exported vs unexported names, import cycles and how to break them, interfaces declared by the consumer, constructors, the compile-time interface check.

**Look in:** the Go spec section on exported identifiers, `go doc` for your own packages (`go doc ./internal/ledger`).

**Done when**
- [ ] `go build ./...` succeeds with the three packages
- [ ] `internal/store` imports `internal/ledger`, never the reverse
- [ ] swapping `FileStore` for `MemoryStore` in `main` changes nothing else

---

## Stage 5: tests

**Goal:** make the ledger provably correct and learn `go test`.

**Build**
- `internal/ledger/ledger_test.go` with table-driven tests for `BalanceFor` and `Balances`, using `MemoryStore`
- A test for `Kobo.Naira` and `Kobo.String` with at least five cases
- A test for `ParseEventType` including an invalid input
- A test that `FileStore` round-trips, using `t.TempDir()`

**Learn:** `testing.T`, `t.Run`, `t.Errorf` vs `t.Fatalf`, anonymous structs for tables, `t.TempDir`, `go test -v`, `go test -cover`, `go test -run Name`.

**Look in:** `testing`, `path/filepath`.

**Done when**
- [ ] `go test ./...` passes
- [ ] changing one rate in the rates map makes a test fail
- [ ] coverage of `internal/ledger` is above 80%

---

## Stage 6: errors done properly

**Goal:** errors that callers can inspect, not just print.

**Build**
- `var ErrUnknownEventType = errors.New("unknown event type")` returned from `ParseEventType`
- `type ValidationError struct { Field, Reason string }` with an `Error()` method, returned when a creator name is empty or an event has no timestamp
- Every error returned from `FileStore` is wrapped: `fmt.Errorf("load events: %w", err)`
- In `main`, use `errors.Is` to detect `ErrUnknownEventType` (exit 2) and `errors.As` to detect `ValidationError` (exit 3); everything else exits 1
- One helper `fail(err error)` in `main` that does the classification and printing

**Learn:** sentinel errors, custom error types, `%w` wrapping, `errors.Is`, `errors.As`, `errors.Unwrap`, the nil-interface trap (return literal `nil`, never a typed nil pointer).

**Look in:** `errors`, `fmt`.

**Done when**
- [ ] a missing file, a bad event name, and an empty creator produce three different messages and exit codes
- [ ] `errors.Is(err, os.ErrNotExist)` works through your wrapping
- [ ] a test asserts the wrapped error with `errors.Is`

---

## Stage 7: concurrency

**Goal:** goroutines, channels, WaitGroup, context, the race detector.

**Build**
- `mystally import <dir>` reads every `*.json` file in the directory concurrently, one goroutine per file, and merges all events into the ledger
- Results come back on a channel as a struct `{events []Event; err error}`; a `sync.WaitGroup` closes the channel when all goroutines finish
- The import takes a `context.Context` with a 10 second timeout; a goroutine checks `ctx.Done()` before starting work and the file read respects it
- Ctrl-C cancels the import cleanly via `signal.NotifyContext`
- `mystally import --workers N` limits concurrency with a buffered channel used as a semaphore
- Put eight sample files in `testdata/import/`

**Learn:** `go`, `chan`, `make`, `close`, `range` over a channel, `select`, `sync.WaitGroup`, `context.WithTimeout`, `defer cancel()`, `signal.NotifyContext`, `go run -race`, goroutine leaks and deadlocks.

**Look in:** `sync`, `context`, `os/signal`, `path/filepath` (`Glob`).

**Done when**
- [ ] importing eight files is visibly faster than reading them in a loop (add a timing print)
- [ ] a one millisecond timeout aborts with `context deadline exceeded` and no goroutine leak
- [ ] `go test -race ./...` and `go run -race ./cmd/mystally import testdata/import` are silent
- [ ] you can explain in one sentence who closes the results channel and why it is safe

---

## Stage 8: HTTP API

**Goal:** the same ledger behind a server, using only `net/http`.

**Build**
- `mystally serve --addr :8080`
- `GET /balances` returns JSON `{"nacho": 1250, ...}`
- `GET /creators/{name}/balance` returns `{"creator": "nacho", "kobo": 1250, "naira": 12.5}`
- `POST /events` accepts an event as JSON, validates it, saves, returns 201; bad JSON returns 400 with the error message as JSON
- A logging middleware that prints method, path, status, and duration
- Pass `r.Context()` into every ledger call
- Graceful shutdown on Ctrl-C using `http.Server.Shutdown`
- Handler tests with `net/http/httptest` and the memory store

**Learn:** `http.HandleFunc` and `http.ServeMux` with method and path patterns (Go 1.22+), `http.Handler` and `http.HandlerFunc` (the function-type trick), middleware as `func(http.Handler) http.Handler`, `json.NewEncoder(w)`, status codes, `httptest.NewRecorder`, `http.Server` with timeouts.

**Look in:** `net/http`, `net/http/httptest`, `encoding/json`, `log/slog`.

**Done when**
- [ ] `curl localhost:8080/balances` returns JSON
- [ ] `curl -X POST -d '{"creator":"nacho","type":"like"}' localhost:8080/events` adds an event
- [ ] each handler has at least one test and they run without a real server or file
- [ ] Ctrl-C finishes in-flight requests then exits

---

## Stage 9: Postgres, generics, a binary

**Goal:** the first third-party package, and a store swap that touches nothing else.

**Build**
- `PostgresStore` in `internal/store` using `github.com/jackc/pgx/v5`, selected when `MYSTALLY_DATABASE_URL` is set
- Schema in a plain `schema.sql` applied by hand with `psql`
- The same store tests as stage 5 run against Postgres when the env var is set, and are skipped otherwise (`t.Skip`)
- One generic helper that is genuinely useful here, for example `func Map[T, U any](xs []T, f func(T) U) []U` or `SumBy`
- `go build -o mystally ./cmd/mystally` and run the binary on another machine or in a container
- `go doc` your own module and fix every exported name that lacks a doc comment

**Learn:** `go get`, `go.sum`, reading a third-party package on pkg.go.dev, connection pools, `t.Skip`, type parameters and constraints, cross-compiling with `GOOS` and `GOARCH`.

**Look in:** pkg.go.dev for pgx (check importers and last release), `go help build`.

**Done when**
- [ ] `MYSTALLY_DATABASE_URL=... go test ./...` passes against a real database
- [ ] nothing in `internal/ledger`, `cmd`, or `internal/httpapi` changed to add Postgres
- [ ] a single binary runs with no Go toolchain present

---

## The graduation test

After stage 9, delete the repo (keep this file) and rebuild Mystally from scratch in one sitting, stages 1 through 8, without looking at the first version. If it takes an afternoon and the tests pass, you know Go fundamentals. The rewrite will be about half the size of the original. That difference is what you learned.

---

## Finding the right package without AI

1. Say the task as a sentence. "Read the whole file." "Split this line on commas." "Turn this string into an int."
2. Pick the noun. File, string, number, JSON, time, HTTP. The noun names the package:

   | noun | package |
   |---|---|
   | print, format | `fmt` |
   | file, env var, args, exit | `os` |
   | string | `strings` |
   | string to number | `strconv` |
   | JSON | `encoding/json` |
   | time, duration | `time` |
   | error | `errors` |
   | path | `path/filepath` |
   | slice helpers | `slices` |
   | map helpers | `maps` |
   | HTTP | `net/http` |
   | cancel, deadline | `context` |
   | goroutine wait, lock | `sync` |
   | stream of bytes | `io`, `bufio` |
   | test | `testing` |

   The full official list is pkg.go.dev/std. Use find-in-page with the noun.
3. Type the package name and a dot in the editor, or run `go doc <pkg>`. Scan the names for one that matches your sentence.
4. Read the signature. `ReadFile(name string) ([]byte, error)` tells you the inputs, the output, and that it can fail.
5. If no name matches, rephrase one level lower. "Append to a file" becomes "open a file in append mode", and `os.OpenFile` with `os.O_APPEND` appears.
6. If the noun is not in `std` at all (Postgres, Redis, Stripe), search pkg.go.dev and judge by importer count and last release date.

## Stuck protocol

1. Read the whole compiler error. It usually names the fix.
2. `go doc` the function you are calling and compare its signature to your call.
3. Find the topic on gobyexample.com and read the one page.
4. Write the smallest separate `main.go` in a scratch folder that reproduces the problem.
5. Only then ask, and ask with the error text, not for the code.

## Things to avoid

- `io/ioutil` (deprecated; use `os` and `io`)
- `sort.Slice` when `slices.SortFunc` exists
- `panic` for anything a user can cause
- `init()` functions and package-level mutable variables
- `any` and `map[string]any` where a struct will do
- A framework for HTTP, logging, or testing before you have felt the limit of the standard library

## Resources (in order of usefulness)

- pkg.go.dev/std for which package, and each package's overview for how
- gobyexample.com for a runnable program per topic
- go.dev/tour for syntax, once, early
- "Learning Go" by Jon Bodner, second edition, read alongside stages 2 through 7
- "100 Go Mistakes and How to Avoid Them" after stage 7
