# Goroutines

## What is a goroutine?

- A lightweight unit of execution managed by the Go runtime.
- Created using the `go` keyword.
- Not the same as an OS thread.

## Key observations

- Goroutines do not start immediately.
- Their execution order is not guaranteed.
- The Go scheduler decides when they run.
- Many goroutines can share a small number of OS threads.
- Sleeping goroutines do not block the scheduler from running other goroutines.