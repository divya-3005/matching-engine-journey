# Go Scheduler

## Overview

The Go Scheduler is part of the Go runtime.

Its responsibility is to efficiently execute many goroutines using a relatively small number of operating system threads.

Without the scheduler, every goroutine would require its own thread, making Go programs much more expensive to run.

---

## Why is a Scheduler Needed?

The CPU can only execute operating system threads.

The CPU has no concept of goroutines.

Therefore, the Go runtime must map goroutines onto operating system threads.

```
Goroutines
     ↓
Go Scheduler
     ↓
OS Threads
     ↓
CPU Cores
```

---

## Responsibilities

The scheduler:

- Chooses which goroutine runs next.
- Maps goroutines to OS threads.
- Suspends blocked goroutines.
- Resumes runnable goroutines.
- Keeps CPU cores busy.

---

## The G-M-P Model

The Go scheduler is built around three concepts.

### G — Goroutine

A goroutine is a lightweight unit of execution.

Examples:

- main()
- worker()
- network handler
- logger

A Go program may create hundreds of thousands of goroutines.

---

### M — Machine

An M represents an operating system thread.

Threads execute machine instructions on CPU cores.

Unlike goroutines, OS threads are expensive to create.

---

### P — Processor

A P is a scheduler context.

It maintains runnable goroutines and provides the resources required to execute Go code.

An M must own a P before it can execute goroutines.

---

## Execution Flow

```
Runnable Goroutines

↓

Processor (P)

↓

Operating System Thread (M)

↓

CPU Core
```

---

## Blocking Operations

Suppose a goroutine executes:

```go
time.Sleep(5 * time.Second)
```

The scheduler marks that goroutine as blocked.

The operating system thread is then free to execute another runnable goroutine.

When the sleep completes, the goroutine becomes runnable again and waits for scheduling.

---

## Benefits

- Millions of goroutines are possible.
- Few operating system threads are required.
- CPU cores remain busy.
- Blocking goroutines do not waste threads.

---

## Summary

- Goroutines are managed by the Go runtime.
- Threads are managed by the operating system.
- The scheduler maps goroutines onto threads.
- The G-M-P model enables efficient concurrency.