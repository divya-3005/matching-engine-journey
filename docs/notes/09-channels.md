# Channels

## What are Channels?

A channel is a communication mechanism between goroutines.

Instead of sharing memory directly, goroutines send values through channels.

```
Goroutine A

↓

Channel

↓

Goroutine B
```

---

## Creating a Channel

```go
ch := make(chan int)
```

This creates a channel that transports integers.

---

## Sending

```go
ch <- 42
```

This sends the value `42` into the channel.

---

## Receiving

```go
value := <-ch
```

This receives a value from the channel.

---

## Blocking Behavior

Channels synchronize goroutines automatically.

If no receiver exists:

```
Sender waits.
```

If no sender exists:

```
Receiver waits.
```

This makes channels useful for coordinating concurrent work.

---

## Buffered Channels

Channels may contain a buffer.

```go
ch := make(chan int, 3)
```

Capacity:

```
3 values
```

The sender blocks only when the buffer becomes full.

The receiver blocks only when the buffer becomes empty.

---

## Closing a Channel

```go
close(ch)
```

Closing tells receivers that no more values will be sent.

A closed channel should never be sent to again.

---

## Range

A common pattern:

```go
for value := range ch {
    fmt.Println(value)
}
```

The loop exits automatically when the channel is closed.

---

## Channels vs Mutexes

Mutexes protect shared memory.

Channels communicate data.

Mutex:

```
Shared Memory

↓

Lock

↓

Modify

↓

Unlock
```

Channel:

```
Producer

↓

Channel

↓

Consumer
```

---

## When to Use Channels

Channels are excellent for:

- Worker pools
- Pipelines
- Producer/consumer systems
- Communication between goroutines

They are not always the fastest option.

For extremely high-throughput systems, custom lock-free data structures are often faster.

---

## Matching Engine

Go channels are convenient and safe.

However, many production matching engines use custom ring buffers instead because they provide:

- Lower latency
- Fewer allocations
- More predictable performance

---

## Summary

- Channels communicate values between goroutines.
- Sending and receiving synchronize automatically.
- Buffered channels reduce blocking.
- Channels are excellent for coordination but are not always the fastest choice.