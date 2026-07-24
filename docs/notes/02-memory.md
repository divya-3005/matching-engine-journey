# Memory

## Overview

Memory is where a running program stores data.

Every variable, object, slice, map, and function call exists somewhere in memory.

Understanding memory is essential for writing efficient concurrent programs.

---

# Memory Layout

A running process typically contains:

```
+---------------------------+
| Code (Text Segment)       |
+---------------------------+
| Global Variables          |
+---------------------------+
| Heap                      |
|                           |
|   dynamically allocated   |
|   objects                 |
+---------------------------+

           Free Space

+---------------------------+
| Stack                     |
|                           |
| Function Calls            |
| Local Variables           |
+---------------------------+
```

The exact layout depends on the operating system.

---

# Stack

Every thread owns its own stack.

The stack stores:

- Function calls
- Local variables
- Parameters
- Return addresses

Example:

```go
func add(a, b int) int {
    c := a + b
    return c
}
```

During execution the stack contains:

```
add()

a = 2

b = 3

c = 5
```

When the function returns, the stack frame disappears.

This is extremely fast.

---

# Heap

The heap stores dynamically allocated objects.

Example:

```go
user := &User{}
```

or

```go
numbers := make([]int, 1000)
```

These objects often outlive the current function.

Unlike the stack, heap memory must eventually be reclaimed by the garbage collector.

---

# Stack vs Heap

| Stack | Heap |
|-------|------|
| Very fast | Slower |
| Automatically freed | Garbage collected |
| Local variables | Shared objects |
| Thread-local | Shared across threads |

---

# Why the Heap is Slower

Heap allocation requires:

- Finding free memory
- Updating allocator metadata
- Garbage collection later

Stack allocation only moves the stack pointer.

---

# Function Call Example

```go
func square(x int) int {
    y := x * x
    return y
}

func main() {
    result := square(5)
}
```

Execution:

```
main()

↓

square()

↓

return

↓

main()
```

Each function creates a new stack frame.

---

# Shared Memory

Threads inside one process share the heap.

Example:

```
Thread A

↓

Heap Object

↑

Thread B
```

This allows communication.

It also creates race conditions.

---

# Why This Matters

Mutexes exist because multiple threads may access the same heap object.

Atomics exist because multiple CPUs may modify the same memory location.

Lock-free programming is largely about coordinating access to shared memory safely.

---

# Escape Analysis

Go tries to keep variables on the stack whenever possible.

Example:

```go
func f() int {
    x := 5
    return x
}
```

`x` lives on the stack.

Example:

```go
func f() *int {
    x := 5
    return &x
}
```

`x` cannot remain on the stack because it is returned.

The compiler moves it to the heap.

This decision is called **escape analysis**.

---

# Why Escape Analysis Matters

Stack allocation is:

- Faster
- Cache friendly
- Does not require garbage collection

Heap allocation is:

- More flexible
- Slower
- Increases GC work

High-performance systems try to minimize heap allocations.

---

# Key Takeaways

- Every running program uses memory.
- The stack stores function calls and local variables.
- The heap stores dynamically allocated objects.
- Threads share heap memory.
- Shared memory creates race conditions.
- Go uses escape analysis to decide stack vs heap allocation.
- Minimizing heap allocations improves performance.