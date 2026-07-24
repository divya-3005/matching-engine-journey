# Mutexes

## Why Do We Need Mutexes?

When multiple goroutines access the same memory at the same time, the program may behave unpredictably.

This problem is called a **race condition**.

A race condition occurs when:

- Two or more goroutines access the same data.
- At least one goroutine modifies the data.
- There is no synchronization.

---

## Example

Suppose two goroutines increment the same counter.

```go
counter++
```

Although it looks like a single operation, it actually consists of three steps:

1. Read the current value.
2. Add one.
3. Write the new value.

If two goroutines perform these steps simultaneously, one update may overwrite the other, producing an incorrect result.

---

## What is a Mutex?

A mutex (Mutual Exclusion) ensures that only one goroutine can execute a critical section of code at a time.

```
Goroutine A
    │
 Lock()
    │
Critical Section
    │
Unlock()
    │

Goroutine B waits until the mutex is unlocked.
```

---

## sync.Mutex

Go provides a mutex in the `sync` package.

```go
var mu sync.Mutex
```

To protect shared data:

```go
mu.Lock()

// Critical section

mu.Unlock()
```

---

## Critical Section

A critical section is any code that accesses shared mutable data.

Only one goroutine should execute a critical section at a time.

---

## Benefits

- Prevents race conditions.
- Protects shared state.
- Makes concurrent programs deterministic.

---

## Costs

Mutexes introduce synchronization overhead.

If many goroutines compete for the same mutex, performance may decrease due to contention.

---

## Summary

- Shared mutable data can cause race conditions.
- Mutexes serialize access to critical sections.
- Use mutexes when multiple goroutines modify the same data.