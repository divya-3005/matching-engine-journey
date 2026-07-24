package orderbook

import (
	"testing"

	"github.com/divya-3005/matching-engine/internal/model"
)

func TestPriceLevelFIFO(t *testing.T) {
	level := NewPriceLevel(100)

	first, err := model.NewOrder(1, "AAPL", model.Buy, model.Limit, 100, 10)
	if err != nil {
		t.Fatal(err)
	}

	second, err := model.NewOrder(2, "AAPL", model.Buy, model.Limit, 100, 20)
	if err != nil {
		t.Fatal(err)
	}

	level.Add(first)
	level.Add(second)

	if level.Len() != 2 {
		t.Fatalf("expected 2 orders, got %d", level.Len())
	}

	if level.Front().ID != 1 {
		t.Fatal("expected first order")
	}

	removed := level.Pop()
	if removed.ID != 1 {
		t.Fatal("wrong order removed")
	}

	if level.Front().ID != 2 {
		t.Fatal("expected second order")
	}
}
