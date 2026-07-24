# Ring Buffer Design

## Requirements

- Single producer
- Single consumer
- Fixed-size buffer
- Constant-time operations
- No dynamic memory allocation

---

## Data Structure

```go
type RingBuffer struct {
    buffer []Order

    writeIndex int
    readIndex  int
}
```

---

## Invariants

Empty

```
write == read
```

Full

```
(write + 1) % capacity == read
```

---

## Producer Responsibilities

- Check if buffer is full
- Write data
- Advance write index

---

## Consumer Responsibilities

- Check if buffer is empty
- Read data
- Advance read index