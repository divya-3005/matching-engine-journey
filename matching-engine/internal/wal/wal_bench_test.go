package wal

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/divya-3005/matching-engine/internal/engine"
	"github.com/divya-3005/matching-engine/internal/model"
)

func BenchmarkReplay(b *testing.B) {
	dir := b.TempDir()
	path := filepath.Join(dir, "events.log")

	wal, err := New(path)
	if err != nil {
		b.Fatal(err)
	}
	defer wal.Close()

	for i := 0; i < 10000; i++ {
		order, err := model.NewOrder(uint64(i+1), "AAPL", model.Buy, model.Limit, 100, 100)
		if err != nil {
			b.Fatal(err)
		}
		if err := wal.LogSubmit(order); err != nil {
			b.Fatal(err)
		}
	}

	if err := wal.Close(); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine := engine.New(20000)
		wal, err := New(path)
		if err != nil {
			b.Fatal(err)
		}

		if err := wal.Replay(engine); err != nil {
			wal.Close()
			b.Fatal(err)
		}

		if err := wal.Close(); err != nil {
			b.Fatal(err)
		}
	}
}
