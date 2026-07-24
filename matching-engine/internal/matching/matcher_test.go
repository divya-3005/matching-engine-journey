package matching

import (
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
