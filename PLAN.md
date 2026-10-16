# Mystally: a creator earnings ledger in Go

A command-line tool that records engagement events for creators, works out what each creator has earned in kobo, and reports balances. Later it grows a concurrent importer, an HTTP API, and a Postgres store.

The product is not the point. The point is that every stage forces one or two Go fundamentals, in the order real Go code uses them. Finish a stage completely before starting the next.

Module path: `github.com/na-cho-dev/mystally`
Binary name: `mystally`

## Contents

- [How to work on this](#how-to-work-on-this)
- [Toolchain](#toolchain-the-whole-thing)
- [Domain](#domain) and [Target layout](#target-layout-you-arrive-here-by-stage-4-not-before)
- [Stage 0: setup](#stage-0-setup)
- [Stage 1: a CLI that reads and writes a text file](#stage-1-a-cli-that-reads-and-writes-a-text-file)
- [Stage 2: real types and JSON](#stage-2-real-types-and-json)
- [Stage 3: the ledger](#stage-3-the-ledger)
- [Stage 4: packages and the Store interface](#stage-4-packages-and-the-store-interface)
- [Stage 5: tests](#stage-5-tests)
- [Stage 6: errors done properly](#stage-6-errors-done-properly)
- [Stage 7: concurrency](#stage-7-concurrency)
- [Stage 8: HTTP API](#stage-8-http-api)
- [Stage 9: Postgres, generics, a binary](#stage-9-postgres-generics-a-binary)
- [The graduation test](#the-graduation-test)
- [Testing by hand](#testing-by-hand) · [When the editor and the terminal disagree](#when-the-editor-and-the-terminal-disagree) · [Mistake log](#mistake-log) · [Go words used in this plan](#go-words-used-in-this-plan)
- [Finding the right package](#finding-the-right-package-without-ai) · [Stuck protocol](#stuck-protocol) · [Things to avoid](#things-to-avoid) · [Resources](#resources-in-order-of-usefulness)

---

## How to work on this

- One stage per one or two evenings. Do not skip ahead.
- No AI-written code. You may ask for a concept to be explained, or ask "why does this error happen" with the error pasted. Never ask for the code.
- No third-party packages until stage 9. The standard library covers every stage before that.
- Look things up in this order: compiler error text, `go doc <pkg>` or the editor dropdown, pkg.go.dev/std, gobyexample.com. See "Finding the right package" at the bottom.
- A stage is done only when every box in its "Done when" list is ticked and `go vet ./...` and `go test ./...` are clean.
- Commit at the end of every stage with a one-line message, for example `stage 3: ledger balances`.

### How each stage is laid out

| Part | What it gives you |
|---|---|
| **Goal / Build / Learn / Look in** | What to build and which tools it needs |
| **Structure** | Which files exist at the end of the stage, and what each function is responsible for. Names and signatures only, never bodies |
| **Examples** | Folded. The same tools used in a *different* program (books, weekdays, notes), so you translate the idea instead of copying text |
| **Gotchas** | Traps, with the real error text. Entries marked *(you hit this)* already happened |
| **Check yourself** | Questions to answer out loud or in a commit message. If you can't answer one, the stage isn't done |
| **Done when** | The gate to the next stage |

### How to use the examples

1. Read **Build** and **Structure**, then try on your own first.
2. Open an example when you're stuck on one tool, or after you finish, to compare.
3. Type it into a scratch folder **outside** this repo. Typing, not pasting, is what makes it stick:

   ```bash
   mkdir -p ~/scratch/go/stage2 && cd ~/scratch/go/stage2
   go mod init scratch
   # create main.go, then:
   go run .
   ```

4. Break it on purpose: delete a `&`, change a receiver, remove a `return`. Read the error.
5. Close the example and write mystally's version without looking.

The examples are reference material, like gobyexample.com. The "no AI-written code" rule still holds for mystally: you write every line of it.

---

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

**Commands need a path.** Without one, `go` works on the folder you're in. From the repo root, `go build` and `go vet` fail with `no Go files in .../mystally` because the code lives in `cmd/mystally`. Pass `./cmd/mystally`, or `./...` (this folder and every folder below it). Point at the **folder**, never at `main.go`: once a package has two files, `go run main.go` compiles only that one and fails with `undefined:` for names in the other.

**Exit codes.** `go run` prints `exit status 3` but always exits with 1 itself. Check exit codes on the built binary: `go build ./cmd/mystally && ./mystally ...; echo $?`.

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
  internal/importer/       stage 7 concurrent import (suggested, see stage 7 Structure)
  internal/httpapi/        stage 8 handlers and middleware
  testdata/                sample JSON files for tests and imports
```

`internal/` is a Go convention: packages under it cannot be imported from outside this module.

Each stage's **Structure** section shows the layout for that stage. It grows like this:

```
stage 0-3   cmd/mystally/ only (one package, more files each stage)
stage 4     + internal/ledger, internal/store
stage 5     + *_test.go next to the code they test
stage 7     + internal/importer, testdata/import
stage 8     + internal/httpapi
stage 9     + schema.sql, internal/store/postgres.go
```

---

## Stage 0: setup

**Goal:** a running hello program, a git repo, and the doc tools in your hands.

**Build**
- `go mod init github.com/<you>/mystally`
- `cmd/mystally/main.go` that prints `mystally 0.1`
- `git init`, a `.gitignore` containing the binary name `mystally`
- Run `go doc os`, `go doc strings`, `go doc fmt.Println` and read what comes back

**Learn:** modules, `package main`, `func main`, how `go run` finds the program.

**Structure**

```
mystally/
  .gitignore            /mystally     the leading slash means "only at the repo root"
  go.mod
  PLAN.md
  cmd/mystally/main.go
```

**Example: the anatomy of `main.go`**

```go
package main // "main" makes this a program you can run, not a library

import "fmt" // standard library: nothing to install, just import

func main() { // ( ) holds the parameters (none here), { } holds the body
	fmt.Println("mystally 0.1")
}
```

**Gotchas**
- A function body goes in `{ }`, not `( )`. `syntax error: unexpected ( in parameter list` means Go is reading your body as a parameter list. *(you hit this)*
- At the top level of a file you can only write declarations: `package`, `import`, `const`, `var`, `type`, `func`. A statement like `fmt.Println(...)` must be inside a function. Otherwise: `non-declaration statement outside function body`.
- The `{` must be on the same line as `func main()`.
- In `.gitignore`, the pattern `mystally` matches any file **or folder** with that name at any depth, including `cmd/mystally/`. Your source silently disappears from `git status`. Use `/mystally`. Check with `git check-ignore -v cmd/mystally/main.go` (no output means it's not ignored). *(you hit this)*
- Standard library packages (`os`, `fmt`, `strings`...) are never installed. If the editor says `could not import os (missing metadata for import of "os")`, see [When the editor and the terminal disagree](#when-the-editor-and-the-terminal-disagree). *(you hit this)*

**Check yourself**
- What is the difference between a module (`go.mod`) and a package (a folder of `.go` files)?
- Why does `go run ./cmd/mystally` work from the repo root but `go run main.go` doesn't?
- What does `./...` mean?

**Done when**
- [ ] `go run ./cmd/mystally` prints the version line
- [ ] `go build ./cmd/mystally && ./mystally` does the same
- [ ] `git ls-files` lists `cmd/mystally/main.go`, and `git status` does not list the `mystally` binary
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

**Structure**

Everything still lives in `cmd/mystally/main.go`, split into small functions with one job each:

| Function | Its only job |
|---|---|
| `main()` | Call `run(os.Args[1:])`. If it returns an error, print it to stderr and `os.Exit(1)`. Nothing else. |
| `run(args []string) error` | `switch` on `args[0]`, call the matching command, return a usage error for anything else |
| `addCmd(args []string) error` | Check the argument count, validate the event name, build the line, append it |
| `listCmd() error` | Read the file, print each non-empty line |
| `const eventsFile = "events.txt"` | The file name in one place. A `const` at package level is fine: it can't change |

Why a separate `run` instead of doing everything in `main`:
- `main` returns nothing, so it can't `return err`. `run` can, so the `if err != nil { return err }` shape works everywhere.
- `os.Exit` skips deferred calls. If only `main` exits, every `defer f.Close()` inside `run` still runs.
- In stage 5 you can call `run` from a test with made-up arguments.

Passing `os.Args[1:]` drops the program path, so inside `run` the command is `args[0]`, the creator `args[1]`, the event `args[2]`. Pick that convention and stick to it. Mixing `os.Args[3]` and `args[2]` in one program is a classic off-by-one.

Add `events.txt` to `.gitignore`. It's data the program makes, not source code.

**Examples**

<details>
<summary><b>One command per <code>case</code>: the run + switch shape</b></summary>

```go
package main

import (
	"errors"
	"fmt"
	"os"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "greet:", err)
		os.Exit(1)
	}
}

// run is the whole program. It returns an error instead of exiting,
// so main is the only place that prints errors and picks the exit code.
func run(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: greet hello <name> | greet version")
	}
	switch args[0] {
	case "hello":
		if len(args) != 2 {
			return errors.New("usage: greet hello <name>")
		}
		fmt.Println("hello,", args[1])
	case "version":
		fmt.Println("greet 0.1")
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
	return nil
}
```

```
$ ./greet              → greet: usage: greet hello <name> | greet version   (exit 1)
$ ./greet hello        → greet: usage: greet hello <name>                   (exit 1)
$ ./greet hello Ada    → hello, Ada                                         (exit 0)
$ ./greet dance        → greet: unknown command "dance"                     (exit 1)
```

`errors.New` and `fmt.Errorf` **create** an error value. They print nothing. `main` prints it.

</details>

<details>
<summary><b>Append one line to a file</b></summary>

```go
package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func main() {
	line := strings.Join([]string{"ada", time.Now().Format(time.RFC3339)}, ",")
	if err := appendLine("visits.log", line); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// appendLine adds one line to the end of path, creating the file if needed.
func appendLine(path, line string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0o644)
	if err != nil {
		return err
	}
	defer f.Close() // runs when appendLine returns, on every path below

	_, err = fmt.Fprintln(f, line) // Fprintln adds the "\n"
	return err
}
```

- `os.O_WRONLY|os.O_APPEND|os.O_CREATE`: `|` combines flags. Exactly one access mode (`O_WRONLY`) plus any behaviours. See `go doc os.O_APPEND`.
- `0o644`: permissions in octal, used only when the file is created. You can read and write; everyone else can read.
- A `*os.File` is something you can write to, like `os.Stderr`, so `fmt.Fprintln` works on both.

</details>

<details>
<summary><b>Read a file back as lines</b></summary>

```go
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

func main() {
	data, err := os.ReadFile("visits.log")
	if errors.Is(err, os.ErrNotExist) {
		fmt.Println("no visits yet")
		return
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	for _, line := range strings.Split(string(data), "\n") {
		if line == "" {
			continue // the file ends in "\n", so the last piece is ""
		}
		fields := strings.Split(line, ",")
		if len(fields) != 2 {
			fmt.Fprintf(os.Stderr, "skipping bad line %q\n", line)
			continue
		}
		fmt.Println(fields[0], "visited at", fields[1])
	}
}
```

```
no visits yet                                  ← before the file exists
ada visited at 2026-10-16T16:38:14+01:00
ada visited at 2026-10-16T16:38:14+01:00
skipping bad line "broken line"
```

`os.ReadFile` returns `[]byte`. `string(data)` converts it. `errors.Is(err, os.ErrNotExist)` tells "no file yet" apart from real failures. You'll learn `errors.Is` properly in stage 6.

</details>

**Gotchas**
- Indexing past the end of a slice doesn't return an error. It **panics**: `index out of range [3] with length 3`. Check `len(args)` before every index. Slicing at the end is fine: `os.Args[1:]` on a one-item slice gives an empty slice, not a panic.
- A missing argument isn't an empty slot in `os.Args`. The slice is just shorter. Looping over it can never find what isn't there. *(you hit this)*
- A `string` is never `nil`; its zero value is `""`. `val == nil` gives `mismatched types string and untyped nil`. *(you hit this)*
- `fmt.Errorf("...")` on its own line prints nothing. `go vet` says `result of fmt.Errorf call not used`. To print, use `fmt.Fprintln(os.Stderr, ...)`. *(you hit this)*
- Writing to a variable doesn't count as using it. Only reading does: `declared and not used: match`. *(you hit this)*
- To check whether a key is in a map, look it up with comma-ok: `rate, ok := rates[name]`. Don't loop over the map. *(you hit this)*
- `strings.Split("a\nb\n", "\n")` gives `["a" "b" ""]`, three pieces. `strings.Split("", "\n")` gives `[""]`, one piece. Skip empty lines.
- `fmt.Println(time.Now())` prints `2026-10-16 14:27:21.675627796 +0100 WAT m=+0.000246591`. Use `time.Now().Format(time.RFC3339)` → `2026-10-16T14:27:21+01:00`. Stage 2 parses it back with the same layout.
- `os.Exit` skips deferred calls. Only `main` should call it.
- A creator named `a,b` would break your comma-separated line. Reject names containing a comma (`go doc strings.Contains`).
- Error messages are lowercase with no final `!` or `.`: `usage: mystally add <creator> <event>`. Go errors get chained (`load events: open events.txt: no such file or directory`), and capitals or punctuation in the middle read badly.

**Test by hand:** see [Testing by hand](#testing-by-hand) for `$?`, `2>/dev/null` and the full checklist command list.

**Check yourself**
- What is in `os.Args[0]`, and why does `run` get `os.Args[1:]`?
- Why does only `main` call `os.Exit`?
- What is the difference between `x := 1` and `x = 1`? When does each one fail to compile?
- Why does splitting your file on `"\n"` give one more piece than there are lines?

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

**Structure**

```
cmd/mystally/
  main.go       the CLI: run, addCmd, listCmd, table printing. No type definitions.
  event.go      Kobo, EventType, Event, their methods, ParseEventType
  storage.go    loadEvents(path string) ([]Event, error)
                saveEvents(path string, events []Event) error
```

- Every `.go` file in one folder is the same package. They see each other's names with no import. This is why you run the **folder**: `go run main.go` would fail with `undefined: Event`.
- `add` changes shape. Stage 1 appended a line. Now it does: load the whole slice → append the new event → save the whole file with `os.WriteFile`. A missing `events.json` loads as an empty slice, not an error.
- New IDs: the highest existing ID + 1, starting at 1.
- Delete the `events.txt` code. Stage 1 lives on in git history.
- Add `events.json` to `.gitignore`.

**Examples**

<details>
<summary><b>A named type with methods: <code>Minutes</code> (the same shape <code>Kobo</code> needs)</b></summary>

```go
// Package minutes shows a named type with methods.
package minutes

import "fmt"

// Minutes is a duration in whole minutes. Storing an integer and only
// converting for display avoids float rounding errors.
type Minutes int64

// Hours converts for display only.
// Value receiver: it only reads m, and m is a small number.
func (m Minutes) Hours() float64 {
	return float64(m) / 60
}

// String makes fmt.Println(Minutes(125)) print "2h05m".
// Value receiver: same reason as Hours.
func (m Minutes) String() string {
	return fmt.Sprintf("%dh%02dm", m/60, m%60)
}
```

- `m/60` is integer division: it truncates. `m%60` is the remainder. `%02d` pads to two digits with zeros, so 5 prints as `05`.
- Any type with a `String() string` method is printed through it by `fmt.Println`, `%v` and `%s`.

</details>

<details>
<summary><b>An <code>iota</code> enum that reads and writes as text: <code>Weekday</code> (the same shape <code>EventType</code> needs)</b></summary>

```go
// Package weekday shows an iota enum that reads and writes as text.
package weekday

import (
	"errors"
	"fmt"
)

type Weekday int

const (
	Unknown   Weekday = iota // 0: the zero value is deliberately invalid
	Monday                   // 1
	Tuesday                  // 2
	Wednesday                // 3
)

var ErrUnknownWeekday = errors.New("unknown weekday")

// Value receiver: String only reads d.
func (d Weekday) String() string {
	switch d {
	case Monday:
		return "monday"
	case Tuesday:
		return "tuesday"
	case Wednesday:
		return "wednesday"
	default:
		return fmt.Sprintf("Weekday(%d)", int(d))
	}
}

func Parse(s string) (Weekday, error) {
	switch s {
	case "monday":
		return Monday, nil
	case "tuesday":
		return Tuesday, nil
	case "wednesday":
		return Wednesday, nil
	}
	return Unknown, fmt.Errorf("%w: %q", ErrUnknownWeekday, s)
}

// MarshalText makes encoding/json write "monday" instead of 1.
// Value receiver: it only reads d.
func (d Weekday) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}

// UnmarshalText makes encoding/json accept "monday".
// Pointer receiver: it must change the caller's variable.
func (d *Weekday) UnmarshalText(b []byte) error {
	parsed, err := Parse(string(b))
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}
```

- `iota` counts up from 0 inside one `const ( )` block. Lines with no `= ...` repeat the previous expression.
- `encoding/json` checks whether a type has `MarshalText`/`UnmarshalText`. If it does, the value is stored as a JSON string. That makes `events.json` readable by eye, and in stage 8 it lets `{"type":"like"}` decode with no extra code.
- `%w` in `fmt.Errorf` wraps the sentinel error. Stage 6 explains it. Using it now costs nothing.

</details>

<details>
<summary><b>Struct tags, a JSON round trip, and an aligned table</b></summary>

```go
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"example.com/learn/s2/weekday"
)

type Book struct {
	ID     int             `json:"id"`
	Title  string          `json:"title"`
	ReadOn weekday.Weekday `json:"read_on"`
	Added  time.Time       `json:"added"`
}

func main() {
	books := []Book{
		{ID: 1, Title: "Things Fall Apart", ReadOn: weekday.Monday,
			Added: time.Date(2026, 10, 1, 9, 30, 0, 0, time.UTC)},
		{ID: 2, Title: "Half of a Yellow Sun", ReadOn: weekday.Wednesday,
			Added: time.Date(2026, 10, 3, 18, 0, 0, 0, time.UTC)},
	}

	// Save: slice -> indented JSON -> file.
	data, err := json.MarshalIndent(books, "", "  ")
	if err != nil {
		fail(err)
	}
	if err := os.WriteFile("books.json", data, 0o644); err != nil {
		fail(err)
	}
	fmt.Println(string(data))

	// Load: file -> JSON -> slice. Unmarshal needs a pointer so it can fill the slice.
	raw, err := os.ReadFile("books.json")
	if err != nil {
		fail(err)
	}
	var loaded []Book
	if err := json.Unmarshal(raw, &loaded); err != nil {
		fail(err)
	}

	// An aligned table: '-' means left-align, the number is the column width.
	fmt.Printf("%-4s %-22s %-10s %s\n", "ID", "TITLE", "DAY", "ADDED")
	for _, b := range loaded {
		fmt.Printf("%-4d %-22s %-10s %s\n", b.ID, b.Title, b.ReadOn, b.Added.Format(time.DateTime))
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
```

Output:

```
[
  {
    "id": 1,
    "title": "Things Fall Apart",
    "read_on": "monday",
    "added": "2026-10-01T09:30:00Z"
  },
  ...
]
ID   TITLE                  DAY        ADDED
1    Things Fall Apart      monday     2026-10-01 09:30:00
2    Half of a Yellow Sun   wednesday  2026-10-03 18:00:00
```

- `time.Time` turns itself into RFC3339 inside JSON. You don't call `Format` for storage, only for display.
- The `%-10s` column for `b.ReadOn` calls its `String()` method.

</details>

**Gotchas**
- If `String()` has a **pointer** receiver, printing a **value** ignores it. A test program printed `1 pointer-string` for `fmt.Println(p, &p)`. Use a value receiver for `String()` on small types like `Kobo` and `EventType`.
- Integer division truncates: `1250/100` is `12` and `1250%100` is `50`. For 1205 kobo, `%02d` is what makes it `₦12.05` and not `₦12.5`.
- Never store money as `float64`. `Naira()` is for display only.
- `encoding/json` silently skips lowercase (unexported) fields. A field named `creator` is never saved and never loaded, and there's no error.
- Without `MarshalText`, `EventType` is saved as a number (`"type": 2`). That's not readable by eye, and reordering the constants later silently changes what old files mean.
- If the first `iota` constant is a real event type, an `Event` whose `Type` you forgot to set silently becomes a `view`. Making 0 an invalid value catches that mistake.
- `json.Unmarshal(data, events)` without `&`: `go vet` says `call of Unmarshal passes non-pointer as second argument`.
- An empty (0-byte) `events.json` fails with `unexpected end of JSON input`. A missing file and an empty file are different cases.
- `fmt` widths count characters, not bytes. `%-8s` pads `₦12.50` correctly even though `₦` is 3 bytes.

**Check yourself**
- Why is `Kobo` an `int64` and not a `float64`?
- For each method you wrote, why did you choose a value or a pointer receiver?
- What is your `EventType`'s zero value, and what does it mean?
- What would `events.json` look like if `EventType` had no `MarshalText`?
- Why does `json.Unmarshal` need `&events` and not `events`?

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

**Structure**

```
cmd/mystally/
  main.go       parse args → call the ledger → print. No arithmetic, no loops that add up money.
  event.go
  storage.go
  ledger.go     Ledger, newLedger, Add, BalanceFor, Balances, the rates
```

```go
type Ledger struct {
	events []Event
	rates  map[EventType]Kobo
}

func newLedger(events []Event) *Ledger // fills in rates. Lowercase until stage 4 exports it.
```

- The rates live in a `Ledger` field that `newLedger` fills in, or in a function that returns the map. Not in a package-level `var` (see Things to avoid).
- Every command follows the same flow: load events → `newLedger(events)` → call a method → save if something changed → print.
- Don't store balances anywhere. Work them out from the events every time, so the events are the single source of truth.
- A quick boundary check: `main.go` should never mention `rates`, and should never add up `Kobo` values itself.

**Examples**

<details>
<summary><b>Value vs pointer receivers, shown with <code>append</code></b></summary>

```go
package main

import "fmt"

type Shelf struct {
	books []string
}

// AddCopy has a value receiver: s is a copy, so the append is lost.
func (s Shelf) AddCopy(title string) {
	s.books = append(s.books, title)
}

// Add has a pointer receiver: s points at the caller's Shelf.
func (s *Shelf) Add(title string) {
	s.books = append(s.books, title)
}

func main() {
	var s Shelf
	s.AddCopy("Purple Hibiscus")
	fmt.Println(len(s.books)) // 0
	s.Add("Purple Hibiscus")
	fmt.Println(len(s.books)) // 1
}
```

This compiles cleanly and fails silently. Only the output shows the bug.

</details>

<details>
<summary><b>Maps: nil maps, <code>make</code>, counting, and sorted output</b></summary>

```go
package main

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

func main() {
	var broken map[string]int       // nil map: reading works, writing panics
	fmt.Println(broken["anything"]) // 0, no panic
	// broken["x"] = 1               // panic: assignment to entry in nil map

	counts := make(map[string]int) // make gives a real, writable map
	for _, w := range strings.Fields("the cat and the hat and the bat") {
		counts[w]++ // a missing key reads as 0, so ++ just works
	}

	// Map order is random on purpose. Sort the keys for stable output.
	for _, w := range slices.Sorted(maps.Keys(counts)) {
		fmt.Printf("%-4s %d\n", w, counts[w])
	}
}
```

```
0
and  2
bat  1
cat  1
hat  1
the  3
```

</details>

**Gotchas**
- `Add` with a value receiver compiles and does nothing. See the `Shelf` example.
- Writing to a nil map panics with `assignment to entry in nil map`. `Balances` must `make` its map before filling it.
- Map iteration order is random, and Go randomises it on purpose. Run `balances` three times: if the order changes, you forgot to sort.
- `maps.Keys` returns an **iterator**, not a slice, so you can't index it or call `len` on it. Wrap it: `slices.Sorted(maps.Keys(m))` or `slices.Collect(maps.Keys(m))`.
- Reading a missing key returns the zero value. An unknown creator gets `0` with no special code.
- **Look in** lists `sort`, but Things to avoid says to use `slices` over `sort.Slice`. Read `sort` once to see what you're avoiding, then use `slices`.

**Check yourself**
- Why must `Add` have a pointer receiver? Could `BalanceFor` have a value receiver?
- What happens if you write to a map declared as `var m map[string]Kobo`?
- Why is an unknown creator's balance 0 without any `if`?

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

**Structure**

```
mystally/
  cmd/mystally/main.go          the CLI and the wiring. Imports both packages below.
  internal/ledger/
    event.go                    Event, EventType, Kobo        (was cmd/mystally/event.go)
    ledger.go                   Ledger, NewLedger, methods    (was cmd/mystally/ledger.go)
    store.go                    the Store interface           (new)
  internal/store/
    file.go                     FileStore, NewFileStore(path) (was cmd/mystally/storage.go)
    memory.go                   MemoryStore                   (new)
```

Import arrows. They must only point one way:

```
cmd/mystally ──► internal/store ──► internal/ledger
      │                                   ▲
      └───────────────────────────────────┘
```

- The first line of each moved file changes from `package main` to `package ledger` or `package store`. The package name matches the folder name.
- Import with the full path, `"github.com/na-cho-dev/mystally/internal/ledger"`, and use the last part as the prefix: `ledger.Event`, `ledger.NewLedger`.
- Capitalise only what other packages need: `NewLedger`, `Event`, `Kobo`, the methods `main` calls. `Ledger`'s fields stay lowercase, and the constructor fills them.
- `FileStore` holds its path: `type FileStore struct { path string }` with `NewFileStore(path string) *FileStore`.
- `Add` now saves through the store, so it can fail. Its signature becomes `Add(e Event) error`. Signatures change when responsibilities move. That's normal.

**Examples**

<details>
<summary><b>A three-package program with a consumer-side interface</b></summary>

```
notify/
  go.mod                         module example.com/notify
  cmd/notify/main.go
  internal/alert/alert.go        declares Sender, uses it
  internal/sender/console.go     implements Sender
  internal/sender/fake.go        implements Sender (for tests)
```

`internal/alert/alert.go`:

```go
// Package alert decides when to send a message. It does not know how
// messages are delivered; that is hidden behind the Sender interface.
package alert

import "fmt"

// Sender is declared here, by the package that needs it (the consumer),
// not by the packages that implement it.
type Sender interface {
	Send(to, msg string) error
}

type Alerter struct {
	sender    Sender // unexported: other packages cannot touch it
	threshold int
}

// NewAlerter is a constructor: the only way other packages can build an
// Alerter with its unexported fields filled in.
func NewAlerter(s Sender, threshold int) *Alerter {
	return &Alerter{sender: s, threshold: threshold}
}

// Check sends an alert when value reaches the threshold.
func (a *Alerter) Check(to string, value int) error {
	if value < a.threshold {
		return nil
	}
	return a.sender.Send(to, fmt.Sprintf("value %d reached threshold %d", value, a.threshold))
}
```

`internal/sender/console.go`:

```go
// Package sender holds ways to deliver messages.
package sender

import (
	"fmt"

	"example.com/notify/internal/alert"
)

// Compile-time check: the build fails here if ConsoleSender stops
// satisfying alert.Sender.
var _ alert.Sender = (*ConsoleSender)(nil)

type ConsoleSender struct{}

func (c *ConsoleSender) Send(to, msg string) error {
	fmt.Printf("to %s: %s\n", to, msg)
	return nil
}
```

`internal/sender/fake.go`:

```go
package sender

import "example.com/notify/internal/alert"

var _ alert.Sender = (*FakeSender)(nil)

// FakeSender records messages instead of delivering them. Tests use it.
type FakeSender struct {
	Sent []string
}

func (f *FakeSender) Send(to, msg string) error {
	f.Sent = append(f.Sent, to+": "+msg)
	return nil
}
```

`cmd/notify/main.go`:

```go
package main

import (
	"fmt"
	"os"

	"example.com/notify/internal/alert"
	"example.com/notify/internal/sender"
)

func main() {
	// Wiring: main picks the concrete Sender. Swapping in
	// &sender.FakeSender{} here changes nothing else.
	a := alert.NewAlerter(&sender.ConsoleSender{}, 10)

	if err := a.Check("ops", 12); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

```
$ go run ./cmd/notify
to ops: value 12 reached threshold 10

$ go doc ./internal/alert
package alert // import "example.com/notify/internal/alert"

Package alert decides when to send a message. It does not know how messages are
delivered; that is hidden behind the Sender interface.

type Alerter struct{ ... }
    func NewAlerter(s Sender, threshold int) *Alerter
type Sender interface{ ... }
```

`alert` never imports `sender`. `ConsoleSender` never says "I implement Sender". It just has the right method. That's what lets `alert` stay ignorant of delivery.

</details>

**Gotchas** (real error text from the example above)
- If `alert` imports `sender` while `sender` imports `alert`:
  ```
  imports example.com/notify/internal/alert from main.go
  imports example.com/notify/internal/sender from oops.go
  imports example.com/notify/internal/alert from console.go: import cycle not allowed
  ```
  The fix is always the same: the interface lives with the consumer, and the implementer imports the consumer.
- If `main` tries to touch an unexported field: `alert.Alerter{}.sender undefined (cannot refer to unexported field sender)`. Use the constructor.
- If an implementation drifts from the interface, the compile-time check catches it at the implementation, not somewhere far away:
  ```
  *ConsoleSender does not implement alert.Sender (wrong type for method Send)
      have Send(string) error
      want Send(string, string) error
  ```
- `var _ ledger.Store = (*FileStore)(nil)` costs nothing at runtime. It's a typed nil that the compiler checks and throws away.
- Slices share memory. If `MemoryStore.Save` keeps the caller's slice and the caller later changes an element, the "stored" data changes too. Store a copy (`slices.Clone`) and return a copy from `Load`.
- After moving files, a leftover `package main` line gives `found packages ledger (event.go) and main (ledger.go)`. Every file in a folder must say the same package.

**Check yourself**
- Why is `Store` declared in `ledger` and not in `store`?
- Draw your import arrows. What error appears if `ledger` imports `store`?
- Why can't `main` set `Ledger`'s fields directly, and why is that a good thing?
- What would you change in `main` to swap `FileStore` for `MemoryStore`? (The answer should be one line.)

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

**Structure**

```
internal/ledger/
  ledger_test.go    package ledger_test   ← must be _test: it imports internal/store
  event_test.go     package ledger        Kobo and ParseEventType tests (either package works)
internal/store/
  file_test.go      package store         FileStore round trip with t.TempDir()
```

**This plan has a built-in trap.** `ledger_test.go` uses `MemoryStore` from `internal/store`, and `internal/store` imports `internal/ledger`. If the test file says `package ledger`, you get this error (shown here with the stage 4 `notify` example's paths, where the same thing happens):

```
imports example.com/notify/internal/sender from cycle_test.go
imports example.com/notify/internal/alert from console.go: import cycle not allowed in test
```

Name the test package `ledger_test`. It's a separate package in the same folder, so the cycle disappears. It can only use exported names, which is exactly what `main` sees. That's usually what you want to test anyway.

**Examples**

<details>
<summary><b>A table-driven test with subtests</b></summary>

Tests for the `Minutes` type from stage 2, in `minutes_test.go` next to `minutes.go`:

```go
package minutes

import "testing"

func TestMinutesString(t *testing.T) {
	tests := []struct {
		name string
		in   Minutes
		want string
	}{
		{"zero", 0, "0h00m"},
		{"under an hour", 5, "0h05m"},
		{"exactly one hour", 60, "1h00m"},
		{"pads minutes", 125, "2h05m"},
		{"large", 6000, "100h00m"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in.String()
			if got != tt.want {
				t.Errorf("Minutes(%d).String() = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
```

```bash
go test ./...                                  # everything
go test -v ./internal/ledger                   # one package, every test name
go test -run 'TestMinutesString/pads' -v ./... # one subtest (spaces become _ in names)
go test -cover ./...                           # coverage percentage
go test -coverprofile=cover.out ./... && go tool cover -func=cover.out   # per function
go tool cover -html=cover.out                  # opens a page showing uncovered lines in red
```

Add `cover.out` to `.gitignore`.

</details>

<details>
<summary><b>An external test package using a fake</b></summary>

Tests for the stage 4 `notify` example, in `internal/alert/alert_test.go`:

```go
// package alert_test (note the _test suffix) is an external test package.
// It can import sender, which imports alert, without an import cycle.
package alert_test

import (
	"testing"

	"example.com/notify/internal/alert"
	"example.com/notify/internal/sender"
)

func TestCheck(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		wantSent int
	}{
		{"below threshold", 9, 0},
		{"at threshold", 10, 1},
		{"above threshold", 50, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &sender.FakeSender{}
			a := alert.NewAlerter(fake, 10)

			if err := a.Check("ops", tt.value); err != nil {
				t.Fatalf("Check returned error: %v", err) // Fatalf: nothing below makes sense now
			}
			if len(fake.Sent) != tt.wantSent {
				t.Errorf("sent %d messages, want %d", len(fake.Sent), tt.wantSent)
			}
		})
	}
}
```

</details>

<details>
<summary><b>A file round trip with <code>t.TempDir</code></b></summary>

The package under test:

```go
// Package shelf saves and loads books as JSON.
package shelf

import (
	"encoding/json"
	"os"
	"time"
)

type Book struct {
	Title string    `json:"title"`
	Added time.Time `json:"added"`
}

func Save(path string, books []Book) error {
	data, err := json.MarshalIndent(books, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func Load(path string) ([]Book, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var books []Book
	if err := json.Unmarshal(data, &books); err != nil {
		return nil, err
	}
	return books, nil
}
```

The test:

```go
package shelf

import (
	"path/filepath"
	"testing"
	"time"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	// t.TempDir gives a fresh folder that is deleted when the test ends.
	path := filepath.Join(t.TempDir(), "books.json")

	// Fixed times, not time.Now(): tests must give the same result every run.
	want := []Book{
		{Title: "Americanah", Added: time.Date(2026, 1, 2, 15, 4, 5, 0, time.UTC)},
		{Title: "Purple Hibiscus", Added: time.Date(2026, 2, 3, 9, 0, 0, 0, time.UTC)},
	}

	if err := Save(path, want); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(got) != len(want) {
		t.Fatalf("got %d books, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].Title != want[i].Title || !got[i].Added.Equal(want[i].Added) {
			t.Errorf("book %d: got %+v, want %+v", i, got[i], want[i])
		}
	}
}
```

</details>

**Gotchas**
- **Write expected values as literal numbers**, like `want: 55`, never as `rates[Like] * 11`. A test that computes its answer the same way the code does will never fail, and then "changing one rate makes a test fail" can't be ticked.
- `t.Errorf` records a failure and keeps going. `t.Fatalf` records it and stops this test. Use `Fatalf` when the lines below would be meaningless, for example after a `Load` error.
- Don't compare times with `==`. A `time.Now()` value saved to JSON and loaded back gave `== says false, Equal says true`. Use `.Equal`, and build test times with `time.Date(...)`.
- Slices can't be compared with `==`. Use `slices.Equal` for slices of simple values, and compare field by field for structs that contain a `time.Time`.
- Test functions must be named `TestXxx(t *testing.T)` and live in files ending `_test.go`, or `go test` silently skips them.
- Tests run with the working directory set to the **package folder**, so `testdata/` paths are relative to `internal/ledger/`, not the repo root.

**Check yourself**
- Why must `ledger_test.go` be `package ledger_test`?
- When do you use `t.Fatalf` instead of `t.Errorf`?
- Why must the expected numbers in your tests be literals?
- Which lines does `go tool cover -html` show in red, and do they matter?

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

**Structure**

```
internal/ledger/errors.go    ErrUnknownEventType, ValidationError and its Error method
internal/ledger/event.go     ParseEventType returns ErrUnknownEventType, wrapped with the bad name
internal/ledger/ledger.go    Add validates the event and returns *ValidationError
internal/store/file.go       every returned error wrapped: "load events: %w", "save events: %w"
cmd/mystally/main.go         main calls run; on error, calls fail(err), which never returns
```

| Error | Detected with | Exit |
|---|---|---|
| `ErrUnknownEventType` (wrapped or not) | `errors.Is` | 2 |
| `*ValidationError` (wrapped or not) | `errors.As` | 3 |
| anything else | n/a | 1 |

**Examples**

<details>
<summary><b>Sentinel errors, custom types, wrapping, and classifying</b></summary>

```go
package main

import (
	"errors"
	"fmt"
	"os"
)

// Sentinel error: one shared value that callers compare against with errors.Is.
var ErrRoomTaken = errors.New("room already taken")

// Custom error type: carries details that callers can read with errors.As.
type ValidationError struct {
	Field, Reason string
}

// Pointer receiver, so the error value is always a *ValidationError.
// errors.As must then look for *ValidationError too.
func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Reason
}

func book(room, guest string) error {
	if guest == "" {
		return &ValidationError{Field: "guest", Reason: "must not be empty"}
	}
	if room == "101" {
		return ErrRoomTaken
	}
	return nil
}

func loadRooms(path string) error {
	if _, err := os.ReadFile(path); err != nil {
		return fmt.Errorf("load rooms: %w", err) // %w keeps err inside the new error
	}
	return nil
}

// exitCode turns an error into an exit code. It looks through any wrapping.
func exitCode(err error) int {
	var ve *ValidationError
	switch {
	case errors.Is(err, ErrRoomTaken):
		return 2
	case errors.As(err, &ve):
		return 3
	default:
		return 1
	}
}

func main() {
	errs := []error{
		book("101", "ada"),
		fmt.Errorf("check-in: %w", book("102", "")),
		loadRooms("missing.json"),
	}
	for _, err := range errs {
		fmt.Printf("exit=%d  notExist=%-5v  %v\n", exitCode(err), errors.Is(err, os.ErrNotExist), err)
	}

	// Go 1.26+: errors.AsType does what errors.As does, without the extra variable.
	if ve, ok := errors.AsType[*ValidationError](errs[1]); ok {
		fmt.Println("bad field:", ve.Field)
	}
}
```

```
exit=2  notExist=false  room already taken
exit=3  notExist=false  check-in: guest: must not be empty
exit=1  notExist=true   load rooms: open missing.json: no such file or directory
bad field: guest
```

Line 2 shows `errors.As` finding the `*ValidationError` **through** the `check-in:` wrapping. Line 3 shows `errors.Is` finding `os.ErrNotExist` through `load rooms:`.

</details>

<details>
<summary><b>The nil-interface trap</b></summary>

```go
package main

import "fmt"

type MyErr struct{}

func (*MyErr) Error() string { return "my error" }

// Wrong: returns a typed nil pointer wrapped in a non-nil error interface.
func check(fail bool) error {
	var p *MyErr
	if fail {
		p = &MyErr{}
	}
	return p
}

// Right: returns a literal nil when there is no error.
func checkFixed(fail bool) error {
	if fail {
		return &MyErr{}
	}
	return nil
}

func main() {
	fmt.Println(check(false) == nil)      // false, even though nothing failed
	fmt.Println(checkFixed(false) == nil) // true
}
```

An interface value is a pair: (type, value). `check(false)` returns `(*MyErr, nil)`. The value is nil, but the type isn't, so the interface isn't nil.

</details>

**Gotchas**
- `%v` turns the inner error into plain text, so `errors.Is` and `errors.As` can't see it any more. Only `%w` keeps it findable.
- The receiver of `Error()` decides what `errors.As` must look for. Pointer receiver → return `&ValidationError{...}` and look for `*ValidationError`. If you mix value and pointer, `errors.As` silently returns false.
- `errors.As(err, ve)` without `&`: `go vet` says `second argument to errors.As must be a non-nil pointer to either a type that implements error, or to any interface type`. At runtime it panics.
- Never compare `err.Error()` strings. Messages change; `errors.Is` and `errors.As` don't break when they do.
- Check the most specific error first in `fail`. Order matters in a `switch`.
- To check exit codes 2 and 3, use the built binary. `go run` turns every failure into 1.

**Check yourself**
- When do you use a sentinel error, and when a custom error type?
- What does `%w` do that `%v` doesn't?
- Why does returning a nil `*ValidationError` as an `error` break `if err != nil`?
- Which exit code does each of your failure cases give? Write them down, then check them with `echo $?`.

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

**Structure**

```
internal/importer/importer.go   Import(ctx context.Context, dir string, workers int) ([]ledger.Event, error)
cmd/mystally/main.go            importCmd: parse flags, build ctx, call Import, add the events, save
testdata/import/                eight sample JSON files
```

Who does what:

```
importCmd ── ctx (Ctrl-C + 10s timeout) ──► importer.Import
                                              ├─ one goroutine per file:
                                              │    wait for a semaphore slot → read + decode → send a result
                                              ├─ closer goroutine: wg.Wait(), then close(results)
                                              └─ for r := range results: collect events, keep the first error
importCmd ◄── []ledger.Event, error ─────────┘
   then adds the events to the ledger and saves, in ONE goroutine (no lock needed)
```

- Each goroutine can reuse your `FileStore` to decode its file: `store.NewFileStore(path).Load()`. Then `importer` imports `store` and `ledger`, and the arrows still point one way.
- Parse `--workers` with `flag.NewFlagSet("import", flag.ContinueOnError)`, one FlagSet per subcommand. Flags go **before** the directory.
- Only one goroutine (the one receiving results) touches the ledger, so the ledger needs no mutex yet. That changes in stage 8.

**Examples**

<details>
<summary><b>Fan-out / fan-in with a semaphore, a timeout and Ctrl-C</b></summary>

Counts lines in every `*.txt` file in a folder, concurrently. Each file pretends to take 200ms, so the speed-up is visible.

```go
package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// result is what each goroutine sends back: a value or an error, never both.
type result struct {
	path  string
	lines int
	err   error
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "lines:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	// A FlagSet per subcommand. Flags must come before the positional args.
	fs := flag.NewFlagSet("lines", flag.ContinueOnError)
	workers := fs.Int("workers", 4, "max files read at once")
	timeout := fs.Duration("timeout", 5*time.Second, "give up after this long")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return errors.New("usage: lines [--workers N] [--timeout D] <dir>")
	}

	// Ctrl-C cancels ctx. Then the timeout is layered on top of it.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	ctx, cancel := context.WithTimeout(ctx, *timeout)
	defer cancel()

	paths, err := filepath.Glob(filepath.Join(fs.Arg(0), "*.txt"))
	if err != nil {
		return err
	}

	before := runtime.NumGoroutine()
	start := time.Now()
	total, err := countAll(ctx, paths, *workers)
	fmt.Printf("took %s, goroutines before %d, after %d\n",
		time.Since(start).Round(time.Millisecond), before, runtime.NumGoroutine())
	if err != nil {
		return err
	}
	fmt.Printf("%d lines in %d files\n", total, len(paths))
	return nil
}

func countAll(ctx context.Context, paths []string, workers int) (int, error) {
	results := make(chan result)
	sem := make(chan struct{}, workers) // a buffered channel with `workers` free slots
	var wg sync.WaitGroup

	for _, p := range paths {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- countOne(ctx, sem, p)
		}()
	}

	// The closer. Only this goroutine closes results, and only after every
	// sender has finished, so nobody can send on a closed channel.
	go func() {
		wg.Wait()
		close(results)
	}()

	// Receive until the channel is closed. We never stop early, so no
	// sender is ever left blocked (that would be a goroutine leak).
	total := 0
	var firstErr error
	for r := range results {
		if r.err != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("%s: %w", r.path, r.err)
			}
			continue
		}
		total += r.lines
	}
	return total, firstErr
}

func countOne(ctx context.Context, sem chan struct{}, path string) result {
	// Take a slot, or give up if cancelled while waiting.
	select {
	case sem <- struct{}{}:
		defer func() { <-sem }() // hand the slot back when done
	case <-ctx.Done():
		return result{path: path, err: ctx.Err()}
	}

	// Pretend each file is slow, so the speed-up is visible.
	select {
	case <-time.After(200 * time.Millisecond):
	case <-ctx.Done():
		return result{path: path, err: ctx.Err()}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return result{path: path, err: err}
	}
	return result{path: path, lines: bytes.Count(data, []byte("\n"))}
}
```

Built with `-race`, run on 8 files:

```
$ ./lines --workers 1 dir     took 1.604s, goroutines before 3, after 3     360 lines in 8 files
$ ./lines dir                 took 402ms,  goroutines before 3, after 3     (default: 4 workers)
$ ./lines --workers 8 dir     took 203ms
$ ./lines --timeout 1ms dir   took 1ms ... f7.txt: context deadline exceeded   (exit 1)
$ ./lines dir --workers 2     usage: lines [--workers N] [--timeout D] <dir>  (flag after the arg!)
```

"Goroutines before 3, after 3" is the leak check: everything started was finished. The race detector printed nothing.

</details>

**Gotchas**
- **Eight small JSON files load in microseconds.** Concurrency may not be visibly faster, and can even be slower, which makes the first "Done when" box hard to tick honestly. Add a hidden `--slow` flag that waits 200ms per file and respects `ctx` (like `countOne` above), or make the sample files large.
- If the receiving loop ever stops early (for example `return` on the first error), every goroutine still trying to send blocks forever. That's a leak. Either keep receiving until the channel closes, or cancel the context and make the senders `select` on `ctx.Done()`.
- Only the sending side closes a channel, and only once. Sending on a closed channel panics, and closing twice panics.
- Call `wg.Add(1)` **before** the `go` statement, not inside the goroutine. Go 1.25+ also has `wg.Go(func() { ... })`, which does the `Add` and `Done` for you.
- Since Go 1.22, each loop iteration gets its own `p`, so `go func() { ... p ... }()` is safe. Old blog posts warn about a bug that no longer exists.
- Always `defer cancel()`. `go vet` warns if a cancel function can leak.
- The flag package stops at the first non-flag argument: `import dir --workers 2` leaves `--workers` unparsed.
- After a timeout, `ctx.Err()` is `context.DeadlineExceeded`. Check it with `errors.Is`, as in stage 6.

**Check yourself**
- Who closes the results channel, and why is it safe? (One sentence.)
- What happens if `Import` returns on the first error without cancelling the context?
- What does the semaphore's buffer size mean?
- Why does merging into the ledger need no mutex in this design?

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

**Structure**

```
internal/httpapi/
  server.go          NewHandler(l *ledger.Ledger) http.Handler: routes and handlers
  middleware.go      logging(next http.Handler) http.Handler
  server_test.go     package httpapi_test: httptest + MemoryStore, no real server, no files
cmd/mystally/main.go serveCmd: --addr flag, http.Server with timeouts, graceful shutdown
```

- **The ledger gets a `sync.Mutex` now.** Every request runs in its own goroutine, so two `POST /events` can call `Add` at the same moment. Stage 7's "one goroutine owns the ledger" no longer holds. `go test -race` will prove it if you forget.
- "Pass `r.Context()` into every ledger call" means ledger methods gain a first parameter: `BalanceFor(ctx context.Context, creator string)`. Give `Store` one too: `Load(ctx)` and `Save(ctx, events)`. `FileStore` and `MemoryStore` can ignore it for now. Postgres in stage 9 needs it, and adding it **now** is what makes stage 9's "nothing in `internal/ledger` changed" box possible.
- Response bodies are structs with JSON tags (`balanceResponse`, `errorResponse`), not `map[string]any`.
- `POST /events` decodes `{"creator":"nacho","type":"like"}`. The server fills in `ID` and `At`, then reuses the validation from stage 6: a `*ValidationError` becomes a 400. If `EventType` has `UnmarshalText` (stage 2), `"like"` decodes with no extra code.

**Examples**

<details>
<summary><b>A small JSON API: store with a mutex, routes, middleware, tests, graceful shutdown</b></summary>

`internal/notes/store.go`:

```go
// Package notes keeps notes in memory, safely shared between goroutines.
package notes

import (
	"slices"
	"sync"
)

type Note struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}

// Store is safe for concurrent use: every HTTP request runs in its own
// goroutine, so two requests can call Add at the same moment.
type Store struct {
	mu    sync.Mutex // guards notes; the zero value is ready to use
	notes []Note
}

func (s *Store) Add(text string) Note {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := Note{ID: len(s.notes) + 1, Text: text}
	s.notes = append(s.notes, n)
	return n
}

func (s *Store) Get(id int) (Note, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, n := range s.notes {
		if n.ID == id {
			return n, true
		}
	}
	return Note{}, false
}

// All returns a copy, so callers cannot change the store's slice
// after the lock is released.
func (s *Store) All() []Note {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Clone(s.notes)
}
```

`internal/httpapi/server.go`:

```go
// Package httpapi exposes notes over HTTP using only net/http.
package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"

	"example.com/notesapi/internal/notes"
)

type server struct {
	store *notes.Store
}

// NewHandler returns every route, wrapped in the logging middleware.
func NewHandler(store *notes.Store) http.Handler {
	s := &server{store: store}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /notes", s.listNotes)
	mux.HandleFunc("GET /notes/{id}", s.getNote)
	mux.HandleFunc("POST /notes", s.createNote)
	return logging(mux)
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json") // headers first
	w.WriteHeader(status)                              // then the status
	json.NewEncoder(w).Encode(v)                       // then the body
}

func (s *server) listNotes(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.store.All())
}

func (s *server) getNote(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{"id must be a number"})
		return
	}
	n, ok := s.store.Get(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, errorResponse{"no such note"})
		return
	}
	writeJSON(w, http.StatusOK, n)
}

func (s *server) createNote(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{err.Error()})
		return // without this return, the code below would run too
	}
	if in.Text == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"text must not be empty"})
		return
	}
	writeJSON(w, http.StatusCreated, s.store.Add(in.Text))
}
```

`internal/httpapi/middleware.go`:

```go
package httpapi

import (
	"log/slog"
	"net/http"
	"time"
)

// statusRecorder wraps a ResponseWriter to remember the status code.
// Embedding http.ResponseWriter gives it every method for free;
// we only override WriteHeader.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// logging is middleware: it takes a handler and returns a handler that
// does something before and after calling it.
func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		slog.Info("request", "method", r.Method, "path", r.URL.Path,
			"status", rec.status, "duration", time.Since(start))
	})
}
```

`internal/httpapi/server_test.go`:

```go
package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"example.com/notesapi/internal/httpapi"
	"example.com/notesapi/internal/notes"
)

func TestCreateThenGet(t *testing.T) {
	h := httpapi.NewHandler(&notes.Store{})

	// No real server: call the handler directly with a fake request
	// and a recorder that captures the response.
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/notes", strings.NewReader(`{"text":"buy milk"}`)))
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST status = %d, want %d; body: %s", rec.Code, http.StatusCreated, rec.Body)
	}

	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/notes/1", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET status = %d, want %d", rec.Code, http.StatusOK)
	}
	var got notes.Note
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.Text != "buy milk" {
		t.Errorf("text = %q, want %q", got.Text, "buy milk")
	}
}

func TestCreateBadJSON(t *testing.T) {
	h := httpapi.NewHandler(&notes.Store{})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/notes", strings.NewReader(`{not json`)))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}
```

`cmd/notesd/main.go`:

```go
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"example.com/notesapi/internal/httpapi"
	"example.com/notesapi/internal/notes"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           httpapi.NewHandler(&notes.Store{}),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
	}

	// ListenAndServe blocks, so run it in a goroutine and report back.
	errc := make(chan error, 1)
	go func() { errc <- srv.ListenAndServe() }()
	slog.Info("listening", "addr", srv.Addr)

	select {
	case err := <-errc: // could not start, e.g. the port is in use
		slog.Error("server failed", "err", err)
		os.Exit(1)
	case <-ctx.Done(): // Ctrl-C
	}

	// A fresh context: ctx is already cancelled, so deriving from it
	// would give Shutdown no time at all.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	slog.Info("shutting down, finishing in-flight requests")
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown", "err", err)
		os.Exit(1)
	}
	if err := <-errc; !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server", "err", err)
	}
	slog.Info("stopped")
}
```

A real run, with curl from another terminal and then Ctrl-C:

```
$ curl -X POST -d '{"text":"buy milk"}' localhost:8080/notes   → {"id":1,"text":"buy milk"}
$ curl localhost:8080/notes/9                                   → {"error":"no such note"}
$ curl -X POST -d '{bad' localhost:8080/notes                   → {"error":"invalid character 'b' looking for beginning of object key string"}
$ curl -X DELETE localhost:8080/notes/1                         → 405 (no route for DELETE)

INFO listening addr=:8080
INFO request method=POST path=/notes status=201 duration=69.063µs
INFO request method=GET path=/notes/9 status=404 duration=84.375µs
INFO request method=DELETE path=/notes/1 status=405 duration=15.622µs
INFO shutting down, finishing in-flight requests
INFO stopped
```

</details>

**Gotchas**
- Forgetting `return` after writing an error response means the handler carries on and writes again. The log shows `http: superfluous response.WriteHeader call`, and the client gets the **first** status.
- Order inside a handler: `w.Header().Set(...)` → `w.WriteHeader(status)` → write the body. Headers set after `WriteHeader` are silently ignored.
- Patterns with a method (`"GET /notes"`) give `405 Method Not Allowed` for other methods automatically.
- The shutdown context must be fresh (`context.Background()`). The signal context is already cancelled by then.
- After `Shutdown`, `ListenAndServe` returns `http.ErrServerClosed`. That's the normal ending, not a failure.
- Handlers run concurrently. Any shared state needs a mutex. Run `go test -race ./...`.
- `json.NewEncoder(w).Encode` adds a trailing newline. Compare test bodies after decoding, not as raw strings.

**Check yourself**
- Why does the ledger need a mutex now, when it didn't in stage 7?
- What does the middleware signature `func(http.Handler) http.Handler` let you do?
- What happens to a request that is half-finished when you press Ctrl-C?
- Why can handler tests run without opening a network port?

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

**Structure**

```
mystally/
  schema.sql
  internal/store/
    postgres.go          PostgresStore, NewPostgresStore(ctx, url) (+ compile-time check)
    store_test.go        testStore(t, s ledger.Store): the shared checks, written once
    memory_test.go       calls testStore with a MemoryStore
    file_test.go         calls testStore with a FileStore in t.TempDir()
    postgres_test.go     t.Skip unless MYSTALLY_DATABASE_URL is set, then calls testStore
  cmd/mystally/main.go   the only change: pick the store from the env var
```

- **Write the store tests once.** A helper `testStore(t *testing.T, s ledger.Store)` holds the checks. Each store's test just builds its store and calls the helper. That's what "the same store tests run against Postgres" means in practice.
- `Save([]Event)` doesn't map neatly onto SQL. One honest version: one transaction, `INSERT ... ON CONFLICT (id) DO NOTHING` for each event.
- Run Postgres locally with Docker (installed on this machine, along with `psql`):

  ```bash
  docker run --name mystally-pg -e POSTGRES_PASSWORD=dev -p 5432:5432 -d postgres:17
  export MYSTALLY_DATABASE_URL='postgres://postgres:dev@localhost:5432/postgres?sslmode=disable'
  psql "$MYSTALLY_DATABASE_URL" -f schema.sql
  go test ./...
  docker stop mystally-pg   # docker start mystally-pg next time
  ```

**Examples**

<details>
<summary><b>Generic helpers, a constraint, and <code>t.Skip</code></b></summary>

```go
// Package gen shows small generic helpers.
package gen

// Map applies f to every element of xs and returns the results.
// T and U are type parameters: the caller's types fill them in.
func Map[T, U any](xs []T, f func(T) U) []U {
	out := make([]U, 0, len(xs))
	for _, x := range xs {
		out = append(out, f(x))
	}
	return out
}

// Number is a constraint: the set of types SumBy accepts.
// ~int64 means "int64, or any named type whose underlying type is int64".
type Number interface {
	~int | ~int64 | ~float64
}

// SumBy adds up f(x) for every x.
func SumBy[T any, N Number](xs []T, f func(T) N) N {
	var total N
	for _, x := range xs {
		total += f(x)
	}
	return total
}
```

```go
package gen

import (
	"os"
	"slices"
	"strconv"
	"testing"
)

type Minutes int64 // a named type: works with SumBy because of ~int64

type Task struct {
	Name string
	Took Minutes
}

func TestMap(t *testing.T) {
	got := Map([]int{1, 2, 3}, strconv.Itoa) // T=int, U=string, inferred
	if want := []string{"1", "2", "3"}; !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestSumBy(t *testing.T) {
	tasks := []Task{{"write", 30}, {"test", 45}}
	got := SumBy(tasks, func(t Task) Minutes { return t.Took })
	if got != 75 {
		t.Errorf("got %d, want 75", got)
	}
}

func TestNeedsDatabase(t *testing.T) {
	url := os.Getenv("EXAMPLE_DATABASE_URL")
	if url == "" {
		t.Skip("EXAMPLE_DATABASE_URL not set; skipping database test")
	}
	// ... connect to url and run the same checks as the memory tests
}
```

```
--- PASS: TestMap
--- PASS: TestSumBy
    gen_test.go:35: EXAMPLE_DATABASE_URL not set; skipping database test
--- SKIP: TestNeedsDatabase
```

Without the `~`, `SumBy` would reject `Minutes`, and it would reject your `Kobo` too.

</details>

<details>
<summary><b>The pgx calls a store needs (pgx v5)</b></summary>

`schema.sql`:

```sql
CREATE TABLE visits (
    id    BIGINT PRIMARY KEY,
    guest TEXT NOT NULL,
    at    TIMESTAMPTZ NOT NULL
);
```

```go
// Package pg shows the pgx calls a Postgres-backed store needs.
package pg

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Visit struct {
	ID    int64     `db:"id"`
	Guest string    `db:"guest"`
	At    time.Time `db:"at"`
}

type Store struct {
	pool *pgxpool.Pool
}

// New opens a connection pool: a set of connections shared by
// concurrent callers and reused between queries.
func New(ctx context.Context, url string) (*Store, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, err
	}
	// pgxpool.New does not connect yet. Ping proves the URL works.
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return &Store{pool: pool}, nil
}

func (s *Store) Close() { s.pool.Close() }

func (s *Store) All(ctx context.Context) ([]Visit, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, guest, at FROM visits ORDER BY id`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[Visit])
}

// SaveAll writes visits that are not stored yet, in one transaction:
// either every row is written or none is.
func (s *Store) SaveAll(ctx context.Context, visits []Visit) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) // harmless after a successful Commit

	for _, v := range visits {
		// $1, $2, $3 are placeholders. Never build SQL with fmt.Sprintf.
		_, err := tx.Exec(ctx,
			`INSERT INTO visits (id, guest, at) VALUES ($1, $2, $3)
			 ON CONFLICT (id) DO NOTHING`,
			v.ID, v.Guest, v.At)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
```

`pgx.CollectRows` reads every row and closes `rows` for you. `RowToStructByName` matches columns to fields by the `db` tag, or by name, ignoring case.

</details>

<details>
<summary><b>Building a binary that runs anywhere</b></summary>

```bash
CGO_ENABLED=0 go build -o mystally ./cmd/mystally                              # this machine, static
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o mystally-arm64 ./cmd/mystally # a Raspberry Pi or ARM server
GOOS=windows GOARCH=amd64 go build -o mystally.exe ./cmd/mystally              # Windows
file mystally                                                                  # shows what you built

# run it in a container that has no Go installed:
docker run --rm -v "$PWD/mystally:/usr/local/bin/mystally:ro" alpine mystally
```

On this machine, a normal build of a program that uses `net/http` came out **dynamically linked** (it depends on the system's C library). With `CGO_ENABLED=0` it came out **statically linked**. Only the static one runs in a minimal container like `alpine` or `scratch`.

</details>

**Gotchas**
- `go get github.com/jackc/pgx/v5` changes `go.mod` and creates `go.sum`. Commit both. `go mod tidy` removes what you no longer use.
- Never build SQL with `fmt.Sprintf` and user input. Use `$1, $2` placeholders.
- `pgxpool.New` doesn't connect. A wrong password only shows up on the first query, so call `Ping` in the constructor.
- Postgres stores time to the **microsecond**, but `time.Now()` has nanoseconds. A round-trip test that uses `time.Now()` fails even with `.Equal`. Use `time.Date(...)` in tests (stage 5 already taught this).
- A normal `go build` may need the system C library at runtime. Use `CGO_ENABLED=0` for a binary you copy elsewhere.
- Don't make things generic just because you can. A generic function earns its place when two or more real types use it.

**Check yourself**
- What is `go.sum` for, and why is it committed?
- Why is the `Store` interface the only thing that had to know Postgres exists?
- What does `~int64` mean in a constraint?
- Why does a static binary matter for a container?

**Done when**
- [ ] `MYSTALLY_DATABASE_URL=... go test ./...` passes against a real database
- [ ] nothing in `internal/ledger`, `cmd`, or `internal/httpapi` changed to add Postgres
- [ ] a single binary runs with no Go toolchain present

---

## The graduation test

After stage 9, delete the repo (keep this file) and rebuild Mystally from scratch in one sitting, stages 1 through 8, without looking at the first version. If it takes an afternoon and the tests pass, you know Go fundamentals. The rewrite will be about half the size of the original. That difference is what you learned.

---

## Testing by hand

Before stage 5 there are no automated tests. These three shell tools are enough.

**Exit code: `echo $?`** prints the exit code of the last command. Use the built binary, because `go run` turns every failure into 1.

```bash
go build ./cmd/mystally && ./mystally add nacho dance; echo $?
```

**Which stream: `2>/dev/null` hides stderr, `>/dev/null` hides stdout.**

```bash
./mystally add nacho dance 2>/dev/null   # should print nothing: the error went to stderr
./mystally list >/dev/null               # should print nothing: the rows went to stdout
```

**The data file: `cat` to inspect, `rm` to reset.**

```bash
cat events.txt
rm events.txt
```

**Stage 1's "Done when" as one command list:**

```bash
go build ./cmd/mystally
rm -f events.txt
./mystally list; echo $?                 # missing file: no panic
./mystally add nacho like
./mystally add ada share
./mystally add nacho comment
./mystally list                          # three rows
./mystally add nacho; echo $?            # clear message, 1
./mystally; echo $?                      # usage, 1
./mystally add nacho dance; echo $?      # message, 1
go vet ./...
```

Write the same kind of list for each later stage from its "Done when" boxes. In stage 5 you'll turn most of them into real tests.

---

## When the editor and the terminal disagree

The red underlines in VS Code come from **gopls**, a background program the Go extension runs. It can get out of date, especially after you create `go.mod` or move files.

1. **Trust the terminal.** `go build ./...` and `go vet ./...` are the real compiler and checker.
2. **Restart gopls:** Ctrl+Shift+P → **Go: Restart Language Server**. If that doesn't help: **Developer: Reload Window**.
3. **Keep gopls current**, built with the same Go version you use. Check with `go version ~/go/bin/gopls`. Update with Ctrl+Shift+P → **Go: Install/Update Tools** → gopls, or `go install golang.org/x/tools/gopls@latest`.

Example: `could not import os (missing metadata for import of "os") compilerBrokenImport` while `go vet` was clean. gopls had started while the project was being created, and never caught up.

---

## Mistake log

When an error takes you more than five minutes, add a row. Reread this table before each new stage.

| # | Stage | What happened | Error or symptom | Lesson |
|---|---|---|---|---|
| 1 | 0 | Used `( )` for `main`'s body | `syntax error: unexpected ( in parameter list` | `( )` holds parameters, `{ }` holds the body |
| 2 | 0 | `.gitignore` line `mystally` hid `cmd/mystally/` | `main.go` missing from `git status` and from the first commit | Root-only patterns need a leading `/` |
| 3 | 0 | Editor said `os` couldn't be imported | `could not import os (missing metadata...)` | Standard packages need no install. Restart gopls; trust the terminal |
| 4 | 1 | Ran `go vet` / `go build` with no path at the repo root | `no Go files in .../mystally` | Pass `./cmd/mystally` or `./...` |
| 5 | 1 | Compared a string to `nil` | `mismatched types string and untyped nil` | Strings are never nil; the zero value is `""` |
| 6 | 1 | Looped over `os.Args` to find a missing argument | the loop never finds it | A missing argument isn't in the slice. Check `len` |
| 7 | 1 | Used `fmt.Errorf` to print a message | nothing printed; vet: `result of fmt.Errorf call not used` | `Errorf` creates a value. `fmt.Fprintln(os.Stderr, ...)` prints |
| 8 | 1 | Set a flag variable but never read it | `declared and not used: match` | Only reading counts as using |
| 9 | 1 | Looped over a map to find a key | works, but slow and long | `v, ok := m[key]` |
| 10 | | | | |

---

## Go words used in this plan

| Word | Meaning |
|---|---|
| **module** | A project with a `go.mod`. Its path (`github.com/na-cho-dev/mystally`) prefixes every import inside it |
| **package** | One folder of `.go` files that share a `package` line. The unit you import |
| **exported** | A name starting with a capital letter. Other packages can use it; lowercase names are private to the package |
| **zero value** | What a variable holds before you set it: `0`, `""`, `false`, `nil`. Every type has one |
| **receiver** | The `(k Kobo)` or `(l *Ledger)` before a method name: the value the method works on |
| **interface** | A set of method signatures. Any type with those methods satisfies it automatically |
| **sentinel error** | A package-level error value (`var ErrX = errors.New(...)`) that callers compare against with `errors.Is` |
| **wrapping** | `fmt.Errorf("context: %w", err)`: a new message that still carries the original error inside |
| **goroutine** | A function running concurrently, started with `go f()` |
| **channel** | A typed pipe between goroutines: `ch <- v` sends, `v := <-ch` receives |
| **context** | A value passed down a call chain that carries cancellation and deadlines |
| **handler** | Anything with `ServeHTTP(w, r)`: the code that answers one HTTP request |
| **middleware** | A function that wraps a handler to add behaviour before and after it |
| **type parameter** | The `[T any]` in a generic function: a placeholder the caller's type fills in |

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
   | command-line flags | `flag` |
   | structured logs | `log/slog` |

   The full official list is pkg.go.dev/std. Use find-in-page with the noun.
3. Type the package name and a dot in the editor, or run `go doc <pkg>`. Scan the names for one that matches your sentence.
4. Read the signature. `ReadFile(name string) ([]byte, error)` tells you the inputs, the output, and that it can fail.
5. If no name matches, rephrase one level lower. "Append to a file" becomes "open a file in append mode", and `os.OpenFile` with `os.O_APPEND` appears.
6. If the noun is not in `std` at all (Postgres, Redis, Stripe), search pkg.go.dev and judge by importer count and last release date.

## Stuck protocol

1. Read the whole compiler error. It usually names the fix.
2. `go doc` the function you are calling and compare its signature to your call.
3. Find the topic on gobyexample.com and read the one page.
4. Open this stage's **Examples** and find the same tool used in another program.
5. Write the smallest separate `main.go` in a scratch folder that reproduces the problem.
6. Only then ask, and ask with the error text, not for the code.

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
