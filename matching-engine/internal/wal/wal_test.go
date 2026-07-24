package wal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/divya-3005/matching-engine/internal/engine"
	"github.com/divya-3005/matching-engine/internal/model"
)

func TestLogSubmit(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "events.log")

	wal, err := New(path)
	if err != nil {
		t.Fatal(err)
	}

	order, err := model.NewOrder(1, "AAPL", model.Buy, model.Limit, 100, 50)
	if err != nil {
		t.Fatal(err)
	}

	if err := wal.LogSubmit(order); err != nil {
		t.Fatal(err)
	}

	if err := wal.Close(); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	expected := "SUBMIT,1,AAPL,BUY,LIMIT,100,50\n"
	if string(data) != expected {
		t.Fatalf("expected %q, got %q", expected, string(data))
	}
}

func TestLogTrade(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "events.log")

	wal, err := New(path)
	if err != nil {
		t.Fatal(err)
	}
	defer wal.Close()

	trade := &model.Trade{
		ID:          1,
		BuyOrderID:  10,
		SellOrderID: 20,
		Symbol:      "AAPL",
		Price:       100,
		Quantity:    50,
	}

	if err := wal.LogTrade(trade); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	expected := "TRADE,1,10,20,AAPL,100,50\n"
	if string(data) != expected {
		t.Fatalf("expected %q, got %q", expected, string(data))
	}
}

func TestLogCancel(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "events.log")

	wal, err := New(path)
	if err != nil {
		t.Fatal(err)
	}
	defer wal.Close()

	if err := wal.LogCancel(42); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	expected := "CANCEL,42\n"
	if string(data) != expected {
		t.Fatalf("expected %q, got %q", expected, string(data))
	}
}

func TestReplaySubmit(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "events.log")

	wal, err := New(path)
	if err != nil {
		t.Fatal(err)
	}

	order1, err := model.NewOrder(1, "AAPL", model.Buy, model.Limit, 100, 50)
	if err != nil {
		t.Fatal(err)
	}
	order2, err := model.NewOrder(2, "AAPL", model.Sell, model.Limit, 101, 20)
	if err != nil {
		t.Fatal(err)
	}

	if err := wal.LogSubmit(order1); err != nil {
		t.Fatal(err)
	}
	if err := wal.LogSubmit(order2); err != nil {
		t.Fatal(err)
	}

	if err := wal.Close(); err != nil {
		t.Fatal(err)
	}

	engine := engine.New(10)
	wal, err = New(path)
	if err != nil {
		t.Fatal(err)
	}
	defer wal.Close()

	if err := wal.Replay(engine); err != nil {
		t.Fatal(err)
	}

	buyLevel, ok := engine.OrderBook().Buys().Level(100)
	if !ok {
		t.Fatal("expected buy level 100")
	}
	if buyLevel.Front().Remaining != 50 {
		t.Fatalf("expected buy quantity 50, got %d", buyLevel.Front().Remaining)
	}

	sellLevel, ok := engine.OrderBook().Sells().Level(101)
	if !ok {
		t.Fatal("expected sell level 101")
	}
	if sellLevel.Front().Remaining != 20 {
		t.Fatalf("expected sell quantity 20, got %d", sellLevel.Front().Remaining)
	}
}
