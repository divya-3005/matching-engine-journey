# Contributing

Thank you for taking the time to read the source and consider contributing.
All improvements — bug fixes, test coverage, documentation clarity — are welcome.

---

## Development Setup

**Requirement:** Go 1.26.5 or the version declared in `matching-engine/go.mod`.

```sh
git clone https://github.com/divya-3005/matching-engine-journey.git
cd matching-engine-journey/matching-engine
go mod download
```

All Go source, tests, and tooling live inside the `matching-engine/` subdirectory.
The `Makefile` is the primary interface for day-to-day development tasks.

---

## Useful Commands

Run these from inside `matching-engine/`.

| Command | Description |
|---|---|
| `make build` | Compile `cmd/server` and `cmd/simulator` into `bin/` |
| `make test` | Run the full test suite |
| `make race` | Run the full test suite with the race detector |
| `make bench` | Run all benchmarks |
| `make fmt` | Format all Go source files with `gofmt` |
| `make vet` | Run `go vet` static analysis |
| `make lint` | Run `fmt` then `vet` |
| `make run-server` | Start the HTTP server on `:8080` |
| `make run-simulator` | Run the synthetic order simulator with default flags |
| `make clean` | Remove the `bin/` directory |
| `make all` | Run `fmt`, `vet`, `test`, `race`, and `build` in sequence |

---

## Coding Guidelines

- **Format before committing.** Run `make fmt` or `gofmt -w .` before every commit.
- **Keep functions focused.** A function that does one thing is easier to test and easier to change.
- **Prefer the standard library.** This project has zero external dependencies by design. Introduce a dependency only when the standard library clearly cannot meet the requirement.
- **Respect package boundaries.** The `orderbook` package stores orders; it does not match them. The `engine` package matches; it does not persist. The `wal` package persists; it does not route. Do not collapse these responsibilities.
- **Keep transport, persistence, and matching logic separated.** The HTTP server (`cmd/server`) and WAL (`internal/wal`) are adapters around a pure computation unit (`internal/engine`). Changes to one should not require changes to the others.
- **Test behaviour, not implementation.** Add or update tests when observable behaviour changes. Tests that depend on internal implementation details become a maintenance burden.
- **Keep benchmarks representative.** If the engine's matching behaviour changes, update the benchmarks in `internal/engine/engine_bench_test.go` and `internal/wal/wal_bench_test.go` so they continue to reflect realistic workloads.

---

## Pull Requests

Before opening a pull request, verify that the following all pass locally:

```sh
make fmt
make vet
make test
make race
```

If any step fails, resolve it before submitting.

**Commit message style:** use the imperative mood, present tense, no trailing period.

```
Add HTTP server
Fix WAL replay for CANCEL records
Improve benchmark documentation
Remove dead placeholder files
```

Each commit should represent one logical change. Avoid mixing unrelated fixes in a single commit.

If a pull request changes observable behaviour — matching rules, API contract, WAL format, snapshot format — update the relevant documentation:

- `README.md` for user-facing changes
- `docs/designs/architecture.md` for internal design changes

---

## Reporting Issues

When reporting a bug, please include:

- **Go version** (`go version`)
- **Operating system**
- **Steps to reproduce** — the minimal sequence of commands or inputs that trigger the issue
- **Expected behaviour** — what you expected to happen
- **Actual behaviour** — what actually happened
- **Logs or error messages**, if applicable
