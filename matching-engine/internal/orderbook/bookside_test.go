package orderbook

import (
	"testing"

	"github.com/divya-3005/matching-engine/internal/model"
)

func TestBookSideAdd(t *testing.T) {
	side := NewBookSide(model.Buy)

	order, err := model.NewOrder(
		1,
		"AAPL",
		model.Buy,
		model.Limit,
		100,
		10,
	)
	if err != nil {
		t.Fatal(err)
	}

	side.Add(order)

	if side.Len() != 1 {
		t.Fatalf("expected 1 price level, got %d", side.Len())
	}

	level, ok := side.Level(100)
	if !ok {
		t.Fatal("expected price level")
	}

	if level.Len() != 1 {
		t.Fatal("expected one order")
	}
}

func TestBookSideSamePrice(t *testing.T) {
	side := NewBookSide(model.Buy)

	first, err := model.NewOrder(1, "AAPL", model.Buy, model.Limit, 100, 10)
	if err != nil {
		t.Fatal(err)
	}

	second, err := model.NewOrder(2, "AAPL", model.Buy, model.Limit, 100, 20)
	if err != nil {
		t.Fatal(err)
	}

	side.Add(first)
	side.Add(second)

	if side.Len() != 1 {
		t.Fatal("expected one price level")
	}

	level, ok := side.Level(100)
	if !ok {
		t.Fatal("expected price level")
	}

	if level.Len() != 2 {
		t.Fatal("expected two orders")
	}
}

func TestBookSideRemove(t *testing.T) {
	side := NewBookSide(model.Buy)

	order, err := model.NewOrder(1, "AAPL", model.Buy, model.Limit, 100, 10)
	if err != nil {
		t.Fatal(err)
	}

	side.Add(order)

	side.Remove(100)

	if side.Len() != 0 {
		t.Fatal("expected zero price levels")
	}
}

func TestBestBuy(t *testing.T) {
	side := NewBookSide(model.Buy)

	o1, err := model.NewOrder(1, "AAPL", model.Buy, model.Limit, 100, 10)
	if err != nil {
		t.Fatal(err)
	}
	o2, err := model.NewOrder(2, "AAPL", model.Buy, model.Limit, 105, 10)
	if err != nil {
		t.Fatal(err)
	}
	o3, err := model.NewOrder(3, "AAPL", model.Buy, model.Limit, 103, 10)
	if err != nil {
		t.Fatal(err)
	}

	side.Add(o1)
	side.Add(o2)
	side.Add(o3)

	best := side.Best()
	if best == nil {
		t.Fatal("expected best level")
	}

	if best.Price() != 105 {
		t.Fatalf("expected 105, got %d", best.Price())
	}
}

func TestBookSideBest(t *testing.T) {
	side := NewBookSide(model.Buy)

	o1, _ := model.NewOrder(1, "AAPL", model.Buy, model.Limit, 100, 10)
	o2, _ := model.NewOrder(2, "AAPL", model.Buy, model.Limit, 105, 10)
	o3, _ := model.NewOrder(3, "AAPL", model.Buy, model.Limit, 103, 10)

	side.Add(o1)
	side.Add(o2)
	side.Add(o3)

	best := side.Best()
	if best == nil {
		t.Fatal("expected best level")
	}

	if best.Price() != 105 {
		t.Fatalf("expected best price 105, got %d", best.Price())
	}
}

func TestBestSell(t *testing.T) {
	side := NewBookSide(model.Sell)

	o1, err := model.NewOrder(1, "AAPL", model.Sell, model.Limit, 105, 10)
	if err != nil {
		t.Fatal(err)
	}
	o2, err := model.NewOrder(2, "AAPL", model.Sell, model.Limit, 101, 10)
	if err != nil {
		t.Fatal(err)
	}
	o3, err := model.NewOrder(3, "AAPL", model.Sell, model.Limit, 103, 10)
	if err != nil {
		t.Fatal(err)
	}

	side.Add(o1)
	side.Add(o2)
	side.Add(o3)

	best := side.Best()
	if best == nil {
		t.Fatal("expected best level")
	}

	if best.Price() != 101 {
		t.Fatalf("expected 101, got %d", best.Price())
	}
}
