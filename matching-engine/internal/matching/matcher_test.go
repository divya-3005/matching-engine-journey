package matching

import (
	"errors"
	"testing"

	"github.com/divya-3005/matching-engine/internal/model"
)

func TestMatch(t *testing.T) {
	m := New()

	buy, err := model.NewOrder(
		1,
		"AAPL",
		model.Buy,
		model.Limit,
		105,
		100,
	)
	if err != nil {
		t.Fatal(err)
	}

	sell, err := model.NewOrder(
		2,
		"AAPL",
		model.Sell,
		model.Limit,
		100,
		100,
	)
	if err != nil {
		t.Fatal(err)
	}

	trade, err := m.Match(buy, sell)
	if err != nil {
		t.Fatal(err)
	}

	if trade == nil {
		t.Fatal("expected trade")
	}

	if trade.Price != 100 {
		t.Fatalf("expected trade at seller price")
	}
}

func TestNoMatch(t *testing.T) {
	m := New()

	buy, err := model.NewOrder(
		1,
		"AAPL",
		model.Buy,
		model.Limit,
		99,
		100,
	)
	if err != nil {
		t.Fatal(err)
	}

	sell, err := model.NewOrder(
		2,
		"AAPL",
		model.Sell,
		model.Limit,
		100,
		100,
	)
	if err != nil {
		t.Fatal(err)
	}

	trade, err := m.Match(buy, sell)
	if err != nil {
		t.Fatal(err)
	}

	if trade != nil {
		t.Fatal("expected no trade")
	}
}

// TestQuantityMismatch verifies that the Matcher returns ErrQuantityMismatch
// when the buy and sell orders have different quantities.
// Note: this constraint is why the Matcher was superseded by the engine's
// inline matching loop, which handles partial fills (mismatched quantities).
func TestQuantityMismatch(t *testing.T) {
	m := New()

	buy, err := model.NewOrder(1, "AAPL", model.Buy, model.Limit, 105, 100)
	if err != nil {
		t.Fatal(err)
	}

	sell, err := model.NewOrder(2, "AAPL", model.Sell, model.Limit, 100, 50)
	if err != nil {
		t.Fatal(err)
	}

	_, err = m.Match(buy, sell)
	if !errors.Is(err, ErrQuantityMismatch) {
		t.Fatalf("expected ErrQuantityMismatch, got %v", err)
	}
}
