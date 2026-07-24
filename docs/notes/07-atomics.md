# Atomics

## What is an Atomic Operation?

An atomic operation is an operation that completes as a single, indivisible step.

No other goroutine can observe the operation halfway through.

For example:

```go
atomic.AddInt64(&counter, 1)
```

The entire increment happens atomically.

---

## Why Not Just Use `counter++`?

The statement:

```go
counter++
```

is not atomic.

It consists of three operations:

```
Read

↓

Add 1

↓

Write
```

Another goroutine may modify the value between these steps, causing a race condition.

---

## Using the sync/atomic Package

Go provides atomic operations in:

```go
sync/atomic
```

Common operations include:

```go
atomic.AddInt64()
atomic.LoadInt64()
atomic.StoreInt64()
atomic.CompareAndSwapInt64()
atomic.SwapInt64()
```

---

## Atomic Increment

```go
var counter int64

atomic.AddInt64(&counter, 1)
```

Unlike:

```go
counter++
```

this operation is safe when accessed by multiple goroutines.

---

## Atomic Load

Reading a shared variable safely:

```go
value := atomic.LoadInt64(&counter)
```

---

## Atomic Store

Writing safely:

```go
atomic.StoreInt64(&counter, 100)
```

---

## Compare-And-Swap (CAS)

CAS is one of the most important atomic operations.

```
Current Value

↓

Is it equal to the expected value?

↓

Yes → Replace it

↓

No → Do nothing
```

In Go:

```go
ok := atomic.CompareAndSwapInt64(
    &counter,
    5,
    6,
)
```

If `counter` is currently `5`, it becomes `6`.

Otherwise nothing happens.

---

## Why Are Atomics Fast?

Mutexes require locking and unlocking.

Atomics use CPU instructions that operate directly on memory.

For simple operations they are usually much faster.

---

## Limitations

Atomics are excellent for:

- Counters
- Flags
- Simple state changes

They are not suitable for protecting complex data structures.

For larger critical sections, mutexes are usually the better choice.

---

## Summary

- Atomic operations execute as one indivisible step.
- They prevent race conditions for simple shared variables.
- They are faster than mutexes for simple operations.
- Compare-And-Swap (CAS) is the foundation of many lock-free algorithms.