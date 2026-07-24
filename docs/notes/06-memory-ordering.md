# Memory Ordering

## Why Memory Ordering Matters

When writing concurrent programs, it is tempting to assume that statements execute exactly in the order they appear.

For example:

```go
a = 1
b = 2
```

We naturally think:

```
a is written first
↓

b is written second
```

However, modern CPUs and compilers may reorder memory operations to improve performance.

This optimization is safe for a single goroutine, but it can affect how other goroutines observe memory.

---

## Why Reordering Happens

Modern processors execute billions of instructions per second.

To maximize performance they:

- Reorder instructions.
- Buffer writes.
- Execute independent instructions in parallel.

As long as a single goroutine observes the same behavior, these optimizations are allowed.

---

## Example

Suppose two shared variables exist.

```go
var data int
var ready bool
```

One goroutine writes:

```go
data = 42
ready = true
```

Another goroutine waits:

```go
if ready {
    fmt.Println(data)
}
```

We expect:

```
data = 42

↓

ready = true
```

But without synchronization, another goroutine is not guaranteed to observe the writes in that order.

---

## Memory Visibility

Each CPU core has caches.

A write performed on one core is not necessarily visible to another core immediately.

Synchronization primitives ensure that memory changes become visible to other goroutines.

---

## Synchronization

The following synchronization mechanisms establish ordering and visibility:

- Mutexes
- Atomic operations
- Channel communication
- WaitGroups

These guarantee that memory updates are observed correctly.

---

## Key Takeaway

Without synchronization:

- The order of memory operations is not guaranteed.
- One goroutine may observe stale values.
- Programs may behave unpredictably.

With synchronization:

- Memory becomes visible in the expected order.
- Concurrent programs behave correctly.