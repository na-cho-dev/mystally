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
| **Goal / End goal** | What the stage is for, and exactly what you can run and explain when it's finished |
| **Build / Learn / Look in** | What to build, which tools it needs, and which packages to read |
| **Structure** | Which files exist at the end of the stage, and what each function is responsible for. Names and signatures only, never bodies |
| **Steps** | The stage broken into small steps, done in order. Each step has four parts (see below) |
| **Gotchas** | Traps, with the real error text. Entries marked *(you hit this)* already happened |
| **Check yourself** | Questions to answer out loud or in a commit message. If you can't answer one, the stage isn't done |
| **Done when** | The gate to the next stage. Every box shows the commands to run and the output you should see |

Every step has the same four parts:

| Part | What it gives you |
|---|---|
| **Achieve** | What changes in mystally during this step, in words |
| **Learn** | The Go concept the step needs, explained from scratch |
| **Example** | Folded. The concept in a *different* program (books, weekdays, notes), with its real output. You translate the idea instead of copying text |
| **Step done when** | The commands to run on mystally and the output you should see before moving to the next step |

### How to work through a step

1. Read **Achieve** and **Learn**, then try on your own first.
2. Open the **Example** when you're stuck on the tool, or after you finish, to compare.
3. Type it into a scratch folder **outside** this repo. Typing, not pasting, is what makes it stick:

   ```bash
   mkdir -p ~/scratch/go/stage2 && cd ~/scratch/go/stage2
   go mod init scratch
   # create main.go, then:
   go run .
   ```

4. Break it on purpose: delete a `&`, change a receiver, remove a `return`. Read the error.
5. Close the example and write mystally's version without looking.
6. Run the step's **Step done when** commands. Move on only when your output matches.

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

**End goal:** When this stage is finished, `go run ./cmd/mystally` and `go build ./cmd/mystally && ./mystally` both print `mystally 0.1`, `git ls-files` lists your source but never the binary, and `go doc fmt.Println` answers a question about a function without a browser. You will know the difference between a module and a package, and how `go run` turns a folder path into a running program.

**Build**
- `go mod init github.com/<you>/mystally`
- `cmd/mystally/main.go` that prints `mystally 0.1`
- `git init`, a `.gitignore` containing the binary name `mystally`
- Run `go doc os`, `go doc strings`, `go doc fmt.Println` and read what comes back

**Learn:** modules, `package main`, `func main`, how `go run` finds the program.

**Look in:** `go help run`, `go help build`, `go help packages`, `go doc fmt`, `git help gitignore`.

### Structure

```
mystally/
  .gitignore            /mystally   (the leading slash means "only at the repo root")
  go.mod                module github.com/na-cho-dev/mystally, go 1.26.2
  PLAN.md
  cmd/
    mystally/
      main.go           package main, func main()
  mystally              the binary that go build makes (ignored, never committed)
```

### Steps

Stage 0 has seven steps. Do them in order, because each one uses what the step before it made. Every step ends with a check: run it, compare with the expected output, and only then move on.

#### Step 0.1: Check your toolchain and editor

**Achieve:**
- Go 1.26.2 is installed and the `go` command works in your terminal.
- VS Code has the Go extension, and the extension has installed its tools (gopls among them).
- Format-on-save is on, so every `.go` file gets the official layout the moment you save it.

**Learn:**
Go is one program, `go`, with subcommands: `go run`, `go build`, `go vet`, `go doc`, `go mod` and more. `go version` says which Go you have. `go env` shows its settings: `GOROOT` is where Go itself lives, and `GOPATH` is where downloaded modules and installed tools (like gopls) go.

Go also ships `gofmt`, the one official layout for Go code. The tool decides tabs, spaces and blank lines, not you. `gofmt -l .` lists files whose layout is off, `gofmt -d file.go` shows the change it would make, and `gofmt -w file.go` makes it.

The Go extension runs **gopls**, the Go language server: a background program that reads your code as you type. It draws the red underlines, offers the `os.` dropdown, and formats on save with gofmt's rules. When the editor and the terminal disagree, trust the terminal (see [When the editor and the terminal disagree](#when-the-editor-and-the-terminal-disagree)).

<details>
<summary><b>Example: what <code>gofmt</code> fixes</b></summary>

```go
package main
import "fmt"
func main(){
fmt.Println( "20°C is", 20*9/5+32, "°F" )}
```

Output:

```
$ go version
go version go1.26.2 linux/amd64
$ gofmt -l .
main.go
$ gofmt -w main.go
$ cat main.go
package main

import "fmt"

func main() {
	fmt.Println("20°C is", 20*9/5+32, "°F")
}
$ gofmt -l .
```

- `gofmt -l .` lists every file whose layout differs from Go's standard layout. No output means everything is formatted.
- Format-on-save runs the same tool every time you save, so you never type the tabs and spaces yourself.

</details>

**Step done when:**

```bash
go version
go env GOROOT GOPATH
```

```
go version go1.26.2 linux/amd64
/usr/local/go
/home/nacho/go
```

(`linux/amd64` is your system and processor; on an Apple Silicon Mac it says `darwin/arm64`. Your paths may differ.) Then the editor check: in any `.go` file (the example's `main.go` will do), change `func main()` to `func main ()` and save. The space must disappear.

#### Step 0.2: Create the module

**Achieve:**
- A `mystally` folder that is a Go module called `github.com/na-cho-dev/mystally`.
- You can read `go.mod` and say what each of its two lines means.

**Learn:**
A **module** is your whole project: a folder with a `go.mod` file at the top. `go mod init <path>` creates that file.

The **module path** (`github.com/na-cho-dev/mystally`) is the project's full name. It does not need to exist online yet. Every package inside the project is imported as module path plus folder, for example `github.com/na-cho-dev/mystally/internal/ledger` in Stage 4.

`go.mod` has two lines for now. `module <path>` names the module. `go 1.26.2` is the minimum Go version the module needs, and it decides which language features are switched on.

A **package** is one folder of `.go` files that share a `package` line. A brand-new module has no packages yet, so `go list ./...` finds nothing.

Every `go` command looks for `go.mod` in the current folder, then in the folder above, and so on upwards. That is how it knows which module you are in, even when you stand in `cmd/mystally/`.

<details>
<summary><b>Example: a new module</b></summary>

Output:

```
$ go mod init example.com/recipes
go: creating new go.mod: module example.com/recipes
$ cat go.mod
module example.com/recipes

go 1.26.2
```

- The first line is the **module path**: the prefix of every import inside this module.
- The `go` line is the Go version the module was written for.

</details>

**Step done when:**

```bash
cat go.mod
go list -m
```

```
module github.com/na-cho-dev/mystally

go 1.26.2
github.com/na-cho-dev/mystally
```

(When you ran `go mod init github.com/na-cho-dev/mystally` it printed `go: creating new go.mod: module github.com/na-cho-dev/mystally`.)

#### Step 0.3: Write the program's entry point

**Achieve:**
- `cmd/mystally/main.go` exists and is a runnable program.
- Running it prints the version line `mystally 0.1` and nothing else.
- The file has gofmt's layout: `gofmt -l .` prints nothing.

**Learn:**
Every Go file starts with a `package` line. `package main` is special: it says "this folder is a program you can run", not a library for other code to import.

A `main` package needs one `func main()`. The program starts there and ends when `main` returns.

`import "fmt"` brings in a package by its import path. Standard library packages (`fmt`, `os`, `strings`) come with Go: nothing to download or install. With more than one import, the usual form is a group, `import ( ... )` with one path per line; the Step 0.6 example and Stage 1 use it.

In `func main() { ... }`, the `( )` holds the parameters (none here) and the `{ }` holds the body: the statements that run.

At the top level of a file you may only write declarations: `package`, `import`, `const`, `var`, `type`, `func`. A statement such as `fmt.Println(...)` must sit inside a function body.

<details>
<summary><b>Example: the smallest program</b></summary>

```go
package main // "main" makes this a program you can run

import "fmt" // standard library: nothing to install

func main() { // ( ) holds the parameters, { } holds the body
	fmt.Println("bookshelf 0.1")
}
```

Output:

```
$ go run .
bookshelf 0.1
```

- Execution always starts at `func main` in `package main`.
- At the top level of a file you may only declare things (`package`, `import`, `func`, ...). Statements such as `fmt.Println` live inside a function.

</details>

**Step done when:**

```bash
go build ./cmd/mystally
./mystally
```

```
mystally 0.1
```

(`go build` compiles the program into a file called `mystally`, and `./mystally` runs that file; Step 0.4 explains both.)

#### Step 0.4: Run it, then build it

**Achieve:**
- Run the program from source with `go run`.
- Build a real binary with `go build`, run it, and know where it landed and why it is called `mystally`.
- Always give `go run`, `go build` and `go vet` a path: `./cmd/mystally` for one package, `./...` for all of them.

**Learn:**
`go run ./cmd/mystally` compiles the package in that folder and runs it straight away. Since Go 1.24 the compiled program is kept in Go's build cache, out of sight, so nothing new appears in your folder.

`go build ./cmd/mystally` compiles the same package and writes a **binary** (the compiled program, a file you can run without Go) into the folder you are in. The binary is named after the package's folder, `mystally`, and you run it with `./mystally`.

The argument is a **package path**: a folder, not a file. The leading `./` means "relative to where I am". Without it, `cmd/mystally` reads as an import path, and Go looks for it inside its own source tree.

With no path at all, `go build` and `go vet` work on the folder you are in. The repo root has no `.go` files, so there they fail with `no Go files in ...`.

`./...` means "this folder and every folder below it". `go vet ./...` runs `go vet` (a checker for code that compiles but is probably wrong) on every package, and `go list ./...` prints every package's import path. See [Toolchain](#toolchain-the-whole-thing) for the full list of commands.

<details>
<summary><b>Example: <code>go run</code> versus <code>go build</code></b></summary>

`cmd/bookshelf/main.go`:

```go
package main

import "fmt"

func main() {
	fmt.Println("bookshelf 0.1")
}
```

Output:

```
$ go run ./cmd/bookshelf
bookshelf 0.1
$ go build ./cmd/bookshelf
$ ls
bookshelf
cmd
go.mod
$ ./bookshelf
bookshelf 0.1
```

- `go run` compiles to a temporary file and runs it. `go build` leaves a binary named after the folder.
- Point both at the folder, `./cmd/bookshelf`, not at `main.go`.

</details>

**Step done when:**

```bash
go build ./cmd/mystally
./mystally
go list ./...
```

```
mystally 0.1
github.com/na-cho-dev/mystally/cmd/mystally
```

(`go build` and `go vet` print nothing when they succeed.)

#### Step 0.5: Start git and ignore only the binary

**Achieve:**
- The `mystally` folder is a git repository.
- `.gitignore` hides the built binary and nothing else.
- You can prove it: git ignores `mystally` at the root but still sees `cmd/mystally/main.go`.

**Learn:**
`git init` turns the current folder into a git repository by creating a hidden `.git` folder. Nothing is tracked until you `git add` it (Step 0.7).

`.gitignore` lists paths git should act as if it can't see. Built binaries and data files go there: your code can recreate them, so they don't belong in history. It only affects files git does not track yet: if a file is already committed, adding it to `.gitignore` does not remove it.

Each line is a pattern. A pattern with no `/` in it matches a file **or folder** with that name at any depth. So `mystally` also matches the folder `cmd/mystally/`, and everything inside it vanishes from git.

A leading `/` anchors the pattern to the folder that holds the `.gitignore`. `/mystally` matches only the binary at the repo root.

Every command ends with an **exit code**: 0 means success, anything else means failure. `echo $?` prints the exit code of the command just before it.

Two commands show what git skips. `git check-ignore -v <path>` prints the file, line number and pattern that ignore a path; for a path that is not ignored it prints nothing and exits with 1. `git status --short --ignored` marks untracked paths (git sees them but does not track them yet) with `??` and ignored ones with `!!`.

<details>
<summary><b>Example: an unanchored pattern hides your source</b></summary>

`cmd/bookshelf/main.go`:

```go
package main

func main() {}
```

Output:

```
$ echo 'bookshelf' > .gitignore
$ git status --short --ignored
?? .gitignore
?? go.mod
!! bookshelf
!! cmd/
$ echo '/bookshelf' > .gitignore
$ git status --short --ignored
?? .gitignore
?? cmd/
?? go.mod
!! bookshelf
```

- `bookshelf` matches every file or folder with that name, so `cmd/bookshelf/` shows as ignored (`!!`).
- `/bookshelf` only matches at the top of the repo: now the binary is ignored and `cmd/` shows as untracked (`??`).

</details>

**Step done when:**

```bash
go build ./cmd/mystally
cat .gitignore
git check-ignore -v mystally
git check-ignore -v cmd/mystally/main.go; echo $?
git status --short --ignored
```

```
/mystally
.gitignore:1:/mystally	mystally
1
?? .gitignore
?? PLAN.md
?? cmd/
?? go.mod
!! mystally
```

(Run this after `git init` and before your first commit.)

#### Step 0.6: Read the docs from the terminal

**Achieve:**
- Read what `go doc os`, `go doc strings` and `go doc fmt.Println` print.
- Find a function's signature and description without opening a browser.

**Learn:**
`go doc` prints documentation from the Go installed on your machine, offline. It shows the same comments as pkg.go.dev.
- `go doc strings` shows the package overview, then one line for each name the package offers.
- `go doc fmt.Println` shows one function: its signature and what it does.
- `go doc -short strings` prints only the one-line list, which works well with `grep`.
- `go doc -all strings` prints everything. Long output reads better through `less` (`go doc -all strings | less`, then `q` to quit).

A **signature** is a function's first line. Read it left to right: `func Println(a ...any) (n int, err error)` takes any number of values of any type (`...any`) and returns two results, the number of bytes written and an error. You will meet the "result plus error" shape in nearly every Stage 1 function.

<details>
<summary><b>Example: from a question to a function</b></summary>

Output:

```
$ go doc -short strings | grep -i upper
func ToUpper(s string) string
func ToUpperSpecial(c unicode.SpecialCase, s string) string
$ go doc strings.ToUpper
package strings // import "strings"

func ToUpper(s string) string
    ToUpper returns s with all Unicode letters mapped to their upper case.
```

- Start from the noun (`strings`), scan the one-line list, then read the one function you need.
- The signature tells you the input, the output and whether it can fail.

</details>

**Step done when:**

```bash
go doc fmt.Println
```

```
package fmt // import "fmt"

func Println(a ...any) (n int, err error)
    Println formats using the default formats for its operands and writes to
    standard output. Spaces are always added between operands and a newline
    is appended. It returns the number of bytes written and any write error
    encountered.

```

(`go doc` also prints one empty line at the end.)

#### Step 0.7: Make the first commit

**Achieve:**
- Commit everything that belongs in the repo: `.gitignore`, `go.mod`, `cmd/mystally/main.go` and `PLAN.md`.
- Check that the commit really contains `cmd/mystally/main.go` and does not contain the binary.

**Learn:**
A **commit** is a saved snapshot of your files, with a message. Making one takes two commands:
- `git add .` stages every file in the folder except ignored ones. Staged means "goes into the next commit".
- `git commit -m "stage 0: setup"` saves the staged files as one snapshot.

Then check what actually went in, not what you meant to put in:
- `git status --short` prints nothing when everything is committed (a clean working tree). Before the commit, `A` marks a staged new file and `??` an untracked one.
- `git ls-files` lists every file git tracks.
- `git log --oneline` lists commits, newest first, one line each.

This is where the `.gitignore` mistake bites: an ignored `main.go` simply stays out of the commit, with no warning. Running `git ls-files` right after the commit catches it.

<details>
<summary><b>Example: the first commit, checked</b></summary>

Output:

```
$ git add .
$ git commit -q -m 'stage 0: setup'
$ git ls-files
.gitignore
cmd/bookshelf/main.go
go.mod
$ git status --short
$ git log --oneline
110a8ab stage 0: setup
```

(your commit hash will differ)

- `git ls-files` proves which files are in the repo. Check that your source is listed.
- An empty `git status --short` means everything is committed.

</details>

**Step done when:**

```bash
git status --short
git ls-files
git log --oneline
```

```
.gitignore
PLAN.md
cmd/mystally/main.go
go.mod
f51e5fb stage 0: setup
```

(Run this after `git add .` and `git commit -m "stage 0: setup"`.)

### Gotchas

- A function body goes in `{ }`, not `( )`. `syntax error: unexpected ( in parameter list; possibly missing comma or )` means Go is reading your body as a parameter list. *(you hit this)*
- At the top level of a file you can only write declarations: `package`, `import`, `const`, `var`, `type`, `func`. A statement like `fmt.Println(...)` must be inside a function. Otherwise: `syntax error: non-declaration statement outside function body`.
- In `.gitignore`, the pattern `mystally` matches any file **or folder** with that name at any depth, including `cmd/mystally/`. Your source silently disappears from `git status`. Use `/mystally`. Check with `git check-ignore -v cmd/mystally/main.go` (no output means it's not ignored). *(you hit this)*
- Standard library packages (`os`, `fmt`, `strings`...) are never installed. If the editor says `could not import os (missing metadata for import of "os")`, see [When the editor and the terminal disagree](#when-the-editor-and-the-terminal-disagree). *(you hit this)*
- `go build` or `go vet` with no path, run from the repo root, fails with `no Go files in /home/nacho/projects/personal/mystally`, because the root holds no `.go` files. Pass `./cmd/mystally` or `./...`. *(you hit this)*
- `gofmt -l .` lists a file even when the only difference is a missing newline at the very end. Your Stage 2 `storage.go` (just `package main`, no final newline) is listed for exactly that. Keep format-on-save on, or run `gofmt -w` on the file. *(you hit this)*

### Check yourself

- What is the difference between a module (`go.mod`) and a package (a folder of `.go` files)?
- Why does `go run ./cmd/mystally` work from the repo root but `go run main.go` doesn't?
- What does `./...` mean?
- Why does the `.gitignore` line `mystally` hide `cmd/mystally/main.go`, while `/mystally` does not?
- `go build ./cmd/mystally` wrote a file called `mystally`. Where did that name come from?

### Done when

- [ ] `go run ./cmd/mystally` prints the version line

  ```bash
  go run ./cmd/mystally
  ```

  ```
  mystally 0.1
  ```

- [ ] `go build ./cmd/mystally && ./mystally` does the same

  ```bash
  go build ./cmd/mystally && ./mystally
  ```

  ```
  mystally 0.1
  ```

- [ ] `git ls-files` lists `cmd/mystally/main.go`, and `git status` does not list the `mystally` binary

  ```bash
  go build ./cmd/mystally
  git ls-files
  git status --short
  git status --short --ignored
  ```

  ```
  .gitignore
  PLAN.md
  cmd/mystally/main.go
  go.mod
  !! mystally
  ```

  (`git status --short` prints nothing.)

- [ ] first commit exists

  ```bash
  git log --oneline
  git ls-tree -r --name-only HEAD
  ```

  ```
  f51e5fb stage 0: setup
  .gitignore
  PLAN.md
  cmd/mystally/main.go
  go.mod
  ```

  (At the end of Stage 0 there is exactly one commit.)

---

## Stage 1: a CLI that reads and writes a text file

**Goal:** the smallest useful program. Arguments in, file on disk, lines out.

**End goal:** you can run `./mystally add nacho like` to append a line such as `nacho,like,2026-10-17T15:28:32+01:00` to `events.txt`, and `./mystally list` to print every saved line. Bad input (no command, a missing argument, an unknown event, a creator name with a comma) prints a message on stderr and exits with code 1. You will understand how a Go program receives its arguments, how errors travel up to `main` as ordinary values, and how to append to a text file and read it back.

**Build**
- `mystally add <creator> <event>` appends one line `creator,event,timestamp` to `events.txt`
- `mystally list` prints every line, one per row
- Any bad input prints a message to stderr and exits with code 1
- Running with no command prints usage

**Learn:** `os.Args`, `os.ReadFile`, `os.OpenFile` with append, `strings.Split`, `strings.Join`, `time.Now`, `for ... range`, slices, the `if err != nil { return err }` shape, `os.Exit`.

**Look in:** `os`, `strings`, `fmt`, `time`.

### Structure

Everything still lives in one file, split into small functions with one job each:

```
mystally/
  .gitignore              /mystally and /events.txt (the leading slash means "repo root only")
  go.mod                  module github.com/na-cho-dev/mystally
  PLAN.md
  cmd/mystally/main.go    the whole program in Stage 1
  events.txt              written by the program, ignored by git
```

### Steps

Stage 1 has six steps. Do them in order: Steps 1.1 to 1.5 each add one piece to `cmd/mystally/main.go`, Step 1.6 tests and commits, and every step ends with a check you run against the built binary before you move on. Steps 1.1 and 1.2 only deal with arguments; the file first appears in Step 1.3.

#### Step 1.1: Route each command through run

**Achieve:**
- `main` only calls `run` with the arguments after the program name, prints any error to stderr, and exits with code 1.
- `run` picks the command from the first argument: `add`, `list`, or anything else.
- Each command checks its own argument count: `add` needs exactly three words (`add`, creator, event), `list` exactly one. No command, an unknown command, or a wrong count returns a usage error.
- For now `addCmd` and `listCmd` can simply return `nil`. Steps 1.3 and 1.4 fill them in.

**Learn:**
- `os.Args` is a `[]string` (a slice: a list of strings that can grow). `os.Args[0]` is the program path, then comes one item per word you typed. A missing argument is not an empty slot: the slice is just shorter, so check `len(args)` before you index into it.
- `os.Args[1:]` is a slice expression meaning "from index 1 to the end". On a one-item slice it gives an empty slice, not a panic (a panic is a crash with a stack trace).
- `switch x { case "a": ... default: ... }` runs the first `case` that matches, or `default` if none does. There is no `break` to forget.
- `errors.New("text")` and `fmt.Errorf("format %q", v)` **create** an `error` value. They print nothing. Printing is `main`'s job.

<details>
<summary><b>Example: one command per <code>case</code></b></summary>

```go
package main

import (
	"errors"
	"fmt"
	"os"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: greet hello <name>")
	}
	switch args[0] {
	case "hello":
		if len(args) != 2 {
			return errors.New("usage: greet hello <name>")
		}
		fmt.Println("hello,", args[1])
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
	return nil
}
```

Output:

```
$ go build -o greet .
$ ./greet hello Ada
hello, Ada
$ ./greet; echo $?
usage: greet hello <name>
1
$ ./greet dance; echo $?
unknown command "dance"
1
```

- `run` returns errors; only `main` prints them and picks the exit code.
- Each `case` checks its own argument count before reading `args[1]`.

</details>

**Step done when:**

```bash
go build ./cmd/mystally
./mystally; echo $?
./mystally dance; echo $?
./mystally add nacho; echo $?
./mystally list extra; echo $?
```

```
usage: mystally add <creator> <event>
usage: mystally list
1
usage: mystally add <creator> <event>
usage: mystally list
1
usage: mystally add <creator> <event>
1
usage: mystally list
1
```

(your wording may differ; every message must start with `usage:` and every exit code must be `1`)

#### Step 1.2: Reject unknown event names

**Achieve:**
- `add` accepts only the five event names from the [Domain](#domain) table.
- An unknown name returns an error that quotes what the user typed and lists the valid names in alphabetical order.
- Nothing is written yet: a valid `add` still does nothing and exits 0.

**Learn:**
- A map literal: `map[string]int{"small": 2, "large": 20}`. Each key appears once. Looking up a missing key gives the value type's zero value (`0` here), so a plain lookup can't tell "missing" from "really 0".
- Comma-ok tells them apart: `kilos, ok := limits[size]`. `ok` is `false` when the key isn't there. If you only need `ok`, write `_, ok := limits[size]` (`_` throws a value away). Never loop over a map to find a key.
- `for k := range m` visits a map's keys in no fixed order, and the order often changes from run to run. For a stable list, use `slices.Sorted(maps.Keys(m))` (packages `slices` and `maps`; both functions arrived in Go 1.23). `maps.Keys` gives an iterator, not a slice. `slices.Sorted` collects it into a sorted `[]string`.
- `strings.Join(names, ", ")` glues a `[]string` into one string with `", "` between the items.

<details>
<summary><b>Example: comma-ok, and a sorted list of choices</b></summary>

```go
package main

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

func main() {
	limits := map[string]int{"small": 2, "medium": 10, "large": 20}

	for _, size := range []string{"medium", "huge"} {
		kilos, ok := limits[size]
		if !ok {
			names := slices.Sorted(maps.Keys(limits))
			fmt.Printf("unknown size %q; choose one of: %s\n", size, strings.Join(names, ", "))
			continue
		}
		fmt.Println(size, "takes up to", kilos, "kg")
	}
}
```

Output:

```
$ go run .
medium takes up to 10 kg
unknown size "huge"; choose one of: large, medium, small
```

- `kilos, ok := limits[size]`: `ok` is false when the key is missing. Never loop over a map to find a key.
- `slices.Sorted(maps.Keys(m))` gives the keys in a stable order; `strings.Join` glues them into one string.

</details>

**Step done when:**

```bash
go build ./cmd/mystally
./mystally add nacho dance; echo $?
./mystally add nacho like; echo $?
```

```
usage: mystally add <creator> <event>
unknown event "dance"; choose one of: comment, like, save, share, view
1
0
```

(your wording may differ; it must contain `"dance"` and the five names in this order, and the valid `add` must exit `0`)

#### Step 1.3: Append one line per add

**Achieve:**
- `addCmd` builds the line `creator,event,timestamp`, with the time in RFC 3339 form.
- It opens `events.txt` for appending, creating it if it's missing, and writes the line plus a newline.
- It returns any error from opening or from writing, so `main` prints the real cause.
- Two `add` calls leave two lines in the file, never one.

**Learn:**
- `os.OpenFile(name, flags, perm)` returns `(*os.File, error)`. Combine flags with `|`: exactly one access mode (`os.O_WRONLY`) plus any behaviours (`os.O_APPEND` writes at the end, `os.O_CREATE` makes the file if it's missing). `perm` (`0o644`, octal; `0644` also works) is only used when the file is created: you can read and write, everyone else can read. See `go doc os.O_APPEND`.
- `defer f.Close()` schedules the close for when the function returns, whichever `return` that is. Put it **after** the `if err != nil` check: when the open fails, `f` is `nil`.
- `:=` declares new variables and sets them. `=` changes variables that already exist. `f, err := os.OpenFile(...)` declares both; a later `_, err = f.WriteString(...)` reuses `err`. `:=` needs at least one new name on its left.
- `time.Now().Format(time.RFC3339)` gives `2026-10-17T15:23:03+01:00`, a layout Stage 2 can parse back.

<details>
<summary><b>Example: append one line per run</b></summary>

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

func appendLine(path, line string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(line + "\n")
	return err
}
```

Output:

```
$ go run .
$ go run .
$ cat visits.log
ada,2026-10-18T08:48:00+01:00
ada,2026-10-18T08:48:01+01:00
```

(your times will differ)

- Write-only, append and create: the three flags together. Without `O_WRONLY` the file opens read-only and every write fails.
- `defer f.Close()` comes after the error check, because a failed open gives a `nil` file.

</details>

**Step done when:**

```bash
go build ./cmd/mystally
rm -f events.txt
./mystally add nacho like
./mystally add ada share
cat events.txt
```

```
nacho,like,2026-10-17T15:21:55+01:00
ada,share,2026-10-17T15:21:55+01:00
```

(your timestamps will differ)

#### Step 1.4: Print every line with list

**Achieve:**
- `listCmd` reads the whole of `events.txt` and prints each line to stdout, one per row.
- Empty pieces (there is one after the final newline) are skipped.
- A missing file must not panic. Either return the error (a message and exit 1) or treat it as "nothing to list" (no output and exit 0). Pick one on purpose.

**Learn:**
- `os.ReadFile(name)` returns `([]byte, error)`: the whole file as bytes.
- Three types look alike. `string` is text. `[]byte` is the raw bytes of text, and printing it shows numbers (`[68 117 110 101]`). `[]string` is a list of strings, and `fmt.Println` shows it as `[a b c]`. Convert bytes to text once with `string(data)`, then use the `strings` package.
- `strings.Split(s, "\n")` returns a `[]string`. A file that ends in `"\n"` gives one extra, empty piece at the end.
- `strings.SplitSeq(s, "\n")` (Go 1.24) gives the same pieces as an iterator, so you can range over it directly without building a slice: `for line := range strings.SplitSeq(text, "\n")`.

<details>
<summary><b>Example: read a file back as lines</b></summary>

```go
package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	data, err := os.ReadFile("shelf.txt")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for line := range strings.SplitSeq(string(data), "\n") {
		if line == "" {
			continue // the file ends in "\n", so the last piece is empty
		}
		fmt.Println("book:", line)
	}
}
```

Output:

```
$ go run .
open shelf.txt: no such file or directory
exit status 1
$ printf 'Dune\nEmma\n' > shelf.txt
$ go run .
book: Dune
book: Emma
```

- A missing file is an error, not a crash: `main` prints it and exits 1.
- `string(data)` turns the bytes into text before splitting.

</details>

**Step done when:**

```bash
go build ./cmd/mystally
rm -f events.txt
./mystally add nacho like
./mystally add ada share
./mystally add nacho comment
./mystally list
rm events.txt
./mystally list; echo $?
```

```
nacho,like,2026-10-17T15:21:55+01:00
ada,share,2026-10-17T15:21:55+01:00
nacho,comment,2026-10-17T15:21:55+01:00
open events.txt: no such file or directory
1
```

(your timestamps will differ; if you chose "a missing file means nothing to list", the last two lines are no output and `0`.)

#### Step 1.5: Refuse creator names that would break the line

**Achieve:**
- `add` rejects a creator name that contains a comma, before the file is opened.
- The error says what is wrong, exits 1, and `events.txt` is left unchanged.

**Learn:**
- Your file format is "fields separated by commas". A creator called `a,b` would be saved as `a,b,like,...`: four fields, and no reader could tell where the name ends. When data contains the separator it is stored with, the format breaks.
- `strings.Contains(s, substr)` reports whether `substr` appears anywhere in `s`: `strings.Contains("a,b", ",")` is `true`.
- `strings.ContainsAny(s, ",\n")` checks for any one of several characters. A newline inside a name would split one row into two, so it is worth refusing too.
- `strings.TrimSpace(s) == ""` catches a name that is empty or only spaces, if you want that check as well.

<details>
<summary><b>Example: refuse the separator inside a field</b></summary>

```go
package main

import (
	"errors"
	"fmt"
	"strings"
)

func checkTitle(title string) error {
	if strings.Contains(title, ",") {
		return errors.New("recipe title must not contain a comma")
	}
	return nil
}

func main() {
	fmt.Println(checkTitle("Jollof rice"))
	fmt.Println(checkTitle("Rice, beans"))
}
```

Output:

```
$ go run .
<nil>
recipe title must not contain a comma
```

- Check before writing: once a bad line is in the file, every reader has to deal with it.
- `<nil>` is how `fmt` prints "no error".

</details>

**Step done when:**

```bash
go build ./cmd/mystally
rm -f events.txt
./mystally add "a,b" like; echo $?
./mystally add ada like
cat events.txt
```

```
creator name must not contain a comma
1
ada,like,2026-10-17T15:21:55+01:00
```

(your wording and timestamps will differ; the rejected call must exit `1`, and no `a,b` line may appear in the file)

#### Step 1.6: Test by hand, then commit

**Achieve:**
- Run the whole Stage 1 checklist against the built binary, checking exit codes and which stream each message goes to.
- `go vet ./...` and `gofmt -l .` print nothing.
- Commit only after the checklist passes, with a message such as `stage 1: add and list events`.

**Learn:**
- Every program ends with an exit code: 0 means success, anything else means failure. `echo $?` prints the code of the last command.
- A program has two output streams. stdout (`fmt.Println`) carries the result; stderr (`fmt.Fprintln(os.Stderr, ...)`) carries errors. `2>/dev/null` hides stderr and `>/dev/null` hides stdout, so you can check each one alone.
- `go run` prints `exit status 3` and then exits with 1 itself. Check exit codes on the binary from `go build`.
- `os.Exit` ends the program immediately and deferred calls do not run. That is why only `main` calls it: by then `run` has returned and every `defer f.Close()` has already run.

<details>
<summary><b>Example: two streams, an exit code, and what <code>os.Exit</code> skips</b></summary>

```go
package main

import (
	"fmt"
	"os"
)

func main() {
	defer fmt.Println("this never prints")
	fmt.Println("result goes to stdout")
	fmt.Fprintln(os.Stderr, "problem goes to stderr")
	os.Exit(3)
}
```

Output:

```
$ go build -o streams .
$ ./streams; echo $?
result goes to stdout
problem goes to stderr
3
$ ./streams 2>/dev/null
result goes to stdout
$ go run .; echo $?
result goes to stdout
problem goes to stderr
exit status 3
1
```

- `os.Exit` ends the program at once, so the deferred line never runs. Only `main` should call it.
- `go run` prints `exit status 3` but itself exits with 1. Check exit codes on the built binary.

</details>

**Step done when:**

```bash
go build ./cmd/mystally
rm -f events.txt
./mystally list; echo $?
./mystally add nacho like
./mystally add ada share
./mystally add nacho comment
./mystally list
./mystally add nacho; echo $?
./mystally; echo $?
./mystally add nacho dance; echo $?
./mystally add nacho dance 2>/dev/null; echo $?
./mystally list >/dev/null; echo $?
```

```
open events.txt: no such file or directory
1
nacho,like,2026-10-17T15:28:32+01:00
ada,share,2026-10-17T15:28:32+01:00
nacho,comment,2026-10-17T15:28:32+01:00
usage: mystally add <creator> <event>
1
usage: mystally add <creator> <event>
usage: mystally list
1
usage: mystally add <creator> <event>
unknown event "dance"; choose one of: comment, like, save, share, view
1
1
0
```

(your wording and timestamps will differ.)

### Gotchas

- Indexing past the end of a slice doesn't return an error. It **panics**: `panic: runtime error: index out of range [3] with length 2`. Check `len(args)` before every index. You hit this when the `add`-only checks ran before the `switch`, so `mystally list` (two items in `os.Args`) asked for item 3. Slicing at the end is fine: `os.Args[1:]` on a one-item slice gives an empty slice, not a panic. *(you hit this)*
- A missing argument isn't an empty slot in `os.Args`. The slice is just shorter. Looping over it can never find what isn't there. *(you hit this)*
- Passing `os.Args[1:]` shifts every index down by one. If you drop the program path but keep the old indexes, every length check fails and every command prints usage. You once committed exactly that change: run the Step 1.6 checklist before every commit. *(you hit this)*
- A `string` is never `nil`; its zero value is `""`. `name == nil` gives `invalid operation: name == nil (mismatched types string and untyped nil)`. *(you hit this)*
- A function that returns `string` can't return `nil`: `cannot use nil as string value in return statement`. Return `""`, and return an `error` alongside it when something went wrong. *(you hit this)*
- `fmt.Errorf("...")` on its own line prints nothing. `go vet` says `result of fmt.Errorf call not used`. To print, use `fmt.Fprintln(os.Stderr, ...)`; to report, `return fmt.Errorf(...)`. *(you hit this)*

### Check yourself

- What is in `os.Args[0]`, and why does `run` get `os.Args[1:]`?
- Why does only `main` call `os.Exit`?
- What is the difference between `x := 1` and `x = 1`? When does each one fail to compile?
- Why does splitting your file on `"\n"` give one more piece than there are lines?
- Why does `defer f.Close()` go after the `if err != nil` check and not before it?
- `maps.Keys` gives the event names in a different order on each run. What do you wrap around it so the error message always lists them the same way, and why can't you pass `maps.Keys(...)` straight to `strings.Join`?

### Done when

- [ ] three `add` calls then `list` shows three rows

  ```bash
  go build ./cmd/mystally
  rm -f events.txt
  ./mystally add nacho like
  ./mystally add ada share
  ./mystally add nacho comment
  ./mystally list
  ./mystally list | wc -l
  ```

  ```
  nacho,like,2026-10-17T15:28:20+01:00
  ada,share,2026-10-17T15:28:20+01:00
  nacho,comment,2026-10-17T15:28:20+01:00
  3
  ```

    (your timestamps will differ)

- [ ] `mystally add` with a missing argument exits 1 with a clear message

  ```bash
  go build ./cmd/mystally
  ./mystally add nacho; echo $?
  ./mystally add; echo $?
  ```

  ```
  usage: mystally add <creator> <event>
  1
  usage: mystally add <creator> <event>
  1
  ```

    (your wording may differ; it must name the missing pieces, `<creator>` and `<event>`)

- [ ] `mystally list` on a missing file does not crash

  ```bash
  go build ./cmd/mystally
  rm -f events.txt
  ./mystally list; echo $?
  ```

  ```
  open events.txt: no such file or directory
  1
  ```

    (if you treat a missing file as empty, you get no output and `0`; either passes, but `panic:` must never appear)

- [ ] `go vet ./...` is clean

  ```bash
  go vet ./...; echo $?
  ```

  ```
  0
  ```

---

## Stage 2: real types and JSON

**Goal:** replace strings with a model that the compiler can check.

**End goal:** you can run `./mystally add nacho like` to add an event to `events.json`, a JSON array you can read by eye, and `./mystally list` to print every event as an aligned table with the columns ID, CREATOR, TYPE and TIME. `./mystally add nacho dance` prints an `unknown event type` message and exits with code 1. You will understand named types, methods and their receivers, `iota` constants, structs with JSON tags, and how `encoding/json` turns your own types into text and back.

**Build**
- `type Kobo int64` with a `Naira() float64` method and a `String()` method that prints `₦12.50`
- `type EventType int` with `iota` constants for the five event types, a `String()` method, and a `ParseEventType(s string) (EventType, error)` function
- `type Event struct` with `ID`, `Creator`, `Type`, `At`, with JSON tags
- Storage moves from `events.txt` to `events.json`: a JSON array of events
- `mystally list` prints an aligned table: id, creator, type, time

**Learn:** named types, methods and receivers, `iota`, struct tags, `encoding/json` Marshal and Unmarshal, `fmt.Sprintf` verbs like `%-10s` and `%6d`, `time.Time` and `time.RFC3339`.

**Look in:** `encoding/json`, `fmt`, `time`, `strconv`.

### Structure

```
mystally/
  .gitignore              /mystally and /events.json
  go.mod                  module github.com/na-cho-dev/mystally
  PLAN.md
  cmd/mystally/           one folder = one package (package main)
    main.go               the CLI: run, addCmd, listCmd, table printing. No type definitions
    event.go              Kobo, EventType and its constants, ParseEventType, Event
    storage.go            loadEvents, saveEvents
  events.json             written by the program, ignored by git
```

### Steps

Stage 2 has ten steps. Do them in order: Steps 2.1 to 2.5 build the types in `event.go` (Step 2.4 also switches the `add` check to `ParseEventType`), Step 2.6 adds `storage.go`, Steps 2.7 and 2.8 move `add` and `list` onto JSON, Step 2.9 saves event types as words, and Step 2.10 tidies up and commits. Each step ends with a check you run before you move on, usually against the built binary.

#### Step 2.1: Give money its own type

**Achieve:**
- `cmd/mystally/event.go` (next to `main.go`, not in `cmd/`) declares `Kobo`, a whole number of kobo.
- `Naira()` converts it to naira for display, and `String()` makes it print as `₦12.50`.
- The package builds again. Your half-written `String` for `EventType` has no return type yet, so delete it or comment it out for now; Step 2.3 brings it back.
- `gofmt -l .` prints nothing. Your `storage.go` was missing its final newline, and `gofmt` lists any file that isn't formatted exactly.

**Learn:**
You have done most of this already. Use the step to check your work and to understand each mistake you fixed on the way.
- `type Kobo int64` declares a **named type**: values are stored as an `int64`, but the compiler treats `Kobo` as a separate type. You convert on purpose: `Kobo(1250)`, `int64(kb)`, `float64(kb)`.
- A **method** is a function attached to a type. The part in brackets before the name is the **receiver**: `func (m Minutes) Hours() float64` means "every `Minutes` has an `Hours` method, and inside it `m` is the value it was called on". You call it with a dot: `film.Hours()`.
- A **value receiver** `(m Minutes)` works on a copy. A **pointer receiver** `(m *Minutes)` gets the address of the caller's variable so it can change it. A method that only reads a small value takes a value receiver.
- Any type with a method exactly `String() string` is printed through it by `fmt.Println`, `%v` and `%s`. `%d` still prints the number underneath.

<details>
<summary><b>Example: a named type with methods</b></summary>

```go
package main

import "fmt"

// Minutes is a duration in whole minutes.
type Minutes int64

// Hours converts for display. Value receiver: it only reads m.
func (m Minutes) Hours() float64 {
	return float64(m) / 60
}

// String makes fmt print "2h05m". Value receiver: it only reads m.
func (m Minutes) String() string {
	return fmt.Sprintf("%dh%02dm", int64(m)/60, int64(m)%60)
}

func main() {
	m := Minutes(125)
	fmt.Println(m, m.Hours())
	fmt.Printf("%v %d %.1f\n", m, m, m.Hours())
}
```

Output:

```
$ go run .
2h05m 2.0833333333333335
2h05m 125 2.1
```

- `fmt.Println(m)` and `%v` call `String()`; `%d` prints the number underneath.
- Inside `String()` the value is converted with `int64(m)` first, so `fmt` never calls `String()` again.

</details>

**Step done when:**

Scratch line: `fmt.Println(Kobo(1250), Kobo(1205), Kobo(7))`

```bash
go build ./cmd/mystally
./mystally 2>/dev/null
```

```
₦12.50 ₦12.05 ₦0.07
```

(`gofmt -l .` and `go vet ./...` print nothing.)

#### Step 2.2: Name the five event types

**Achieve:**
- `EventType` is a named type over `int`, with one constant per event type in a `const ( ... )` block that uses `iota`.
- The constants start with a capital letter: `View`, `Like`, `Save`, `Comment`, `Share`. Rename your lowercase `view`, `like`, ... now.
- You decide what the number 0 means. Recommended: put `Unknown` first, so 0 means "not set" and a forgotten `Type` can't quietly become a view.

**Learn:**
- `const` declares a **constant**: a value fixed when the program is compiled, which can never change. Parentheses group several constants.
- `iota` is a counter that only exists inside a `const ( )` block. It is 0 on the first line, 1 on the second, 2 on the third, and so on.
- A line with only a name repeats the line above it: after `Unknown Weekday = iota`, the line `Monday` means `Monday Weekday = iota`, and by then `iota` is 1.
- A named type doesn't mix with plain `int` variables: `var d Weekday = n` (with `n` an `int`) does not compile. A plain number written in the code, like `3`, is fine, and so is an explicit conversion like `Weekday(n)`.

<details>
<summary><b>Example: <code>iota</code> constants</b></summary>

```go
package main

import "fmt"

type Weekday int

const (
	Unknown Weekday = iota // 0: means "not set"
	Monday                 // 1: a bare name repeats "Weekday = iota"
	Tuesday                // 2
)

func main() {
	var d Weekday // never set: holds the zero value
	fmt.Println(Unknown, Monday, Tuesday, d == Unknown)
	fmt.Printf("%T\n", Tuesday)
}
```

Output:

```
$ go run .
0 1 2 true
main.Weekday
```

- `iota` counts 0, 1, 2 down the block. They print as numbers until the type has a `String` method.
- A variable you never set holds 0, which is why `Unknown` comes first.

</details>

**Step done when:**

Scratch line: `fmt.Println(View, Like, Save, Comment, Share)`

```bash
go build ./cmd/mystally
./mystally 2>/dev/null
```

```
1 2 3 4 5
```

(without `Unknown` first you get `0 1 2 3 4`; either is fine if you chose it on purpose.)

#### Step 2.3: Print an event type as its name

**Achieve:**
- `EventType` gets a method with the exact signature `String() string`. It uses a `switch` with one `case` per constant and returns the lowercase name a user types: `like`.
- Any other number (`Unknown`, or a value made with `EventType(9)`) returns a fallback such as `EventType(9)`, never an empty string.
- `String` returns the text. It never prints anything.

**Learn:**
- In `func (t EventType) String() string`, the last `string` is the **return type**. Leave it out and the function returns nothing, so `return "like"` fails with `too many return values`.
- A function with a return type must end with a `return` on every path. A `switch` with a `default:` that returns, or one `return` after the switch, covers "anything else". Forget it and Go says `missing return`.
- In the fallback, format the number with `int(t)`, not `t` with `%v` or `%s`. You saw in Step 2.1 what happens when `String` formats its own receiver as text.
- `fmt.Sprintf` builds a string and returns it. `fmt.Sprintln()` with no arguments returns just `"\n"`, so even with `string` added to the signature, your current body would hand back a newline, not a name.

<details>
<summary><b>Example: <code>String</code> with a fallback</b></summary>

```go
package main

import "fmt"

type Weekday int

const (
	Unknown Weekday = iota
	Monday
	Tuesday
)

// String returns the name a user would type. Value receiver: it only reads d.
func (d Weekday) String() string {
	switch d {
	case Monday:
		return "monday"
	case Tuesday:
		return "tuesday"
	}
	return fmt.Sprintf("Weekday(%d)", int(d))
}

func main() {
	fmt.Println(Monday, Tuesday, Unknown, Weekday(9))
}
```

Output:

```
$ go run .
monday tuesday Weekday(0) Weekday(9)
```

- The signature must be exactly `String() string`, and the method returns the text; it never prints.
- Every other value falls through to the fallback, formatted with `int(d)`, so a bad value is visible instead of blank.

</details>

**Step done when:**

Scratch line: `fmt.Println(View, Like, Save, Comment, Share, EventType(0), EventType(9))`

```bash
go build ./cmd/mystally
./mystally 2>/dev/null
```

```
view like save comment share EventType(0) EventType(9)
```

(without `Unknown` first, `EventType(0)` prints `view`.)

#### Step 2.4: Turn what the user typed into an event type

**Achieve:**
- `ParseEventType(s string) (EventType, error)` turns `"like"` into `Like`. For any other text it returns the zero value and an error whose text contains `unknown event type` and quotes what the user typed.
- `add` checks the event name with `ParseEventType` instead of the Stage 1 `map[string]int`. Delete the map, and the imports only it used (`maps`, `slices`).
- Call it where your map lookup was (in `run`'s `add` case), or at the top of `addCmd`. Step 2.7 needs the value it returns inside `addCmd`, so the top of `addCmd` saves a move later.
- `add` still appends to `events.txt` for now. Only the check changes in this step.

**Learn:**
- The "value plus error" shape from Stage 1 again: on success return the value and `nil`; on failure return the zero value (`Unknown`, if you added it) and an error. The caller checks `err` before using the value.
- Two ways to write it. A `switch` on the string spells each name twice (here and in `String`). A loop over the constants that compares `t.String() == s` spells each name once. The loop works because `iota` numbers are consecutive: `for d := Monday; d <= Wednesday; d++` in the examples. If you loop up to your last constant, move that end when you add a type.
- The same loop can collect the valid names in constant order for the error message, so the user sees what to type instead.
- `%q` quotes the input, so an empty argument shows up as `""`.

<details>
<summary><b>Example: parse text by looping over the constants</b></summary>

```go
package main

import "fmt"

type Weekday int

const (
	Unknown Weekday = iota
	Monday
	Tuesday
)

func (d Weekday) String() string {
	switch d {
	case Monday:
		return "monday"
	case Tuesday:
		return "tuesday"
	}
	return fmt.Sprintf("Weekday(%d)", int(d))
}

func ParseWeekday(s string) (Weekday, error) {
	for d := Monday; d <= Tuesday; d++ {
		if d.String() == s {
			return d, nil
		}
	}
	return Unknown, fmt.Errorf("unknown weekday %q", s)
}

func main() {
	fmt.Println(ParseWeekday("tuesday"))
	fmt.Println(ParseWeekday("friday"))
}
```

Output:

```
$ go run .
tuesday <nil>
Weekday(0) unknown weekday "friday"
```

- `d++` works because a `Weekday` is an `int` underneath; each name is written once, in `String`.
- On failure it returns `Unknown` (0) with the error, so a forgotten error check can't hand back a real day.

</details>

**Step done when:**

```bash
go build ./cmd/mystally
rm -f events.txt
./mystally add nacho dance; echo $?
./mystally add nacho like; echo $?
cat events.txt
```

```
unknown event type "dance"; choose one of: view, like, save, comment, share
1
0
nacho,like,2026-10-17T15:41:59+01:00
```

(your wording and timestamp will differ; the message must contain `unknown event type`, and the valid `add` must exit `0`)

#### Step 2.5: Describe one event with a struct

**Achieve:**
- `event.go` declares `type Event struct` with the fields `ID`, `Creator`, `Type` and `At`, tagged `json:"id"`, `json:"creator"`, `json:"type"` and `json:"at"`.
- `Type` is an `EventType` and `At` is a `time.Time`, not strings, so the compiler checks them for you.
- Every field name starts with a capital letter. `encoding/json` can't see lowercase fields (Step 2.6).

**Learn:**
- A **struct** groups named fields into one value: `type Book struct { ID int; Title string }`.
- A **composite literal** builds one: `Book{ID: 1, Title: "Dune"}`. Fields you leave out get their zero value. Always write the field names; `Book{1, "Dune"}` breaks as soon as you add a field.
- Read and change fields with a dot: `b.Title`, `b.Pages = 215`.
- `%+v` prints a struct with its field names. It is the quickest way to see what a struct holds.

<details>
<summary><b>Example: a struct, its zero values, and a copy</b></summary>

```go
package main

import "fmt"

type Book struct {
	ID    int
	Title string
	Pages int
}

func main() {
	b := Book{ID: 1, Title: "Dune"} // Pages is left out: it gets 0
	fmt.Printf("%+v\n", b)

	c := b // copies every field
	c.Pages = 412
	fmt.Println(b.Pages, c.Pages)
}
```

Output:

```
$ go run .
{ID:1 Title:Dune Pages:0}
0 412
```

- Name the fields in a literal; any you leave out get their zero value.
- `%+v` prints the field names, the quickest way to see what a struct holds.

</details>

**Step done when:**

Scratch line: `fmt.Printf("%+v\n", Event{ID: 1, Creator: "nacho", Type: Like})`

```bash
go build ./cmd/mystally
./mystally 2>/dev/null
go doc -cmd -u ./cmd/mystally Event
```

```
{ID:1 Creator:nacho Type:like At:0001-01-01 00:00:00 +0000 UTC}
package main // import "github.com/na-cho-dev/mystally/cmd/mystally"

type Event struct {
	ID      int       `json:"id"`
	Creator string    `json:"creator"`
	Type    EventType `json:"type"`
	At      time.Time `json:"at"`
}
    Event is one engagement event.
```

(`go doc` reads your own code: `-cmd` lets it show a `package main`, and `-u` includes lowercase names.)

#### Step 2.6: Load and save events as JSON

**Achieve:**
- `storage.go` gets `loadEvents(path string) ([]Event, error)` and `saveEvents(path string, events []Event) error`.
- `loadEvents` reads the whole file and decodes it. A missing file is not an error: it returns an empty slice and `nil`.
- `saveEvents` encodes the whole slice as indented JSON and replaces the file.
- Nothing calls them yet. Step 2.7 does.

**Learn:**
- `json.MarshalIndent(v, "", "  ")` turns a value into JSON bytes, indenting each level by two spaces. `json.Marshal(v)` does the same on one line.
- `json.Unmarshal(data, &v)` fills `v` from JSON bytes. It needs the **address** of your variable (`&v`) so it can write into it.
- `os.WriteFile(path, data, 0o644)` creates or replaces the whole file in one call. `os.ReadFile` reads it all back.
- `errors.Is(err, os.ErrNotExist)` spots "no file yet", the check Stage 1 introduced (your Stage 1 `listCmd` returns the error instead). Check it before the general `err != nil`.

<details>
<summary><b>Example: load and save a JSON list</b></summary>

```go
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type Todo struct {
	Text string `json:"text"`
}

func load(path string) ([]Todo, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil // no file yet: an empty list
	}
	if err != nil {
		return nil, err
	}
	var todos []Todo
	err = json.Unmarshal(data, &todos) // & lets it fill todos
	return todos, err
}

func save(path string, todos []Todo) error {
	data, err := json.MarshalIndent(todos, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func main() {
	todos, err := load("todos.json")
	fmt.Println(len(todos), err)
	todos = append(todos, Todo{Text: "buy rice"})
	fmt.Println(save("todos.json", todos))
}
```

Output:

```
$ go run .
0 <nil>
<nil>
$ go run .
1 <nil>
<nil>
$ cat todos.json
[
  {
    "text": "buy rice"
  },
  {
    "text": "buy rice"
  }
]
```

- A missing file is the normal "nothing saved yet" case, checked with `errors.Is` before the general `err != nil`.
- The whole file is rewritten every time: a JSON array can't be appended to like a text file.

</details>

**Step done when:**

Scratch lines (paste all three, in this order):
- `fmt.Println(loadEvents("events.json"))`
- `fmt.Println(saveEvents("events.json", []Event{{ID: 1, Creator: "nacho", Type: Like}}))`
- `fmt.Println(loadEvents("events.json"))`

```bash
go build ./cmd/mystally
rm -f events.json
./mystally 2>/dev/null
cat events.json
rm events.json
go doc -cmd -u -short ./cmd/mystally
```

```
[] <nil>
<nil>
[{1 nacho like 0001-01-01 00:00:00 +0000 UTC}] <nil>
[
  {
    "id": 1,
    "creator": "nacho",
    "type": 2,
    "at": "0001-01-01T00:00:00Z"
  }
]
const eventsFile = "events.txt"
func addCmd(args []string) error
func listCmd() error
func main()
func run(args []string) error
func saveEvents(path string, events []Event) error
type Event struct{ ... }
    func loadEvents(path string) ([]Event, error)
type EventType int
    const Unknown EventType = iota ...
    func ParseEventType(s string) (EventType, error)
type Kobo int64
```

(`Println` accepts a call that returns two values when it is the only argument.)

#### Step 2.7: Rewrite add as load, append, save

**Achieve:**
- `eventsFile` becomes `"events.json"`. `list` keeps its Stage 1 body for now, so it prints the raw lines of `events.json` until Step 2.8 rewrites it.
- `add` now: parse the event type → `loadEvents` → work out the next ID (the highest existing ID + 1, or 1 for an empty file) → append a new `Event` stamped with `time.Now()` → `saveEvents`.
- Delete the Stage 1 text-file code in `addCmd` (`os.OpenFile`, joining the line with commas). Stage 1 lives on in git history. The comma check on the creator name is no longer needed (JSON quotes every string), but keeping it does no harm.
- In `.gitignore`, replace `/events.txt` with `/events.json`, and delete your old `events.txt`.

**Learn:**
- To change a JSON file you rewrite all of it: load the whole slice, change it in memory, save the whole slice.
- `append` returns the longer slice, and you must store it: `items = append(items, x)`. A bare `append(items, x)` on its own line doesn't compile: Go says its value `is not used`.
- Use the highest ID + 1, not `len(events) + 1`. If an event is ever removed, `len + 1` hands out an ID that is still in use. A loop that remembers the biggest ID seen is enough.
- `time.Time` saves itself in JSON as RFC 3339 text with nanoseconds and your UTC offset: `"2026-10-17T15:42:31.783824293+01:00"`. If you prefer whole seconds, store `time.Now().Truncate(time.Second)`.

<details>
<summary><b>Example: the next number is the highest + 1</b></summary>

```go
package main

import "fmt"

type Ticket struct {
	ID    int
	Plate string
}

func nextID(tickets []Ticket) int {
	highest := 0
	for _, t := range tickets {
		highest = max(highest, t.ID)
	}
	return highest + 1
}

func main() {
	tickets := []Ticket{{ID: 1, Plate: "LAG-1"}, {ID: 3, Plate: "ABJ-7"}}
	tickets = append(tickets, Ticket{ID: nextID(tickets), Plate: "KAN-2"})
	fmt.Printf("%+v\n", tickets)
}
```

Output:

```
$ go run .
[{ID:1 Plate:LAG-1} {ID:3 Plate:ABJ-7} {ID:4 Plate:KAN-2}]
```

- `len(tickets) + 1` would give 3 again; the highest ID + 1 never repeats one.
- `append` returns the longer slice, so store it: `tickets = append(tickets, ...)`.

</details>

**Step done when:**

```bash
go build ./cmd/mystally
rm -f events.json events.txt
./mystally add nacho like
./mystally add ada share
./mystally add nacho dance; echo $?
cat events.json
git check-ignore -v events.json
```

```
unknown event type "dance"; choose one of: view, like, save, comment, share
1
[
  {
    "id": 1,
    "creator": "nacho",
    "type": 2,
    "at": "2026-10-17T15:42:31.783824293+01:00"
  },
  {
    "id": 2,
    "creator": "ada",
    "type": 5,
    "at": "2026-10-17T15:42:31.795936752+01:00"
  }
]
.gitignore:2:/events.json	events.json
```

(your wording and timestamps will differ; the `dance` line must contain `unknown event type` and be followed by `1`.)

#### Step 2.8: Print the list as an aligned table

**Achieve:**
- `list` loads with `loadEvents`, prints a header row `ID CREATOR TYPE TIME`, then one row per event, with every column lined up.
- TYPE shows the name (`like`) through your `String` method. TIME is easy to read, for example `2026-10-17 15:42:58`.
- A missing `events.json` prints only the header (or a message such as `no events yet`) and exits 0.

**Learn:**
- A number between `%` and the verb is a minimum **width**. `%-10s` makes text at least 10 characters wide, lined up on the left (that is what the `-` does). `%5d` makes a number at least 5 characters wide, lined up on the right. Use the same widths in the header and in the rows.
- A value longer than its width is never cut: it pushes the rest of the row out of line. Pick widths that fit your usual values.
- Widths count characters, not bytes. `₦` and `é` take more than one byte but count as one character, so columns still line up.
- `%s` and `%-8s` on a value with a `String` method use that method, so `e.Type` prints `like` with no extra code.

<details>
<summary><b>Example: an aligned table</b></summary>

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	added := time.Date(2026, 10, 17, 9, 5, 0, 0, time.UTC)
	fmt.Printf("%-4s %-12s %5s  %s\n", "ID", "RECIPE", "MINS", "ADDED")
	fmt.Printf("%-4d %-12s %5d  %s\n", 1, "jollof", 45, added.Format(time.DateTime))
	fmt.Printf("%-4d %-12s %5d  %s\n", 12, "moi moi", 90, added.Format("2 Jan 15:04"))
}
```

Output:

```
$ go run .
ID   RECIPE        MINS  ADDED
1    jollof          45  2026-10-17 09:05:00
12   moi moi         90  17 Oct 09:05
```

- The number is the minimum width; `-` lines text up on the left. Widths count characters, so `₦` is one.
- Time layouts are written with Go's reference time: `2006-01-02 15:04:05`, or ready-made ones like `time.DateTime`.

</details>

**Step done when:**

```bash
go build ./cmd/mystally
rm -f events.json
./mystally list; echo $?
./mystally add nacho like
./mystally add ada share
./mystally add nacho comment
./mystally list
```

```
ID   CREATOR    TYPE     TIME
0
ID   CREATOR    TYPE     TIME
1    nacho      like     2026-10-17 15:42:58
2    ada        share    2026-10-17 15:42:58
3    nacho      comment  2026-10-17 15:42:58
```

(your widths, header wording and timestamps may differ; the columns must line up, TYPE must show names, and the empty list must exit `0`.)

#### Step 2.9: Store event types as words in events.json

**Achieve:**
- `EventType` gets `MarshalText() ([]byte, error)` with a value receiver, and `UnmarshalText(b []byte) error` with a pointer receiver.
- `events.json` now says `"type": "like"`, which you can read by eye. Reordering the constants later can't silently change what an old file means.
- `UnmarshalText` uses `ParseEventType`, so a file with an unknown type fails to load with the same `unknown event type` error.
- Files saved before this step hold numbers and no longer load. Delete `events.json`: there is no real data in it yet.

**Learn:**
- When `encoding/json` saves a value whose type has a `MarshalText` method, it calls it and stores the result as a JSON string. When it loads into a type with `UnmarshalText`, it passes the text between the quotes. [Stage 4](#stage-4-packages-and-the-store-interface) explains the mechanism (interfaces); for now, the exact method names and signatures are what count.
- `MarshalText` only reads, so it takes a value receiver. A pointer receiver here only works when `json` can reach the field's address, which depends on how `Marshal` was called. That is a bug waiting to happen.
- `UnmarshalText` must change the caller's variable, so it takes a pointer receiver. With a value receiver it would change a copy, and the loaded value would stay 0 with no error.
- `*t = parsed` means "store `parsed` in the variable that `t` points to".

<details>
<summary><b>Example: a type that saves itself as a word in JSON</b></summary>

```go
package main

import (
	"encoding/json"
	"fmt"
)

type Weekday int

const (
	Unknown Weekday = iota
	Monday
)

// MarshalText saves the name. Value receiver: it only reads d.
func (d Weekday) MarshalText() ([]byte, error) {
	if d == Monday {
		return []byte("monday"), nil
	}
	return nil, fmt.Errorf("cannot save Weekday(%d)", int(d))
}

// UnmarshalText loads the name. Pointer receiver: it changes the caller's value.
func (d *Weekday) UnmarshalText(b []byte) error {
	if string(b) != "monday" {
		return fmt.Errorf("unknown weekday %q", b)
	}
	*d = Monday
	return nil
}

type Shift struct {
	Day Weekday `json:"day"`
}

func main() {
	data, _ := json.Marshal(Shift{Day: Monday})
	fmt.Println(string(data))

	var s Shift
	fmt.Println(json.Unmarshal([]byte(`{"day":"monday"}`), &s), s.Day == Monday)
	fmt.Println(json.Unmarshal([]byte(`{"day":"funday"}`), &s))
}
```

Output:

```
$ go run .
{"day":"monday"}
<nil> true
unknown weekday "funday"
```

- `encoding/json` calls these methods for you, so the file holds `"monday"` instead of `1`.
- `UnmarshalText` needs a pointer receiver; `*d = Monday` stores into the caller's variable.

</details>

**Step done when:**

```bash
go build ./cmd/mystally
./mystally list; echo $?
rm events.json
./mystally add nacho like
./mystally add ada share
grep '"type"' events.json
./mystally list
sed -i 's/"share"/"dance"/' events.json
./mystally list; echo $?
rm events.json
```

```
json: cannot unmarshal number into Go struct field Event.type of type main.EventType
1
    "type": "like",
    "type": "share",
ID   CREATOR    TYPE     TIME
1    nacho      like     2026-10-17 15:43:06
2    ada        share    2026-10-17 15:43:06
unknown event type "dance"; choose one of: view, like, save, comment, share
1
```

(your wording, column widths and timestamps will differ.)

#### Step 2.10: Explain every receiver, then check and commit

**Achieve:**
- Every method has a one-line comment saying why its receiver is a value or a pointer. Make it the last line of the method's doc comment, directly above `func`.
- Doc comments start with the name of the thing they describe: `// Naira converts kb to naira, for display only.`
- `gofmt -l .` and `go vet ./...` print nothing, and the whole **Done when** list below passes.
- Commit with a message such as `stage 2: real types and JSON`.

**Learn:**
- The rule: a **pointer** receiver when the method must change the value (`UnmarshalText`) or the value is big to copy; a **value** receiver when it only reads a small value (`Naira`, `String`, `MarshalText`).
- A value-receiver method works on a value and on a pointer; Go follows the pointer for you. A pointer-receiver method works on a pointer, and on a variable (Go takes its address for you), but not on a copy handed to `fmt`. That is why a pointer-receiver `String` is ignored when you print a plain value.
- A **doc comment** is the comment directly above a declaration. `go doc` shows it, and so does your editor when you hover.
- `grep -B1 '^func (' cmd/mystally/*.go` prints every method line with the line just above it, so you can check every receiver comment at once.

<details>
<summary><b>Example: receiver comments that say why</b></summary>

```go
package main

import "fmt"

type Counter struct {
	n int
}

// Inc adds one. Pointer receiver: it changes the counter.
func (c *Counter) Inc() { c.n++ }

// Value returns the count. Value receiver: it only reads a small struct.
func (c Counter) Value() int { return c.n }

func main() {
	var c Counter
	c.Inc() // Go takes &c for you
	c.Inc()
	fmt.Println(c.Value())
}
```

Output:

```
$ go run .
2
$ grep -B1 '^func (' main.go
// Inc adds one. Pointer receiver: it changes the counter.
func (c *Counter) Inc() { c.n++ }
--
// Value returns the count. Value receiver: it only reads a small struct.
func (c Counter) Value() int { return c.n }
```

- Pointer receiver when the method changes the value; value receiver when it only reads something small.
- `grep -B1 '^func ('` shows each method with the comment above it, so you can check them all at once.

</details>

**Step done when:**

```bash
grep -B1 '^func (' cmd/mystally/*.go
```

```
cmd/mystally/event.go-// Value receiver: ...
cmd/mystally/event.go:func (kb Kobo) Naira() float64 {
--
cmd/mystally/event.go-// Value receiver: ...
cmd/mystally/event.go:func (kb Kobo) String() string {
--
cmd/mystally/event.go-// Value receiver: ...
cmd/mystally/event.go:func (t EventType) String() string {
--
cmd/mystally/event.go-// Value receiver: ...
cmd/mystally/event.go:func (t EventType) MarshalText() ([]byte, error) {
--
cmd/mystally/event.go-// Pointer receiver: ...
cmd/mystally/event.go:func (t *EventType) UnmarshalText(b []byte) error {
```

(your wording may differ; `...` stands for your own reason.)

### Gotchas

- `event.go` and `storage.go` were first created in `cmd/`, not `cmd/mystally/`. A different folder is a different package, so `main.go` couldn't see their names: `cmd/mystally/main.go:5:27: undefined: Kobo`. `go build ./...` also tried to build `cmd/` on its own and said `function main is undeclared in the main package`. Keep every file of the program in `cmd/mystally/`. *(you hit this)*
- A method that should hand back text must say so in its signature. `func (ev EventType) String()` with `return ...` in the body gives `too many return values` with `have (string)` and `want ()`. A body that only prints (with `string` in the signature) gives `missing return`. *(you hit this)*
- Formatting the receiver with `%s` or `%v` inside its own `String` loops forever. `go vet` says ``fmt.Sprintf format %s with arg kb causes recursive (github.com/na-cho-dev/mystally/cmd/mystally.Kobo).String method call``, and the program dies with `runtime: goroutine stack exceeds 1000000000-byte limit` and `fatal error: stack overflow`. Format a plain number: `kb.Naira()` or `int(t)`. *(you hit this)*
- `%b` is not "two decimals": for a `float64` it prints a binary exponent, so `Kobo(1250)` printed `₦7036874417766400p-49`. Use `%.2f`. *(you hit this)*
- With a pointer receiver, `kb` is an address, not a number: `cannot convert kb (variable of type *Kobo) to type float64`. `Naira` only reads, so give it a value receiver. *(you hit this)*
- 100 kobo is 1 naira, so you divide. Multiplying made `Kobo(1250)` print `₦125000.00`. *(you hit this)*

### Check yourself

- Why is `Kobo` an `int64` and not a `float64`?
- For each method you wrote, why did you choose a value or a pointer receiver?
- What is your `EventType`'s zero value, and what does it mean?
- What would `events.json` look like if `EventType` had no `MarshalText`?
- Why does `json.Unmarshal` need `&events` and not `events`?
- Why is a new ID the highest ID + 1 and not `len(events) + 1`?
- `String` and `MarshalText` return the same word. Which one does `fmt` call, which one does `encoding/json` call, and why keep both?

### Done when

- [ ] `events.json` is readable by eye and round-trips through load and save

  ```bash
  go build ./cmd/mystally
  rm -f events.json
  ./mystally add nacho like
  ./mystally add ada share
  cat events.json
  cp events.json before.json
  ./mystally add nacho comment
  diff before.json events.json
  rm before.json
  ```

  ```
  [
    {
      "id": 1,
      "creator": "nacho",
      "type": "like",
      "at": "2026-10-17T15:43:28.511575105+01:00"
    },
    {
      "id": 2,
      "creator": "ada",
      "type": "share",
      "at": "2026-10-17T15:43:28.517668894+01:00"
    }
  ]
  12a13,18
  >   },
  >   {
  >     "id": 3,
  >     "creator": "nacho",
  >     "type": "comment",
  >     "at": "2026-10-17T15:43:28.53052162+01:00"
  ```

  (your timestamps will differ.)

- [ ] `mystally add nacho like` works; `mystally add nacho dance` fails with "unknown event type"

  ```bash
  go build ./cmd/mystally
  ./mystally add nacho like; echo $?
  ./mystally add nacho dance; echo $?
  ./mystally add nacho dance 2>/dev/null; echo $?
  ./mystally list
  ```

  ```
  0
  unknown event type "dance"; choose one of: view, like, save, comment, share
  1
  1
  ID   CREATOR    TYPE     TIME
  1    nacho      like     2026-10-17 15:43:28
  2    ada        share    2026-10-17 15:43:28
  3    nacho      comment  2026-10-17 15:43:28
  4    nacho      like     2026-10-17 15:43:43
  ```

  (run this right after the box above, which leaves three events in `events.json`.)

- [ ] `fmt.Println(Kobo(1250))` prints `₦12.50`

  Scratch line: `fmt.Println(Kobo(1250))`

  ```bash
  go build ./cmd/mystally
  ./mystally 2>/dev/null
  ```

  ```
  ₦12.50
  ```

    (delete the scratch line and rebuild afterwards)

- [ ] every method receiver choice (value or pointer) has a one-line comment saying why

  ```bash
  grep -B1 '^func (' cmd/mystally/*.go
  ```

  ```
  cmd/mystally/event.go-// Value receiver: ...
  cmd/mystally/event.go:func (kb Kobo) Naira() float64 {
  --
  cmd/mystally/event.go-// Value receiver: ...
  cmd/mystally/event.go:func (kb Kobo) String() string {
  --
  cmd/mystally/event.go-// Value receiver: ...
  cmd/mystally/event.go:func (t EventType) String() string {
  --
  cmd/mystally/event.go-// Value receiver: ...
  cmd/mystally/event.go:func (t EventType) MarshalText() ([]byte, error) {
  --
  cmd/mystally/event.go-// Pointer receiver: ...
  cmd/mystally/event.go:func (t *EventType) UnmarshalText(b []byte) error {
  ```

  (your wording may differ; `...` stands for your own reason.)

---

## Stage 3: the ledger

**Goal:** put the logic in one struct with methods, away from the CLI.

**End goal:** you can run `./mystally balances` to print every creator and what they have earned, sorted by name, and `./mystally balance nacho` to print one creator. After `add nacho like`, `add nacho share` and `add ada comment`, they report `ada ₦0.20` and `nacho ₦0.55`, and a creator with no events gets `₦0.00`. You will understand why a method that changes its struct needs a pointer receiver, how Go maps behave (nil until made, zero for a missing key, random order), and why `main.go` only parses and prints while the ledger does every sum.

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

### Structure

Still one package in one folder. Stage 3 adds one file, `ledger.go`, and two commands:

```
mystally/
  .gitignore            /mystally, /events.json
  go.mod
  PLAN.md
  cmd/mystally/
    main.go             parse args → call the ledger → print. No arithmetic, no loops that add up money.
    event.go            Kobo, EventType, Event, ParseEventType (Stage 2, unchanged)
    storage.go          loadEvents, saveEvents (Stage 2, unchanged)
    ledger.go           Ledger, newLedger, Add, BalanceFor, Balances, Events, the rates   (new)
  events.json           written by the program, ignored by git
```

### Steps

Stage 3 has six steps. Do them in order: each one adds one piece to `ledger.go` or `main.go`, and each ends with a check you run against the built binary before you move on. Most checks delete `events.json` right after building, so they give the same numbers every time you run them.

#### Step 3.1: Create the Ledger and fill in its rates

**Achieve:**
- A new file `cmd/mystally/ledger.go`, in the same folder as `event.go`, declares `Ledger` with its two lowercase fields: the events and the rates.
- `newLedger` keeps the events it is given and builds the rates map with the five rates from the [Domain](#domain) table, keyed by your `EventType` constants (whatever you named them in Stage 2).
- If the Stage 1 `map[string]int` of event names is still in `main.go`, delete it now: `ParseEventType` checks names, and the rates belong to the ledger.
- Nothing calls `newLedger` yet, so every command behaves exactly as it did at the end of Stage 2.

**Learn:**
- A struct field can be any type, including a slice or a map. Lowercase field names are private to the package. That starts to matter in Stage 4.
- A **constructor** is an ordinary function, by convention named `newX` (or `NewX` once it is exported), that builds a value with every field ready to use. It usually returns a pointer: `return &Garden{...}`. The `&` gives the address of the new struct, so every caller shares one value instead of copying it.
- A map literal makes a map and fills it in one go: `map[Plant]int{Fern: 2, Basil: 3}`. The keys can be your own constants. An untyped number like `5` takes on the map's value type, so a `map[EventType]Kobo` literal needs no conversions.
- The zero value of a map is `nil`. A field you never set stays `nil`. Reading a nil map is allowed: every lookup gives the zero value (`0`) and `len` is 0. Step 3.4 shows what writing to one does.

<details>
<summary><b>Example: a constructor that fills a map field</b></summary>

```go
package main

import "fmt"

type Garden struct {
	cups map[string]int // cups of water per plant
}

func newGarden() *Garden {
	return &Garden{cups: map[string]int{"fern": 2, "basil": 3}}
}

func main() {
	g := newGarden()
	fmt.Println(g.cups["basil"], g.cups["cactus"])

	var empty Garden // no constructor: cups is a nil map
	fmt.Println(empty.cups == nil, len(empty.cups))
}
```

Output:

```
$ go run .
3 0
true 0
```

- The constructor makes every field ready to use and returns a pointer.
- A missing key reads as 0, and a nil map can be read (but not written).

</details>

**Step done when:**

```bash
go build ./cmd/mystally
go doc -cmd -u ./cmd/mystally Ledger
rm -f events.json
./mystally add nacho like
./mystally add ada comment
./mystally list
```

```
package main // import "github.com/na-cho-dev/mystally/cmd/mystally"

type Ledger struct {
	events []Event
	rates  map[EventType]Kobo
}

func newLedger(events []Event) *Ledger
ID   CREATOR    TYPE     TIME
1    nacho      like     2026-10-17 15:43:16
2    ada        comment  2026-10-17 15:43:16
```

(`go doc -cmd -u` works as in Stage 2: it shows a `package main` and its unexported names.)

#### Step 3.2: Record new events through Add

**Achieve:**
- `Add` gives the event the next ID (the highest existing ID + 1, or 1 for the very first event) and appends it to the ledger's events.
- `Events` hands the events back, so a command can save or print them.
- `add` in `main.go` now follows the flow: load events → `newLedger(events)` → `Add` → save what `Events` returns. The next-ID calculation leaves `main.go`, because it is arithmetic.
- `list` goes through the ledger too: load → `newLedger` → print what `Events` returns.

**Learn:**
- Stage 2 introduced value and pointer receivers. This is where the difference shows: a value receiver, `(s Shelf)`, works on a copy of the struct, so any change is lost when the method returns. A pointer receiver, `(s *Shelf)`, works on the caller's struct, so changes stick.
- `append` may return a new slice, so you always store its result: `s.books = append(s.books, title)`. With a value receiver, that result is stored into the copy.
- Go adds the `&` or `*` for you when you call a method. A pointer can call value-receiver methods, and a variable can call pointer-receiver methods. That is why the wrong receiver still compiles.
- Rule of thumb: a method that changes its struct gets a pointer receiver, and once one method of a type has one, give them all one. Small read-only types like `Kobo` keep value receivers.

<details>
<summary><b>Example: value versus pointer receivers with <code>append</code></b></summary>

```go
package main

import "fmt"

type Shelf struct {
	books []string
}

// AddCopy has a value receiver: it appends to a copy.
func (s Shelf) AddCopy(title string) { s.books = append(s.books, title) }

// Add has a pointer receiver: it appends to the caller's Shelf.
func (s *Shelf) Add(title string) { s.books = append(s.books, title) }

func main() {
	var s Shelf
	s.AddCopy("Dune")
	fmt.Println(len(s.books))
	s.Add("Dune")
	fmt.Println(len(s.books))
}
```

Output:

```
$ go run .
0
1
```

- Both versions compile; only the output shows that the value receiver lost the book.
- A method that changes its struct gets a pointer receiver.

</details>

**Step done when:**

```bash
go build ./cmd/mystally
rm -f events.json
./mystally add nacho like
./mystally add nacho share
./mystally add ada comment
./mystally list
./mystally add ada view
./mystally list
```

```
ID   CREATOR    TYPE     TIME
1    nacho      like     2026-10-17 15:43:17
2    nacho      share    2026-10-17 15:43:17
3    ada        comment  2026-10-17 15:43:17
ID   CREATOR    TYPE     TIME
1    nacho      like     2026-10-17 15:43:17
2    nacho      share    2026-10-17 15:43:17
3    ada        comment  2026-10-17 15:43:17
4    ada        view     2026-10-17 15:43:17
```

(your table layout and times will differ; the IDs must be 1, 2, 3 and then 4.)

#### Step 3.3: Work out one creator's balance

**Achieve:**
- `BalanceFor` goes through every event, keeps the ones that belong to the creator, and adds up the rate for each event's type.
- A creator with no events gets 0, which prints as `₦0.00`, with no special case for "not found".
- `mystally balance <creator>` loads the events, builds the ledger, and prints the creator's name and balance. Any other argument count is a usage error with exit code 1.

**Learn:**
- `for _, e := range events` visits each element of a slice in turn. `_` drops the index, and `e` is a copy of the element.
- `m[key]` gives the value, or the zero value when the key is missing. You don't need comma-ok here: a missing rate and a rate of 0 both add nothing. Never loop over a map to find a key; index it.
- Start a running total at its type's zero value: `var total Minutes` is 0. A named type only adds to the same type: `Minutes + Minutes` works, `Minutes + int` doesn't compile. Untyped constants like `20` fit either.
- `BalanceFor` only reads, so a value receiver would work too. Keep the pointer receiver anyway: it matches `Add`, and the struct isn't copied on every call.

<details>
<summary><b>Example: add up one member's minutes</b></summary>

```go
package main

import "fmt"

type Visit struct {
	Member string
	Class  string
}

func minutesFor(member string, visits []Visit, length map[string]int) int {
	total := 0
	for _, v := range visits {
		if v.Member == member {
			total += length[v.Class]
		}
	}
	return total
}

func main() {
	length := map[string]int{"yoga": 60, "spin": 45}
	visits := []Visit{{"ada", "yoga"}, {"ada", "spin"}, {"bola", "yoga"}}
	fmt.Println(minutesFor("ada", visits, length), minutesFor("chidi", visits, length))
}
```

Output:

```
$ go run .
105 0
```

- Loop over the slice, index the map. A member with no visits simply adds nothing: 0.
- The total starts at its zero value, so there is no special case for "not found".

</details>

**Step done when:**

```bash
go build ./cmd/mystally
rm -f events.json
./mystally add nacho like
./mystally add nacho share
./mystally add ada comment
for e in view like save comment share; do ./mystally add bola "$e"; done
./mystally balance nacho
./mystally balance ada
./mystally balance bola
./mystally balance zainab; echo $?
./mystally balance; echo $?
```

```
nacho      ₦0.55
ada        ₦0.20
bola       ₦0.86
zainab     ₦0.00
0
usage: mystally balance <creator>
1
```

(the `for` line runs `add bola` once per event type.)

#### Step 3.4: Add up every creator at once

**Achieve:**
- `Balances` returns a map from creator name to total, with one entry for each creator who has at least one event.
- It walks the events once, adding each event's rate to its creator's entry.
- `mystally balances` loads the events, builds the ledger, and for now prints the whole map with one `fmt.Println`. Step 3.5 replaces that line with sorted output, one creator per line.

**Learn:**
- `make(map[string]int)` creates an empty map you can write to. An empty literal, `map[string]int{}`, does the same. `var m map[string]int` only declares a nil map: reading works, writing panics with `assignment to entry in nil map`.
- `m[k] += v` reads the current value (0 when `k` is missing), adds `v`, and stores the result. No "is it there yet?" check needed.
- A map value is a handle to the data inside it. Returning a map, or passing one to a function, does not copy its entries.
- `fmt.Println(m)` prints `map[k:v ...]` with the keys sorted (fmt has done this since Go 1.12). A `for k, v := range m` loop does not sort: Go starts each loop at a random place, on purpose, so that no code can rely on the order.

<details>
<summary><b>Example: totals for everyone in one map</b></summary>

```go
package main

import "fmt"

type Loan struct {
	Borrower string
	Days     int
}

func main() {
	loans := []Loan{{"ada", 14}, {"bola", 21}, {"ada", 7}}

	days := make(map[string]int) // a nil map would panic on the first write
	for _, l := range loans {
		days[l.Borrower] += l.Days // a missing key starts from 0
	}
	fmt.Println(days)
}
```

Output:

```
$ go run .
map[ada:21 bola:21]
```

- `make` creates a map you can write to; `var m map[string]int` is nil and panics on the first write: `assignment to entry in nil map`.
- `fmt.Println` sorts a map's keys, but a `range` loop over it does not.

</details>

**Step done when:**

```bash
go build ./cmd/mystally
rm -f events.json
./mystally balances
./mystally add nacho like
./mystally add nacho share
./mystally add ada comment
for e in view like save comment share; do ./mystally add bola "$e"; done
./mystally balances
./mystally balances
./mystally balance bola
```

```
map[]
map[ada:₦0.20 bola:₦0.86 nacho:₦0.55]
map[ada:₦0.20 bola:₦0.86 nacho:₦0.55]
bola       ₦0.86
```

(the `map[...]` lines are a temporary check.)

#### Step 3.5: Print balances sorted by name

**Achieve:**
- `balances` prints one line per creator, the name then the amount, in alphabetical order by name. It replaces the temporary `fmt.Println` from Step 3.4.
- The order is the same on every run, however many creators there are.
- With no events, `balances` prints nothing and exits 0.

**Learn:**
- `maps.Keys(m)` (Go 1.23) returns an **iterator**, of type `iter.Seq[string]`: something a `for range` loop can pull keys from one at a time. It is not a slice, so you can't index it or call `len` on it.
- `slices.Sorted(it)` (Go 1.23) collects an iterator into a new slice and sorts it. `slices.Collect(it)` (Go 1.23) only collects; `slices.Sort(s)` (Go 1.21) sorts an existing slice in place. You already used `slices.Sorted(maps.Keys(...))` for the event names in Stage 1.
- Strings sort by byte value, and every uppercase letter comes before every lowercase one, so `Zaria` sorts before `abuja`.
- `%-10s` pads a value on the right to 10 characters, so the amounts line up. If you pad the amount too, `fmt` counts characters, not bytes, so `₦` takes one column.

<details>
<summary><b>Example: print a map in key order</b></summary>

```go
package main

import (
	"fmt"
	"maps"
	"slices"
)

func main() {
	days := map[string]int{"chidi": 7, "ada": 21, "bola": 14}
	for _, name := range slices.Sorted(maps.Keys(days)) {
		fmt.Printf("%-8s %3d days\n", name, days[name])
	}
}
```

Output:

```
$ go run .
ada       21 days
bola      14 days
chidi      7 days
```

- `maps.Keys` gives an iterator, not a slice; `slices.Sorted` collects and sorts it.
- Ranging over the map itself would give a different order on different runs.

</details>

**Step done when:**

```bash
go build ./cmd/mystally
rm -f events.json
./mystally add nacho like
./mystally add nacho share
./mystally add ada comment
./mystally add zainab save
./mystally add chidi view
./mystally add bola share
./mystally balances
for i in 1 2 3 4 5 6 7 8 9 10; do ./mystally balances | head -n 1; done | uniq -c
```

```
ada        ₦0.20
bola       ₦0.50
chidi      ₦0.01
nacho      ₦0.55
zainab     ₦0.10
     10 ada        ₦0.20
```

(your spacing may differ.)

#### Step 3.6: Keep main.go to parsing and printing

**Achieve:**
- `main.go` never mentions `rates` and never adds numbers. Every sum lives in `ledger.go`.
- Every command has the same shape: check the arguments → load → `newLedger` → one method → save if something changed → print. If the "load, then build the ledger" lines repeat in several commands, a small helper in `main.go` that does both is fine.
- The full Stage 3 check passes, `go vet ./...` and `gofmt -l .` print nothing, and only then do you commit (for example `stage 3: ledger balances`).

**Learn:**
- `main.go` and `ledger.go` are the same package, so `main.go` *can* read `l.rates` and `l.events`. Lowercase names only hide things from other packages. Until Stage 4 moves the ledger into a package of its own, the boundary is up to you, and a search is the quickest way to check it.
- `grep -nE 'rates|\+' file` prints every line that contains `rates` or a `+`, with its line number (`-n`). `-E` lets you write `|` for "or", and `\+` means a literal plus sign.
- grep's exit code says what it found: 0 means at least one match, 1 means none. Here, 1 is the result you want.
- At the start of this stage the same search would most likely have found the next-ID line (the `+ 1`) in `addCmd`. Since Step 3.2 that calculation lives in `Add`.

<details>
<summary><b>Example: find arithmetic that leaked into main</b></summary>

```go
package main

import "fmt"

func main() {
	pages := []int{120, 80}
	fmt.Println("read", pages[0]+pages[1], "pages") // this sum belongs in a method
}
```

Output:

```
$ grep -nE 'pages\[|\+' main.go; echo $?
7:	fmt.Println("read", pages[0]+pages[1], "pages") // this sum belongs in a method
0
```

- `grep` exits with 0 when it finds a match and 1 when it finds none. For your `main.go`, 1 is the answer you want.
- Not every `+` is arithmetic (`"a" + "b"` joins strings), so read each match.

</details>

**Step done when:**

```bash
go build ./cmd/mystally
grep -nE 'rates|\+' cmd/mystally/main.go; echo $?
rm -f events.json
./mystally add nacho like
./mystally add nacho share
./mystally add ada comment
./mystally list
./mystally balances
./mystally balance nacho
```

```
1
ID   CREATOR    TYPE     TIME
1    nacho      like     2026-10-17 15:44:13
2    nacho      share    2026-10-17 15:44:13
3    ada        comment  2026-10-17 15:44:13
ada        ₦0.20
nacho      ₦0.55
nacho      ₦0.55
```

(the `1` is grep's exit code: no match.)

### Gotchas

- Create `ledger.go` in `cmd/mystally/`, next to `event.go`. In Stage 2 you made `event.go` and `storage.go` in `cmd/` by mistake: a different folder is a different package. With `ledger.go` in `cmd/`, `go build ./cmd/mystally` fails with `undefined: Ledger` and `undefined: newLedger`, and `go vet ./...` adds `cmd/ledger.go:4:11: undefined: Event`. *(you hit this)*
- Pointer receivers are for methods that change their struct, like `Add`. `Kobo`'s methods only read, so they keep value receivers: with `(kb *Kobo)` you got `cannot convert kb (variable of type *Kobo) to type float64`. *(you hit this)*
- `balance` needs its creator, so check `len(args)` before reading `args[1]`. Without the check, `./mystally balance` panics with `panic: runtime error: index out of range [1] with length 1` and exits 2. You hit the same panic in Stage 1 when `add`'s checks ran for `list`. *(you hit this)*
- To find one rate, index the map. Don't loop over it looking for the key; that was Stage 1's comma-ok lesson. *(you hit this)*
- `gofmt -l .` lists a file whose only problem is a missing newline at the very end. Run it before you commit. *(you hit this)*
- Run the whole Step 3.6 check before you commit, not after. In Stage 1 you committed a change that made every command print usage. *(you hit this)*

### Check yourself

- Why must `Add` have a pointer receiver? Could `BalanceFor` have a value receiver?
- What happens if you write to a map declared as `var m map[string]Kobo`?
- Why is an unknown creator's balance 0 without any `if`?
- `main.go` and `ledger.go` are one package, so `main.go` could read `l.rates`. What stops it today, and what will stop it in Stage 4?
- In Step 3.4, printing the whole balances map with one `fmt.Println` showed its keys in order. Why did `balances` still need `slices.Sorted` in Step 3.5?

### Done when

- [ ] `main.go` no longer contains any arithmetic; it only parses args and prints

  ```bash
  grep -nE 'rates|\+' cmd/mystally/main.go; echo $?
  ```

  ```
  1
  ```

  (no lines, and grep's exit code `1` means no match.)

- [ ] balances output is deterministic (sorted) on every run

  ```bash
  go build ./cmd/mystally
  rm -f events.json
  ./mystally add nacho like
  ./mystally add nacho share
  ./mystally add ada comment
  ./mystally add zainab save
  ./mystally add chidi view
  ./mystally add bola share
  ./mystally balances
  for i in 1 2 3 4 5 6 7 8 9 10; do ./mystally balances | head -n 1; done | uniq -c
  ```

  ```
  ada        ₦0.20
  bola       ₦0.50
  chidi      ₦0.01
  nacho      ₦0.55
  zainab     ₦0.10
       10 ada        ₦0.20
  ```

    (your spacing may differ; the names must be in alphabetical order, and the last line must be a single line with a count of 10)

- [ ] an unknown creator reports a zero balance, not a crash

  ```bash
  go build ./cmd/mystally
  rm -f events.json
  ./mystally balance zainab; echo $?
  ./mystally balances; echo $?
  ./mystally add nacho like
  ./mystally balance zainab; echo $?
  ```

  ```
  zainab     ₦0.00
  0
  0
  zainab     ₦0.00
  0
  ```

  (with no file, `balances` prints nothing and exits `0`.)

---

## Stage 4: packages and the Store interface

**Goal:** split the program into packages and introduce the first interface.

**End goal:** `go build ./...` builds three packages (`cmd/mystally`, `internal/ledger` and `internal/store`), and `./mystally add`, `list`, `balances` and `balance` behave exactly as they did in Stage 3, now loading and saving through a `FileStore`. `go list -f '{{.ImportPath}}: {{join .Imports " "}}' ./...` shows `internal/store` importing `internal/ledger` and never the reverse, and changing one line in `main.go` swaps in a `MemoryStore`. You will understand how folders, imports and capital letters decide who can see what, why an interface belongs to the code that uses it, and how one `var _` line makes the compiler confirm that a type satisfies an interface.

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

### Structure

Stage 4 adds two packages under `internal/`. Two files move, one file turns into a type with methods, and two files are new:

```
mystally/
  .gitignore            /mystally, /events.json
  go.mod                module github.com/na-cho-dev/mystally
  PLAN.md
  cmd/mystally/
    main.go             the CLI and the wiring. Imports both packages below
  internal/ledger/
    event.go            Kobo, EventType and its constants, Event, ParseEventType   (was cmd/mystally/event.go)
    ledger.go           Ledger, NewLedger, Load, Add, BalanceFor, Balances, Events  (was cmd/mystally/ledger.go)
    store.go            the Store interface                                         (new)
  internal/store/
    file.go             FileStore, NewFileStore, Load, Save                         (was cmd/mystally/storage.go)
    memory.go           MemoryStore, NewMemoryStore, Load, Save                     (new)
  events.json           written by FileStore, ignored by git
```

```
cmd/mystally ──► internal/store ──► internal/ledger
      │                                   ▲
      └───────────────────────────────────┘
```

### Steps

Stage 4 has eight steps. Do them in order: Steps 4.1 to 4.5 move code into packages without changing what mystally does, Step 4.6 hands the store to the ledger, and Steps 4.7 and 4.8 prove the design works. Each step ends with a check you run before you move on.

#### Step 4.1: Move the event types and the ledger into internal/ledger

**Achieve:**
- `internal/ledger/` exists at the repo root, next to `go.mod`, and `event.go` and `ledger.go` live in it. Create the folder with `mkdir -p internal/ledger`, then move both files with `git mv`.
- Both files start with `package ledger`. Every name another package will need starts with a capital letter: `NewLedger` (was `newLedger`), and your event constants if they are still lowercase (`View`, `Like`, `Save`, `Comment`, `Share`). `main` may not use the constants, but Stage 5's tests live in another package and will. `Ledger`'s fields stay lowercase.
- `NewLedger` still takes the events for now; Step 4.6 changes that. Move first, change behaviour later: if something breaks, you know which of the two caused it.
- `main.go` and `storage.go` stay in `cmd/mystally` and stop compiling. Step 4.2 fixes them.

**Learn:**
- A package is one folder of `.go` files that share a `package` line (Stage 0). By convention the name is the folder's name: `internal/ledger` holds `package ledger`. `package main` is special: it makes a program you can run, and nothing can import it.
- A name is **exported** (usable from other packages) when it starts with an upper-case letter. The rule is the same for types, functions, constants, package-level variables, struct fields and methods. The Go spec calls this "Exported identifiers". Every lowercase name is private to its package.
- `go build ./internal/ledger` compiles one package. For a package that isn't `main` it writes no file: it only tells you whether the package compiles.
- `go doc ./internal/ledger` lists the exported names, which is exactly what another package can see. `go doc -u` adds the unexported ones.

<details>
<summary><b>Example: a package with exported and unexported names</b></summary>

`internal/weather/weather.go`:

```go
// Package weather summarises temperature readings.
package weather

// Hottest returns the highest reading. Exported: other packages can call it.
func Hottest(readings []int) int {
	return highest(readings)
}

// highest is unexported: only package weather can call it.
func highest(readings []int) int {
	best := readings[0]
	for _, r := range readings[1:] {
		best = max(best, r)
	}
	return best
}
```

Output:

```
$ go build ./internal/weather
$ go doc ./internal/weather
package weather // import "example.com/forecast/internal/weather"

Package weather summarises temperature readings.

func Hottest(readings []int) int
$ go doc -u ./internal/weather
package weather // import "example.com/forecast/internal/weather"

Package weather summarises temperature readings.

func Hottest(readings []int) int
func highest(readings []int) int
```

- A capital first letter exports a name. `go doc` shows exactly what other packages can see; `-u` adds the rest.
- `go build` on a package that isn't `main` writes nothing: it only proves the package compiles.

</details>

**Step done when:**

```bash
git status --short
go list ./...
go build ./internal/ledger
go vet ./internal/ledger
go doc ./internal/ledger
```

```
RM cmd/mystally/event.go -> internal/ledger/event.go
RM cmd/mystally/ledger.go -> internal/ledger/ledger.go
github.com/na-cho-dev/mystally/cmd/mystally
github.com/na-cho-dev/mystally/internal/ledger
package ledger // import "github.com/na-cho-dev/mystally/internal/ledger"

Package ledger records engagement events and works out what each creator has
earned.

type Event struct{ ... }
type EventType int
    const Unknown EventType = iota ...
    func ParseEventType(s string) (EventType, error)
type Kobo int64
type Ledger struct{ ... }
    func NewLedger(events []Event) *Ledger
```

(`R` means git recorded a move, and `M` that the file changed after it.)

#### Step 4.2: Import the ledger from main

**Achieve:**
- `main.go` and `storage.go` import `"github.com/na-cho-dev/mystally/internal/ledger"` and put `ledger.` in front of every name that moved: `ledger.Event`, `ledger.NewLedger`, `ledger.ParseEventType`, and `*ledger.Ledger` wherever you wrote the type.
- Every command behaves exactly as it did at the end of Stage 3.
- Your Stage 3 `events.json` still loads: moving a type to another package doesn't change its JSON.

**Learn:**
- An **import path** is the module path from `go.mod` plus the folder: `github.com/na-cho-dev/mystally` + `/internal/ledger`. Always write the whole path. `"internal/ledger"` on its own makes Go look in the standard library, and it fails with `package internal/ledger is not in std`.
- In code, the prefix is the package's name, which is the last part of the path: `weather.Hottest`, `ledger.Event`. It tells every reader where a name comes from.
- Put standard-library imports first, then a blank line, then your module's packages. `gofmt` sorts each group and keeps the groups apart.
- An import you don't use is a compile error, like an unused variable: `"example.com/forecast/internal/weather" imported and not used`.

<details>
<summary><b>Example: <code>main</code> imports the package</b></summary>

`cmd/forecast/main.go`:

```go
package main

import (
	"fmt"

	"example.com/forecast/internal/weather"
)

func main() {
	fmt.Println("hottest:", weather.Hottest([]int{24, 31, 28}))
}
```

Output:

```
$ go run ./cmd/forecast
hottest: 31
```

(`internal/weather/weather.go` is the file from Step 4.1.)

- The import path is the module path plus the folder; in code you use the last part: `weather.Hottest`.
- Standard-library imports first, a blank line, then your own packages.

</details>

**Step done when:**

```bash
go build ./cmd/mystally
./mystally list
rm -f events.json
./mystally add nacho like
./mystally add nacho share
./mystally add ada comment
./mystally list
./mystally balances
./mystally balance nacho
./mystally add nacho dance; echo $?
```

```
ID   CREATOR    TYPE     TIME
1    bola       view     2026-10-17 15:44:15
2    bola       like     2026-10-17 15:44:15
3    bola       save     2026-10-17 15:44:15
4    bola       comment  2026-10-17 15:44:15
5    bola       share    2026-10-17 15:44:15
6    nacho      like     2026-10-17 15:44:15
7    nacho      share    2026-10-17 15:44:15
8    ada        comment  2026-10-17 15:44:15
ID   CREATOR    TYPE     TIME
1    nacho      like     2026-10-17 15:58:31
2    nacho      share    2026-10-17 15:58:31
3    ada        comment  2026-10-17 15:58:31
ada        ₦0.20
nacho      ₦0.55
nacho      ₦0.55
unknown event type "dance"; choose one of: view, like, save, comment, share
1
```

(the first `list` prints whatever your Stage 3 `events.json` held; if it loads, the JSON survived the move.)

#### Step 4.3: Declare the Store interface where it is used

**Achieve:**
- A new file, `internal/ledger/store.go`, declares `Store` exactly as in **Build**: a `Load` that returns the events and an error, and a `Save` that takes the events and returns an error.
- Nothing satisfies `Store` or uses it yet, and mystally behaves as it did in Step 4.2.

**Learn:**
- An **interface** type is a list of method signatures. Any type that has all of those methods **satisfies** the interface automatically. Go has no `implements` keyword.
- The package that *calls* the methods (the **consumer**) declares the interface, not the packages that provide them. `ledger` will call `Load` and `Save`, so `Store` lives in `ledger`, and `store` will import `ledger` to satisfy it. That keeps the import arrows pointing one way.
- List only the methods the consumer calls. A small interface is easy to satisfy, and easy to fake in a test.
- Parameter names are optional in an interface: `Save([]Event) error` is complete.

<details>
<summary><b>Example: the consumer declares the interface</b></summary>

```go
package main

import "fmt"

// Sender is declared here, by the code that calls Send.
type Sender interface {
	Send(to, msg string) error
}

type Alerter struct {
	sender Sender
	limit  int
}

func (a *Alerter) Check(to string, value int) error {
	if value < a.limit {
		return nil
	}
	return a.sender.Send(to, fmt.Sprintf("value %d reached %d", value, a.limit))
}

// Console never says "I implement Sender". Having Send is enough.
type Console struct{}

func (Console) Send(to, msg string) error {
	fmt.Printf("to %s: %s\n", to, msg)
	return nil
}

func main() {
	a := &Alerter{sender: Console{}, limit: 10}
	fmt.Println(a.Check("ops", 9))
	fmt.Println(a.Check("ops", 12))
}
```

Output:

```
$ go run .
<nil>
to ops: value 12 reached 10
<nil>
```

- `Alerter` only knows the interface; any type with a matching `Send` method fits.
- List only the methods the consumer calls. Small interfaces are easy to satisfy and to fake.

</details>

**Step done when:**

```bash
go build ./...
go doc ./internal/ledger Store
```

```
package ledger // import "github.com/na-cho-dev/mystally/internal/ledger"

type Store interface {
	Load() ([]Event, error)
	Save([]Event) error
}
```

(`go build` prints nothing.)

#### Step 4.4: Move loading and saving into FileStore

**Achieve:**
- `cmd/mystally/storage.go` moves to `internal/store/file.go` (`mkdir -p internal/store`, then `git mv`). It starts with `package store` and declares `FileStore` and `NewFileStore` as in **Structure**.
- Stage 2's `loadEvents` and `saveEvents` become `FileStore`'s `Load` and `Save` methods. The path comes from the struct's field instead of a parameter, and the behaviour doesn't change: a missing file still means no events.
- `var _ ledger.Store = (*FileStore)(nil)` sits near the top of `file.go`.
- For this step only, `main.go` loads and saves through a `FileStore` built from `eventsFile`, and still hands the loaded events to `NewLedger`. Step 4.6 moves that job into the ledger.

**Learn:**
- The implementer imports the consumer: `store` imports `ledger`. Inside `store`, names from `ledger` need the prefix, so the methods take and return `[]ledger.Event`.
- A type satisfies an interface just by having methods with exactly the right names and signatures. Nothing in the code says "FileStore is a Store", so a typo in a signature only shows up where a `FileStore` is first used as a `Store`, possibly in another package.
- The **compile-time check** `var _ alert.Sender = (*Console)(nil)` moves that error next to the type. `(*Console)(nil)` converts `nil` into a nil pointer of type `*Console`, and assigning it to a `Sender` makes the compiler compare the methods. `_`, the **blank identifier**, throws the variable away.
- Use the pointer form `(*T)(nil)`. Methods with pointer receivers belong to `*T`, not to `T`, and a `*T` also has every value-receiver method, so the pointer form works either way.

<details>
<summary><b>Example: an implementation with a compile-time check</b></summary>

`internal/sender/console.go`:

```go
// Package sender delivers messages.
package sender

import (
	"fmt"

	"example.com/notify/internal/alert"
)

// The build fails on this line if Console stops satisfying alert.Sender.
var _ alert.Sender = (*Console)(nil)

// Console prints messages.
type Console struct{}

// Send prints msg.
func (c *Console) Send(to, msg string) error {
	fmt.Printf("to %s: %s\n", to, msg)
	return nil
}
```

Output:

```
$ go run ./cmd/notify
to ops: backup finished
$ sed -i 's/) Send(to/) Deliver(to/' internal/sender/console.go
$ go build ./...
# example.com/notify/internal/sender
internal/sender/console.go:11:22: cannot use (*Console)(nil) (value of type *Console) as alert.Sender value in variable declaration: *Console does not implement alert.Sender (missing method Send)
```

- `sender` imports `alert`, never the reverse. `(*Console)(nil)` is a typed nil pointer that costs nothing at run time.
- After the `sed` renames `Send`, the build fails on the check line, next to the type, instead of somewhere far away.

</details>

**Step done when:**

```bash
go build ./cmd/mystally
ls cmd/mystally
go doc ./internal/store FileStore
rm -f events.json
./mystally list
./mystally add nacho like
./mystally add ada comment
./mystally list
```

```
main.go
package store // import "github.com/na-cho-dev/mystally/internal/store"

type FileStore struct {
	// Has unexported fields.
}

func NewFileStore(path string) *FileStore
func (f *FileStore) Load() ([]ledger.Event, error)
func (f *FileStore) Save(events []ledger.Event) error
ID   CREATOR    TYPE     TIME
ID   CREATOR    TYPE     TIME
1    nacho      like     2026-10-17 15:58:43
2    ada        comment  2026-10-17 15:58:43
```

(`main.go` is the only file left in `cmd/mystally`.)

#### Step 4.5: Add a MemoryStore

**Achieve:**
- A new file, `internal/store/memory.go`, declares `MemoryStore` and `NewMemoryStore` as in **Structure**, with `Load` and `Save` methods that satisfy `ledger.Store`. Neither can fail, so both always return a nil error.
- `Save` keeps a copy of the events it is given, and `Load` returns a copy, so the caller and the store never share a slice.
- `var _ ledger.Store = (*MemoryStore)(nil)` sits next to it.
- Nothing uses `MemoryStore` yet. Step 4.7 swaps it in for one run, and Stage 5's tests use it for real.

**Learn:**
- A slice value is a small header that points at an array. Storing a slice stores that pointer, not the elements, so the caller and the store share one array: when the caller changes an element, the "stored" data changes too.
- `slices.Clone(s)` (Go 1.21) makes a new array holding the same elements. Clone on the way in (`Save`) and on the way out (`Load`).
- The clone is shallow: each element is copied with `=`. That is enough for structs of strings, numbers and times, which is what an event is.
- `Save` changes the store, so it needs a pointer receiver (the Stage 3 rule). With a value receiver it still compiles and still passes the compile-time check, and every save is lost. See **Gotchas**.

<details>
<summary><b>Example: copy a slice in and out</b></summary>

```go
package main

import (
	"fmt"
	"slices"
)

type Memory struct {
	notes []string
}

func (m *Memory) Save(notes []string) { m.notes = slices.Clone(notes) }

func (m *Memory) Load() []string { return slices.Clone(m.notes) }

func main() {
	notes := []string{"milk", "eggs"}
	var m Memory
	m.Save(notes)
	notes[0] = "CHANGED" // the caller keeps editing its own slice
	fmt.Println(m.Load(), notes)
}
```

Output:

```
$ go run .
[milk eggs] [CHANGED eggs]
```

- A slice points at shared memory. Without `slices.Clone`, the stored copy would say `CHANGED` too.
- `Save` changes the struct, so it has a pointer receiver.

</details>

**Step done when:**

```bash
go doc ./internal/store
go doc ./internal/store MemoryStore
grep -n 'var _ ledger.Store' internal/store/*.go
```

```
package store // import "github.com/na-cho-dev/mystally/internal/store"

Package store holds the places events can be kept.

type FileStore struct{ ... }
    func NewFileStore(path string) *FileStore
type MemoryStore struct{ ... }
    func NewMemoryStore() *MemoryStore
package store // import "github.com/na-cho-dev/mystally/internal/store"

type MemoryStore struct {
	// Has unexported fields.
}

func NewMemoryStore() *MemoryStore
func (m *MemoryStore) Load() ([]ledger.Event, error)
func (m *MemoryStore) Save(events []ledger.Event) error
internal/store/file.go:11:var _ ledger.Store = (*FileStore)(nil)
internal/store/memory.go:9:var _ ledger.Store = (*MemoryStore)(nil)
```

(`go vet` prints nothing.)

#### Step 4.6: Give the ledger its store

**Achieve:**
- `Ledger` gains a lowercase field that holds a `Store`. `NewLedger(s Store) *Ledger` keeps the store and makes the rates map. It no longer takes events.
- `Load() error` asks the store for the events and keeps them in the ledger.
- `Add(e Event) error` gives the event its ID and appends it as before, then saves all the events through the store and returns whatever error the store returned.
- `main.go` builds the ledger in one place: a `FileStore` for `eventsFile`, handed to `NewLedger`, then `Load`. The `add` command returns `Add`'s error. `main.go` no longer calls the store's `Load` or `Save` itself.

**Learn:**
- A constructor that takes an interface lets the caller choose the implementation. Handing a type the things it depends on, instead of letting it build them, is called **dependency injection**. In Go it is just a constructor parameter.
- Keep work that can fail out of a constructor that can't. `NewLedger` only stores values, so it returns no error. Loading can fail, so it is a separate method that returns `error`.
- When a method starts calling something that can fail, it returns `error` too, and every caller has to deal with it. That is why `Add`'s signature changes.
- The zero value of an interface is `nil`, and calling a method on a nil interface panics with `invalid memory address or nil pointer dereference`. The constructor guarantees the field is set, and because the field is lowercase, another package has no other way to set it.

<details>
<summary><b>Example: a constructor that takes an interface</b></summary>

```go
package main

import (
	"errors"
	"fmt"
)

type Saver interface {
	Save(items []string) error
}

type Board struct {
	saver Saver
	items []string
}

func NewBoard(s Saver) *Board { return &Board{saver: s} }

// Add can fail now, because saving can fail.
func (b *Board) Add(item string) error {
	b.items = append(b.items, item)
	return b.saver.Save(b.items)
}

type screen struct{}

func (screen) Save(items []string) error { fmt.Println("saved", items); return nil }

type fullDisk struct{}

func (fullDisk) Save([]string) error { return errors.New("disk full") }

func main() {
	fmt.Println(NewBoard(screen{}).Add("milk"))
	fmt.Println(NewBoard(fullDisk{}).Add("milk"))
}
```

Output:

```
$ go run .
saved [milk]
<nil>
disk full
```

- The caller chooses the implementation; `Board` only stores it.
- Once a method calls something that can fail, it returns an `error` too, and passes the store's error on unchanged.

</details>

**Step done when:**

```bash
go build ./cmd/mystally
grep -c 'NewFileStore' cmd/mystally/main.go
grep -c '\.Save(' cmd/mystally/main.go
rm -f events.json
./mystally list
./mystally add nacho like
./mystally add nacho share
./mystally add ada comment
./mystally balances
chmod a-w events.json
./mystally add ada view; echo $?
chmod u+w events.json
./mystally add ada view; echo $?
./mystally list
./mystally balance ada
```

```
1
0
ID   CREATOR    TYPE     TIME
ada        ₦0.20
nacho      ₦0.55
open events.json: permission denied
1
0
ID   CREATOR    TYPE     TIME
1    nacho      like     2026-10-17 15:57:41
2    nacho      share    2026-10-17 15:57:41
3    ada        comment  2026-10-17 15:57:41
4    ada        view     2026-10-17 15:57:41
ada        ₦0.21
```

(`1`: one place in `main.go` builds the store.)

#### Step 4.7: Swap in MemoryStore for one run

**Achieve:**
- In `main.go`, passing `store.NewMemoryStore()` where you passed `store.NewFileStore(eventsFile)` is the whole change. It builds, and nothing else needs editing.
- You see what an in-memory store means for a CLI: every run starts empty, and nothing reaches `events.json`.
- `git restore` puts the `FileStore` back.

**Learn:**
- **Wiring** means choosing the concrete types and plugging them together. Keep it in one place, in `main`; everything else only knows the interface.
- Each run of `./mystally` is a new process. A `MemoryStore` lives in that process's memory, so whatever `add` saved is gone when it exits. Useless for the CLI, ideal for Stage 5's tests.
- An unused constant at package level is not an error (only unused imports and unused local variables are), so `eventsFile` can stay while the swap is in place.
- `git add -A` stages every change you have made. After that, `git diff` shows only what you change next, and `git restore <file>` puts a file back to its staged version.

<details>
<summary><b>Example: swap the implementation in one place</b></summary>

```go
package main

import "fmt"

type Sender interface {
	Send(msg string) error
}

type Console struct{}

func (Console) Send(msg string) error { fmt.Println("console:", msg); return nil }

type Recorder struct{ sent []string }

func (r *Recorder) Send(msg string) error { r.sent = append(r.sent, msg); return nil }

func notify(s Sender) error { return s.Send("backup finished") }

func main() {
	fmt.Println(notify(Console{})) // wiring: main picks the implementation
	rec := &Recorder{}
	fmt.Println(notify(rec)) // swapped: notify didn't change
	fmt.Println("recorded:", rec.sent)
}
```

Output:

```
$ go run .
console: backup finished
<nil>
<nil>
recorded: [backup finished]
```

- Only `main` names a concrete type. Everything else works through the interface.
- A recorder like this keeps everything in memory: ideal for tests, useless across separate runs.

</details>

**Step done when:**

```bash
git add -A
```

Now edit `main.go`: pass `store.NewMemoryStore()` where you passed `store.NewFileStore(eventsFile)`. Then run:

```bash
git diff --stat
go build ./cmd/mystally
rm -f events.json
./mystally add nacho like; echo $?
./mystally add ada comment; echo $?
./mystally list
ls events.json
git restore cmd/mystally/main.go
go build ./cmd/mystally
./mystally add nacho like
./mystally list
```

```
 cmd/mystally/main.go | 2 +-
 1 file changed, 1 insertion(+), 1 deletion(-)
0
0
ID   CREATOR    TYPE     TIME
ls: cannot access 'events.json': No such file or directory
ID   CREATOR    TYPE     TIME
1    nacho      like     2026-10-17 15:57:43
```

(both `add`s exit 0: they saved, into memory.)

#### Step 4.8: Check the import arrows and commit

**Achieve:**
- `go list` confirms the arrows: `internal/store` imports `internal/ledger`, and `internal/ledger` imports nothing from this module.
- `go doc` shows each package the way another programmer sees it, and only the names you meant to export appear.
- `go vet ./...` and `gofmt -l .` are clean, **Done when** boxes 1 to 4 pass, then you commit (for example `stage 4: packages and the Store interface`), and box 5 confirms the commit.

**Learn:**
- Go forbids **import cycles**: if A imports B, B can't import A, directly or through other packages. The build stops with `import cycle not allowed` and prints the chain of imports.
- The fix always has the same shape: the interface lives with the consumer, and the implementer imports the consumer. If `ledger` ever seems to need `store`, the code that needs it is wiring, and it belongs in `main`.
- `go list ./...` prints every package in the module. `-f` takes a template: `{{.ImportPath}}` is the package's path, and `{{join .Imports " "}}` is its direct imports joined with spaces.
- `| grep mystally` keeps only your module's packages. grep exits with 1 when nothing matched, which is what you want for `ledger`.

<details>
<summary><b>Example: the cycle you get by reversing an import</b></summary>

This does not compile:

`internal/alert/alert.go`:

```go
package alert

import "example.com/cycle/internal/sender"

func Notify() { sender.Send() }
```

`internal/sender/sender.go`:

```go
package sender

import "example.com/cycle/internal/alert"

func Send() { alert.Notify() }
```

Output:

```
$ go build ./...
package example.com/cycle/internal/alert
	imports example.com/cycle/internal/sender from alert.go
	imports example.com/cycle/internal/alert from sender.go: import cycle not allowed
```

- Go refuses import cycles and prints the chain. The fix: the interface lives with the consumer, and only the implementer imports it.
- Check your own arrows with `go list -f '{{.ImportPath}}: {{join .Imports " "}}' ./...`.

</details>

**Step done when:**

```bash
go list -f '{{.ImportPath}}: {{join .Imports " "}}' ./...
go list -f '{{join .Imports "\n"}}' ./internal/ledger | grep mystally; echo $?
go doc ./internal/ledger Ledger
```

```
github.com/na-cho-dev/mystally/cmd/mystally: errors fmt github.com/na-cho-dev/mystally/internal/ledger github.com/na-cho-dev/mystally/internal/store maps os slices strings time
github.com/na-cho-dev/mystally/internal/ledger: fmt time
github.com/na-cho-dev/mystally/internal/store: encoding/json errors github.com/na-cho-dev/mystally/internal/ledger os slices
1
package ledger // import "github.com/na-cho-dev/mystally/internal/ledger"

type Ledger struct {
	// Has unexported fields.
}

func NewLedger(s Store) *Ledger
func (l *Ledger) Add(e Event) error
func (l *Ledger) BalanceFor(creator string) Kobo
func (l *Ledger) Balances() map[string]Kobo
func (l *Ledger) Events() []Event
func (l *Ledger) Load() error
```

(the standard-library names on each `go list` line depend on your code; what matters is which `github.com/na-cho-dev/mystally/...` paths appear.)

### Gotchas

- `&ledger.Ledger{}` with no fields *does* compile, and its store is nil. The first `Load` panics with `invalid memory address or nil pointer dereference` and exits 2: the same panic you got in Stage 1 from calling `file.Name()` on a nil `*os.File`. Build every `Ledger` with `NewLedger`. *(you hit this)*
- In Stage 1 you moved the event map into another function and got `declared and not used: events` where it was made and `undefined: events` where it was used: a variable declared inside a function exists only inside that function. Stage 4 adds two more rules for names in another package: you need the import and the `pkg.` prefix, and the name must start with a capital letter. *(you hit this)*
- `internal/` goes at the repo root, next to `go.mod`. A different folder is a different package with a different import path: in Stage 2 you created `event.go` in `cmd/` instead of `cmd/mystally/`. *(you hit this)*
- The import path is the whole module path plus the folder. `"internal/ledger"` gives `package internal/ledger is not in std (/usr/local/go/src/internal/ledger)`. Commands need a path too: `go build internal/ledger` fails the same way, and plain `go build` at the repo root still says `no Go files in .../mystally`. Write `./internal/ledger` or `./...`. *(you hit this)*
- Return the store's error as it is. In Stage 1 you printed a fixed message instead of `err`, and it hid the real cause, `permission denied`. *(you hit this)*
- After you move folders, the editor may keep underlining imports that `go build ./...` accepts. Trust the terminal and restart gopls (see [When the editor and the terminal disagree](#when-the-editor-and-the-terminal-disagree)). In Stage 0 a stale gopls reported `could not import os`. *(you hit this)*

### Check yourself

- Why is `Store` declared in `ledger` and not in `store`?
- Draw your import arrows. What error appears if `ledger` imports `store`?
- Why can't `main` set `Ledger`'s fields directly, and why is that a good thing?
- What would you change in `main` to swap `FileStore` for `MemoryStore`? (The answer should be one line.)
- Why does `MemoryStore` copy the slice in both `Save` and `Load`?
- With `var _ ledger.Store = (*FileStore)(nil)`, where is a typo in `Save`'s signature reported? Where would it be reported without that line?

### Done when

- [ ] `go build ./...` succeeds with the three packages

  ```bash
  go build ./...; echo $?
  go list ./...
  ```

  ```
  0
  github.com/na-cho-dev/mystally/cmd/mystally
  github.com/na-cho-dev/mystally/internal/ledger
  github.com/na-cho-dev/mystally/internal/store
  ```

- [ ] `internal/store` imports `internal/ledger`, never the reverse

  ```bash
  go list -f '{{.ImportPath}}: {{join .Imports " "}}' ./...
  go list -f '{{join .Imports "\n"}}' ./internal/store | grep mystally
  go list -f '{{join .Imports "\n"}}' ./internal/ledger | grep mystally; echo $?
  ```

  ```
  github.com/na-cho-dev/mystally/cmd/mystally: errors fmt github.com/na-cho-dev/mystally/internal/ledger github.com/na-cho-dev/mystally/internal/store maps os slices strings time
  github.com/na-cho-dev/mystally/internal/ledger: fmt time
  github.com/na-cho-dev/mystally/internal/store: encoding/json errors github.com/na-cho-dev/mystally/internal/ledger os slices
  github.com/na-cho-dev/mystally/internal/ledger
  1
  ```

  (the standard-library names on the first three lines depend on your code.)

- [ ] swapping `FileStore` for `MemoryStore` in `main` changes nothing else

  ```bash
  git add -A
  ```

  Now edit `main.go`: pass `store.NewMemoryStore()` where you passed `store.NewFileStore(eventsFile)`. Then run:

  ```bash
  git diff --stat
  go build ./cmd/mystally
  rm -f events.json
  ./mystally add nacho like; echo $?
  ./mystally add ada comment; echo $?
  ./mystally list
  ls events.json
  git restore cmd/mystally/main.go
  go build ./cmd/mystally
  ./mystally add nacho like
  ./mystally list
  ```

  ```
   cmd/mystally/main.go | 2 +-
   1 file changed, 1 insertion(+), 1 deletion(-)
  0
  0
  ID   CREATOR    TYPE     TIME
  ls: cannot access 'events.json': No such file or directory
  ID   CREATOR    TYPE     TIME
  1    nacho      like     2026-10-17 15:57:43
  ```

  (`git diff --stat` must report one file, one insertion and one deletion.)

---

## Stage 5: tests

**Goal:** make the ledger provably correct and learn `go test`.

**End goal:** you can run `go test ./...` and get `ok` for `internal/ledger` and `internal/store`, `go test -v -run TestBalanceFor ./internal/ledger` to watch one test's cases run by name, and `go test -cover ./internal/ledger` to see that more than 80% of the ledger's statements run under test. If you change the like rate from 5 to 6, a test fails and its message names the case and both amounts. You will understand how `go test` finds tests, how to write table-driven tests with `t.Run`, when to stop a test with `t.Fatalf`, why the ledger's tests live in a separate `ledger_test` package, and how to test file code without touching your real `events.json`.

**Build**
- `internal/ledger/ledger_test.go` with table-driven tests for `BalanceFor` and `Balances`, using `MemoryStore`
- A test for `Kobo.Naira` and `Kobo.String` with at least five cases
- A test for `ParseEventType` including an invalid input
- A test that `FileStore` round-trips, using `t.TempDir()`

**Learn:** `testing.T`, `t.Run`, `t.Errorf` vs `t.Fatalf`, anonymous structs for tables, `t.TempDir`, `go test -v`, `go test -cover`, `go test -run Name`.

**Look in:** `testing`, `path/filepath`.

### Structure

The program code doesn't change in this stage, unless a test finds a bug in it. Stage 5 adds three test files, each in the same folder as the code it tests, and one line to `.gitignore`:

```
mystally/
  .gitignore              /mystally, /events.json (from Stage 2), and now /cover.out
  go.mod
  PLAN.md
  cmd/mystally/
    main.go               unchanged. No tests here yet: go test says [no test files]
  internal/ledger/
    event.go              unchanged
    ledger.go             unchanged
    store.go              unchanged
    event_test.go         package ledger        Kobo, ParseEventType, EventType text   (new)
    ledger_test.go        package ledger_test   Ledger tests through MemoryStore       (new)
  internal/store/
    file.go               unchanged
    memory.go             unchanged
    file_test.go          package store         FileStore round trip in t.TempDir()     (new)
```

```
ledger_test.go (package ledger_test) ──► internal/store ──► internal/ledger
        │                                                         ▲
        └─────────────────────────────────────────────────────────┘
event_test.go (package ledger) is part of internal/ledger: it imports no mystally package.
file_test.go  (package store)  is part of internal/store, which already imports internal/ledger.
```

### Steps

Stage 5 has eight steps. Do them in order: Steps 5.1 to 5.3 test `event.go` from inside the `ledger` package, Steps 5.4 and 5.5 test the ledger from outside it, Step 5.6 tests `FileStore`, and Steps 5.7 and 5.8 measure the tests and prove that they catch mistakes. Each step ends with a check you run before you move on, and from now on most checks are `go test` commands instead of `./mystally` runs.

#### Step 5.1: Write your first test

**Achieve:**
- A new file `internal/ledger/event_test.go`, in the same folder as `event.go`, starts with `package ledger`.
- It holds one test, `TestKoboString`, that checks one amount: `Kobo(1250)` prints as `₦12.50`.
- `go test ./...` runs it and prints `ok` for `internal/ledger`.

**Learn:**
- A **test** is a function `func TestXxx(t *testing.T)` in a file whose name ends in `_test.go`. `go test` finds it by those two rules; nothing in your code calls it. The letter after `Test` must not be lowercase: `TestKoboString`, not `Testkobo`.
- `t` is a `*testing.T`: your handle on this one test. `t.Errorf` formats a message like `fmt.Sprintf`, records it and marks the test failed. A test that returns with no failure recorded passes.
- Go has no `assert` functions. You write an ordinary `if got != want { t.Errorf(...) }`. The usual message shape is `Func(input) = <what came back>, want <what you expected>`, and `%q` puts quotes around strings so spaces and empty strings show.
- `go test ./...` tests every package in this folder and below. `ok` means every test passed, `FAIL` means at least one failed, and `[no test files]` means the folder has no `_test.go` file.

<details>
<summary><b>Example: a first test</b></summary>

`temps.go`:

```go
package temps

// Celsius is a temperature.
type Celsius float64

// Fahrenheit converts c.
func (c Celsius) Fahrenheit() float64 { return float64(c)*9/5 + 32 }
```

`temps_test.go`:

```go
package temps

import "testing"

func TestFahrenheit(t *testing.T) {
	got := Celsius(100).Fahrenheit()
	if got != 212 {
		t.Errorf("Celsius(100).Fahrenheit() = %v, want 212", got)
	}
}
```

Output:

```
$ go test -v .
=== RUN   TestFahrenheit
--- PASS: TestFahrenheit (0.00s)
PASS
ok  	example.com/temps	0.007s
```

(your time will differ)

- A test is `func TestXxx(t *testing.T)` in a `_test.go` file. `go test` finds it; nothing calls it.
- No assert functions: an ordinary `if` and `t.Errorf` in the shape `Func(input) = got, want X`.

</details>

**Step done when:**

```bash
go test ./...
go test -v ./internal/ledger
```

```
?   	github.com/na-cho-dev/mystally/cmd/mystally	[no test files]
ok  	github.com/na-cho-dev/mystally/internal/ledger	0.005s
?   	github.com/na-cho-dev/mystally/internal/store	[no test files]
=== RUN   TestKoboString
--- PASS: TestKoboString (0.00s)
PASS
ok  	github.com/na-cho-dev/mystally/internal/ledger	0.005s
```

(your times will differ, or show `(cached)`.)

#### Step 5.2: Turn the test into a table

**Achieve:**
- `TestKoboString` checks at least five amounts from one table. For example 0, 1, 55, 100 and 1250 kobo must print `₦0.00`, `₦0.01`, `₦0.55`, `₦1.00` and `₦12.50`.
- A new `TestKoboNaira` does the same for `Naira()`, with at least five cases. For example 7 kobo must give `0.07` and 1250 kobo `12.5`.
- Each case has a short name and runs as its own subtest, so `go test -v` lists every case.

**Learn:**
- A **table-driven test** puts its cases in a slice and runs the same check on each one. Adding a case is one line.
- Each element is an **anonymous struct**: a struct type written in place, with no `type` name, because only this test uses it. `[]struct{ name string; in Minutes; want string }{ {...}, {...} }` is a slice of them.
- `t.Run(name, func(t *testing.T) { ... })` runs a **subtest**: a test inside a test, with its own name and its own pass or fail. Use the inner `t` inside it, not the outer one.
- In the output a subtest is called `TestMinutesString/pads_minutes`: the parent's name, a `/`, and the case name with spaces turned into `_`.

<details>
<summary><b>Example: a table of cases with subtests</b></summary>

`minutes_test.go`:

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
		{"pads minutes", 125, "2h05m"},
		{"large", 6000, "100h00m"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.in.String(); got != tt.want {
				t.Errorf("Minutes(%d).String() = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
```

Output:

```
$ go test -v .
=== RUN   TestMinutesString
=== RUN   TestMinutesString/zero
=== RUN   TestMinutesString/pads_minutes
=== RUN   TestMinutesString/large
--- PASS: TestMinutesString (0.00s)
    --- PASS: TestMinutesString/zero (0.00s)
    --- PASS: TestMinutesString/pads_minutes (0.00s)
    --- PASS: TestMinutesString/large (0.00s)
PASS
ok  	example.com/minutes	0.008s
$ go test -run 'TestMinutesString/pads' -v .
=== RUN   TestMinutesString
=== RUN   TestMinutesString/pads_minutes
--- PASS: TestMinutesString (0.00s)
    --- PASS: TestMinutesString/pads_minutes (0.00s)
PASS
ok  	example.com/minutes	0.007s
```

(`minutes.go` holds the `Minutes` type from Stage 2. Your times will differ.)

- Each case is one line in the table. `t.Run` gives it its own name and pass or fail.
- Spaces in a case name become `_`, and `-run` picks tests by name.

</details>

**Step done when:**

```bash
go test -v -run TestKobo ./internal/ledger
```

```
=== RUN   TestKoboString
=== RUN   TestKoboString/zero
=== RUN   TestKoboString/one_kobo
=== RUN   TestKoboString/under_one_naira
=== RUN   TestKoboString/whole_naira
=== RUN   TestKoboString/naira_and_kobo
=== RUN   TestKoboString/large
--- PASS: TestKoboString (0.00s)
    --- PASS: TestKoboString/zero (0.00s)
    --- PASS: TestKoboString/one_kobo (0.00s)
    --- PASS: TestKoboString/under_one_naira (0.00s)
    --- PASS: TestKoboString/whole_naira (0.00s)
    --- PASS: TestKoboString/naira_and_kobo (0.00s)
    --- PASS: TestKoboString/large (0.00s)
=== RUN   TestKoboNaira
=== RUN   TestKoboNaira/zero
=== RUN   TestKoboNaira/one_kobo
=== RUN   TestKoboNaira/seven_kobo
=== RUN   TestKoboNaira/whole_naira
=== RUN   TestKoboNaira/naira_and_kobo
--- PASS: TestKoboNaira (0.00s)
    --- PASS: TestKoboNaira/zero (0.00s)
    --- PASS: TestKoboNaira/one_kobo (0.00s)
    --- PASS: TestKoboNaira/seven_kobo (0.00s)
    --- PASS: TestKoboNaira/whole_naira (0.00s)
    --- PASS: TestKoboNaira/naira_and_kobo (0.00s)
PASS
ok  	github.com/na-cho-dev/mystally/internal/ledger	0.008s
```

(your case names and times will differ; each test must list at least five `--- PASS` lines under it)

#### Step 5.3: Test the error path too

**Achieve:**
- `TestParseEventType` checks that each of the five names turns into its constant with a `nil` error.
- It also checks inputs that must be rejected: at least `dance`, and ideas like `Like` (capital letter, unless you chose in Stage 2 to accept any case) and the empty string. For those, the error must not be `nil` and its text must contain `unknown event type`. If you added `Unknown` in Stage 2, also check that the value returned with the error is `Unknown`.
- It uses `t.Fatalf` where a failure makes the rest of the case meaningless, and `t.Errorf` everywhere else.

**Learn:**
- Add a `wantErr bool` field to the table. In the subtest, branch on it: when an error is wanted, check that there is one, then `return`; otherwise check that there isn't one, then compare the value.
- `t.Errorf` records a failure and keeps running the test. `t.Fatalf` records it and stops this test, or this subtest, at once. The other subtests still run.
- Use `Fatalf` when the lines below would be meaningless or would crash, for example reading a result that came back `nil` together with an error.
- An `error` is compared with `nil`, never with `""`. To check its text, use `strings.Contains(err.Error(), "unknown event type")`. Check a fragment, not the whole message, so rewording the rest doesn't break the test. [Stage 6](#stage-6-errors-done-properly) replaces this text check with `errors.Is`.

<details>
<summary><b>Example: test the good input and the bad</b></summary>

`weekday_test.go`:

```go
package weekday

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		in      string
		want    Weekday
		wantErr bool
	}{
		{"monday", Monday, false},
		{"Monday", Unknown, true},
		{"", Unknown, true},
	}
	for _, tt := range tests {
		got, err := Parse(tt.in)
		if tt.wantErr {
			if err == nil || !strings.Contains(err.Error(), "unknown weekday") {
				t.Errorf("Parse(%q) error = %v, want unknown weekday", tt.in, err)
			}
			continue
		}
		if err != nil {
			t.Fatalf("Parse(%q) unexpected error: %v", tt.in, err)
		}
		if got != tt.want {
			t.Errorf("Parse(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}
```

Output:

```
$ go test -v .
=== RUN   TestParse
--- PASS: TestParse (0.00s)
PASS
ok  	example.com/weekday	0.006s
```

(`weekday.go` holds the `Weekday` code from Stage 2. Your time will differ.)

- A `wantErr` field splits each case into "must fail" and "must succeed". Check a fragment of the message, not all of it.
- `t.Fatalf` stops the test when continuing makes no sense; `t.Errorf` records the failure and carries on.

</details>

**Step done when:**

```bash
go test -v -run TestParseEventType ./internal/ledger
```

```
=== RUN   TestParseEventType
=== RUN   TestParseEventType/view
=== RUN   TestParseEventType/like
=== RUN   TestParseEventType/save
=== RUN   TestParseEventType/comment
=== RUN   TestParseEventType/share
=== RUN   TestParseEventType/not_an_event
=== RUN   TestParseEventType/capital_letter
=== RUN   TestParseEventType/empty
--- PASS: TestParseEventType (0.00s)
    --- PASS: TestParseEventType/view (0.00s)
    --- PASS: TestParseEventType/like (0.00s)
    --- PASS: TestParseEventType/save (0.00s)
    --- PASS: TestParseEventType/comment (0.00s)
    --- PASS: TestParseEventType/share (0.00s)
    --- PASS: TestParseEventType/not_an_event (0.00s)
    --- PASS: TestParseEventType/capital_letter (0.00s)
    --- PASS: TestParseEventType/empty (0.00s)
PASS
ok  	github.com/na-cho-dev/mystally/internal/ledger	0.005s
```

(your case names and times will differ; at least one case must be an input that has to be rejected, and it must pass)

#### Step 5.4: Test the ledger from outside its package

**Achieve:**
- A new file `internal/ledger/ledger_test.go` starts with `package ledger_test`. It imports `internal/ledger` and `internal/store` with their full paths, like `main` does.
- Each case builds its own ledger on a fresh `MemoryStore`, loads it, then adds each event, with `t.Fatalf` on any error.
- `TestBalanceFor` is a table with at least these cases: no events (0); one like (5); a like and a share (55); ada's comment plus nacho's view, asked for nacho (1, so other creators don't count); bola with one event of each type (86); a creator with no events (0).
- Every test event has a creator and a fixed time made with `time.Date`. [Stage 6](#stage-6-errors-done-properly) makes `Add` reject an event with a zero time, and these tests should still pass then.

**Learn:**
- A test file may use the folder's package name (`package ledger`, an **internal test**) or that name plus `_test` (`package ledger_test`, an **external test package**). Any other name in that folder is an error.
- The external test package is a separate package, so it imports `ledger` like any other code and sees only exported names: `ledger.NewLedger`, `ledger.Like`. A lowercase name gives `undefined: ledger.like (but have Like)`, and a lowercase field gives `l.rates undefined (cannot refer to unexported field rates)`.
- Why it's needed here: `internal/store` imports `internal/ledger`. A test in `package ledger` that imports `store` would make `ledger` import itself through `store`. Go refuses with `import cycle not allowed in test`. `go build ./...` never notices, because it doesn't compile test files.
- A **fake** is a simple stand-in for a real dependency. `MemoryStore` keeps events in a slice, so each case is fast, needs no file, and starts empty.

<details>
<summary><b>Example: an external test package with a fake store</b></summary>

`internal/shelf/shelf_test.go`:

```go
package shelf_test // a separate package: it may import memstore

import (
	"testing"

	"example.com/library/internal/memstore"
	"example.com/library/internal/shelf"
)

func TestAdd(t *testing.T) {
	m := &memstore.Memory{}
	s := shelf.New(m)
	if err := s.Add("Dune"); err != nil {
		t.Fatal(err)
	}
	got, _ := m.Load()
	if len(got) != 1 || got[0] != "Dune" {
		t.Errorf("titles = %v, want [Dune]", got)
	}
}
```

Output:

```
$ go test -v ./internal/shelf
=== RUN   TestAdd
--- PASS: TestAdd (0.00s)
PASS
ok  	example.com/library/internal/shelf	0.006s
```

(your time will differ)

- `memstore` imports `shelf`, so a test written as `package shelf` couldn't import `memstore`: that is an import cycle. `package shelf_test` is a separate package, so it can.
- The test sees only exported names, exactly like any other caller.

</details>

**Step done when:**

```bash
grep -m1 '^package' internal/ledger/ledger_test.go
go test -v -run TestBalanceFor ./internal/ledger
```

```
package ledger_test
=== RUN   TestBalanceFor
=== RUN   TestBalanceFor/no_events
=== RUN   TestBalanceFor/one_like
=== RUN   TestBalanceFor/like_and_share
=== RUN   TestBalanceFor/other_creators_ignored
=== RUN   TestBalanceFor/one_of_each
=== RUN   TestBalanceFor/unknown_creator
--- PASS: TestBalanceFor (0.00s)
    --- PASS: TestBalanceFor/no_events (0.00s)
    --- PASS: TestBalanceFor/one_like (0.00s)
    --- PASS: TestBalanceFor/like_and_share (0.00s)
    --- PASS: TestBalanceFor/other_creators_ignored (0.00s)
    --- PASS: TestBalanceFor/one_of_each (0.00s)
    --- PASS: TestBalanceFor/unknown_creator (0.00s)
PASS
ok  	github.com/na-cho-dev/mystally/internal/ledger	0.006s
```

(your case names and times will differ; the package line must be `package ledger_test`)

#### Step 5.5: Compare whole maps and slices

**Achieve:**
- The setup from Step 5.4 (fresh `MemoryStore`, `NewLedger`, `Load`, `Add` each event) moves into one helper, `newTestLedger`, that every ledger test calls.
- `TestBalances` compares the whole map from `Balances()` with the map you expect: an empty map for no events; `nacho` 55 and `ada` 20 after nacho's like and share and ada's comment.
- `TestAddGivesNextID` adds three events and checks, through `Events()`, that their IDs are 1, 2 and 3.

**Learn:**
- `==` doesn't work on maps or slices: `invalid operation: got != want (map can only be compared to nil)`, and the same for a slice. Use `maps.Equal(a, b)` (Go 1.21): same keys with the same values, in any order. Use `slices.Equal(a, b)` (Go 1.21): same length and the same elements in the same order.
- Both treat `nil` and empty as equal: `maps.Equal` of a nil map and `map[string]int{}` is `true`.
- `%v` prints a map with its keys sorted, so a failure message is easy to read: `Balances() = map[ada:₦0.20 nacho:₦0.56], want ...`.
- A **test helper** is an ordinary function that takes `t *testing.T`. Calling `t.Helper()` on its first line makes failures inside it report the line in the test that called it.

<details>
<summary><b>Example: compare whole maps, with a helper</b></summary>

`loans_test.go`:

```go
package loans

import (
	"maps"
	"testing"
)

func checkDays(t *testing.T, got, want map[string]int) {
	t.Helper() // failures point at the caller's line
	if !maps.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestDaysByBorrower(t *testing.T) {
	loans := []Loan{{"ada", 14}, {"bola", 21}, {"ada", 7}}
	checkDays(t, DaysByBorrower(loans), map[string]int{"ada": 21, "bola": 21})
	checkDays(t, DaysByBorrower(nil), map[string]int{})
}
```

Output:

```
$ go test -v .
=== RUN   TestDaysByBorrower
--- PASS: TestDaysByBorrower (0.00s)
PASS
ok  	example.com/loans	0.005s
```

(your time will differ)

- Maps and slices can't be compared with `==`; `maps.Equal` and `slices.Equal` compare their contents, and treat nil and empty as equal.
- `t.Helper()` makes a failure inside the helper report the test line that called it.

</details>

**Step done when:**

```bash
go test -v -run 'TestBalances|TestAddGivesNextID' ./internal/ledger
```

```
=== RUN   TestBalances
=== RUN   TestBalances/no_events
=== RUN   TestBalances/two_creators
--- PASS: TestBalances (0.00s)
    --- PASS: TestBalances/no_events (0.00s)
    --- PASS: TestBalances/two_creators (0.00s)
=== RUN   TestAddGivesNextID
--- PASS: TestAddGivesNextID (0.00s)
PASS
ok  	github.com/na-cho-dev/mystally/internal/ledger	0.008s
```

(your names and times will differ)

#### Step 5.6: Round-trip FileStore through a temporary folder

**Achieve:**
- A new file `internal/store/file_test.go`, `package store`, saves two or three events through a `FileStore` whose file is inside `t.TempDir()`, loads them back, and checks every field of every event.
- Times are compared with `Equal`, never `==`.
- A second test loads a file that doesn't exist and expects no events and a `nil` error, as Stage 2 promised.
- After `go test`, no new file has appeared in `internal/store/` or anywhere in the repo.

**Learn:**
- `t.TempDir()` makes a new, empty folder for this test and deletes it when the test ends. Build a path in it with `filepath.Join(dir, "events.json")`: `path/filepath` puts in the right separator for your system.
- Tests run with the working directory set to the **package folder** (`internal/store/`), so a plain `"events.json"` would write next to `file.go`.
- Build test times with `time.Date(2026, 10, 17, 9, 0, 0, 0, time.UTC)`: the same value on every run. `time.Now()` changes every run.
- `t1 == t2` compares how a `time.Time` is stored (clock readings and a location), not the moment it means, and a JSON round trip changes some of that. `t1.Equal(t2)` compares the moment. A struct that holds a `time.Time`, like your `Event`, has the same problem with `==`, so compare it field by field.

<details>
<summary><b>Example: a file round trip in a temporary folder</b></summary>

`shelf_test.go`:

```go
package shelf

import (
	"path/filepath"
	"testing"
	"time"
)

func TestSaveLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "books.json") // deleted after the test
	want := []Book{{Title: "Dune", Added: time.Date(2026, 10, 17, 9, 0, 0, 0, time.UTC)}}

	if err := Save(path, want); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Title != "Dune" || !got[0].Added.Equal(want[0].Added) {
		t.Errorf("Load = %+v, want %+v", got, want)
	}
}
```

Output:

```
$ go test -v .
=== RUN   TestSaveLoad
--- PASS: TestSaveLoad (0.00s)
PASS
ok  	example.com/shelf	0.007s
```

(your time will differ)

- `t.TempDir()` gives a fresh folder per test and removes it afterwards, so tests never touch your real files.
- Fixed times from `time.Date`, compared with `Equal`: `time.Now()` changes every run, and `==` compares how a time is stored, not the moment.

</details>

**Step done when:**

```bash
go test -v ./internal/store
ls -1 internal/store
```

```
=== RUN   TestFileStoreRoundTrip
--- PASS: TestFileStoreRoundTrip (0.00s)
=== RUN   TestFileStoreMissingFile
--- PASS: TestFileStoreMissingFile (0.00s)
PASS
ok  	github.com/na-cho-dev/mystally/internal/store	0.006s
file.go
file_test.go
memory.go
```

(your names and time will differ; `ls` must show no `events.json`)

#### Step 5.7: Measure what the tests cover

**Achieve:**
- `go test -cover ./...` reports more than 80% for `internal/ledger`.
- You have read the per-function report and know which functions no test runs. If the ledger is at 80% or less, or a function that matters is at `0.0%`, add a test for it. A likely pair is `MarshalText` and `UnmarshalText`: one test that turns each event type into text and back, and rejects `dance`, covers both.
- `.gitignore` has a line `/cover.out` (and `/cover.html` if you write that file too).

**Learn:**
- **Coverage** is the share of a package's statements that ran at least once during its tests. It shows what isn't tested at all. It can't show whether the tested lines are checked properly; Step 5.8 does that.
- `go test -cover ./...` adds a percentage to each package line. Since Go 1.22, a package with no tests shows `coverage: 0.0% of statements`.
- `-coverprofile=cover.out` saves the details to a file. `go tool cover -func=cover.out` lists each function with its percentage. `go tool cover -html=cover.out` opens a page in your browser with every line coloured: green ran, red never ran, grey isn't a statement. `-o cover.html` writes the page to a file instead.
- Each package is measured with its own tests only. `FileStore`'s tests run `MarshalText` through `encoding/json`, but that doesn't count for `internal/ledger`.

<details>
<summary><b>Example: what coverage shows</b></summary>

`points.go`:

```go
package fitness

// Points scores a workout.
func Points(minutes int) int { return minutes / 10 }

// Badge names a level.
func Badge(points int) string {
	if points >= 100 {
		return "gold"
	}
	return "none"
}
```

`points_test.go`:

```go
package fitness

import "testing"

func TestPoints(t *testing.T) {
	if got := Points(45); got != 4 {
		t.Errorf("Points(45) = %d, want 4", got)
	}
}
```

Output:

```
$ go test -cover .
ok  	example.com/fitness	0.012s	coverage: 25.0% of statements
$ go test -coverprofile=cover.out . >/dev/null
$ go tool cover -func=cover.out
example.com/fitness/points.go:4:	Points		100.0%
example.com/fitness/points.go:7:	Badge		0.0%
total:					(statements)	25.0%
```

(your time will differ)

- `Badge` has no test, so it shows 0.0%. Coverage finds code no test runs at all.
- It can't tell you whether the tested lines are checked properly; Step 5.8 covers that.

</details>

**Step done when:**

```bash
go test -cover ./...
go test -coverprofile=cover.out ./internal/ledger
go tool cover -func=cover.out
git check-ignore cover.out
```

```
	github.com/na-cho-dev/mystally/cmd/mystally		coverage: 0.0% of statements
ok  	github.com/na-cho-dev/mystally/internal/ledger	0.009s	coverage: 95.2% of statements
ok  	github.com/na-cho-dev/mystally/internal/store	0.007s	coverage: 61.1% of statements
ok  	github.com/na-cho-dev/mystally/internal/ledger	(cached)	coverage: 95.2% of statements
github.com/na-cho-dev/mystally/internal/ledger/event.go:12:	Naira		100.0%
github.com/na-cho-dev/mystally/internal/ledger/event.go:15:	String		100.0%
github.com/na-cho-dev/mystally/internal/ledger/event.go:28:	String		85.7%
github.com/na-cho-dev/mystally/internal/ledger/event.go:44:	ParseEventType	100.0%
github.com/na-cho-dev/mystally/internal/ledger/event.go:53:	MarshalText	100.0%
github.com/na-cho-dev/mystally/internal/ledger/event.go:55:	UnmarshalText	100.0%
github.com/na-cho-dev/mystally/internal/ledger/ledger.go:11:	NewLedger	100.0%
github.com/na-cho-dev/mystally/internal/ledger/ledger.go:25:	Load		80.0%
github.com/na-cho-dev/mystally/internal/ledger/ledger.go:35:	Add		100.0%
github.com/na-cho-dev/mystally/internal/ledger/ledger.go:48:	BalanceFor	100.0%
github.com/na-cho-dev/mystally/internal/ledger/ledger.go:59:	Balances	100.0%
github.com/na-cho-dev/mystally/internal/ledger/ledger.go:68:	Events		100.0%
total:								(statements)	95.2%
cover.out
```

(your times, line numbers, function list and percentages will differ; the `internal/ledger` figure must be above 80%.)

#### Step 5.8: Prove the tests catch a wrong rate, then commit

**Achieve:**
- Change one rate in the ledger's rates map (like: 5 to 6), run the tests, and watch them fail with messages that name each broken case and both amounts. Then change it back and watch them pass.
- If nothing fails, a test somewhere computes its expected value instead of writing it down. Fix that test, then try again.
- The CLI works exactly as it did at the end of Stage 4.
- `go vet ./...` and `gofmt -l .` print nothing. Only then, commit (for example `stage 5: tests`).

**Learn:**
- A test is only useful if it can fail. Making it fail on purpose, once, is the quickest proof.
- Write expected values as literal numbers (`want: 55`), worked out by hand from the [Domain](#domain) table. A test that computes `want` the same way the code does agrees with any rate, right or wrong.
- From `package ledger_test` you can't even see the rates map, which pushes you toward literals. That is one more reason the ledger tests live outside the package.
- Read a failing run from the bottom up: `FAIL github.com/...` names the package, each `--- FAIL:` line names a test or case, and the indented `ledger_test.go:54:` line is your own message.

<details>
<summary><b>Example: a literal expected value catches a changed rate</b></summary>

`weight.go`:

```go
package parcel

var gramsPerBox = map[string]int{"small": 500, "medium": 2500}

// Weight is the weight in grams of count boxes of one size.
func Weight(size string, count int) int { return gramsPerBox[size] * count }
```

`weight_test.go`:

```go
package parcel

import "testing"

func TestWeight(t *testing.T) {
	if got := Weight("medium", 3); got != 7500 { // worked out by hand
		t.Errorf("Weight(medium, 3) = %d, want 7500", got)
	}
}
```

Output:

```
$ go test .
ok  	example.com/parcel	0.006s
$ sed -i 's/2500/2600/' weight.go
$ go test .
--- FAIL: TestWeight (0.00s)
    weight_test.go:7: Weight(medium, 3) = 7800, want 7500
FAIL
FAIL	example.com/parcel	0.007s
FAIL
```

(your times will differ)

- `7500` was worked out by hand, so a wrong rate fails the test. A `want` computed as `gramsPerBox["medium"] * 3` would pass with any rate.
- Read a failure from the bottom up: the package, the test, then your own message.

</details>

**Step done when:**

```bash
# 1. edit internal/ledger/ledger.go: change the like rate from 5 to 6, and save
go test ./internal/ledger; echo $?
# 2. change it back to 5, and save
go test ./internal/ledger; echo $?
go build ./cmd/mystally
rm -f events.json
./mystally add nacho like
./mystally add nacho share
./mystally add ada comment
./mystally balances
```

```
--- FAIL: TestBalanceFor (0.00s)
    --- FAIL: TestBalanceFor/one_like (0.00s)
        ledger_test.go:54: BalanceFor("nacho") = ₦0.06, want ₦0.05
    --- FAIL: TestBalanceFor/like_and_share (0.00s)
        ledger_test.go:54: BalanceFor("nacho") = ₦0.56, want ₦0.55
    --- FAIL: TestBalanceFor/one_of_each (0.00s)
        ledger_test.go:54: BalanceFor("bola") = ₦0.87, want ₦0.86
--- FAIL: TestBalances (0.00s)
    --- FAIL: TestBalances/two_creators (0.00s)
        ledger_test.go:75: Balances() = map[ada:₦0.20 nacho:₦0.56], want map[ada:₦0.20 nacho:₦0.55]
FAIL
FAIL	github.com/na-cho-dev/mystally/internal/ledger	0.008s
FAIL
1
ok  	github.com/na-cho-dev/mystally/internal/ledger	0.006s
0
ada        ₦0.20
nacho      ₦0.55
```

(your case names, message wording, line numbers, times and spacing will differ, and the second run may show `(cached)`; each failure must show the wrong amount and the expected one.)

### Gotchas

- Using `fmt.Errorf` instead of `t.Errorf` in a test makes the test pass whatever happens: `fmt.Errorf` only builds an error value, and nothing reads it. `go test` stays quiet; only `go vet` says `result of fmt.Errorf call not used`. You did the same thing to print a message in Stage 1. *(you hit this)*
- `t.Error` doesn't format. `t.Error("got %q", got)` stops `go test` before it runs anything: `(*testing.common).Error call has possible Printf formatting directive %q`. Use `t.Errorf`. In Stage 1 the same check caught `fmt.Fprintln` with `%q`. *(you hit this)*
- `got, err := ParseEventType("like")` with `err` never read fails with `declared and not used: err`. Check every error, in tests too. *(you hit this)*
- A test file in the wrong folder is a different package. `internal/event_test.go`, one level too high, fails with `internal/event_test.go:6:12: undefined: Kobo` and adds a package `github.com/na-cho-dev/mystally/internal` to the output. In Stage 2 you created `event.go` in `cmd/` the same way. *(you hit this)*
- `go test` with no path at the repo root fails with `no Go files in /.../mystally` and `FAIL . [setup failed]`, just like `go build` and `go vet` did in Stage 1. Use `go test ./...` or a package path. *(you hit this)*
- `gofmt -l .` lists a test file whose only problem is a missing newline at the very end. Run it before you commit. *(you hit this)*

### Check yourself

- Why must `ledger_test.go` be `package ledger_test`?
- When do you use `t.Fatalf` instead of `t.Errorf`?
- Why must the expected numbers in your tests be literals?
- Which lines does `go tool cover -html` show in red, and do they matter?
- Why do the ledger tests use `MemoryStore` and the `FileStore` tests use `t.TempDir()`, instead of your real `events.json`?
- Two `time.Time` values mean the same moment, but `==` says false. Why, and what do you use instead?

### Done when

- [ ] `go test ./...` passes

  ```bash
  go vet ./...
  go test -count=1 ./...
  ```

  ```
  ?   	github.com/na-cho-dev/mystally/cmd/mystally	[no test files]
  ok  	github.com/na-cho-dev/mystally/internal/ledger	0.017s
  ok  	github.com/na-cho-dev/mystally/internal/store	0.009s
  ```

    (your times will differ; `go vet` prints nothing, and every package line must start with `ok` or `?`)

- [ ] changing one rate in the rates map makes a test fail

  ```bash
  # edit internal/ledger/ledger.go: change the like rate from 5 to 6, and save
  go test ./internal/ledger; echo $?
  # change it back to 5, and save
  go test ./internal/ledger; echo $?
  ```

  ```
  --- FAIL: TestBalanceFor (0.00s)
      --- FAIL: TestBalanceFor/one_like (0.00s)
          ledger_test.go:54: BalanceFor("nacho") = ₦0.06, want ₦0.05
      --- FAIL: TestBalanceFor/like_and_share (0.00s)
          ledger_test.go:54: BalanceFor("nacho") = ₦0.56, want ₦0.55
      --- FAIL: TestBalanceFor/one_of_each (0.00s)
          ledger_test.go:54: BalanceFor("bola") = ₦0.87, want ₦0.86
  --- FAIL: TestBalances (0.00s)
      --- FAIL: TestBalances/two_creators (0.00s)
          ledger_test.go:75: Balances() = map[ada:₦0.20 nacho:₦0.56], want map[ada:₦0.20 nacho:₦0.55]
  FAIL
  FAIL	github.com/na-cho-dev/mystally/internal/ledger	0.010s
  FAIL
  1
  ok  	github.com/na-cho-dev/mystally/internal/ledger	0.006s
  0
  ```

  (your case names, message wording, line numbers and times will differ, and the second run may show `(cached)`; each failure must show the wrong amount and the expected one.)

- [ ] coverage of `internal/ledger` is above 80%

  ```bash
  go test -cover ./internal/ledger
  ```

  ```
  ok  	github.com/na-cho-dev/mystally/internal/ledger	0.007s	coverage: 95.2% of statements
  ```

    (your time and percentage will differ; the percentage must be above 80.0%)

---

## Stage 6: errors done properly

**Goal:** errors that callers can inspect, not just print.

**End goal:** when Stage 6 is finished, `./mystally add nacho dance` exits with code 2, `./mystally add "" like` exits with code 3, and storage and usage failures, such as a save that can't create its file, exit with code 1, each with its own message on stderr. `go test ./...` includes tests that check errors with `errors.Is` and `errors.AsType`, and they fail if you swap a `%w` for `%v`. You will understand sentinel errors, custom error types, wrapping with `%w`, how `errors.Is` and `errors.As` look through the wrapping, and why a function that returns `error` must return a literal `nil`.

**Build**
- `var ErrUnknownEventType = errors.New("unknown event type")` returned from `ParseEventType`
- `type ValidationError struct { Field, Reason string }` with an `Error()` method, returned when a creator name is empty or an event has no timestamp
- Every error returned from `FileStore` is wrapped: `fmt.Errorf("load events: %w", err)`
- In `main`, use `errors.Is` to detect `ErrUnknownEventType` (exit 2) and `errors.As` to detect `ValidationError` (exit 3); everything else exits 1
- One helper `fail(err error)` in `main` that does the classification and printing

**Learn:** sentinel errors, custom error types, `%w` wrapping, `errors.Is`, `errors.As`, `errors.Unwrap`, the nil-interface trap (return literal `nil`, never a typed nil pointer).

**Look in:** `errors`, `fmt`.

### Structure

Stage 6 adds one file, `internal/ledger/errors.go`. Everything else stays where Stage 5 left it; only the way errors are made, wrapped and handled changes:

```
mystally/
  .gitignore              unchanged
  go.mod
  PLAN.md
  cmd/mystally/
    main.go               main calls run; on error, calls fail(err), which never returns
  internal/ledger/
    errors.go             ErrUnknownEventType, ValidationError and its Error method    (new)
    event.go              ParseEventType returns ErrUnknownEventType, wrapped with the bad name
    event_test.go         the bad-name checks use errors.Is
    ledger.go             Add validates the event and returns *ValidationError
    ledger_test.go        + Add rejects an empty creator and a zero time
    store.go              unchanged
  internal/store/
    file.go               every returned error wrapped: "load events: %w", "save events: %w"
    file_test.go          + errors.Is(err, os.ErrNotExist) through the wrapping
    memory.go             unchanged
```

### Steps

Stage 6 has six steps. Do them in order: Steps 6.1 to 6.3 give each kind of failure its own error, Step 6.4 teaches `main` to tell them apart, Step 6.5 checks for the nil-interface trap, and Step 6.6 pins everything down with tests; until Step 6.4 every failure still exits 1. Each step ends with a check you run before you move on.

#### Step 6.1: Give the unknown event type a name

**Achieve:**
- A new file, `internal/ledger/errors.go`, declares `ErrUnknownEventType` exactly as in **Build**.
- `ParseEventType` still returns an error for a bad name, but that error now wraps `ErrUnknownEventType` with `%w`, and it still quotes what the user typed. Put the `%w` where the words `unknown event type` used to be in your format string, and pass `ErrUnknownEventType` as its argument.
- What the user sees doesn't change: the same message, and still exit code 1. `UnmarshalText` calls `ParseEventType`, so it gets the new error for free.

**Learn:**
- A **sentinel error** is one error value, made once in a package-level `var`, that callers compare against. `errors.New("...")` makes it. By convention its name starts with `Err`, like `os.ErrNotExist` in the standard library.
- Go constants can only be numbers, strings and booleans, so a sentinel is a `var`. Never assign to it.
- `fmt.Errorf` builds a new error from a format string. The verb `%w` ("wrap") prints the inner error's text, just like `%v`, **and** keeps the inner error inside the new one.
- `errors.Is(err, target)` reports whether `err` is `target` or wraps it, however many layers deep. Plain `==` only looks at the outer error, so it stops matching as soon as anything wraps the sentinel.

<details>
<summary><b>Example: a sentinel error, wrapped and found again</b></summary>

```go
package main

import (
	"errors"
	"fmt"
)

var ErrUnknownDay = errors.New("unknown weekday")

func parse(s string) error {
	if s != "monday" {
		return fmt.Errorf("%w: %q", ErrUnknownDay, s) // %w keeps the sentinel inside
	}
	return nil
}

func main() {
	err := parse("funday")
	fmt.Println(err)
	fmt.Println(err == ErrUnknownDay, errors.Is(err, ErrUnknownDay))

	flat := fmt.Errorf("%v: %q", ErrUnknownDay, "funday") // %v only copies the text
	fmt.Println(flat, errors.Is(flat, ErrUnknownDay))
}
```

Output:

```
$ go run .
unknown weekday: "funday"
false true
unknown weekday: "funday" false
```

- `==` stops matching as soon as anything wraps the sentinel; `errors.Is` looks inside every layer.
- `%v` and `%w` print the same text. Only `%w` keeps the sentinel findable.

</details>

**Step done when:**

```bash
go build ./cmd/mystally
./mystally add nacho dance; echo $?
go doc ./internal/ledger ErrUnknownEventType
go test ./internal/ledger
```

```
unknown event type "dance"; choose one of: view, like, save, comment, share
1
package ledger // import "github.com/na-cho-dev/mystally/internal/ledger"

var ErrUnknownEventType = errors.New("unknown event type")
    ErrUnknownEventType is returned (wrapped) by ParseEventType.

ok  	github.com/na-cho-dev/mystally/internal/ledger	0.016s
```

(your wording may differ; the message must still contain `unknown event type` and `dance`, and the code stays 1 until Step 6.4.)

#### Step 6.2: Return a ValidationError from Add

**Achieve:**
- `errors.go` also declares `ValidationError` exactly as in **Build**, with an `Error` method on a pointer receiver that builds the message from `Field` and `Reason`, for example `invalid creator: must not be empty`.
- `Add` checks the event before it does anything else. A creator that is empty or only spaces, or an `At` time that was never set, returns a `*ValidationError` naming that field. A rejected event gets no ID, isn't appended and isn't saved.
- If `main.go` checks for an empty creator itself, delete that check. The rule now lives in the ledger, so every caller gets it, including Stage 8's HTTP API.
- Any Stage 5 test event without a time now fails `Add`. Give each one a fixed time with `time.Date`, as Stage 5 asked.

**Learn:**
- `error` is a built-in interface with one method: `type error interface { Error() string }`. Any type with an `Error() string` method is an error.
- A **custom error type** is usually a struct, so the error carries details that callers can read as fields (`Field`, `Reason`), not only as text.
- With a pointer receiver, `func (e *ValidationError) Error() string`, only `*ValidationError` has the method. So you return `&ValidationError{...}`, and callers look for `*ValidationError`.
- `var _ error = (*ValidationError)(nil)` is Stage 4's compile-time check, pointed at the built-in `error` interface.

<details>
<summary><b>Example: an error type that names the bad field</b></summary>

```go
package main

import (
	"errors"
	"fmt"
)

type FieldError struct {
	Field, Problem string
}

// Error makes *FieldError an error. Pointer receiver: callers look for *FieldError.
func (e *FieldError) Error() string { return e.Field + " " + e.Problem }

func book(guest string, nights int) error {
	if guest == "" {
		return &FieldError{Field: "guest", Problem: "must not be empty"}
	}
	if nights < 1 {
		return &FieldError{Field: "nights", Problem: "must be at least 1"}
	}
	return nil
}

func main() {
	err := fmt.Errorf("check-in: %w", book("ada", 0))
	fmt.Println(err)

	var fe *FieldError
	if errors.As(err, &fe) {
		fmt.Println("bad field:", fe.Field)
	}
	fmt.Println(book("ada", 2))
}
```

Output:

```
$ go run .
check-in: nights must be at least 1
bad field: nights
<nil>
```

- Any type with `Error() string` is an `error`. A struct lets callers read details as fields, not just text.
- `errors.As` finds the first `*FieldError` inside the chain and stores it in `fe`.

</details>

**Step done when:**

```bash
go build ./cmd/mystally
rm -f events.json
./mystally add "" like; echo $?
./mystally add "   " like; echo $?
ls events.json
./mystally add nacho like; echo $?
go test ./...
```

```
invalid creator: must not be empty
1
invalid creator: must not be empty
1
ls: cannot access 'events.json': No such file or directory
0
?   	github.com/na-cho-dev/mystally/cmd/mystally	[no test files]
ok  	github.com/na-cho-dev/mystally/internal/ledger	0.009s
ok  	github.com/na-cho-dev/mystally/internal/store	0.009s
```

(your wording may differ; it must name the creator.)

#### Step 6.3: Wrap every FileStore error with context

**Achieve:**
- `FileStore.Load` wraps every error it returns with `fmt.Errorf("load events: %w", err)`: the error from reading the file and the error from `encoding/json`.
- `FileStore.Save` does the same with `save events: %w`, for the JSON error and the write error.
- A missing file is still not an error. `Load` checks for it before it wraps anything, and returns an empty slice and `nil`, as in Stage 4.
- The ledger and `main` pass these errors on unchanged, so each message holds exactly one `load events:` or `save events:`.

**Learn:**
- A good error message reads from what you were doing down to what broke: `load recipes: open recipes.json: no such file or directory`. Each layer adds what *it* was doing, once.
- The `os` error already names the operation and the file (`open recipes.json`), so your prefix doesn't repeat the path.
- `errors.Unwrap(err)` returns the error one layer inside `err`, or `nil` when there is none. You rarely call it yourself: `errors.Is` and `errors.As` unwrap layer by layer for you.
- `os.ErrNotExist` is a sentinel from the `os` package. You already use `errors.Is(err, os.ErrNotExist)` in `FileStore.Load`, and it keeps working through any number of `%w` layers.

<details>
<summary><b>Example: add context, keep the cause</b></summary>

```go
package main

import (
	"errors"
	"fmt"
	"os"
)

func loadRecipes(path string) error {
	if _, err := os.ReadFile(path); err != nil {
		return fmt.Errorf("load recipes: %w", err)
	}
	return nil
}

func main() {
	err := loadRecipes("recipes.json")
	fmt.Println(err)
	fmt.Println(errors.Is(err, os.ErrNotExist))
	fmt.Println(errors.Unwrap(err))
}
```

Output:

```
$ go run .
load recipes: open recipes.json: no such file or directory
true
open recipes.json: no such file or directory
```

- Each layer adds what *it* was doing. The `os` error already names the file, so the prefix doesn't repeat it.
- `errors.Is` still finds `os.ErrNotExist` through the wrapping; `errors.Unwrap` peels off one layer.

</details>

**Step done when:**

```bash
go build ./cmd/mystally
echo '[{"id": 1,' > events.json
./mystally list; echo $?
rm events.json
ln -s no-such-folder/events.json events.json
./mystally list; echo $?
./mystally add nacho like; echo $?
rm events.json
./mystally add nacho like; echo $?
```

```
load events: unexpected end of JSON input
1
ID   CREATOR    TYPE     TIME
0
save events: open events.json: no such file or directory
1
0
```

(the `echo` line writes half a JSON array, to fake a damaged file.)

#### Step 6.4: Classify errors in fail

**Achieve:**
- `main.go` gets `fail(err error)`. It prints `err` to stderr, then exits with 2 when `errors.Is` finds `ledger.ErrUnknownEventType`, with 3 when `errors.As` (or `errors.AsType`) finds a `*ledger.ValidationError`, and with 1 for anything else.
- `main` calls `fail` when `run` returns an error. `fail` is now the only place in mystally that calls `os.Exit`.
- Usage errors and storage errors still exit 1.

**Learn:**
- `errors.Is(err, target)` checks `err` and every error inside it for one equal to `target`. Use it for sentinel **values**: `ErrUnknownEventType`, `os.ErrNotExist`.
- `errors.As(err, &target)` checks the same chain for the first error of `target`'s **type**. If it finds one, it stores it in `target` and returns true. Declare the variable first (`var ve *ValidationError`), then pass its address, `&ve`.
- Go 1.26 added `errors.AsType[*ValidationError](err)`. It returns the error and a `bool`, with no variable to declare first. The part in square brackets is a **type argument**; Stage 9 explains generics.
- `switch { case ...: }` with nothing after `switch` runs the first case that is true, so put the most specific check first. `errors.Is` and `errors.As` each return one `bool`, so they fit in a `case`. `errors.AsType` returns two values, so use it in an `if ...; ok` instead.

<details>
<summary><b>Example: turn errors into exit codes</b></summary>

```go
package main

import (
	"errors"
	"fmt"
)

var ErrRoomTaken = errors.New("room already taken")

type FieldError struct{ Field string }

func (e *FieldError) Error() string { return e.Field + " is missing" }

func exitCode(err error) int {
	var fe *FieldError
	switch {
	case errors.Is(err, ErrRoomTaken):
		return 2
	case errors.As(err, &fe):
		return 3
	default:
		return 1
	}
}

func main() {
	errs := []error{
		fmt.Errorf("book: %w", ErrRoomTaken),
		fmt.Errorf("book: %w", &FieldError{Field: "guest"}),
		errors.New("disk full"),
	}
	for _, err := range errs {
		fmt.Println(exitCode(err), err)
	}
	fe, ok := errors.AsType[*FieldError](errs[1]) // Go 1.26: no variable needed
	fmt.Println(fe.Field, ok)
}
```

Output:

```
$ go run .
2 book: room already taken
3 book: guest is missing
1 disk full
guest true
```

- A `switch` with nothing after it runs the first true `case`, so put the most specific check first.
- `errors.AsType` (Go 1.26) returns the error and a `bool` without declaring a variable first.

</details>

**Step done when:**

```bash
go build ./cmd/mystally
rm -f events.json
./mystally add nacho dance; echo $?
./mystally add "" like; echo $?
./mystally add "" dance; echo $?
./mystally list extra; echo $?
./mystally add nacho like; echo $?
go run ./cmd/mystally add nacho dance; echo $?
```

```
unknown event type "dance"; choose one of: view, like, save, comment, share
2
invalid creator: must not be empty
3
unknown event type "dance"; choose one of: view, like, save, comment, share
2
usage: mystally list
1
0
unknown event type "dance"; choose one of: view, like, save, comment, share
exit status 2
1
```

(your wording may differ; the codes must be 2, 3, 2, 1 and 0.)

#### Step 6.5: Return a literal nil when nothing is wrong

**Achieve:**
- Every mystally function whose result type is `error` returns a literal `nil` when nothing went wrong, never a `*ValidationError` variable that happens to be nil.
- If you moved `Add`'s checks into a helper, give the helper the result type `error` and let it `return nil` itself.
- Break it once on purpose, so you recognise it later: in `Add` (or your helper), declare a `var` of type `*ValidationError`, set it only when a check fails, and return that variable where you now return `nil`. Run the check below, read the symptom, then put it back.

**Learn:**
- An interface value is a pair: (type, value). It equals `nil` only when both are empty.
- `var p *ValidationError` is a nil pointer. `var err error = p` stores (`*ValidationError`, nil) in `err`: the value is nil but the type isn't, so `err != nil` is true.
- That happens whenever a variable of a concrete error type leaves a function through a result of type `error`. `return nil` stores (no type, no value), which is what callers expect.
- `go vet` doesn't catch it. In mystally the trap shows up as a good `add` that prints `<nil>` and exits 3, while your Stage 5 tests that add events fail and print `<nil>` where the error should be. `errors.As` finds the type `*ValidationError`, so `fail` picks 3. Printing calls `Error` on a nil pointer, which panics when it reads `e.Field`, and `fmt` catches that panic and prints `<nil>`.

<details>
<summary><b>Example: the nil-interface trap</b></summary>

```go
package main

import "fmt"

type FieldError struct{}

func (*FieldError) Error() string { return "bad field" }

// Wrong: a nil *FieldError inside an error is not a nil error.
func check(ok bool) error {
	var fe *FieldError
	if !ok {
		fe = &FieldError{}
	}
	return fe
}

// Right: return a literal nil when nothing is wrong.
func checkFixed(ok bool) error {
	if !ok {
		return &FieldError{}
	}
	return nil
}

func main() {
	fmt.Println(check(true) == nil, checkFixed(true) == nil)
}
```

Output:

```
$ go run .
false true
```

- An interface value is a pair (type, value). `check(true)` returns (`*FieldError`, nil): not nil, so `err != nil` is true.
- `go vet` doesn't catch it. Return a literal `nil` on success.

</details>

**Step done when:**

```bash
go build ./cmd/mystally
rm -f events.json
./mystally add nacho like; echo $?
./mystally add ada share; echo $?
./mystally balances
go test ./...
```

```
0
0
ada        ₦0.50
nacho      ₦0.05
?   	github.com/na-cho-dev/mystally/cmd/mystally	[no test files]
ok  	github.com/na-cho-dev/mystally/internal/ledger	0.011s
ok  	github.com/na-cho-dev/mystally/internal/store	0.011s
```

(your spacing may differ; the names must be in alphabetical order with ₦0.50 and ₦0.05.)

#### Step 6.6: Assert the wrapped errors in tests

**Achieve:**
- `event_test.go`: every check on a bad name uses `errors.Is(err, ErrUnknownEventType)`. Replace the Stage 5 `strings.Contains` check, and add a small `TestParseEventTypeUnknown` with one job: `dance` must give an error that wraps the sentinel.
- `ledger_test.go`: a new `TestAddInvalid` table with an empty creator, a creator of only spaces, and a zero time. Each must return a `*ledger.ValidationError` with the right `Field`, and the ledger must still hold no events.
- `file_test.go`: a new `TestFileStoreSaveMissingFolder`. A `FileStore` whose path is inside a folder that doesn't exist must fail to `Save` with an error that satisfies `errors.Is(err, os.ErrNotExist)`.
- The names are suggestions. Any name that starts with `Test` and a capital letter works.

**Learn:**
- Test what callers rely on: `errors.Is` for a sentinel, `errors.As` or `errors.AsType` for a type. A check on the text still passes when the wrapping is gone, because `%v` and `%w` print the same thing.
- `errors.AsType` hands you the error with its real type, so a test can check its fields too, as the example checks `pe.Op`.
- `event_test.go` is `package ledger`, so it writes `ErrUnknownEventType`. `ledger_test.go` is `package ledger_test`, so it writes `ledger.ValidationError`.
- A path such as `filepath.Join(t.TempDir(), "no-such-folder", "catalogue.json")` points into a folder that doesn't exist. Saving there fails with a real "no such file or directory", with no setup and nothing to clean up.

<details>
<summary><b>Example: tests that check the wrapping, not the wording</b></summary>

`recipes_test.go`:

```go
package recipes

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMissing(t *testing.T) {
	err := Load(filepath.Join(t.TempDir(), "none.json"))
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Load error = %v, want os.ErrNotExist inside", err)
	}
	if pe, ok := errors.AsType[*fs.PathError](err); !ok || pe.Op != "open" {
		t.Errorf("want a *fs.PathError from open, got %v", err)
	}
	t.Log(err)
}
```

Output:

```
$ go test -v -trimpath .
=== RUN   TestLoadMissing
    recipes_test.go:19: load recipes: open /tmp/TestLoadMissing1786638312/001/none.json: no such file or directory
--- PASS: TestLoadMissing (0.00s)
PASS
ok  	example.com/recipes	0.008s
```

(the logged path and your time will differ)

- Assert with `errors.Is` and `errors.AsType`: a text check still passes if someone swaps `%w` for `%v`.
- `t.Log` shows the message under `-v` without asserting on its wording.

</details>

**Step done when:**

```bash
go test -v -run 'TestParseEventTypeUnknown|TestAddInvalid|TestFileStoreSaveMissingFolder' ./...
```

```
?   	github.com/na-cho-dev/mystally/cmd/mystally	[no test files]
=== RUN   TestParseEventTypeUnknown
--- PASS: TestParseEventTypeUnknown (0.00s)
=== RUN   TestAddInvalid
=== RUN   TestAddInvalid/empty_creator
=== RUN   TestAddInvalid/only_spaces
=== RUN   TestAddInvalid/zero_time
--- PASS: TestAddInvalid (0.00s)
    --- PASS: TestAddInvalid/empty_creator (0.00s)
    --- PASS: TestAddInvalid/only_spaces (0.00s)
    --- PASS: TestAddInvalid/zero_time (0.00s)
PASS
ok  	github.com/na-cho-dev/mystally/internal/ledger	0.014s
=== RUN   TestFileStoreSaveMissingFolder
    file_test.go:54: save events: open /tmp/TestFileStoreSaveMissingFolder3156405424/001/no-such-folder/events.json: no such file or directory
--- PASS: TestFileStoreSaveMissingFolder (0.00s)
PASS
ok  	github.com/na-cho-dev/mystally/internal/store	0.018s
```

(put your own test names in `-run`; `|` means "or".)

### Gotchas

- `%v` turns the inner error into plain text, so `errors.Is` and `errors.As` can't see it any more. Only `%w` keeps it findable. `go vet` doesn't warn you; only a test catches it.
- Exit code 2 is also what Go uses when a program panics. In Stage 1 your `list` panicked with `index out of range [3] with length 2`, and a panic like that always exits with code 2. If you see 2 together with a stack trace, it is a panic, not an unknown event type. *(you hit this)*
- Formatting `e` itself inside its own `Error` method calls `Error` again, forever. `go vet` says `fmt.Sprintf format %v with arg e causes recursive (*github.com/na-cho-dev/mystally/internal/ledger.ValidationError).Error method call`, and running it ends with `goroutine stack exceeds 1000000000-byte limit`. It's your Stage 2 `String` mistake in a new place: build the message from `e.Field` and `e.Reason`. *(you hit this)*
- `fmt.Errorf` only builds an error. On a line of its own it does nothing, and `go vet` says `result of fmt.Errorf call not used`, the message you met in Stage 1 when you tried to print with it. Write `return fmt.Errorf(...)`. *(you hit this)*
- Error strings start with a lowercase letter and end without punctuation or `"\n"`, because they get wrapped into the middle of longer lines, as in `save events: open events.json: ...`. In Stage 1, messages ending in `"\n"` printed a blank line. *(you hit this)*
- Wrap; don't replace. In Stage 1 you printed a fixed message instead of `err`, and it hid the real cause, `permission denied`. `%w` adds your context and keeps the cause. *(you hit this)*

### Check yourself

- When do you use a sentinel error, and when a custom error type?
- What does `%w` do that `%v` doesn't?
- Why does returning a nil `*ValidationError` as an `error` break `if err != nil`?
- Which exit code does each of your failure cases give? Write them down, then check them with `echo $?`.
- Why does a damaged `events.json` with an unknown type exit 2, and would you change that?
- Why is `fail` the only function in mystally that calls `os.Exit`?

### Done when

- [ ] a missing file, a bad event name, and an empty creator produce three different messages and exit codes

  ```bash
  go build ./cmd/mystally
  rm -f events.json
  ln -s no-such-folder/events.json events.json
  ./mystally add nacho like; echo $?
  rm events.json
  ./mystally add nacho dance; echo $?
  ./mystally add "" like; echo $?
  ./mystally add nacho like; echo $?
  ```

  ```
  save events: open events.json: no such file or directory
  1
  unknown event type "dance"; choose one of: view, like, save, comment, share
  2
  invalid creator: must not be empty
  3
  0
  ```

  (a missing `events.json` is not an error for mystally, it just means no events yet, so this check uses the link from Step 6.3 to make the save itself hit a missing folder.)

- [ ] `errors.Is(err, os.ErrNotExist)` works through your wrapping

  ```bash
  go build ./cmd/mystally
  go test -v -run TestFileStoreSaveMissingFolder ./internal/store
  rm -f events.json
  ./mystally list; echo $?
  ```

  ```
  === RUN   TestFileStoreSaveMissingFolder
      file_test.go:54: save events: open /tmp/TestFileStoreSaveMissingFolder2132500468/001/no-such-folder/events.json: no such file or directory
  --- PASS: TestFileStoreSaveMissingFolder (0.00s)
  PASS
  ok  	github.com/na-cho-dev/mystally/internal/store	0.011s
  ID   CREATOR    TYPE     TIME
  0
  ```

  (the test name is yours; your temporary folder name and times will differ.)

- [ ] a test asserts the wrapped error with `errors.Is`

  ```bash
  go test ./...
  # now change %w to %v in ParseEventType's fmt.Errorf, then:
  go vet ./internal/ledger
  go test ./internal/ledger
  # change it back to %w, then:
  go test ./...
  ```

  ```
  ?   	github.com/na-cho-dev/mystally/cmd/mystally	[no test files]
  ok  	github.com/na-cho-dev/mystally/internal/ledger	0.008s
  ok  	github.com/na-cho-dev/mystally/internal/store	0.009s
  --- FAIL: TestParseEventType (0.00s)
      --- FAIL: TestParseEventType/not_an_event (0.00s)
          event_test.go:75: error "unknown event type \"dance\"; choose one of: view, like, save, comment, share" does not wrap ErrUnknownEventType
      --- FAIL: TestParseEventType/capital_letter (0.00s)
          event_test.go:75: error "unknown event type \"Like\"; choose one of: view, like, save, comment, share" does not wrap ErrUnknownEventType
      --- FAIL: TestParseEventType/empty (0.00s)
          event_test.go:75: error "unknown event type \"\"; choose one of: view, like, save, comment, share" does not wrap ErrUnknownEventType
  --- FAIL: TestParseEventTypeUnknown (0.00s)
      event_test.go:112: ParseEventType(dance) error = unknown event type "dance"; choose one of: view, like, save, comment, share, want one wrapping ErrUnknownEventType
  FAIL
  FAIL	github.com/na-cho-dev/mystally/internal/ledger	0.007s
  FAIL
  ?   	github.com/na-cho-dev/mystally/cmd/mystally	[no test files]
  ok  	github.com/na-cho-dev/mystally/internal/ledger	(cached)
  ok  	github.com/na-cho-dev/mystally/internal/store	(cached)
  ```

  (your times, or `(cached)`, will differ, and the failure lines are whatever your tests print.)

---

## Stage 7: concurrency

**Goal:** goroutines, channels, WaitGroup, context, the race detector.

**End goal:** you can run `./mystally import testdata/import` to merge eight sample files into your ledger, and `./mystally import --workers 1 --slow testdata/import` then `--workers 8 --slow` to watch the same import drop from about 1.6 seconds to about 0.2 seconds. A one millisecond timeout or a Ctrl-C stops an import cleanly, saves nothing and leaves no goroutine behind, and `go test -race ./...` passes with nothing reported. You will understand how goroutines hand results back over a channel, who closes that channel and why, how a context tells goroutines to stop, and how the race detector finds bugs that normal runs hide.

**Build**
- `mystally import <dir>` reads every `*.json` file in the directory concurrently, one goroutine per file, and merges all events into the ledger
- Results come back on a channel as a struct `{events []Event; err error}`; a `sync.WaitGroup` closes the channel when all goroutines finish
- The import takes a `context.Context` with a 10 second timeout; a goroutine checks `ctx.Done()` before starting work and the file read respects it
- Ctrl-C cancels the import cleanly via `signal.NotifyContext`
- `mystally import --workers N` limits concurrency with a buffered channel used as a semaphore
- Put eight sample files in `testdata/import/`

**Learn:** `go`, `chan`, `make`, `close`, `range` over a channel, `select`, `sync.WaitGroup`, `context.WithTimeout`, `defer cancel()`, `signal.NotifyContext`, `go run -race`, goroutine leaks and deadlocks.

**Look in:** `sync`, `context`, `os/signal`, `path/filepath` (`Glob`).

### Structure

Stage 7 adds one package, one command and a folder of sample files. `internal/ledger` and `internal/store` don't change.

```
mystally/
  .gitignore                unchanged: /mystally, /events.json, /cover.out
  go.mod
  cmd/mystally/
    main.go                 + an "import" case in run, + importCmd
  internal/importer/        (new package)
    importer.go             Import, ImportWithDelay, loadFile, a small result struct
    importer_test.go        package importer_test, run with -race             (Step 7.8)
  internal/ledger/          unchanged
  internal/store/           unchanged
  testdata/import/          (new) day-1.json ... day-8.json, five events each
```

```
importCmd ── ctx (Ctrl-C + 10s timeout) ──► importer.ImportWithDelay
                                              ├─ one goroutine per file (loadFile):
                                              │    wait for a semaphore slot → check ctx → wait out --slow → Load → send a result
                                              ├─ closer goroutine: wg.Wait(), then close(results)
                                              └─ for r := range results: collect events, keep the first error
importCmd ◄── []ledger.Event (sorted), error ─┘
   then adds the events to the ledger and saves, in ONE goroutine (no lock needed)
```

### Steps

Stage 7 has eight steps. Steps 7.1 and 7.2 build a plain one-file-at-a-time import and measure it, Steps 7.3 and 7.4 make it concurrent and prove it's correct, Steps 7.5 to 7.7 add the three ways to stop or slow it (a deadline, a worker limit, Ctrl-C), and Step 7.8 tests it under the race detector. Do them in order: each step ends with a check you run before you move on.

#### Step 7.1: Read the files one after another

**Achieve:**
- Eight sample files in `testdata/import/`, `day-1.json` to `day-8.json`. Each holds five events in the same JSON format as `events.json`. The first commands under "Step done when" make them.
- A new package `internal/importer` with `Import(ctx, dir, workers)`. For now it ignores `ctx` and `workers`, reads the files in a plain loop with `store.NewFileStore(path).Load()`, and returns all the events sorted by time. An error names the file it came from.
- A new command `mystally import <dir>` that times the `Import` call, adds every event to the ledger with `l.Add`, and prints two lines: how long the import took, and how many events it added. The timing line prints even when the import fails.
- A missing or empty folder gives an error, not a silent "imported 0 events".

**Learn:**
- `filepath.Glob(pattern)` returns the paths that match a pattern such as `testdata/import/*.json`, sorted by name. Build the pattern with `filepath.Join(dir, "*.json")`, never with `+ "/"`.
- Glob only returns an error for a malformed pattern. A folder that doesn't exist simply matches nothing, so check `len(paths) == 0` yourself.
- To time something: `start := time.Now()` before it, `time.Since(start)` after. The result is a `time.Duration`, which prints itself like `1.604936s` or `363µs`. `.Round(time.Microsecond)` drops the nanoseconds.
- `ctx` and `workers` are in the signature from the start, so nothing that calls `Import` has to change later. Until Step 7.5, pass `context.Background()`: an empty context that never ends. Go doesn't complain about unused parameters.

<details>
<summary><b>Example: count lines in every file, one after another</b></summary>

```go
package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func main() {
	start := time.Now()
	paths, err := filepath.Glob(filepath.Join("notes", "*.txt"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	total := 0
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		total += bytes.Count(data, []byte("\n"))
	}
	fmt.Println(paths, total, "lines in", time.Since(start).Round(time.Microsecond))
}
```

Output:

```
$ go run .
[notes/mon.txt notes/tue.txt] 3 lines in 355µs
```

(your time will differ)

- `filepath.Glob` returns matching paths sorted by name. A missing folder isn't an error: it matches nothing.
- `time.Since(start)` measures; this sequential version is the baseline the concurrent one must beat.

</details>

**Step done when:**

First make the sample files, from the repo root. The `cat` command writes `day-1.json`; the loop copies it seven times, and `sed "s/old/new/"` changes the date in each copy:

```bash
mkdir -p testdata/import
cat > testdata/import/day-1.json <<'EOF'
[
  {"id": 1, "creator": "nacho", "type": "like",    "at": "2026-10-01T09:00:00Z"},
  {"id": 2, "creator": "nacho", "type": "share",   "at": "2026-10-01T09:05:00Z"},
  {"id": 3, "creator": "ada",   "type": "comment", "at": "2026-10-01T10:00:00Z"},
  {"id": 4, "creator": "bola",  "type": "view",    "at": "2026-10-01T11:30:00Z"},
  {"id": 5, "creator": "ada",   "type": "save",    "at": "2026-10-01T12:00:00Z"}
]
EOF
for d in 2 3 4 5 6 7 8; do sed "s/2026-10-01/2026-10-0$d/" testdata/import/day-1.json > testdata/import/day-$d.json; done
```

Each file pays nacho 55 kobo (a like and a share), ada 30 (a comment and a save) and bola 1 (a view), using the rates in [Domain](#domain). All eight together pay nacho ₦4.40, ada ₦2.40 and bola ₦0.08. Commit these files with the code.

Then build mystally and check:

```bash
go build ./cmd/mystally
rm -f events.json
head -3 testdata/import/day-8.json
./mystally import testdata/import
./mystally balances
./mystally list | head -4
./mystally list | tail -1
./mystally import nowhere; echo $?
./mystally import; echo $?
```

```
[
  {"id": 1, "creator": "nacho", "type": "like",    "at": "2026-10-08T09:00:00Z"},
  {"id": 2, "creator": "nacho", "type": "share",   "at": "2026-10-08T09:05:00Z"},
import took 1.154ms
imported 40 events
ada        ₦2.40
bola       ₦0.08
nacho      ₦4.40
ID   CREATOR    TYPE     TIME
1    nacho      like     2026-10-01 09:00:00
2    nacho      share    2026-10-01 09:05:00
3    ada        comment  2026-10-01 10:00:00
40   ada        save     2026-10-08 12:00:00
import took 55µs
no .json files in nowhere
1
usage: mystally import <dir>
1
```

(your time will differ.)

#### Step 7.2: Give import its own flags

**Achieve:**
- `mystally import [--slow] <dir>` parses its flags with its own `flag.FlagSet`. The folder is the one argument left after the flags.
- `--slow` pauses 200 milliseconds before each file, so eight files take about 1.6 seconds in a loop. That is the time Step 7.3 has to beat. Write it down.
- The pause goes in the importer: add `ImportWithDelay(ctx, dir, workers, delay)` next to `Import`, and make `Import` call it with a delay of 0. For now a plain `time.Sleep(delay)` before each file is enough; Step 7.5 replaces it.
- A flag after the folder, or a flag that doesn't exist, gives an error and exit code 1.

**Learn:**
- The `flag` package parses options like `--slow` and `--workers 4`. `flag.NewFlagSet("import", flag.ContinueOnError)` makes a parser for one subcommand.
- Define each flag on it before parsing. `fs.Bool("slow", false, "help text")` returns a `*bool` and `fs.Int("workers", 4, "help text")` returns an `*int`: pointers, because the value is only filled in by `Parse`. Read them with `*slow` afterwards.
- `fs.Parse(args)` reads flags from the front of `args` and stops at the first argument that isn't a flag. What's left is `fs.Args()`. `fs.NArg()` counts it, and `fs.Arg(0)` is the first one (or `""` if there is none: no panic).
- Give `Parse` only what comes after the command name. In `run`, where `args[0]` is `"import"`, that's `args[1:]`.

<details>
<summary><b>Example: flags for one subcommand</b></summary>

```go
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

func run(args []string) error {
	fs := flag.NewFlagSet("count", flag.ContinueOnError)
	slow := fs.Bool("slow", false, "pause per file")
	workers := fs.Int("workers", 4, "files read at once")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return errors.New("usage: count [--slow] [--workers N] <dir>")
	}
	fmt.Println("dir:", fs.Arg(0), "slow:", *slow, "workers:", *workers)
	return nil
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

Output:

```
$ go build -o count .
$ ./count --workers 2 notes
dir: notes slow: false workers: 2
$ ./count notes --slow
usage: count [--slow] [--workers N] <dir>
```

- `fs.Bool` and `fs.Int` return pointers that `Parse` fills in, hence `*slow`.
- Parsing stops at the first argument that isn't a flag: in `notes --slow`, `--slow` is left over as a second argument.

</details>

**Step done when:**

```bash
go build ./cmd/mystally
rm -f events.json
./mystally import testdata/import
./mystally import --slow testdata/import
./mystally import testdata/import --slow; echo $?
./mystally import --fast testdata/import; echo $?
./mystally balances
```

```
import took 2.264ms
imported 40 events
import took 1.607702s
imported 40 events
usage: mystally import [--slow] <dir>
1
flag provided but not defined: -fast
Usage of import:
  -slow
    	wait 200ms per file, to make the timing visible
flag provided but not defined: -fast
1
ada        ₦4.80
bola       ₦0.16
nacho      ₦8.80
```

(your times will differ; the `--slow` run must take about 1.6 seconds.)

#### Step 7.3: Read every file in its own goroutine

**Achieve:**
- `ImportWithDelay` starts one goroutine per file. Each one waits out `delay`, loads its file and sends one result (its path, plus its events or its error) on a channel.
- A `sync.WaitGroup` counts those goroutines, and a separate closer goroutine closes the channel when they have all finished.
- The importer receives with `for r := range results` until the channel is closed. It collects the events, keeps only the first error (with its file name), and still sorts the events by time before returning them.
- `--slow` now takes about 200ms instead of 1.6 seconds. A broken file still fails the whole import, and nothing is added to the ledger.

**Learn:**
- A **goroutine** is a function running at the same time as the rest of your program. `go f()` starts one and carries on at once, without waiting for `f`. When `main` returns, the program ends, whether the goroutines have finished or not.
- `go func() { ... }()` starts a function with no name (a **function literal**); the `()` at the end calls it. It can use variables of the function around it, such as `p` and `results`: a function that does that is called a **closure**.
- A **channel** is a typed pipe between goroutines. `results := make(chan result)` makes one, `results <- r` sends, `r := <-results` receives. On an **unbuffered** channel (no size given to `make`), a send waits until another goroutine receives, and a receive waits for a send: the two goroutines meet.
- `close(results)` means "no more values are coming". `for r := range results` receives value after value and ends when the channel is closed and empty.

<details>
<summary><b>Example: one goroutine each, results on a channel</b></summary>

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

type result struct {
	name string
	n    int
}

func main() {
	start := time.Now()
	results := make(chan result)
	var wg sync.WaitGroup
	for _, name := range []string{"ada", "bola", "chidi"} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(100 * time.Millisecond) // pretend to read a file
			results <- result{name, len(name)}
		}()
	}
	go func() { wg.Wait(); close(results) }() // the closer

	for r := range results {
		fmt.Println(r.name, r.n)
	}
	fmt.Println("took", time.Since(start).Round(100*time.Millisecond))
}
```

Output:

```
$ go run .
bola 4
chidi 5
ada 3
took 100ms
```

(the order of the three names may differ)

- Three 100ms jobs took about 100ms in total, because they ran at the same time.
- Only the closer calls `close`, and only after `wg.Wait()`, so nobody can send on a closed channel; `range` then ends.

</details>

**Step done when:**

```bash
go build ./cmd/mystally
rm -f events.json
./mystally import --slow testdata/import
./mystally balances
./mystally list | head -4
mkdir broken && cp testdata/import/*.json broken/
echo '[' > broken/day-9.json
./mystally import broken; echo $?
echo '[{"id": 1, "creator": "ada", "type": "dance", "at": "2026-10-09T09:00:00Z"}]' > broken/day-9.json
./mystally import broken; echo $?
./mystally balances
rm -r broken
```

```
import took 202.628ms
imported 40 events
ada        ₦2.40
bola       ₦0.08
nacho      ₦4.40
ID   CREATOR    TYPE     TIME
1    nacho      like     2026-10-01 09:00:00
2    nacho      share    2026-10-01 09:05:00
3    ada        comment  2026-10-01 10:00:00
import took 2.591ms
broken/day-9.json: load events: unexpected end of JSON input
1
import took 2.104ms
broken/day-9.json: load events: unknown event type "dance"; choose one of: view, like, save, comment, share
2
ada        ₦2.40
bola       ₦0.08
nacho      ₦4.40
```

(your times will differ; `--slow` must take about 200ms now.)

#### Step 7.4: Catch races and leaks

**Achieve:**
- The timing line also says how many goroutines exist just before and just after the import, using `runtime.NumGoroutine()`. The same number twice means every goroutine the import started has finished.
- A `-race` build of mystally imports the sample files with no `WARNING: DATA RACE`.
- Once, on purpose: make the receive loop `return` on the first error, import a folder with a broken file, and watch "after" jump. Then put the loop back.

**Learn:**
- A **data race** happens when two goroutines use the same variable at the same time and at least one of them writes to it. The result can be wrong with no error at all, and only on some runs.
- The **race detector** watches every memory access while the program runs. Turn it on with `-race`: `go run -race`, `go build -race`, `go test -race`. It only reports races that actually happen during that run, so run the code you care about. A race build is slower and pauses about a second before it exits.
- The cure is not to share. Each goroutine sends its own result on the channel, and only the receiving goroutine adds things up. The Go saying is "don't communicate by sharing memory; share memory by communicating".
- A **goroutine leak** is a goroutine that never finishes, usually because it waits forever on a send or receive that nothing will match. There's no error and no crash; it just stays in memory.

<details>
<summary><b>Example: a race that a normal run hides</b></summary>

```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	total := 0
	var wg sync.WaitGroup
	for range 50 {
		wg.Go(func() { total++ }) // many goroutines write total at once
	}
	wg.Wait()
	fmt.Println(total)
}
```

Output:

```
$ go run .
50
$ go run -race -trimpath .
==================
WARNING: DATA RACE
Read at 0x00c00009c028 by goroutine 9:
  main.main.func1()
      example.com/race/main.go:12 +0x2e

Previous write at 0x00c00009c028 by goroutine 8:
...
```

(shortened; the full report shows both goroutines' stacks. Your numbers may differ.)

- The normal run looks fine. `-race` watches every memory access and reports the unsafe write.
- The cure is not to share: each goroutine sends its result, and one goroutine adds them up.

</details>

**Step done when:**

```bash
go build -race ./cmd/mystally
rm -f events.json
./mystally import --slow testdata/import
./mystally import testdata/import
./mystally balances
go build ./cmd/mystally
```

```
import took 204.518ms, goroutines before 1, after 1
imported 40 events
import took 6.815ms, goroutines before 1, after 1
imported 40 events
ada        ₦4.80
bola       ₦0.16
nacho      ₦8.80
```

Then break it on purpose, look, and put it back:

```bash
# make the receive loop return on the first error, then:
go build ./cmd/mystally
mkdir broken && cp testdata/import/*.json broken/
echo '[' > broken/day-0.json
./mystally import broken
./mystally import broken
rm -r broken
# put the loop back the way it was
go build ./cmd/mystally
```

```
import took 1.32ms, goroutines before 1, after 3
broken/day-0.json: load events: unexpected end of JSON input
import took 738µs, goroutines before 1, after 10
broken/day-0.json: load events: unexpected end of JSON input
```

(your times will differ.)

#### Step 7.5: Give up after a deadline

**Achieve:**
- `importCmd` makes a context that ends after 10 seconds and passes it to the importer instead of `context.Background()`.
- Each goroutine checks the context before it starts work. If the context has already ended, it sends `ctx.Err()` as its error and does nothing else.
- The `--slow` pause becomes a wait that also ends when the context ends, so a timeout doesn't sit out the rest of the 200ms.
- With the timeout set to one millisecond for a moment, the import fails with `context deadline exceeded`, the goroutine counts still match, and nothing is saved.

**Learn:**
- A **context** (`context.Context`) carries a "stop now" signal down through function calls. Pass it as the first parameter and call it `ctx`. Code that may take a while watches it.
- `ctx, cancel := context.WithTimeout(parent, 10*time.Second)` returns a context that ends after 10 seconds, or earlier if you call `cancel()`. Write `defer cancel()` on the next line, always: it frees the timer when the work finishes early.
- `ctx.Done()` returns a channel that is closed when the context ends. Then `ctx.Err()` says why: `context.DeadlineExceeded` (it prints `context deadline exceeded`) or `context.Canceled` (`context canceled`). Before the end, `ctx.Err()` is `nil`. Test for them with `errors.Is`, as in Stage 6.
- `select` waits on several channel operations at once and runs the case that's ready first. With a `default:` case it never waits, which makes a quick "has ctx ended yet?" check.

<details>
<summary><b>Example: give up when a deadline passes</b></summary>

```go
package main

import (
	"context"
	"fmt"
	"time"
)

func slowCount(ctx context.Context) (int, error) {
	select {
	case <-time.After(200 * time.Millisecond): // the work
		return 42, nil
	case <-ctx.Done(): // the deadline won
		return 0, ctx.Err()
	}
}

func main() {
	for _, limit := range []time.Duration{time.Second, 50 * time.Millisecond} {
		ctx, cancel := context.WithTimeout(context.Background(), limit)
		n, err := slowCount(ctx)
		cancel()
		fmt.Println(limit, n, err)
	}
}
```

Output:

```
$ go run .
1s 42 <nil>
50ms 0 context deadline exceeded
```

- `select` runs whichever case is ready first: the work, or `ctx.Done()`.
- Always call `cancel` (usually `defer cancel()`); it frees the timer as soon as you're done.

</details>

**Step done when:**

```bash
go build ./cmd/mystally
rm -f events.json
./mystally import --slow testdata/import
# edit cmd/mystally/main.go: change 10*time.Second to time.Millisecond, and save
go build ./cmd/mystally
./mystally import --slow testdata/import; echo $?
./mystally balances
# change it back to 10*time.Second, and save
go build ./cmd/mystally
./mystally import --slow testdata/import
```

```
import took 202.733ms, goroutines before 1, after 1
imported 40 events
import took 1.264ms, goroutines before 1, after 1
testdata/import/day-4.json: context deadline exceeded
1
ada        ₦2.40
bola       ₦0.08
nacho      ₦4.40
import took 203.056ms, goroutines before 1, after 1
imported 40 events
```

(your times and the file in the error will differ.)

#### Step 7.6: Limit the workers with a semaphore

**Achieve:**
- `--workers N` (default 4) sets how many files are read at the same time. There is still one goroutine per file; the others wait their turn.
- The limit is a buffered channel with N slots, used as a **semaphore**. Each goroutine waits for a free slot or for the context to end, whichever comes first, and gives its slot back when it's done.
- After a goroutine gets its slot, it checks the context once more before it starts work.
- `--workers 0` (or less) is an error, not a hang.

**Learn:**
- A **buffered channel** has room for some values: `make(chan T, n)` holds up to `n`. A send succeeds at once while there's room and waits when the buffer is full. A receive takes a value out and frees a place.
- A **semaphore** is a counter of free slots. A buffered channel makes one: send to take a slot (`sem <- struct{}{}`), receive to give it back (`<-sem`). The buffer size is how many goroutines may work at once.
- `struct{}` is a type with no fields, and `struct{}{}` is its only value. It takes no memory. Use it when only the fact of sending matters, not the value.
- Take the slot inside a `select` with a `case <-ctx.Done():`, so a goroutine that is still waiting for a slot can give up. Give the slot back with `defer func() { <-sem }()` straight after taking it.

<details>
<summary><b>Example: a buffered channel as a semaphore</b></summary>

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	for _, workers := range []int{1, 4} {
		sem := make(chan struct{}, workers) // this many free slots
		var wg sync.WaitGroup
		start := time.Now()
		for range 4 {
			wg.Go(func() {
				sem <- struct{}{}        // take a slot, or wait
				defer func() { <-sem }() // give it back
				time.Sleep(100 * time.Millisecond)
			})
		}
		wg.Wait()
		fmt.Println(workers, "workers:", time.Since(start).Round(100*time.Millisecond))
	}
}
```

Output:

```
$ go run .
1 workers: 400ms
4 workers: 100ms
```

- A send waits while the buffer is full, so at most `workers` jobs run at once.
- `struct{}{}` takes no memory: only the fact of sending matters.

</details>

**Step done when:**

```bash
go build ./cmd/mystally
rm -f events.json
./mystally import --workers 1 --slow testdata/import
./mystally import --workers 2 --slow testdata/import
./mystally import --slow testdata/import
./mystally import --workers 8 --slow testdata/import
./mystally import --workers 0 testdata/import; echo $?
./mystally import testdata/import --workers 2; echo $?
./mystally balances
rm events.json
```

```
import took 1.61015s, goroutines before 1, after 1
imported 40 events
import took 815.555ms, goroutines before 1, after 1
imported 40 events
import took 403.711ms, goroutines before 1, after 1
imported 40 events
import took 203.031ms, goroutines before 1, after 1
imported 40 events
import took 13µs, goroutines before 1, after 1
workers must be at least 1, got 0
1
usage: mystally import [--workers N] [--slow] <dir>
1
ada        ₦9.60
bola       ₦0.32
nacho      ₦17.60
```

(your times will differ; they must fall from about 1.6 seconds to about 200ms as the workers go up.)

#### Step 7.7: Cancel cleanly on Ctrl-C

**Achieve:**
- Pressing Ctrl-C during an import cancels the context instead of killing mystally. The goroutines stop, the timing line prints, the error says `context canceled`, the exit code is 1, and nothing is saved.
- The 10 second timeout still applies. It is now layered on top of the Ctrl-C context.

**Learn:**
- Ctrl-C makes the terminal send the **interrupt signal** (`os.Interrupt`, also called SIGINT) to your program. By default a Go program dies on the spot: no deferred calls, no message. The shell shows exit code 130 (128 plus signal number 2).
- Try it on your Step 7.6 build before you change anything: `timeout --preserve-status -s INT 0.5 ./mystally import --workers 1 --slow testdata/import; echo $?` prints only `130`. `timeout` sends the interrupt after half a second, and `--preserve-status` keeps the program's own exit code, so you don't have to race your finger.
- `ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)` turns the signal into a cancelled context. `defer stop()` puts the normal Ctrl-C behaviour back when the command finishes.
- Contexts nest. `context.WithTimeout(ctx, 10*time.Second)` on top of the signal context ends at the timeout or at Ctrl-C, whichever comes first. The importer neither knows nor cares which.

<details>
<summary><b>Example: Ctrl-C as a cancelled context</b></summary>

```go
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	select {
	case <-time.After(5 * time.Second):
		fmt.Println("finished")
	case <-ctx.Done():
		fmt.Println("stopped:", ctx.Err())
		fmt.Println("cause:", context.Cause(ctx))
	}
}
```

Output:

```
$ go build -o wait .
$ timeout --preserve-status -s INT 0.3 ./wait; echo $?
stopped: context canceled
cause: interrupt signal received
0
```

- `timeout -s INT 0.3` sends the same signal as Ctrl-C after 0.3 seconds. The program stops cleanly and exits 0.
- Without `NotifyContext` the program would die on the spot, skipping every `defer`.

</details>

**Step done when:**

```bash
go build ./cmd/mystally
rm -f events.json
# run the next line, then press Ctrl-C while it runs: you have about 1.6 seconds
./mystally import --workers 1 --slow testdata/import; echo $?
ls events.json
timeout --preserve-status -s INT 0.5 ./mystally import --workers 1 --slow testdata/import; echo $?
ls events.json
./mystally import --slow testdata/import
./mystally balances
rm events.json
```

```
^Cimport took 691.007ms, goroutines before 3, after 2
testdata/import/day-5.json: context canceled
1
ls: cannot access 'events.json': No such file or directory
import took 494.698ms, goroutines before 3, after 2
testdata/import/day-4.json: context canceled
1
ls: cannot access 'events.json': No such file or directory
import took 403.957ms, goroutines before 3, after 3
imported 40 events
ada        ₦2.40
bola       ₦0.08
nacho      ₦4.40
```

(your times and file names will differ.)

#### Step 7.8: Test the importer under the race detector

**Achieve:**
- A new `internal/importer/importer_test.go` in package `importer_test`, with at least these tests (the names are suggestions):

  | Test | What it proves |
  |---|---|
  | `TestImport` | Importing `../../testdata/import` with 4 workers gives 40 events, sorted by time |
  | `TestImportCancelled` | With a context that is already cancelled, the error wraps `context.Canceled` |
  | `TestImportEmptyDir` | An empty folder from `t.TempDir()` gives an error |
  | `TestImportZeroWorkers` | 0 workers gives an error |

- `go test -race ./...` passes and the race detector reports nothing.
- `go run -race ./cmd/mystally import testdata/import` prints its two normal lines and nothing else.

**Learn:**
- `t.Context()` (Go 1.24+) gives each test its own context, cancelled when the test ends. Use it in tests instead of `context.Background()`.
- For a context that's already cancelled: `ctx, cancel := context.WithCancel(t.Context())`, then call `cancel()` before the call you're testing. Check the error with `errors.Is(err, context.Canceled)`, not by its text.
- Tests run in their package's folder, so from `internal/importer` the sample files are at `../../testdata/import` (the same gotcha as Stage 5).
- `slices.IsSortedFunc(s, cmp)` reports whether a slice is already in order, using the same kind of comparison function as `slices.SortStableFunc`.

<details>
<summary><b>Example: test concurrent code under the race detector</b></summary>

`count_test.go`:

```go
package count

import (
	"context"
	"errors"
	"testing"
)

func TestCount(t *testing.T) {
	got, err := Count(t.Context(), []string{"ada", "bola"})
	if err != nil || got != 7 {
		t.Errorf("Count = %d, %v; want 7, nil", got, err)
	}
}

func TestCountCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel() // already cancelled before the call
	if _, err := Count(ctx, []string{"ada"}); !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v, want context.Canceled", err)
	}
}
```

Output:

```
$ go test -race -v .
=== RUN   TestCount
--- PASS: TestCount (0.00s)
=== RUN   TestCountCancelled
--- PASS: TestCountCancelled (0.00s)
PASS
ok  	example.com/count	1.058s
```

(your times will differ)

- `t.Context()` (Go 1.24) is cancelled when the test ends. Cancel a child context first to test the "already stopped" path.
- `-race` slows the run a little and fails it if any access is unsafe.

</details>

**Step done when:**

```bash
go test -race -count=1 ./...
go test -race -v -run TestImport ./internal/importer
rm -f events.json
go run -race ./cmd/mystally import testdata/import
go run -race ./cmd/mystally import --workers 1 --slow testdata/import
go run -race ./cmd/mystally balances
rm events.json
```

```
?   	github.com/na-cho-dev/mystally/cmd/mystally	[no test files]
ok  	github.com/na-cho-dev/mystally/internal/importer	1.056s
ok  	github.com/na-cho-dev/mystally/internal/ledger	1.046s
ok  	github.com/na-cho-dev/mystally/internal/store	1.047s
=== RUN   TestImport
--- PASS: TestImport (0.00s)
=== RUN   TestImportCancelled
--- PASS: TestImportCancelled (0.00s)
=== RUN   TestImportEmptyDir
--- PASS: TestImportEmptyDir (0.00s)
=== RUN   TestImportZeroWorkers
--- PASS: TestImportZeroWorkers (0.00s)
PASS
ok  	github.com/na-cho-dev/mystally/internal/importer	1.049s
import took 7.433ms, goroutines before 3, after 3
imported 40 events
import took 1.613756s, goroutines before 3, after 3
imported 40 events
ada        ₦4.80
bola       ₦0.16
nacho      ₦8.80
```

(your times and test names may differ. "Silent" means no `WARNING: DATA RACE` and no `FAIL` anywhere; the normal output still prints.)

### Gotchas

- Always `defer cancel()`. `go vet` warns if a cancel function can leak: `the cancel function returned by context.WithTimeout should be called, not discarded, to avoid a context leak`. And `ctx, cancel := ...` with `cancel` never used doesn't compile: `declared and not used: cancel`. You met `declared and not used` in Stage 1 for a flag you only wrote to and an `err` you never checked. *(you hit this)*
- Give `fs.Parse` the arguments **after** the command name. With `"import"` still at the front, parsing stops at once: `--slow` stays `false` and `fs.Args()` is `[import --slow testdata/import]`. It's the same index slip as Stage 1, when you removed `args = os.Args` without shifting the indexes and every command printed usage. *(you hit this)*
- Keep everything that belongs to `import` inside `importCmd`: the FlagSet, the argument count check, the context. In Stage 1, `add`-only checks placed before the `switch` made `list` panic with `index out of range [3] with length 2`. *(you hit this)*
- The new package lives in its own folder, `internal/importer/`, and its files start with `package importer`. With the file one level too high, in `internal/`, the build says `no required module provides package github.com/na-cho-dev/mystally/internal/importer; to add it:` and suggests `go get`. Don't run it: move the file. With `package main` at the top instead, it says `import "github.com/na-cho-dev/mystally/internal/importer" is a program, not an importable package`. In Stage 2 you created `event.go` in `cmd/` instead of `cmd/mystally/`, a different folder and so a different package. *(you hit this)*
- Print the real error, not a fixed message like "import failed". `context deadline exceeded`, `context canceled` and `unexpected end of JSON input` each tell you something different. In Stage 1 a fixed message hid a `permission denied`. *(you hit this)*
- `fmt.Println` doesn't format. `fmt.Println("baked in %s", d)` printed `baked in %s 20ms`, and `go vet` says `fmt.Println call has possible Printf formatting directive %s`. Use `fmt.Printf` with a `\n`. In Stage 1 the same check caught `Fprintln` with `%q`. *(you hit this)*

### Check yourself

- Who closes the results channel, and why is it safe? (One sentence.)
- What happens if `Import` returns on the first error without cancelling the context?
- What does the semaphore's buffer size mean?
- Why does merging into the ledger need no mutex in this design?
- Why does `--workers 1 --slow` take as long as the plain loop from Step 7.2?
- What do "goroutines before" and "after" tell you, and why is "after" one lower after a Ctrl-C?

### Done when

- [ ] importing eight files is visibly faster than reading them in a loop (add a timing print)

  ```bash
  go build ./cmd/mystally
  rm -f events.json
  ./mystally import --workers 1 --slow testdata/import
  ./mystally import --workers 8 --slow testdata/import
  rm events.json
  ```

  ```
  import took 1.60814s, goroutines before 3, after 3
  imported 40 events
  import took 202.624ms, goroutines before 3, after 3
  imported 40 events
  ```

    (your times will differ; `--workers 1` reads one file at a time, like the loop from Step 7.2, and must take about 1.6 seconds, `--workers 8` about 200ms)

- [ ] a one millisecond timeout aborts with `context deadline exceeded` and no goroutine leak

  ```bash
  # edit cmd/mystally/main.go: change 10*time.Second to time.Millisecond, and save
  go build ./cmd/mystally
  rm -f events.json
  ./mystally import --slow testdata/import; echo $?
  ls events.json
  # change it back to 10*time.Second, save, and rebuild
  go build ./cmd/mystally
  ```

  ```
  import took 1.518ms, goroutines before 3, after 3
  testdata/import/day-4.json: context deadline exceeded
  1
  ls: cannot access 'events.json': No such file or directory
  ```

  (your time and the file in the error will differ; before and after must be equal, and nothing may be saved.)

- [ ] `go test -race ./...` and `go run -race ./cmd/mystally import testdata/import` are silent

  ```bash
  rm -f events.json
  go test -race ./...
  go run -race ./cmd/mystally import testdata/import
  rm events.json
  ```

  ```
  ?   	github.com/na-cho-dev/mystally/cmd/mystally	[no test files]
  ok  	github.com/na-cho-dev/mystally/internal/importer	1.119s
  ok  	github.com/na-cho-dev/mystally/internal/ledger	1.099s
  ok  	github.com/na-cho-dev/mystally/internal/store	1.112s
  import took 12.622ms, goroutines before 3, after 3
  imported 40 events
  ```

  (your times will differ; no `WARNING: DATA RACE` and no `FAIL` may appear.)

- [ ] you can explain in one sentence who closes the results channel and why it is safe

  Write your sentence as the second paragraph of the Stage 7 commit message, then read it back:

  ```bash
  go vet ./...
  gofmt -l .
  git add -A
  git commit -q -m "stage 7: concurrent import" -m "<your one sentence>"
  git log -1 --format=%B
  git status --short
  ```

  ```
  stage 7: concurrent import

  <your one sentence>
  ```

  (`<your one sentence>` stands for the sentence you wrote; it must name the goroutine that closes the channel and say why no send can come after the close.)

---

## Stage 8: HTTP API

**Goal:** the same ledger behind a server, using only `net/http`.

**End goal:** when Stage 8 is finished, `./mystally serve --addr 127.0.0.1:8080` puts your ledger behind a web server. `curl localhost:8080/balances` returns every balance as JSON, `curl localhost:8080/creators/nacho/balance` returns one creator's, and `curl -X POST -d '{"creator":"nacho","type":"like"}' localhost:8080/events` saves a new event and answers `201 Created`. Ctrl-C lets a request that is still running finish before the server exits, and `go test -race ./...` tests every route with no network and no file. You will understand handlers and `ServeMux` patterns, JSON in and out of a request, middleware, why shared data needs a mutex once requests arrive together, and how a request's context travels down to the store.

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

### Structure

Stage 8 adds one package, `internal/httpapi`, and one command, `serve`. Every ledger and `Store` method that does real work gains a `ctx` first parameter, and the ledger gets a lock:

```
mystally/
  .gitignore                + /serve.log (the checks in this stage write the server's log there)
  go.mod
  cmd/mystally/
    main.go                 + a "serve" case in run, + serveCmd; the other commands pass a ctx on
  internal/httpapi/         (new package)
    server.go               NewHandler, one handler per route, writeJSON, the request and response structs
    middleware.go           logging and its statusRecorder
    server_test.go          package httpapi_test: httptest + MemoryStore, no real server, no files
  internal/importer/
    importer.go             loadFile passes its ctx on: store.NewFileStore(path).Load(ctx)
    importer_test.go        unchanged
  internal/ledger/
    store.go                Store's two methods take ctx first
    ledger.go               + mu sync.Mutex; Load, Add, BalanceFor and Balances take ctx first
    ledger_test.go          t.Context() in every call; BalanceFor and Balances now return an error too
    event.go, errors.go     unchanged
    event_test.go           unchanged
  internal/store/
    file.go, memory.go      Load(ctx) and Save(ctx, events); both may ignore ctx for now
    file_test.go            t.Context() in every call
  testdata/import/          unchanged
```

```
cmd/mystally ──► internal/httpapi ──────────────────────► internal/ledger
cmd/mystally ──► internal/importer ──► internal/store ──► internal/ledger
cmd/mystally ──► internal/store
```

The routes:

| Request | Success | Failure |
|---|---|---|
| `GET /balances` | `200`, `{"ada":20,"nacho":55}`, or `{}` when there are no events | `500`, `{"error":"internal error"}` |
| `GET /creators/{name}/balance` | `200`, `{"creator":"nacho","kobo":55,"naira":0.55}`; a creator with no events gets `0` | `500` |
| `POST /events` with `{"creator":"nacho","type":"like"}` | `201`, the creator's new balance in the same shape | `400` for broken JSON, an unknown or missing type, or an empty creator, with the message in `{"error":"..."}`; `500` if the save fails, with the real error in the log |
| any other method or path | `405` or `404`, from `ServeMux` itself | |

### Steps

Stage 8 has nine steps. Do them in order: Step 8.1 only changes signatures, Steps 8.2 to 8.5 build the server one route at a time, Step 8.6 makes the ledger safe when requests arrive together, and Steps 8.7 to 8.9 add logging, tests and a clean shutdown. Each step ends with a check you run before you move on.

#### Step 8.1: Pass a context through the ledger and the store

**Achieve:**
- `Store`'s `Load` and `Save` take a `context.Context` first, exactly as in **Structure**. `FileStore` and `MemoryStore` follow, and may ignore it.
- The ledger's `Load`, `Add`, `BalanceFor` and `Balances` take `ctx` first and hand it to the store. `BalanceFor` and `Balances` also return an error: `ctx.Err()` when the context is already done, `nil` otherwise.
- The CLI commands pass `context.Background()`. `importCmd` already has a context from Stage 7, so it passes that one to the ledger, and `loadFile` passes its own to `Load`.
- Nothing you can see changes: every command prints what it printed at the end of Stage 7. Your tests pass `t.Context()` and check the new errors.

**Learn:**
- A `context.Context` carries a "stop" signal, and maybe a deadline, down a chain of calls. You met it in Stage 7. By convention it is the **first** parameter and is named `ctx`.
- A web server gives every request its own context, `r.Context()`, which is cancelled when the client hangs up. Passing it all the way down lets slow work (Stage 9's database queries) stop early instead of finishing for nobody.
- A method that can be cancelled can fail, so it returns an `error` too. `ctx.Err()` is `nil` while the context is live, and `context.Canceled` or `context.DeadlineExceeded` once it's done.
- Unused **parameters** are fine in Go. Only unused local variables are errors (your Stage 1 `declared and not used`). So `FileStore.Load(ctx context.Context)` may ignore `ctx` today.

<details>
<summary><b>Example: a context passed all the way down</b></summary>

```go
package main

import (
	"context"
	"fmt"
)

type Store struct{}

func (Store) Load(ctx context.Context) ([]string, error) {
	if err := ctx.Err(); err != nil { // stop early if nobody is waiting
		return nil, err
	}
	return []string{"buy rice"}, nil
}

type Notes struct{ store Store }

func (n *Notes) Count(ctx context.Context) (int, error) {
	items, err := n.store.Load(ctx)
	return len(items), err
}

func main() {
	n := &Notes{}
	fmt.Println(n.Count(context.Background()))

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // like a client hanging up
	fmt.Println(n.Count(ctx))
}
```

Output:

```
$ go run .
1 <nil>
0 context canceled
```

- `ctx` is always the first parameter. A method that can be cancelled returns an `error` too.
- In a server, each request's `r.Context()` is cancelled when the client goes away.

</details>

**Step done when:**

```bash
go build ./cmd/mystally
rm -f events.json
./mystally add nacho like
./mystally add nacho share
./mystally add ada comment
./mystally balance nacho
./mystally import testdata/import
./mystally balances
go doc ./internal/ledger Ledger
go test ./...
rm events.json
```

```
nacho      ₦0.55
import took 1.759ms, goroutines before 3, after 3
imported 40 events
ada        ₦2.60
bola       ₦0.08
nacho      ₦4.95
package ledger // import "github.com/na-cho-dev/mystally/internal/ledger"

type Ledger struct {
	// Has unexported fields.
}
    Ledger holds the events and the rate each event type pays.

func NewLedger(s Store) *Ledger
func (l *Ledger) Add(ctx context.Context, e Event) error
func (l *Ledger) BalanceFor(ctx context.Context, creator string) (Kobo, error)
func (l *Ledger) Balances(ctx context.Context) (map[string]Kobo, error)
func (l *Ledger) Events() []Event
func (l *Ledger) Load(ctx context.Context) error
?   	github.com/na-cho-dev/mystally/cmd/mystally	[no test files]
ok  	github.com/na-cho-dev/mystally/internal/importer	0.010s
ok  	github.com/na-cho-dev/mystally/internal/ledger	0.009s
ok  	github.com/na-cho-dev/mystally/internal/store	0.008s
```

(your times will differ.)

#### Step 8.2: Start a server with mystally serve

**Achieve:**
- A new package, `internal/httpapi`, with `NewHandler(l *ledger.Ledger) http.Handler`. For now it returns an empty `http.NewServeMux()`, so every path answers 404.
- A new command, `mystally serve [--addr :8080]`: a `serve` case in `run`, and `serveCmd(args []string) error`. It parses `--addr` with its own `FlagSet`, as `import` does, loads the ledger, builds an `http.Server` with the four timeouts from the example, logs the address with `slog.Info` and returns whatever `ListenAndServe` returns.
- The usage text gets a `usage: mystally serve [--addr :8080]` line, and `.gitignore` gets `/serve.log`.
- If the port is taken, `serve` exits 1 with the error. Stopping cleanly comes in Step 8.9; until then, `kill %1` or Ctrl-C simply ends the program.

**Learn:**
- `net/http` is Go's web server and client, in the standard library. An `http.Server` listens on an address written `host:port`, such as `127.0.0.1:8080`, and hands each request to its `Handler`.
- `srv.ListenAndServe()` opens the port and then **blocks**: it only returns when the server fails or is shut down. An error straight away usually means the port is in use.
- An `http.Server` without timeouts waits forever for a slow client. Set them every time: `ReadHeaderTimeout` (time to send the request line and headers), `ReadTimeout` (the whole request, body included), `WriteTimeout` (time to write the response) and `IdleTimeout` (how long an unused connection may stay open for the next request).
- `http.NewServeMux()` makes a **router**: it picks the code to run from the request's method and path. With no routes, it answers `404 page not found` to everything.

<details>
<summary><b>Example: a server with an address flag and timeouts</b></summary>

```go
package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	flag.Parse()

	srv := &http.Server{
		Addr:              *addr,
		Handler:           http.NewServeMux(), // no routes yet: every path is 404
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	slog.Info("listening", "addr", srv.Addr)
	if err := srv.ListenAndServe(); err != nil { // blocks until it fails
		slog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
```

Output:

```
$ go build -o notesd .
$ ./notesd --addr 127.0.0.1:18080 & sleep 0.3; curl -s -o /dev/null -w '%{http_code}\n' 127.0.0.1:18080/notes; kill $!
2026/10/18 08:58:49 INFO listening addr=127.0.0.1:18080
404
```

(your time will differ)

- `ListenAndServe` blocks while the server runs. Always set the timeouts, or a slow client can hold a connection forever.
- An empty `ServeMux` answers `404 page not found` to everything.

</details>

**Step done when:**

```bash
go build ./cmd/mystally
./mystally serve extra; echo $?
rm -f events.json
./mystally serve --addr 127.0.0.1:8080 2>serve.log &
sleep 1
curl -i localhost:8080/balances
./mystally serve --addr 127.0.0.1:8080; echo $?
kill %1
wait
cat serve.log
git check-ignore serve.log
```

```
usage: mystally serve [--addr :8080]
1
HTTP/1.1 404 Not Found
Content-Type: text/plain; charset=utf-8
X-Content-Type-Options: nosniff
Date: Sat, 17 Oct 2026 17:38:51 GMT
Content-Length: 19

404 page not found
2026/10/17 18:38:51 INFO listening addr=127.0.0.1:8080
listen tcp 127.0.0.1:8080: bind: address already in use
1
2026/10/17 18:38:50 INFO listening addr=127.0.0.1:8080
serve.log
```

(your dates and times will differ.)

#### Step 8.3: Answer GET /balances with JSON

**Achieve:**
- `NewHandler` registers `GET /balances`. Its handler calls `Balances(r.Context())` and sends the map as JSON with status 200.
- One helper, `writeJSON`, writes every response: the `Content-Type: application/json` header, then the status, then the body.
- If `Balances` returns an error, the handler sends `500` with `{"error":"internal error"}` and returns.
- Any other method on `/balances` gets `405` from the `ServeMux`, with no code from you.

**Learn:**
- A **handler** is anything with a method `ServeHTTP(w http.ResponseWriter, r *http.Request)`: that is the `http.Handler` interface. `w` is where you write the response, `r` is the request.
- `mux.HandleFunc(pattern, f)` registers a plain function with that signature. A method value such as `s.listNotes` works too: it remembers `s`, so the handler can reach the store.
- `http.HandleFunc`, without a mux, registers on one shared default mux, `http.DefaultServeMux`:
  ```go
  http.HandleFunc("GET /notes", listNotes)
  http.ListenAndServe(addr, nil) // nil means DefaultServeMux
  ```
  This plan uses `mux.HandleFunc` on its own `ServeMux` from `http.NewServeMux()` instead, so tests don't share routes: every test gets a fresh mux.
- Since Go 1.22, a pattern can start with a method and a space: `"GET /balances"`. A `GET` pattern also answers `HEAD`, and any other method gets `405 Method Not Allowed` with an `Allow` header.
- `json.NewEncoder(w).Encode(v)` writes `v` as JSON straight into the response, followed by a newline. A `map[string]Kobo` becomes a JSON object with its keys sorted. `Kobo` has no `MarshalText`, so each amount is a plain number of kobo (your `String` method is ignored).
- Write a response in this order: headers with `w.Header().Set(...)`, then the status with `w.WriteHeader(code)`, then the body. Once the status has gone, header changes are ignored.
- Use the named constants: `http.StatusOK` (200), `http.StatusCreated` (201), `http.StatusBadRequest` (400), `http.StatusNotFound` (404), `http.StatusInternalServerError` (500).

<details>
<summary><b>Example: a GET route that answers with JSON</b></summary>

```go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
)

type server struct{ minutes map[string]int }

func (s *server) totals(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json") // headers first
	w.WriteHeader(http.StatusOK)                       // then the status
	json.NewEncoder(w).Encode(s.minutes)               // then the body
}

func main() {
	s := &server{minutes: map[string]int{"bola": 45, "ada": 60}}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /totals", s.totals)

	for _, method := range []string{"GET", "POST"} {
		rec := httptest.NewRecorder() // records the response, no network needed
		mux.ServeHTTP(rec, httptest.NewRequest(method, "/totals", nil))
		fmt.Print(method, " ", rec.Code, " ", rec.Body.String())
	}
}
```

Output:

```
$ go run .
GET 200 {"ada":60,"bola":45}
POST 405 Method Not Allowed
```

- `"GET /totals"` only matches GET (and HEAD); other methods get `405` automatically.
- `httptest` calls the handler without a network, which is how Step 8.8 tests it.

</details>

**Step done when:**

```bash
go build ./cmd/mystally
rm -f events.json
./mystally serve --addr 127.0.0.1:8080 2>serve.log &
sleep 1
curl localhost:8080/balances
kill %1
wait
./mystally add nacho like
./mystally add nacho share
./mystally add ada comment
./mystally serve --addr 127.0.0.1:8080 2>serve.log &
sleep 1
curl -i localhost:8080/balances
curl -i -X POST localhost:8080/balances
kill %1
wait
```

```
{}
HTTP/1.1 200 OK
Content-Type: application/json
Date: Sat, 17 Oct 2026 17:39:01 GMT
Content-Length: 22

{"ada":20,"nacho":55}
HTTP/1.1 405 Method Not Allowed
Allow: GET, HEAD
Content-Type: text/plain; charset=utf-8
X-Content-Type-Options: nosniff
Date: Sat, 17 Oct 2026 17:39:01 GMT
Content-Length: 19

Method Not Allowed
```

(your dates will differ.)

#### Step 8.4: Read the creator's name from the path

**Achieve:**
- `GET /creators/{name}/balance` answers with a `balanceResponse`: `{"creator":"nacho","kobo":55,"naira":0.55}`.
- The name comes from `r.PathValue("name")`. A creator with no events gets `0`, just as `./mystally balance zainab` prints `₦0.00`.
- `balanceResponse` is a struct with JSON tags whose three fields give exactly those names. `naira` comes from `Kobo.Naira()`.

**Learn:**
- `{name}` in a pattern is a **wildcard** (Go 1.22+). It matches one segment of the path, the part between two `/`. `r.PathValue("name")` returns the text it matched, and `""` for a name that isn't in the pattern.
- `/creators/nacho` is missing the `balance` segment, and `/creators/nacho/balance/extra` has a segment the pattern doesn't have, so both match nothing and get 404.
- A **response struct** fixes the shape of the JSON: the names, the types and the order. A tag such as `` `json:"kobo"` `` gives a field its lowercase JSON name. Only exported (capitalised) fields are encoded.
- Prefer a struct to `map[string]any`: the compiler checks the field names, and the shape is written down in one place (see [Things to avoid](#things-to-avoid)).

<details>
<summary><b>Example: a wildcard in the path and a response struct</b></summary>

```go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
)

type minutesResponse struct {
	Member  string `json:"member"`
	Minutes int    `json:"minutes"`
}

func main() {
	minutes := map[string]int{"ada": 60}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /members/{name}/minutes", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(minutesResponse{Member: name, Minutes: minutes[name]})
	})

	for _, path := range []string{"/members/ada/minutes", "/members/zed/minutes", "/members/ada"} {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", path, nil))
		fmt.Print(rec.Code, " ", rec.Body.String())
	}
}
```

Output:

```
$ go run .
200 {"member":"ada","minutes":60}
200 {"member":"zed","minutes":0}
404 404 page not found
```

- `{name}` matches one path segment, and `r.PathValue("name")` returns it. A path with a missing segment matches nothing: 404.
- A struct with tags fixes the JSON's names and order; the compiler checks the fields.

</details>

**Step done when:**

```bash
go build ./cmd/mystally
rm -f events.json
./mystally add nacho like
./mystally add nacho share
./mystally add ada comment
./mystally serve --addr 127.0.0.1:8080 2>serve.log &
sleep 1
curl localhost:8080/creators/nacho/balance
curl localhost:8080/creators/zainab/balance
curl localhost:8080/creators/nacho
curl localhost:8080/creators/nacho/balance/extra
kill %1
wait
```

```
{"creator":"nacho","kobo":55,"naira":0.55}
{"creator":"zainab","kobo":0,"naira":0}
404 page not found
404 page not found
```

(`zainab` has no events, so `kobo` must be `0`; the last two must be `404 page not found`.)

#### Step 8.5: Accept new events with POST /events

**Achieve:**
- `POST /events` decodes `{"creator":"nacho","type":"like"}` into a small request struct with only those two fields. The type field is a `ledger.EventType`, so your `UnmarshalText` from Stage 2 parses `"like"` and rejects `"dance"` while decoding.
- The handler builds an `Event` with the current time and calls `Add(r.Context(), e)`. On success it answers `201` with the creator's new balance, in the `balanceResponse` shape.
- A body that won't decode is a `400` with the decoder's message. A `*ValidationError` from `Add` is a `400` with its message. Any other error from `Add` is a `500` with `{"error":"internal error"}`, and the real error goes to the log with `slog.Error`.
- `Add` now also rejects an event whose type isn't one of the five, with a `*ValidationError` for the type field. A body without `"type"` leaves the field at its zero value, and that must never be saved (see **Gotchas** for what happens if it is). If your `EventType` still starts with `View` at 0, add `Unknown` as the first constant now. `events.json` stores type names, not numbers, so files you already have still load. Run the Step 8.1 commands again afterwards.
- `Add` saves the longer list first and keeps it only if the save worked. A failed save must not show up in `GET /balances`.

**Learn:**
- `json.NewDecoder(r.Body).Decode(&in)` reads the request body into `in`. Pass a pointer, as with `json.Unmarshal`. The server closes `r.Body` for you.
- Decode into a struct with only the fields a client may send. Fields the client sends but the struct doesn't have (an `id`, say) are silently ignored, so nobody can pick their own ID.
- If a field's type has `UnmarshalText`, the decoder calls it for that field, and its error comes back from `Decode` unchanged, so `errors.Is` still finds your sentinel.
- Who made the mistake decides the status. The client's mistakes (broken JSON, a bad value) are `4xx`, with a message that helps them fix it. Your mistakes (a file that won't save) are `5xx`, with a plain message for the client and the real error in your log.

<details>
<summary><b>Example: decode a POST, then 201 or 400</b></summary>

```go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
)

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func createNote(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{err.Error()})
		return // without this, the code below runs too
	}
	if in.Text == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"text must not be empty"})
		return
	}
	writeJSON(w, http.StatusCreated, in)
}

func main() {
	for _, body := range []string{`{"text":"buy rice"}`, `{"text":""}`, `{bad`} {
		rec := httptest.NewRecorder()
		createNote(rec, httptest.NewRequest("POST", "/notes", strings.NewReader(body)))
		fmt.Print(rec.Code, " ", rec.Body.String())
	}
}
```

Output:

```
$ go run .
201 {"text":"buy rice"}
400 {"error":"text must not be empty"}
400 {"error":"invalid character 'b' looking for beginning of object key string"}
```

- Decode into a struct with only the fields a client may send; extra fields (an `id`) are ignored.
- Client mistakes are `400` with a message that helps them fix it. Always `return` after writing an error.

</details>

**Step done when:**

```bash
go build ./cmd/mystally
rm -f events.json
./mystally serve --addr 127.0.0.1:8080 2>serve.log &
sleep 1
curl -i -X POST -d '{"creator":"nacho","type":"like"}' localhost:8080/events
curl -w '%{http_code}\n' -X POST -d '{"creator":"nacho","type":"dance"}' localhost:8080/events
curl -w '%{http_code}\n' -X POST -d '{"creator":"","type":"like"}' localhost:8080/events
curl -w '%{http_code}\n' -X POST -d '{"creator":"ada"}' localhost:8080/events
curl -w '%{http_code}\n' -X POST -d '{bad' localhost:8080/events
chmod a-w events.json
curl -w '%{http_code}\n' -X POST -d '{"creator":"ada","type":"comment"}' localhost:8080/events
chmod u+w events.json
curl localhost:8080/balances
kill %1
wait
cat serve.log
./mystally list
```

```
HTTP/1.1 201 Created
Content-Type: application/json
Date: Sat, 17 Oct 2026 15:50:15 GMT
Content-Length: 42

{"creator":"nacho","kobo":5,"naira":0.05}
{"error":"unknown event type \"dance\"; choose one of: view, like, save, comment, share"}
400
{"error":"invalid creator: must not be empty"}
400
{"error":"invalid type: must be one of: view, like, save, comment, share"}
400
{"error":"invalid character 'b' looking for beginning of object key string"}
400
{"error":"internal error"}
500
{"nacho":5}
2026/10/17 16:50:14 INFO listening addr=127.0.0.1:8080
2026/10/17 16:50:16 ERROR add event err="save events: open events.json: permission denied"
ID   CREATOR    TYPE     TIME
1    nacho      like     2026-10-17 16:50:15
```

(your dates and times will differ, and your wording may differ.)

#### Step 8.6: Lock the ledger

**Achieve:**
- `Ledger` gets a field `mu sync.Mutex`. `Load`, `Add`, `BalanceFor`, `Balances` and `Events` each lock it before they touch `events`, and unlock it with `defer`.
- `Add` holds the lock from choosing the ID until the save is done, so two requests can't pick the same ID or save over each other.
- `Events` returns a copy (`slices.Clone`), made while it holds the lock, so nobody reads the slice after the lock is released.
- No ledger method calls another locking method while it holds the lock (see **Gotchas**).

**Learn:**
- The server starts a new goroutine for each connection, so requests from different clients run **at the same time**, and they all share your one `*ledger.Ledger`.
- Two `Add`s at once are a **data race** (see Stage 7, Step 7.4): both can read the same highest ID and both append, so events go missing or IDs repeat.
- A `sync.Mutex` ("mutual exclusion") lets one goroutine through at a time. `mu.Lock()` waits until the mutex is free and takes it, and `mu.Unlock()` gives it back. `defer mu.Unlock()` right after the `Lock` unlocks on every return path.
- The zero value of `sync.Mutex` is an unlocked mutex, so no constructor is needed. Put it next to the fields it guards and say so in a comment.

<details>
<summary><b>Example: a mutex around shared data</b></summary>

```go
package main

import (
	"fmt"
	"sync"
)

type Notes struct {
	mu    sync.Mutex // guards items
	items []string
}

func (n *Notes) Add(s string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.items = append(n.items, s)
}

func (n *Notes) Len() int {
	n.mu.Lock() // reads need the lock too
	defer n.mu.Unlock()
	return len(n.items)
}

func main() {
	var n Notes
	var wg sync.WaitGroup
	for range 100 {
		wg.Go(func() { n.Add("x") }) // 100 "requests" at once
	}
	wg.Wait()
	fmt.Println(n.Len())
}
```

Output:

```
$ go run -race .
100
```

- Every request runs in its own goroutine, so shared data needs a lock. The race detector stays silent here.
- The zero `sync.Mutex` is ready to use. Put it next to the fields it guards.

</details>

**Step done when:**

```bash
go build -race ./cmd/mystally
rm -f events.json
./mystally serve --addr 127.0.0.1:8080 2>serve.log &
sleep 1
seq 50 | xargs -P 50 -I{} curl -s -o /dev/null -X POST -d '{"creator":"nacho","type":"like"}' localhost:8080/events
curl localhost:8080/creators/nacho/balance
kill %1
wait
grep -c 'WARNING: DATA RACE' serve.log
./mystally list | tail -n 2
```

```
{"creator":"nacho","kobo":250,"naira":2.5}
0
49   nacho      like     2026-10-17 16:50:19
50   nacho      like     2026-10-17 16:50:19
```

(your times will differ.)

#### Step 8.7: Log every request with middleware

**Achieve:**
- `internal/httpapi/middleware.go` holds `logging(next http.Handler) http.Handler` and a small `statusRecorder`.
- `NewHandler` returns `logging(mux)`, so every request, 404s and 405s included, is logged once with its method, path, status and duration.
- No handler changes.

**Learn:**
- `http.HandlerFunc` is a **function type** with a method: `type HandlerFunc func(ResponseWriter, *Request)`, and its `ServeHTTP` just calls the function. So `http.HandlerFunc(f)` turns any function with the right signature into an `http.Handler`. That is the "function-type trick", and it is what `HandleFunc` does for you behind the scenes.
- **Middleware** is a function that takes a handler and returns a new one: `func(next http.Handler) http.Handler`. The new handler does something, calls `next.ServeHTTP(w, r)`, then does something else. The result is again a `Handler`, so middleware can wrap a whole mux, one route, or other middleware.
- A `ResponseWriter` doesn't tell you which status was written. To log it, wrap `w` in your own type that **embeds** `http.ResponseWriter`, so it has every method for free, and override only `WriteHeader` to remember the code.
- Start the remembered status at `200`. A handler that only calls `Write` never calls your `WriteHeader`; the real writer sends 200 by itself.

<details>
<summary><b>Example: logging middleware that records the status</b></summary>

```go
package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"time"
)

type statusRecorder struct {
	http.ResponseWriter // embedded: every method for free
	status              int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		slog.Info("request", "method", r.Method, "path", r.URL.Path,
			"status", rec.status, "duration", time.Since(start))
	})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /hello", func(w http.ResponseWriter, r *http.Request) { fmt.Fprintln(w, "hi") })
	h := logging(mux)
	for _, path := range []string{"/hello", "/nope"} {
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", path, nil))
	}
}
```

Output:

```
$ go run .
2026/10/18 08:55:47 INFO request method=GET path=/hello status=200 duration=46.664µs
2026/10/18 08:55:47 INFO request method=GET path=/nope status=404 duration=19.953µs
```

(your times and durations will differ)

- Middleware takes a handler and returns one that does something before and after calling it.
- The recorder starts at 200, because a handler that only writes a body never calls `WriteHeader`.

</details>

**Step done when:**

```bash
go build ./cmd/mystally
rm -f events.json
./mystally serve --addr 127.0.0.1:8080 2>serve.log &
sleep 1
curl -s -o /dev/null -X POST -d '{"creator":"nacho","type":"like"}' localhost:8080/events
curl -s -o /dev/null -X POST -d '{"creator":"nacho","type":"dance"}' localhost:8080/events
curl -s -o /dev/null localhost:8080/balances
curl -s -o /dev/null localhost:8080/creators/nacho/balance
curl -s -o /dev/null localhost:8080/nope
curl -s -o /dev/null -X DELETE localhost:8080/balances
kill %1
wait
cat serve.log
```

```
2026/10/17 16:50:22 INFO listening addr=127.0.0.1:8080
2026/10/17 16:50:23 INFO request method=POST path=/events status=201 duration=618.79µs
2026/10/17 16:50:23 INFO request method=POST path=/events status=400 duration=339.331µs
2026/10/17 16:50:23 INFO request method=GET path=/balances status=200 duration=226.617µs
2026/10/17 16:50:24 INFO request method=GET path=/creators/nacho/balance status=200 duration=157.142µs
2026/10/17 16:50:24 INFO request method=GET path=/nope status=404 duration=25.12µs
2026/10/17 16:50:24 INFO request method=DELETE path=/balances status=405 duration=64.705µs
```

(your times and durations will differ; there must be one `request` line per request, with the statuses 201, 400, 200, 200, 404 and 405.)

#### Step 8.8: Test every handler without a network

**Achieve:**
- `internal/httpapi/server_test.go`, package `httpapi_test`, tests every route through `NewHandler`, on a ledger built from `store.NewMemoryStore()`.
- At least: `GET /balances` after a few posts, `GET /creators/{name}/balance`, a good `POST /events` (201), and a table of bad bodies that must each get 400 and leave the balances empty.
- One test sends 50 `POST`s from 50 goroutines and checks the balance is 250. Run it with `-race`.
- No test opens a port or touches `events.json`.

**Learn:**
- A handler is a method you can call. `httptest.NewRequest(method, path, body)` builds a request in memory, `httptest.NewRecorder()` is a `ResponseWriter` that keeps everything written to it, and `h.ServeHTTP(rec, req)` runs the handler. No network at all.
- Afterwards, `rec.Code` is the status and `rec.Body` the body. Decode the body with `json.NewDecoder(rec.Body).Decode(&got)` and compare the fields.
- Test through `NewHandler`, not through the handler methods. Then the patterns, the methods in them and the middleware are tested too.
- Package `httpapi_test` is Stage 5's external test package: it sees only what other packages see. It may import `store` for `MemoryStore`, and `httpapi` itself still doesn't.

<details>
<summary><b>Example: handler tests without a network</b></summary>

`notes_test.go`:

```go
package notes_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"example.com/notes"
)

func TestCreateThenList(t *testing.T) {
	h := notes.NewHandler()

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("POST", "/notes", strings.NewReader(`{"text":"buy rice"}`)))
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST status = %d, want %d", rec.Code, http.StatusCreated)
	}

	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/notes", nil))
	var got []string
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != "buy rice" {
		t.Errorf("GET /notes = %v, want [buy rice]", got)
	}
}
```

Output:

```
$ go test -race -v .
=== RUN   TestCreateThenList
--- PASS: TestCreateThenList (0.00s)
PASS
ok  	example.com/notes	1.048s
```

(`notes.go` holds the handlers. Your time will differ.)

- `httptest.NewRequest` builds a request in memory and `httptest.NewRecorder` keeps the response: no port, no server.
- Test through `NewHandler`, so the routes and methods are tested too.

</details>

**Step done when:**

```bash
go test -race -v ./internal/httpapi
```

```
=== RUN   TestBalances
--- PASS: TestBalances (0.00s)
=== RUN   TestPostEvent
--- PASS: TestPostEvent (0.00s)
=== RUN   TestCreatorBalance
--- PASS: TestCreatorBalance (0.00s)
=== RUN   TestPostEventBad
=== RUN   TestPostEventBad/not_json
=== RUN   TestPostEventBad/empty_body
=== RUN   TestPostEventBad/unknown_type
=== RUN   TestPostEventBad/no_type
=== RUN   TestPostEventBad/empty_creator
--- PASS: TestPostEventBad (0.00s)
    --- PASS: TestPostEventBad/not_json (0.00s)
    --- PASS: TestPostEventBad/empty_body (0.00s)
    --- PASS: TestPostEventBad/unknown_type (0.00s)
    --- PASS: TestPostEventBad/no_type (0.00s)
    --- PASS: TestPostEventBad/empty_creator (0.00s)
=== RUN   TestPostEventsAtOnce
--- PASS: TestPostEventsAtOnce (0.01s)
PASS
ok  	github.com/na-cho-dev/mystally/internal/httpapi	1.051s
```

Then prove the concurrency test can fail: comment out the two lines in `Add` that lock and unlock `mu`, run the command below, and put them back.

```bash
go test -race ./internal/httpapi | tail -n 7
```

```
--- FAIL: TestPostEventsAtOnce (0.02s)
    server_test.go:131: body = {"creator":"nacho","kobo":100,"naira":1}
        , want kobo 250
    testing.go:1712: race detected during execution of test
FAIL
FAIL	github.com/na-cho-dev/mystally/internal/httpapi	0.066s
FAIL
```

(your test names, times and amounts will differ; every test must `PASS` with the lock, and the concurrency test must `FAIL` without it.)

#### Step 8.9: Shut down gracefully on Ctrl-C

**Achieve:**
- `serveCmd` builds its context with `signal.NotifyContext(context.Background(), os.Interrupt)`, as `importCmd` does, and runs `ListenAndServe` in a goroutine that sends its error on a channel.
- It waits for whichever comes first: that error (the server couldn't start, so return it) or Ctrl-C.
- On Ctrl-C it logs that it is shutting down, calls `srv.Shutdown` with a **new** 10 second context, treats `http.ErrServerClosed` as the normal ending, logs `stopped`, and returns `nil`: exit code 0.
- A request that is still running when you press Ctrl-C gets its answer, and new connections are refused.

**Learn:**
- `srv.Shutdown(ctx)` first closes the listener, so new connections are refused, then closes idle connections, then waits for active requests to finish. If `ctx` ends first, it gives up and returns `ctx.Err()`.
- When `Shutdown` starts, `ListenAndServe` returns `http.ErrServerClosed` straight away. That is the normal ending, not a failure: check for it with `errors.Is` and don't report it.
- `ListenAndServe` blocks, so it runs in its own goroutine, and a channel with room for one value, `make(chan error, 1)`, carries its result back. The buffer lets that goroutine finish even if nobody is receiving any more.
- `select` waits for whichever channel is ready first (Stage 7): the server's error or `ctx.Done()`.

<details>
<summary><b>Example: finish the request in flight, then stop</b></summary>

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /slow", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		fmt.Fprintln(w, "slow request done")
	})
	srv := &http.Server{Addr: "127.0.0.1:18081", Handler: mux, ReadHeaderTimeout: 5 * time.Second}

	errc := make(chan error, 1)
	go func() { errc <- srv.ListenAndServe() }()

	select {
	case err := <-errc: // could not start
		slog.Error("server failed", "err", err)
		os.Exit(1)
	case <-ctx.Done(): // Ctrl-C
	}

	slog.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // ctx is already cancelled
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown", "err", err)
	}
	if err := <-errc; !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server", "err", err)
	}
	slog.Info("stopped")
}
```

Output:

```
$ go build -o graceful .
$ ./graceful & sleep 0.3; curl -s 127.0.0.1:18081/slow & sleep 0.1; kill -INT %1; wait
2026/10/18 08:55:55 INFO shutting down
slow request done
2026/10/18 08:55:55 INFO stopped
```

(your times will differ)

- `Shutdown` stops accepting new connections and waits for running requests: the slow request still finished.
- The signal context is already cancelled by then, so `Shutdown` gets a fresh one with its own timeout.

</details>

**Step done when:**

```bash
go build ./cmd/mystally
rm -f events.json
./mystally serve --addr 127.0.0.1:8080 2>serve.log &
sleep 1
(sleep 2; echo '{"creator":"nacho","type":"like"}') | curl -sS -w '%{http_code}\n' -X POST -T - localhost:8080/events &
sleep 1
kill -INT %1
curl localhost:8080/balances
wait %1; echo $?
wait
cat serve.log
./mystally list
```

```
curl: (7) Failed to connect to localhost port 8080 after 0 ms: Couldn't connect to server
{"creator":"nacho","kobo":5,"naira":0.05}
201
0
2026/10/17 18:39:10 INFO listening addr=127.0.0.1:8080
2026/10/17 18:39:12 INFO shutting down, finishing in-flight requests
2026/10/17 18:39:13 INFO request method=POST path=/events status=201 duration=1.987101528s
2026/10/17 18:39:13 INFO stopped
ID   CREATOR    TYPE     TIME
1    nacho      like     2026-10-17 18:39:13
```

(your times will differ, and the curl output and the `0` may swap places.)

### Gotchas

- Forgetting `return` after writing an error response means the handler carries on and writes again. The log shows `http: superfluous response.WriteHeader call`, and the client gets the **first** status. The full log line names the handler, for example `http: superfluous response.WriteHeader call from main.getRecipe (main.go:17)`.
- Order inside a handler: `w.Header().Set(...)` → `w.WriteHeader(status)` → write the body. Headers set after `WriteHeader` are silently ignored. A handler that set `Content-Type: application/json` too late sent `Content-Type: text/plain; charset=utf-8`.
- Patterns with a method (`"GET /notes"`) give `405 Method Not Allowed` for other methods automatically.
- A pattern needs a space between the method and the path. `"GET/balances"` doesn't fail: Go reads `GET` as a host name, and every request gets `404 page not found`.
- An `err` that is never read doesn't compile: `declared and not used: err`. You met it in Stage 1. *(you hit this)* But the compiler only catches a new variable. In a handler with two calls, a second `x, err := ...` reuses the first `err` once that one has been checked, so leaving the second unchecked compiles and `go vet` stays quiet. Every call that returns an error needs its own `if err != nil { writeJSON(...); return }`.
- For a 500, the client gets `internal error`, but the log must get the real `err`. In Stage 1 a fixed message hid a `permission denied`; here the `slog.Error` line is the only place the real cause shows up. *(you hit this)*

### Check yourself

- Why does the ledger need a mutex now, when it didn't in stage 7?
- What does the middleware signature `func(http.Handler) http.Handler` let you do?
- What happens to a request that is half-finished when you press Ctrl-C?
- Why can handler tests run without opening a network port?
- Which `POST /events` failures are the client's fault (400), and which are yours (500)? Why does the client see the real message for one but only `internal error` for the other?
- Why must `Add` save before it keeps the new event, now that the ledger lives as long as the server?

### Done when

- [ ] `curl localhost:8080/balances` returns JSON

  ```bash
  go build ./cmd/mystally
  rm -f events.json
  ./mystally add nacho like
  ./mystally add nacho share
  ./mystally add ada comment
  ./mystally serve --addr 127.0.0.1:8080 2>serve.log &
  sleep 1
  curl -i localhost:8080/balances
  kill -INT %1
  wait
  ```

  ```
  HTTP/1.1 200 OK
  Content-Type: application/json
  Date: Sat, 17 Oct 2026 15:50:51 GMT
  Content-Length: 22

  {"ada":20,"nacho":55}
  ```

    (your date will differ; the status must be `200 OK`, the `Content-Type` must be `application/json`, and the body must be `{"ada":20,"nacho":55}`.)

- [ ] `curl -X POST -d '{"creator":"nacho","type":"like"}' localhost:8080/events` adds an event

  ```bash
  go build ./cmd/mystally
  rm -f events.json
  ./mystally serve --addr 127.0.0.1:8080 2>serve.log &
  sleep 1
  curl -X POST -d '{"creator":"nacho","type":"like"}' localhost:8080/events
  curl localhost:8080/creators/nacho/balance
  kill -INT %1
  wait
  ./mystally list
  cat events.json
  ```

  ```
  {"creator":"nacho","kobo":5,"naira":0.05}
  {"creator":"nacho","kobo":5,"naira":0.05}
  ID   CREATOR    TYPE     TIME
  1    nacho      like     2026-10-17 16:50:53
  [
    {
      "id": 1,
      "creator": "nacho",
      "type": "like",
      "at": "2026-10-17T16:50:53.747902796+01:00"
    }
  ]
  ```

    (your timestamps will differ; the event must be in `events.json`, and `list` must show it after the server has stopped.)

- [ ] each handler has at least one test and they run without a real server or file

  ```bash
  rm -f events.json
  go test -count=1 -v ./internal/httpapi
  grep -ohE 'httptest\.NewRecorder|store\.NewMemoryStore|ListenAndServe|httptest\.NewServer|NewFileStore' internal/httpapi/server_test.go | sort -u
  ls events.json internal/httpapi/events.json
  ```

  ```
  === RUN   TestBalances
  --- PASS: TestBalances (0.00s)
  === RUN   TestPostEvent
  --- PASS: TestPostEvent (0.00s)
  === RUN   TestCreatorBalance
  --- PASS: TestCreatorBalance (0.00s)
  === RUN   TestPostEventBad
  === RUN   TestPostEventBad/not_json
  === RUN   TestPostEventBad/empty_body
  === RUN   TestPostEventBad/unknown_type
  === RUN   TestPostEventBad/no_type
  === RUN   TestPostEventBad/empty_creator
  --- PASS: TestPostEventBad (0.00s)
      --- PASS: TestPostEventBad/not_json (0.00s)
      --- PASS: TestPostEventBad/empty_body (0.00s)
      --- PASS: TestPostEventBad/unknown_type (0.00s)
      --- PASS: TestPostEventBad/no_type (0.00s)
      --- PASS: TestPostEventBad/empty_creator (0.00s)
  === RUN   TestPostEventsAtOnce
  --- PASS: TestPostEventsAtOnce (0.00s)
  PASS
  ok  	github.com/na-cho-dev/mystally/internal/httpapi	0.015s
  httptest.NewRecorder
  store.NewMemoryStore
  ls: cannot access 'events.json': No such file or directory
  ls: cannot access 'internal/httpapi/events.json': No such file or directory
  ```

  (your test names will differ; there must be at least one test for each of the three routes, and all must pass.)

- [ ] Ctrl-C finishes in-flight requests then exits

  ```bash
  go build ./cmd/mystally
  rm -f events.json
  ./mystally serve --addr 127.0.0.1:8080 2>serve.log &
  sleep 1
  (sleep 2; echo '{"creator":"nacho","type":"like"}') | curl -sS -w '%{http_code}\n' -X POST -T - localhost:8080/events &
  sleep 1
  kill -INT %1
  curl localhost:8080/balances
  wait %1; echo $?
  wait
  cat serve.log
  ./mystally list
  ```

  ```
  curl: (7) Failed to connect to localhost port 8080 after 0 ms: Couldn't connect to server
  {"creator":"nacho","kobo":5,"naira":0.05}
  201
  0
  2026/10/17 18:39:21 INFO listening addr=127.0.0.1:8080
  2026/10/17 18:39:23 INFO shutting down, finishing in-flight requests
  2026/10/17 18:39:24 INFO request method=POST path=/events status=201 duration=1.982353498s
  2026/10/17 18:39:24 INFO stopped
  ID   CREATOR    TYPE     TIME
  1    nacho      like     2026-10-17 18:39:24
  ```

  (your times will differ, and the curl output and the `0` may swap places.)

---

## Stage 9: Postgres, generics, a binary

**Goal:** the first third-party package, and a store swap that touches nothing else.

**End goal:** with the Docker container running and `MYSTALLY_DATABASE_URL` exported, `./mystally add`, `list`, `balances`, `import` and `serve` read and write Postgres instead of `events.json`, and `MYSTALLY_DATABASE_URL="$TEST_DB" go test -count=1 ./...` runs the same store tests against Postgres that plain `go test ./...` runs against memory and a file. `git diff --stat stage-8 postgres -- internal/ledger internal/httpapi cmd` shows that `cmd/mystally/main.go` is the only file in those folders that changed to get there, and `CGO_ENABLED=0 go build -o mystally ./cmd/mystally` makes one file that runs in an `alpine` container with no Go inside. You will understand what `go.mod` and `go.sum` record, what a connection pool and a transaction are, how type parameters and a constraint let one function add up both `Kobo` and `int`, and why a static binary is the one to copy elsewhere.

**Build**
- `PostgresStore` in `internal/store` using `github.com/jackc/pgx/v5`, selected when `MYSTALLY_DATABASE_URL` is set
- Schema in a plain `schema.sql` applied by hand with `psql`
- The same store tests as stage 5 run against Postgres when the env var is set, and are skipped otherwise (`t.Skip`)
- One generic helper that is genuinely useful here, for example `func Map[T, U any](xs []T, f func(T) U) []U` or `SumBy`
- `go build -o mystally ./cmd/mystally` and run the binary on another machine or in a container
- `go doc` your own module and fix every exported name that lacks a doc comment

**Learn:** `go get`, `go.sum`, reading a third-party package on pkg.go.dev, connection pools, `t.Skip`, type parameters and constraints, cross-compiling with `GOOS` and `GOARCH`.

**Look in:** pkg.go.dev for pgx (check importers and last release), `go help build`.

### Structure

Stage 9 adds one store, four test files, a schema and one generic helper. For Postgres, the only file outside `internal/store` that changes is `cmd/mystally/main.go`:

```
mystally/
  .gitignore                + /dist/ (the cross-compiled binaries)                          (Step 9.9)
  go.mod                    + require github.com/jackc/pgx/v5, and the modules it needs     (Step 9.1)
  go.sum                    (new) a checksum for every module in the build; committed        (Step 9.1)
  schema.sql                (new) CREATE TABLE events, applied by hand with psql             (Step 9.2)
  cmd/mystally/
    main.go                 + openStore: Postgres or events.json, from MYSTALLY_DATABASE_URL (Step 9.3)
                            + a doc comment above "package main"                             (Step 9.8)
  internal/ledger/
    sum.go                  (new) the Number constraint and SumBy                            (Step 9.7)
    sum_test.go             (new) TestSumBy, with Kobo and with int                          (Step 9.7)
    ledger.go               BalanceFor adds up with SumBy                                    (Step 9.7)
    every .go file          doc comments only                                                (Step 9.8)
  internal/store/
    postgres.go             (new) PostgresStore and its methods                              (Steps 9.3 to 9.5)
    store_test.go           (new) testStore: the shared checks, written once                 (Step 9.6)
    memory_test.go          (new) TestMemoryStore calls testStore                            (Step 9.6)
    file_test.go            TestFileStore calls testStore; TestFileStoreSaveMissingFolder stays (Step 9.6)
    postgres_test.go        (new) TestPostgresStore: skip, or empty the table and call testStore (Step 9.6)
  internal/httpapi/         doc comments only                                                (Step 9.8)
  internal/importer/        doc comments only                                                (Step 9.8)
  dist/                     (new, ignored by git) mystally-linux-arm64, mystally.exe         (Step 9.9)
```

```
cmd/mystally ──► internal/store ──► internal/ledger
                     │
                     └──► github.com/jackc/pgx/v5/pgxpool (and github.com/jackc/pgx/v5)
internal/ledger, internal/httpapi, internal/importer: no pgx import, no Postgres code
```

### Steps

Stage 9 has nine steps, done in order: Steps 9.1 and 9.2 get the tools ready (the pgx module and a running database), Steps 9.3 to 9.5 build `PostgresStore` and switch mystally to it, Step 9.6 runs one set of store tests against all three stores and proves nothing else changed, and Steps 9.7 to 9.9 add a generic helper, doc comments and a binary that runs anywhere. Each step ends with a check you run before you move on. Outputs that need the database were worked out from the code, not captured from a run, and depend on your container and the rows in it: read them as what you should see, not as exact text.

#### Step 9.1: Add pgx to the module

**Achieve:**
- Before you change anything, tag the Stage 8 commit `stage-8`. Step 9.6 compares against it.
- Read pgx's page on pkg.go.dev, as the Look in line says, before you depend on it.
- Start `internal/store/postgres.go` (`package store`) with only the `PostgresStore` struct and its one field, `pool *pgxpool.Pool`, plus a doc comment. The build fails until the module knows about pgx.
- Make it build with `go get github.com/jackc/pgx/v5` and then `go mod tidy`. The first time, `go get` prints these lines (your version numbers may be newer), and `go mod tidy` prints nothing:

  ```
  go: added github.com/jackc/pgpassfile v1.0.0
  go: added github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761
  go: added github.com/jackc/pgx/v5 v5.11.0
  go: added golang.org/x/text v0.29.0
  ```

  `go.mod` changes and `go.sum` appears; both go into the Stage 9 commit.

**Learn:**
- A **third-party package** is code from outside the standard library. Its import path says where it lives: `github.com/jackc/pgx/v5/pgxpool`. From version 2 on, the major version is part of the path (`/v5`).
- pgx is one **module**, `github.com/jackc/pgx/v5`, holding several packages: `pgx`, `pgxpool`, `pgtype` and more. `go get` takes a module or a package path and records the module in `go.mod`; `import` always names a package.
- `go get github.com/jackc/pgx/v5` downloads the latest release into your module cache (`~/go/pkg/mod`) and adds a `require` line to `go.mod`.
- `go.sum` holds a checksum for every module your build uses. Every later download is checked against it, so nobody can swap the code under you. Commit it, never edit it. `go mod verify` re-checks the cache and prints `all modules verified`.

<details>
<summary><b>Example: add a third-party module</b></summary>

```go
package main

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	_, err := pgxpool.ParseConfig("postgres://localhost/parcels")
	fmt.Println(err)
}
```

Output:

```
$ go get github.com/jackc/pgx/v5@v5.11.0
go: added github.com/jackc/pgpassfile v1.0.0
go: added github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761
go: added github.com/jackc/pgx/v5 v5.11.0
go: added golang.org/x/text v0.29.0
$ go mod tidy
$ cat go.mod
module example.com/parcels

go 1.26.2

require github.com/jackc/pgx/v5 v5.11.0

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/text v0.29.0 // indirect
)
$ go run .
<nil>
```

- `go get` downloads the module into your module cache and records it in `go.mod`; `go.sum` holds checksums. Commit both.
- `go mod tidy` adds what your imports need and removes the rest. `// indirect` marks modules your dependencies need.

</details>

**Step done when:**

```bash
git tag --list stage-8
go build ./... && go vet ./...
head -5 go.mod
git status --short
```

```
stage-8
module github.com/na-cho-dev/mystally

go 1.26.2

require github.com/jackc/pgx/v5 v5.11.0
 M go.mod
?? go.sum
?? internal/store/postgres.go
```

(your version numbers may be newer.)

#### Step 9.2: Start Postgres and create the table

**Achieve:**
- `schema.sql` at the repo root with one `CREATE TABLE events` statement and four columns: `id BIGINT PRIMARY KEY`, `creator TEXT NOT NULL`, `type TEXT NOT NULL`, `at TIMESTAMPTZ NOT NULL`.
- A Postgres 17 server in a Docker container named `mystally-pg`, on port 5433 of your machine.
- The table in two databases: `postgres`, the default one, for running mystally; and `mystally_test`, which the tests in Step 9.6 empty on every run.
- `MYSTALLY_DATABASE_URL` exported for the first database, and a shell variable `TEST_DB` for the second.

**Learn:**
- A **table** has named, typed **columns**. Each **row** is one record: here, one event. `CREATE TABLE` makes the table; it starts empty.
- Column types: `BIGINT` is a 64-bit integer (`int64` in Go, and also `int` on your machine), `TEXT` a string of any length, `TIMESTAMPTZ` an instant in time (`time.Time`), kept to the microsecond.
- `PRIMARY KEY` makes `id` unique: Postgres refuses a second row with the same id. Step 9.5's `ON CONFLICT (id)` relies on that. `NOT NULL` refuses a row with that column missing.
- **Docker** runs a program in a **container**, a small isolated system made from an **image** (`postgres:17`). `--name` names the container, `-e` sets an environment variable inside it (the image reads `POSTGRES_PASSWORD`, and `POSTGRES_USER` and `POSTGRES_DB` if you give them), `-p 5433:5432` connects port 5433 on your machine to port 5432 inside, and `-d` runs it in the background. `docker stop mystally-pg` and `docker start mystally-pg` stop and restart it; the data stays.

  ```bash
  docker run --name mystally-pg -e POSTGRES_PASSWORD=dev -p 5433:5432 -d postgres:17
  export MYSTALLY_DATABASE_URL='postgres://postgres:dev@localhost:5433/postgres?sslmode=disable'
  TEST_DB='postgres://postgres:dev@localhost:5433/mystally_test?sslmode=disable'
  until pg_isready -q -d "$MYSTALLY_DATABASE_URL"; do sleep 1; done
  psql "$MYSTALLY_DATABASE_URL" -f schema.sql
  psql "$MYSTALLY_DATABASE_URL" -c 'CREATE DATABASE mystally_test'
  psql "$TEST_DB" -f schema.sql
  ```

  The first `docker run` downloads the `postgres:17` image, which takes a minute, then prints the new container's long ID. The `until` line asks `pg_isready` once a second until the server accepts connections. The three `psql` lines print `CREATE TABLE`, `CREATE DATABASE` and `CREATE TABLE`. Port 5432, Postgres's usual port, is already taken on your machine: a PostgreSQL 16 server, installed alongside `psql`, runs there as a system service. That's why the container uses 5433.
- **Every new terminal, from now on.** The checks in the rest of this stage need the database. The container and the table stay, but the variables live in one terminal, so in each new one run:

  ```bash
  docker start mystally-pg
  export MYSTALLY_DATABASE_URL='postgres://postgres:dev@localhost:5433/postgres?sslmode=disable'
  TEST_DB='postgres://postgres:dev@localhost:5433/mystally_test?sslmode=disable'
  ```

<details>
<summary><b>Example: a table and a throwaway Postgres</b></summary>

`schema.sql`:

```sql
CREATE TABLE parcels (
    id   BIGINT PRIMARY KEY,
    dest TEXT NOT NULL,
    sent TIMESTAMPTZ NOT NULL
);
```

In a terminal:

```bash
docker run --name parcels-pg -e POSTGRES_PASSWORD=dev -p 5432:5432 -d postgres:17
export PARCELS_DATABASE_URL='postgres://postgres:dev@localhost:5432/postgres?sslmode=disable'
psql "$PARCELS_DATABASE_URL" -f schema.sql
psql "$PARCELS_DATABASE_URL" -c '\d parcels'
```

(not run here: it needs Docker. `psql` prints `CREATE TABLE`, then the three columns.)

- `PRIMARY KEY` makes `id` unique; `NOT NULL` refuses a missing value.
- The connection URL reads `postgres://USER:PASSWORD@HOST:PORT/DATABASE`.

</details>

**Step done when:**

```bash
pg_isready -d "$MYSTALLY_DATABASE_URL"
psql "$MYSTALLY_DATABASE_URL" -Atc 'SELECT count(*) FROM events'
psql "$TEST_DB" -Atc 'SELECT count(*) FROM events'
git status --short schema.sql
```

```
localhost:5433 - accepting connections
0
0
?? schema.sql
```

(both databases have an empty `events` table.)

#### Step 9.3: Choose the store from MYSTALLY_DATABASE_URL

**Achieve:**
- `NewPostgresStore(ctx, url)` opens a pool with `pgxpool.New` and checks it with `Ping`. If `Ping` fails, it closes the pool and returns the error, wrapped with a short prefix of your choice. `Close()` closes the pool.
- `Load` and `Save` exist but, for now, only return an error that says they aren't written yet. That's enough for `*PostgresStore` to satisfy `ledger.Store`: add the compile-time check next to the type.
- In `main.go`, `openStore(ctx)` returns the store and a function that closes it: Postgres when `MYSTALLY_DATABASE_URL` is not empty, otherwise the `FileStore` and a function that does nothing. Every command that loads the ledger (`serve` too) uses it and defers the close function.
- Nothing in `internal/ledger` or `internal/httpapi` changes.

**Learn:**
- `os.Getenv("NAME")` returns a `string`, and `""` when the variable isn't set: never `nil`. `os.LookupEnv` also tells you whether it was set at all, but here empty and unset mean the same thing: use the file.
- `NAME=value ./mystally list` sets a variable for that one command. `NAME= ./mystally list` sets it to empty for one command, which here means "use `events.json`". `unset NAME` removes an exported variable.
- `pgxpool.New(ctx, url)` reads the URL and makes a pool, but connects to nothing yet. A stopped server or a wrong password would only show up at the first query. `pool.Ping(ctx)` makes the first connection now.
- A **connection pool** keeps a few connections open and lends one to each query, so a query doesn't pay for a new connection every time. It is safe for concurrent use: `serve`'s goroutines can all share one pool. Its default size is the larger of 4 and your CPU count.

<details>
<summary><b>Example: an environment variable picks the store</b></summary>

```go
package main

import (
	"fmt"
	"os"
)

type Store interface{ Name() string }

type fileStore struct{}

func (fileStore) Name() string { return "file" }

type dbStore struct{ url string }

func (s dbStore) Name() string { return "postgres at " + s.url }

func openStore() Store {
	if url := os.Getenv("SHELF_DATABASE_URL"); url != "" {
		return dbStore{url: url}
	}
	return fileStore{}
}

func main() {
	fmt.Println("using", openStore().Name())
}
```

Output:

```
$ go build -o pick .
$ ./pick
using file
$ SHELF_DATABASE_URL=postgres://localhost/shelf ./pick
using postgres at postgres://localhost/shelf
$ SHELF_DATABASE_URL= ./pick
using file
```

- `os.Getenv` returns `""` when the variable is missing, so unset and empty mean the same thing.
- `NAME=value ./prog` sets a variable for that one command only.

</details>

**Step done when:**

```bash
go build ./cmd/mystally
rm -f events.json
MYSTALLY_DATABASE_URL= ./mystally add nacho like
MYSTALLY_DATABASE_URL= ./mystally balances
MYSTALLY_DATABASE_URL='postgres://postgres:dev@localhost:5499/postgres?sslmode=disable' ./mystally balances; echo $?
./mystally balances; echo $?
```

```
nacho      ₦0.05
connect to postgres: failed to connect to `user=postgres database=postgres`: 127.0.0.1:5499 (localhost): dial error: dial tcp 127.0.0.1:5499: connect: connection refused
1
load events: not written yet
1
```

(your wording may differ.)

#### Step 9.4: Load events with a query

**Achieve:**
- `Load` runs one `SELECT` of the four columns, `ORDER BY id`, and turns the rows into `[]ledger.Event` with `pgx.CollectRows`.
- The `type` column is `TEXT`: scan it into a `string`, then `ParseEventType`. A row with a type mystally doesn't know makes `Load` fail, and the error is wrapped as `load events: %w` like `FileStore`'s.
- An empty table gives no events and no error, like a missing `events.json`.
- With the variable set, `list`, `balances` and `serve` now read Postgres. `add` still fails, because `Save` is still a stub.

**Learn:**
- `pool.Query(ctx, sql)` sends a query and returns `pgx.Rows`, which must be closed when you're done. `pgx.CollectRows(rows, fn)` does the loop for you: it calls `fn` once per row, collects the results in a slice, closes `rows` and returns any error.
- `CollectRows` is a generic function, `CollectRows[T any](rows Rows, fn RowToFunc[T]) ([]T, error)`: Go works out `T` from what your `fn` returns. Step 9.7 shows how that works.
- Inside `fn`, `row.Scan(&a, &b, &c, &d)` copies the columns, in `SELECT` order, into the variables you point at. `BIGINT` fits an `int`, `TEXT` a `string`, `TIMESTAMPTZ` a `time.Time`.
- pgx can't scan `TEXT` straight into a named integer type like `EventType`: it doesn't know your names. Scan into a `string`, then parse it.

<details>
<summary><b>Example: read rows with <code>Query</code> and <code>CollectRows</code></b></summary>

`load.go`:

```go
package parcels

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Parcel struct {
	ID   int
	Dest string
}

// Load returns every parcel, in ID order.
func Load(ctx context.Context, pool *pgxpool.Pool) ([]Parcel, error) {
	rows, err := pool.Query(ctx, `SELECT id, dest FROM parcels ORDER BY id`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (Parcel, error) {
		var p Parcel
		err := row.Scan(&p.ID, &p.Dest) // columns in SELECT order
		return p, err
	})
}
```

(compiled and vetted here; running it needs the database from Step 9.2)

- `CollectRows` loops over the rows for you, calls your function once per row, and closes `rows`.
- `Scan` copies each column into the variable you point at, in `SELECT` order.

</details>

**Step done when:**

Put two events in the table by hand, then read them through mystally:

```bash
go build ./cmd/mystally
psql "$MYSTALLY_DATABASE_URL" -c "INSERT INTO events VALUES (1, 'nacho', 'like', '2026-10-17 09:00:00+01'), (2, 'ada', 'share', '2026-10-17 09:30:00+01')"
./mystally list
./mystally balances
./mystally add bola comment; echo $?
```

```
INSERT 0 2
ID   CREATOR    TYPE     TIME
1    nacho      like     2026-10-17 09:00:00
2    ada        share    2026-10-17 09:30:00
ada        ₦0.50
nacho      ₦0.05
save events: not written yet
1
```

(your wording for the last error may differ.)

#### Step 9.5: Save events in a transaction

**Achieve:**
- `Save` begins a transaction, runs one `INSERT ... ON CONFLICT (id) DO NOTHING` per event with `$1` to `$4` placeholders and the type as `e.Type.String()`, and commits. A `defer tx.Rollback(ctx)` right after `Begin` undoes everything on any early return.
- `Ledger.Add` passes every event each time, so the rows already in the table are skipped and only the new one is written.
- `add`, `import` and `POST /events` now write to Postgres, and `events.json` is not touched at all.

**Learn:**
- A **transaction** groups statements so that either all of them happen or none do. `tx, err := pool.Begin(ctx)` starts one, statements run on `tx`, and `tx.Commit(ctx)` makes them all permanent at once.
- `defer tx.Rollback(ctx)` right after `Begin` is the safety net: if any `return` comes before `Commit`, every row written so far is undone. After a `Commit`, `Rollback` does nothing (it returns `pgx.ErrTxClosed`, "tx is closed", which you ignore).
- `$1, $2, $3, $4` are **placeholders**. pgx sends the SQL and the values separately, so a value can never become SQL. Building the statement with `fmt.Sprintf` would let a creator named `x'); DROP TABLE events; --` run its own SQL. That attack is called **SQL injection**.
- `ON CONFLICT (id) DO NOTHING`: when a row with that `id` already exists, skip this one instead of failing. It only works because `id` is the `PRIMARY KEY`.

<details>
<summary><b>Example: save a batch in one transaction</b></summary>

`save.go`:

```go
package parcels

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Parcel struct {
	ID   int
	Dest string
}

// Save writes parcels that aren't stored yet. All of them, or none.
func Save(ctx context.Context, pool *pgxpool.Pool, parcels []Parcel) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) // undoes everything unless Commit ran

	for _, p := range parcels {
		_, err := tx.Exec(ctx,
			`INSERT INTO parcels (id, dest) VALUES ($1, $2) ON CONFLICT (id) DO NOTHING`,
			p.ID, p.Dest) // values go separately: never build SQL with Sprintf
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
```

(compiled and vetted here; running it needs the database from Step 9.2)

- `$1, $2` are placeholders: the values travel separately, so a value can never become SQL.
- `ON CONFLICT (id) DO NOTHING` skips rows that already exist; it works because `id` is the primary key.

</details>

**Step done when:**

```bash
go build ./cmd/mystally
rm -f events.json
./mystally add bola comment
./mystally list
./mystally import testdata/import | tail -1
./mystally balances
psql "$MYSTALLY_DATABASE_URL" -Atc 'SELECT count(*) FROM events'
ls events.json
```

```
ID   CREATOR    TYPE     TIME
1    nacho      like     2026-10-17 09:00:00
2    ada        share    2026-10-17 09:30:00
3    bola       comment  2026-10-17 17:16:42
imported 40 events
ada        ₦2.90
bola       ₦0.28
nacho      ₦4.45
43
ls: cannot access 'events.json': No such file or directory
```

(your timestamp for event 3 will differ.)

#### Step 9.6: Run one set of store tests against every store

**Achieve:**
- `internal/store/store_test.go` (`package store`) holds `testStore(t, s ledger.Store)`, with the three checks from Structure. It begins with `t.Helper()` and builds its events with `time.Date`.
- `TestMemoryStore` and `TestFileStore` each build their store and call `testStore`. `TestFileStore` replaces Stage 5's round-trip and missing-file tests; keep `TestFileStoreSaveMissingFolder` from Stage 6.
- `TestPostgresStore` skips when `MYSTALLY_DATABASE_URL` is empty, and skips when the database's name doesn't end in `_test` (`pgxpool.ParseConfig(url)`, then `cfg.ConnConfig.Database`, as in Step 9.1's example). Otherwise it connects, registers `t.Cleanup(s.Close)`, empties the table with `TRUNCATE events` through `s.pool`, and calls `testStore`.
- The tests pass against `mystally_test`. Then commit, tag the commit `postgres`, and check what changed since `stage-8`.

**Learn:**
- `t.Skip("reason")` stops a test and marks it skipped: not a failure. `go test` still prints `ok`. With `-v` you see `--- SKIP` and the reason. `t.Skipf` takes a format, like `t.Errorf`.
- `t.Cleanup(f)` runs `f` when the test finishes, like a `defer` for the whole test. `t.Cleanup(s.Close)` closes the pool even if `testStore` stops the test with `Fatal`.
- A **contract test** is one helper that takes the interface type, so any implementation fits: memory, file and Postgres all go through the same lines.
- `postgres_test.go` must be `package store` to reach the unexported `s.pool`. `TRUNCATE events` deletes every row at once. Empty the table at the **start** of the test, not the end, so a failed run leaves its rows for you to look at with `psql`.

<details>
<summary><b>Example: one contract test for every implementation</b></summary>

`shelf_test.go`:

```go
package shelf

import (
	"os"
	"testing"
)

func testStore(t *testing.T, s Store) {
	t.Helper()
	if err := s.Save([]string{"Dune"}); err != nil {
		t.Fatal(err)
	}
	got, err := s.Load()
	if err != nil || len(got) != 1 || got[0] != "Dune" {
		t.Errorf("Load = %v, %v; want [Dune], nil", got, err)
	}
}

func TestMemory(t *testing.T) { testStore(t, &Memory{}) }

func TestPostgres(t *testing.T) {
	if os.Getenv("SHELF_DATABASE_URL") == "" {
		t.Skip("SHELF_DATABASE_URL not set")
	}
	// open the Postgres store here, then: testStore(t, store)
}
```

Output:

```
$ go test -v .
=== RUN   TestMemory
--- PASS: TestMemory (0.00s)
=== RUN   TestPostgres
    shelf_test.go:23: SHELF_DATABASE_URL not set
--- SKIP: TestPostgres (0.00s)
PASS
ok  	example.com/shelf	0.009s
```

(your time will differ)

- One helper takes the interface, so memory, file and Postgres stores all pass through the same checks.
- `t.Skip` marks a test skipped, not failed: `go test` still says `ok` without a database.

</details>

**Step done when:**

```bash
go test ./...
go test -v -run 'Store$' ./internal/store
MYSTALLY_DATABASE_URL="$TEST_DB" go test -count=1 -v -run 'Store$' ./internal/store
psql "$TEST_DB" -Atc 'SELECT id, creator, type FROM events ORDER BY id'
git add -A
git commit -q -m "stage 9: postgres store"
git tag postgres
git diff --stat stage-8 postgres -- internal/ledger internal/httpapi cmd
```

```
?   	github.com/na-cho-dev/mystally/cmd/mystally	[no test files]
ok  	github.com/na-cho-dev/mystally/internal/httpapi	0.043s
ok  	github.com/na-cho-dev/mystally/internal/importer	0.013s
ok  	github.com/na-cho-dev/mystally/internal/ledger	0.012s
ok  	github.com/na-cho-dev/mystally/internal/store	0.012s
=== RUN   TestFileStore
--- PASS: TestFileStore (0.00s)
=== RUN   TestMemoryStore
--- PASS: TestMemoryStore (0.00s)
=== RUN   TestPostgresStore
    postgres_test.go:21: database "postgres" is not a test database (name must end in _test); skipping
--- SKIP: TestPostgresStore (0.00s)
PASS
ok  	github.com/na-cho-dev/mystally/internal/store	0.012s
=== RUN   TestFileStore
--- PASS: TestFileStore (0.00s)
=== RUN   TestMemoryStore
--- PASS: TestMemoryStore (0.00s)
=== RUN   TestPostgresStore
--- PASS: TestPostgresStore (0.05s)
PASS
ok  	github.com/na-cho-dev/mystally/internal/store	0.071s
1|nacho|like
2|ada|share
3|bola|comment
 cmd/mystally/main.go | 45 +++++++++++++++++++++++++++++++++++----------
 1 file changed, 35 insertions(+), 10 deletions(-)
```

(your times, line numbers, skip wording and the rows your `testStore` saved will differ.)

#### Step 9.7: Write one generic helper

**Achieve:**
- `internal/ledger/sum.go`: a constraint `Number` that allows `Kobo`, `int` and `float64`, and `SumBy[T any, N Number](xs []T, f func(T) N) N` with a doc comment starting `// SumBy ...`.
- `BalanceFor` adds up with `SumBy` instead of its own loop: the function you pass returns the rate for the creator's events and 0 for everyone else's. `Balances` keeps its map loop.
- `internal/ledger/sum_test.go`: `TestSumBy` calls `SumBy` once with a function that returns `Kobo` and once with one that returns `int`. Two real types are what make it worth being generic.

**Learn:**
- A **type parameter** is a placeholder for a type, written in square brackets after the function's name: `func SumBy[T any, N Number](...)`. The caller's types fill it in. Usually Go **infers** them from the arguments, so you call `SumBy(events, f)` with no brackets.
- A **constraint** says which types a type parameter accepts. `any` accepts every type. `Number` is an interface listing types: `~int | ~int64 | ~float64`. The function body may only do what every listed type can do, and here that's `+`.
- `~int64` means "`int64`, or any named type whose underlying type is `int64`". `Kobo` is declared as `int64`, so `~int64` lets it in. Plain `int64` would not.
- `var total N` is the zero value of whatever `N` turns out to be: `0` as a `Kobo` or as an `int`. The result has the caller's type, so `SumBy` over `Kobo` gives back a `Kobo` that still prints as `₦0.55`.

<details>
<summary><b>Example: <code>SumBy</code> and <code>Map</code> with a named type</b></summary>

```go
package main

import "fmt"

// Number is a constraint: the types SumBy accepts.
type Number interface {
	~int | ~int64 | ~float64
}

func SumBy[T any, N Number](xs []T, f func(T) N) N {
	var total N
	for _, x := range xs {
		total += f(x)
	}
	return total
}

func Map[T, U any](xs []T, f func(T) U) []U {
	out := make([]U, 0, len(xs))
	for _, x := range xs {
		out = append(out, f(x))
	}
	return out
}

type Minutes int64 // a named type: allowed by ~int64

type Task struct {
	Name string
	Took Minutes
}

func main() {
	tasks := []Task{{"write", 30}, {"test", 45}}
	fmt.Println(SumBy(tasks, func(t Task) Minutes { return t.Took }))
	fmt.Println(Map(tasks, func(t Task) string { return t.Name }))
}
```

Output:

```
$ go run .
75
[write test]
```

- Go infers `T` and `N` from the arguments, and the result keeps the caller's type (`Minutes`).
- Without the `~`, `int64` alone would reject `Minutes`, and it would reject your `Kobo` too.

</details>

**Step done when:**

```bash
go test -v -run SumBy ./internal/ledger
go test ./...
go doc ./internal/ledger SumBy
go build ./cmd/mystally
rm -f events.json
MYSTALLY_DATABASE_URL= ./mystally add nacho like
MYSTALLY_DATABASE_URL= ./mystally add nacho share
MYSTALLY_DATABASE_URL= ./mystally add ada comment
MYSTALLY_DATABASE_URL= ./mystally balance nacho
MYSTALLY_DATABASE_URL= ./mystally balance ada
rm events.json
```

```
=== RUN   TestSumBy
--- PASS: TestSumBy (0.00s)
PASS
ok  	github.com/na-cho-dev/mystally/internal/ledger	0.007s
?   	github.com/na-cho-dev/mystally/cmd/mystally	[no test files]
ok  	github.com/na-cho-dev/mystally/internal/httpapi	0.017s
ok  	github.com/na-cho-dev/mystally/internal/importer	0.015s
ok  	github.com/na-cho-dev/mystally/internal/ledger	0.017s
ok  	github.com/na-cho-dev/mystally/internal/store	0.019s
package ledger // import "github.com/na-cho-dev/mystally/internal/ledger"

func SumBy[T any, N Number](xs []T, f func(T) N) N
    SumBy adds up f(x) for every x in xs.
nacho      ₦0.55
ada        ₦0.20
```

(your times and doc sentence will differ.)

#### Step 9.8: Document every exported name

**Achieve:**
- Every package has a package comment: `// Package ledger ...` above `package ledger` in one of its files, and the same for `store`, `importer` and `httpapi`. `main.go` gets one above `package main` that starts with the program's name, `Mystally`.
- Every exported type, function, method, constant group and variable has a doc comment that starts with its own name.
- `go doc` shows a sentence under every exported name in the module.

**Learn:**
- A doc comment (Stage 2) is the `//` comment directly above a declaration, with no blank line between them. `go doc`, pkg.go.dev and the editor's hover all show it.
- Start it with the name and write full sentences: `// Load reads every event, in ID order.` A comment that starts with the name still reads well when it's shown on its own.
- One file per package carries the package comment: `// Package store ...`. For a command, the comment describes the program and starts with its name: `// Mystally records ...`.
- `go doc ./internal/store` lists a package's exported names with its package comment. `go doc -all ./internal/store` prints every declaration with its comment indented under it: a declaration with nothing under it has no doc comment. `go doc ./internal/store FileStore.Load` shows one method.

<details>
<summary><b>Example: spot a missing doc comment</b></summary>

`recipe.go`:

```go
// Package recipe scales recipes.
package recipe

// Recipe is a dish and how many it serves.
type Recipe struct {
	Name   string
	Serves int
}

// Double returns the recipe for twice as many people.
func (r Recipe) Double() Recipe { r.Serves *= 2; return r }

func (r Recipe) Half() Recipe { r.Serves /= 2; return r }
```

Output:

```
$ go doc -all .
package recipe // import "example.com/recipe"

Package recipe scales recipes.

TYPES

type Recipe struct {
	Name   string
	Serves int
}
    Recipe is a dish and how many it serves.

func (r Recipe) Double() Recipe
    Double returns the recipe for twice as many people.

func (r Recipe) Half() Recipe
```

- Every documented name has its comment indented under it. `Half` has nothing: that's the one to fix.
- Start a doc comment with the name and write a full sentence: `// Half returns ...`.

</details>

**Step done when:**

```bash
go doc ./cmd/mystally
go doc ./internal/store
go doc -all ./internal/store
```

```
Mystally records creators' engagement events and reports what each creator has
earned. Events go to events.json, or to Postgres when MYSTALLY_DATABASE_URL is
set.
package store // import "github.com/na-cho-dev/mystally/internal/store"

Package store loads and saves ledger events: in memory, in a JSON file, or in
Postgres.

type FileStore struct{ ... }
    func NewFileStore(path string) *FileStore
type MemoryStore struct{ ... }
    func NewMemoryStore() *MemoryStore
type PostgresStore struct{ ... }
    func NewPostgresStore(ctx context.Context, url string) (*PostgresStore, error)
package store // import "github.com/na-cho-dev/mystally/internal/store"

Package store loads and saves ledger events: in memory, in a JSON file, or in
Postgres.

TYPES

type FileStore struct {
	// Has unexported fields.
}
    FileStore keeps events in one JSON file.

func NewFileStore(path string) *FileStore
    NewFileStore returns a FileStore for the file at path. The file need not
    exist yet.

func (f *FileStore) Load(ctx context.Context) ([]ledger.Event, error)
    Load reads every event from the file. A missing file means no events.

func (f *FileStore) Save(ctx context.Context, events []ledger.Event) error
    Save replaces the file's contents with events.

type MemoryStore struct {
	// Has unexported fields.
}
    MemoryStore keeps events in a slice. Tests use it.

func NewMemoryStore() *MemoryStore
    NewMemoryStore returns an empty MemoryStore.

func (m *MemoryStore) Load(ctx context.Context) ([]ledger.Event, error)
    Load returns a copy of the stored events.

func (m *MemoryStore) Save(ctx context.Context, events []ledger.Event) error
    Save keeps a copy of events, replacing what was there.

type PostgresStore struct {
	// Has unexported fields.
}
    PostgresStore keeps events in the events table (see schema.sql).

func NewPostgresStore(ctx context.Context, url string) (*PostgresStore, error)
    NewPostgresStore opens a connection pool to url and checks that it works.

func (s *PostgresStore) Close()
    Close closes every connection in the pool.

func (s *PostgresStore) Load(ctx context.Context) ([]ledger.Event, error)
    Load reads every event, in ID order.

func (s *PostgresStore) Save(ctx context.Context, events []ledger.Event) error
    Save inserts the events that are not stored yet, in one transaction.

```

(your sentences and receiver names will differ; every declaration must have one under it, starting with its own name.)

#### Step 9.9: Build one binary that runs anywhere

**Achieve:**
- `CGO_ENABLED=0 go build -o mystally ./cmd/mystally` makes a **statically linked** binary, and `file` says so.
- Two cross-compiled binaries in `dist/`: one for Linux on ARM (`linux/arm64`, a Raspberry Pi or an ARM server) and one for Windows (`windows/amd64`). Add `/dist/` to `.gitignore`.
- The static binary runs in an `alpine` container, which has no Go inside (the third "Done when" box).
- Commit the rest of Stage 9.

**Learn:**
- **cgo** lets Go code call C. A normal build on your machine uses it in the `net` package (which `net/http` and pgx import) to look up host names through the system's C library, so the binary is **dynamically linked**: it needs the system's C library (`libc.so.6`) when it starts. `CGO_ENABLED=0` switches cgo off and Go's own code takes over. The result is **statically linked**: one file that needs nothing else.
- `file mystally` says which kind you built. `ldd mystally` lists the libraries a binary needs, or says `not a dynamic executable`. `go version -m mystally` shows the settings it was built with, `CGO_ENABLED` among them.
- `GOOS` (the operating system) and `GOARCH` (the processor) choose what to build for. `go tool dist list` prints every pair Go supports. A cross build switches cgo off by itself; writing `CGO_ENABLED=0` anyway makes it obvious.
- The first build for a new `GOOS`/`GOARCH` takes a minute or more: Go compiles the standard library for that target once, then reuses it.

<details>
<summary><b>Example: dynamic, static and cross-compiled builds</b></summary>

```go
package main

import (
	"fmt"
	"net/http"
)

func main() {
	fmt.Println(http.StatusText(http.StatusOK))
}
```

Output:

```
$ go build -o app . && file -b app | grep -o '[a-z]*ally linked'
dynamically linked
$ CGO_ENABLED=0 go build -o app . && file -b app | grep -o '[a-z]*ally linked'
statically linked
$ GOOS=linux GOARCH=arm64 go build -o app-arm64 . && file -b app-arm64 | cut -d, -f2
 ARM aarch64
```

- A normal build that uses `net` links to the system's C library. `CGO_ENABLED=0` gives a static binary that runs in any Linux container.
- `GOOS` and `GOARCH` choose the target; a cross build is static by default.

</details>

**Step done when:**

```bash
CGO_ENABLED=0 go build -o mystally ./cmd/mystally
file mystally
ldd mystally
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o dist/mystally-linux-arm64 ./cmd/mystally
GOOS=windows GOARCH=amd64 go build -o dist/mystally.exe ./cmd/mystally
file dist/*
git check-ignore dist/mystally.exe
```

```
mystally: ELF 64-bit LSB executable, x86-64, version 1 (SYSV), statically linked, Go BuildID=L-iOixXaAQBXBfSwmqw2/b2QYMOP6_P55Otto5CEb/ZQp7Jf-kf55ksN9qFYTr/s_5vRi8OJjHoEdJM_YiQ, BuildID[sha1]=c79d765988906049582c0c5fdf6534948ae1a7b7, with debug_info, not stripped
	not a dynamic executable
dist/mystally.exe:         PE32+ executable (console) x86-64, for MS Windows, 16 sections
dist/mystally-linux-arm64: ELF 64-bit LSB executable, ARM aarch64, version 1 (SYSV), statically linked, Go BuildID=JdBK_N94SHf9IRviG-nx/ZyH9_G-9LOlRDBBLMkWc/pcjVqLYQAOkTP0IYNA4T/3sHOXg59DaRiiV1NLp0t, BuildID[sha1]=5a0221e6be0bc4c611921ae720d7aafe3383bd25, with debug_info, not stripped
dist/mystally.exe
```

(your BuildIDs will differ.)

### Gotchas

- After `go get`, the editor may keep underlining the pgx import for a while. That's gopls running behind, like the `could not import os` you saw in Stage 0: trust `go build ./...`, and restart the language server if it lasts. *(you hit this)*
- `tag, err := tx.Exec(...)` with `tag` never used doesn't compile: `declared and not used: tag`. Write `_, err := tx.Exec(...)`. In Stage 1 you met `declared and not used` for a flag you only wrote to and an `err` you never checked. *(you hit this)*
- `os.Getenv` returns a `string`. `url == nil` gives `invalid operation: url == nil (mismatched types string and untyped nil)`, as in Stage 1. Compare with `""`. *(you hit this)*
- Print the real error. `connection refused` (container stopped), `password authentication failed` (wrong password or port), `database "mystally_tset" does not exist` (a typo) and `relation "events" does not exist` (schema not applied) need four different fixes. In Stage 1, a fixed message hid a `permission denied` from you. *(you hit this)*
- `postgres_test.go` must say `package store` to reach `s.pool` for the `TRUNCATE`. As `package store_test`, vet stops with `s.pool undefined (cannot refer to unexported field pool)`. The package line, like the folder (your `event.go` in `cmd/` in Stage 2), decides which names a file can see. *(you hit this)*
- `go build -o mystally` with no package path at the repo root fails with `no Go files in /home/nacho/projects/personal/mystally`. Name the folder: `./cmd/mystally`. *(you hit this)*

### Check yourself

- What is `go.sum` for, and why is it committed?
- Why is the `Store` interface the only thing that had to know Postgres exists?
- What does `~int64` mean in a constraint?
- Why does a static binary matter for a container?
- Why does `Save` run in a transaction, and what does the deferred `Rollback` do after `Commit` has run?
- Why does `./mystally list` show the imported events an hour later with Postgres than with `events.json`, and why is that not a bug?

### Done when

- [ ] `MYSTALLY_DATABASE_URL=... go test ./...` passes against a real database

  ```bash
  docker start mystally-pg
  MYSTALLY_DATABASE_URL="$TEST_DB" go test -count=1 ./...
  MYSTALLY_DATABASE_URL="$TEST_DB" go test -count=1 -v -run Postgres ./internal/store
  ```

  ```
  mystally-pg
  ?   	github.com/na-cho-dev/mystally/cmd/mystally	[no test files]
  ok  	github.com/na-cho-dev/mystally/internal/httpapi	0.027s
  ok  	github.com/na-cho-dev/mystally/internal/importer	0.019s
  ok  	github.com/na-cho-dev/mystally/internal/ledger	0.012s
  ok  	github.com/na-cho-dev/mystally/internal/store	0.084s
  === RUN   TestPostgresStore
  --- PASS: TestPostgresStore (0.05s)
  PASS
  ok  	github.com/na-cho-dev/mystally/internal/store	0.071s
  ```

  (your times will differ.)

- [ ] nothing in `internal/ledger`, `cmd`, or `internal/httpapi` changed to add Postgres

  ```bash
  git diff --stat stage-8 postgres -- internal/ledger internal/httpapi cmd
  git diff --stat stage-8 postgres -- internal/ledger internal/httpapi
  ```

  ```
   cmd/mystally/main.go | 45 +++++++++++++++++++++++++++++++++++----------
   1 file changed, 35 insertions(+), 10 deletions(-)
  ```

  (your numbers will differ.)

- [ ] a single binary runs with no Go toolchain present

  ```bash
  CGO_ENABLED=0 go build -o mystally ./cmd/mystally
  rm -f events.json
  MYSTALLY_DATABASE_URL= ./mystally add nacho like
  MYSTALLY_DATABASE_URL= ./mystally add ada share
  docker run --rm -v "$PWD:/work" -w /work alpine ./mystally balances
  docker run --rm alpine sh -c 'command -v go || echo "no go in this container"'
  rm events.json
  ```

  ```
  ada        ₦0.50
  nacho      ₦0.05
  no go in this container
  ```

  (the first `docker run` also prints Docker's lines about downloading the `alpine` image.)

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
| 10 | 1 | Put `%q` in a `Fprintln` call | vet: `fmt.Fprintln call has possible Printf formatting directive %q` | Only the `f` functions (`Printf`, `Fprintf`, `Sprintf`, `Errorf`) fill in verbs |
| 11 | 1 | Passed a map to `%s` | vet: `format %s has arg events of wrong type map[string]int` | `%s` wants a string. Build one first (`slices.Sorted(maps.Keys(m))` + `strings.Join`) |
| 12 | 1 | Opened the file with `os.O_APPEND` only | `no such file or directory`, then `bad file descriptor` | Combine an access mode and behaviours: `O_WRONLY\|O_APPEND\|O_CREATE` |
| 13 | 1 | Printed a hard-coded reason instead of `err` | said "no such file" when the real cause was `permission denied` | Print `err` itself; it already says what went wrong |
| 14 | 1 | Called `file.Name()` after `OpenFile` failed | `panic: runtime error: invalid memory address or nil pointer dereference` | On error the `*os.File` is `nil`; everything is in `err` |
| 15 | 1 | Wrote a list as `[a, b, c]` | `unexpected comma; expecting ]` | A slice literal is type plus braces: `[]string{a, b, c}` |
| 16 | 1 | Passed `fmt.Fprintln`'s results into `WriteString` | `too many arguments ... have (int, error) want (string)` | `Fprintln` already writes; use one or the other |
| 17 | 1 | Wrote `*args []string` | `expected type argument list` | `*` goes on the type, and slices rarely need a pointer |
| 18 | 1 | Used `=` for a variable that didn't exist yet | `undefined: content` | `:=` declares, `=` only assigns |
| 19 | 1 | Wrote `x != "add" \|\| x != "list"` | vet: `suspect or` (always true) | "Neither A nor B" is `x != A && x != B` |
| 20 | 1 | Returned `nil` from a function returning `string` | `cannot use nil as string value in return statement` | A string's empty value is `""` |
| 21 | 1 | Swapped the returns inside `if err != nil` | errors swallowed, content lost | Inside `if err != nil`, return the error |
| 22 | 1 | Ran an add-only check before the `switch` | `panic: runtime error: index out of range [3] with length 2` on `list` | A command's checks go inside its own `case` |
| 23 | 1 | Wrote `byte(",")` | `cannot convert "," (untyped string constant) to type byte` | Double quotes make a string, single quotes a character: `','` |
| 24 | 1 | Ended an error message with `\n` | an extra blank line | Whoever prints an error adds the newline |
| 25 | 1 | Removed `args = os.Args` without shifting the indexes, then committed | every command printed usage | `os.Args[1:]` moves every index down by one. Run the checklist before each commit |
| 26 | 2 | Created `event.go` in `cmd/` instead of `cmd/mystally/` | `main.go` couldn't see `Kobo` | A different folder is a different package |
| 27 | 2 | Wrote `String()` that printed, then one with no return type | `fmt` ignored it / `too many return values: have (string) want ()` | The exact signature is `String() string`, and it returns the text |
| 28 | 2 | Formatted the value with `%s` inside its own `String()` | vet: `causes recursive String method call`; runtime stack overflow | Format plain numbers inside `String()` |
| 29 | 2 | Used `%b` for a float | `₦7036874417766400p-49` | `%b` is binary. `%.2f` gives 2 decimals |
| 30 | 2 | Converted a `*Kobo` receiver to `float64` | `cannot convert kb (variable of type *Kobo) to type float64` | Use a value receiver for methods that only read |
| 31 | | | | |

---

## Go words used in this plan

| Word | Meaning |
|---|---|
| **module** | A project with a `go.mod`. Its path (`github.com/na-cho-dev/mystally`) prefixes every import inside it |
| **package** | One folder of `.go` files that share a `package` line. The unit you import |
| **exported** | A name starting with a capital letter. Other packages can use it; lowercase names are private to the package |
| **zero value** | What a variable holds before you set it: `0`, `""`, `false`, `nil`. Every type has one |
| **receiver** | The `(kb Kobo)` or `(l *Ledger)` before a method name: the value the method works on |
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
4. Open the **Example** in the step you're on, and find the same tool used in another program.
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
