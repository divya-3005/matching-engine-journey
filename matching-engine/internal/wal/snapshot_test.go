package wal

import (
	"path/filepath"
	"testing"

	"github.com/divya-3005/matching-engine/internal/model"
	"github.com/divya-3005/matching-engine/internal/orderbook"
)

func TestSaveAndLoadSnapshot(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "snapshot.json")

	book := orderbook.New()

	buy1, err := model.NewOrder(1, "AAPL", model.Buy, model.Limit, 100, 50)
	if err != nil {
		t.Fatal(err)
	}
	buy2, err := model.NewOrder(2, "AAPL", model.Buy, model.Limit, 100, 30)
	if err != nil {
		t.Fatal(err)
	}
	if err := func() error {
		book.Add(buy1)
		book.Add(buy2)
		return nil
	}(); err != nil {
		t.Fatal(err)
	}

	sell1, err := model.NewOrder(3, "AAPL", model.Sell, model.Limit, 101, 20)
	if err != nil {
		t.Fatal(err)
	}
	if err := func() error {
		book.Add(sell1)
		return nil
	}(); err != nil {
		t.Fatal(err)
	}

	if err := SaveSnapshot(path, book); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadSnapshot(path)
	if err != nil {
		t.Fatal(err)
	}

	buyLevels := loaded.BuyLevels()
	if len(buyLevels) != 1 {
		t.Fatalf("expected 1 buy level, got %d", len(buyLevels))
	}
	if buyLevels[0].Price() != 100 {
		t.Fatalf("expected buy price 100, got %d", buyLevels[0].Price())
	}
	if buyLevels[0].Len() != 2 {
		t.Fatalf("expected 2 buy orders, got %d", buyLevels[0].Len())
	}
	if buyLevels[0].Orders()[0].ID != 1 || buyLevels[0].Orders()[0].Remaining != 50 {
		t.Fatalf("unexpected first buy order: %#v", buyLevels[0].Orders()[0])
	}
	if buyLevels[0].Orders()[1].ID != 2 || buyLevels[0].Orders()[1].Remaining != 30 {
		t.Fatalf("unexpected second buy order: %#v", buyLevels[0].Orders()[1])
	}

	sellLevels := loaded.SellLevels()
	if len(sellLevels) != 1 {
		t.Fatalf("expected 1 sell level, got %d", len(sellLevels))
	}
	if sellLevels[0].Price() != 101 {
		t.Fatalf("expected sell price 101, got %d", sellLevels[0].Price())
	}
	if sellLevels[0].Len() != 1 {
		t.Fatalf("expected 1 sell order, got %d", sellLevels[0].Len())
	}
	if sellLevels[0].Orders()[0].ID != 3 || sellLevels[0].Orders()[0].Remaining != 20 {
		t.Fatalf("unexpected sell order: %#v", sellLevels[0].Orders()[0])
	}
}
