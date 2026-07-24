# Program Execution

## Overview

Every program follows the same high-level lifecycle:

```
Source Code
    ↓
Compiler
    ↓
Executable
    ↓
Operating System
    ↓
CPU Executes Instructions
```

Understanding this pipeline is the foundation of systems programming.

---

# Source Code

A program begins as source code written in a programming language such as Go.

Example:

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello")
}
```

Humans understand source code.

The CPU does not.

---

# Compiler

The compiler translates source code into machine instructions.

For Go:

```
go build
```

produces an executable binary.

The compiler also performs:

- Type checking
- Optimization
- Dead code elimination
- Function inlining
- Escape analysis

---

# Executable

The output of the compiler is a binary executable.

It contains:

- Machine instructions
- Metadata
- Program sections
- Symbols

The executable is stored on disk until it is started.

---

# Operating System

When you run:

```
./matching-engine
```

the operating system:

- Creates a process
- Allocates virtual memory
- Loads the executable
- Creates the initial thread
- Starts execution

The operating system is responsible for managing CPU time and memory.

---

# Process

A process is a running instance of a program.

A process owns:

- Virtual memory
- Threads
- Open files
- Network sockets
- Heap
- Stack

Multiple processes are isolated from one another.

---

# Thread

A thread is a sequence of instructions being executed.

A process may contain one or many threads.

Each thread has its own:

- Stack
- Program counter
- CPU registers

Threads inside the same process share:

- Heap
- Global variables
- Open files

---

# CPU Execution

The CPU repeatedly performs the instruction cycle:

```
Fetch

↓

Decode

↓

Execute

↓

Repeat
```

This happens billions of times every second.

---

# Memory Hierarchy

Programs do not access RAM directly every time.

The hierarchy is:

```
CPU Registers

↓

L1 Cache

↓

L2 Cache

↓

L3 Cache

↓

RAM

↓

Disk
```

The closer memory is to the CPU, the faster it is.

---

# Why This Matters

Everything we build later depends on understanding this execution model.

Examples:

- Goroutines run on threads.
- Threads execute on CPUs.
- CPUs access memory through caches.
- Cache behavior affects latency.
- Matching engines are optimized around CPU and memory behavior.

Understanding program execution explains why some designs are dramatically faster than others.