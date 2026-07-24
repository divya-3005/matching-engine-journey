# Matching Engine

A production-style matching engine built in Go.

## Overview

This repository contains a limit order book and matching engine with:

- FIFO price-time priority matching
- Partial fills and multi-level matching
- Market and limit orders
- Order cancellation
- Write-ahead logging (WAL)
- WAL replay for recovery
- JSON snapshot save/load for fast restart
- Benchmarks for core engine performance

## Project structure

- `cmd/` — command-line applications and experimental runners
- `internal/engine/` — matching engine and order processing
- `internal/orderbook/` — book structure, price levels, and side management
- `internal/model/` — order and trade domain types
- `internal/ringbuffer/` — simple queue used for engine ingestion
- `internal/wal/` — WAL persistence, replay, and snapshots

## Getting started

Run tests:

```sh
go test ./...
```

Run benchmarks:

```sh
go test -bench=. -benchmem ./internal/engine
```

## Benchmarks

The engine is fast for basic workloads:

- `BenchmarkProcessLimitOrders`
- `BenchmarkSubmit`
- `BenchmarkMatch`
- `BenchmarkCancel`

Use `-benchtime=3s` for shorter runs when iterating.

## Continuous integration

A GitHub Actions workflow runs:

- `go test ./...`
- `go test -race ./...`
- `go vet ./...`

## Next improvements

Future work includes:

- HTTP API for order submission and book queries
- Simulator for synthetic order flow and throughput measurement
- Performance profiling and index optimizations
- O(1) cancellation with an order lookup index
- Multi-symbol order books
