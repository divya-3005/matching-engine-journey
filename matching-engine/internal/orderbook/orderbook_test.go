package orderbook

import (
	"testing"

	"github.com/divya-3005/matching-engine/internal/model"
)

func mustOrder(
	t *testing.T,
	id uint64,
	side model.Side,
	price uint64,
	qty uint64,
) *model.Order {
	t.Helper()

	order, err := model.NewOrder(
		id,
		"AAPL",
		side,
		model.Limit,
		price,
		qty,
	)
	if err != nil {
		t.Fatal(err)
	}

	return order
}

func TestAddBuyOrder(t *testing.T) {
	ob := New()

	order, err := model.NewOrder(
		1,
		"AAPL",
		model.Buy,
		model.Limit,
		15000,
		100,
	)
	if err != nil {
		t.Fatal(err)
	}

	ob.Add(order)

	if len(ob.BuyOrders()) != 1 {
		t.Fatal("expected one buy order")
	}

	if len(ob.SellOrders()) != 0 {
		t.Fatal("expected zero sell orders")
	}
}

func TestAddSellOrder(t *testing.T) {
	ob := New()

	order, err := model.NewOrder(
		2,
		"AAPL",
		model.Sell,
		model.Limit,
		15100,
		50,
	)
	if err != nil {
		t.Fatal(err)
	}

	ob.Add(order)

	if len(ob.SellOrders()) != 1 {
		t.Fatal("expected one sell order")
	}

	if len(ob.BuyOrders()) != 0 {
		t.Fatal("expected zero buy orders")
	}
}

func TestCancelOrder(t *testing.T) {
	ob := New()

	o1 := mustOrder(t, 1, model.Buy, 100, 100)
	o2 := mustOrder(t, 2, model.Buy, 101, 50)
	o3 := mustOrder(t, 3, model.Buy, 100, 70)

	ob.Add(o1)
	ob.Add(o2)
	ob.Add(o3)

	cancelled, err := ob.Cancel(2)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled.ID != 2 {
		t.Fatalf("expected cancelled order 2, got %d", cancelled.ID)
	}

	if _, ok := ob.Buys().Level(101); ok {
		t.Fatal("expected price level 101 removed")
	}

	level, ok := ob.Buys().Level(100)
	if !ok {
		t.Fatal("expected price level 100 to remain")
	}

	first := level.Front()
	if first == nil || first.ID != 1 {
		t.Fatalf("expected first order 1, got %v", first)
	}

	level.Pop()
	second := level.Front()
	if second == nil || second.ID != 3 {
		t.Fatalf("expected second order 3, got %v", second)
	}
}

func TestCancelMiddleOrderSamePrice(t *testing.T) {
	ob := New()

	o1 := mustOrder(t, 1, model.Buy, 100, 100)
	o2 := mustOrder(t, 2, model.Buy, 100, 50)
	o3 := mustOrder(t, 3, model.Buy, 100, 70)

	ob.Add(o1)
	ob.Add(o2)
	ob.Add(o3)

	cancelled, err := ob.Cancel(2)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled.ID != 2 {
		t.Fatalf("expected cancelled order 2, got %d", cancelled.ID)
	}

	level, ok := ob.Buys().Level(100)
	if !ok {
		t.Fatal("expected price level 100 to remain")
	}

	first := level.Front()
	if first == nil || first.ID != 1 {
		t.Fatalf("expected first order 1, got %v", first)
	}

	level.Pop()
	second := level.Front()
	if second == nil || second.ID != 3 {
		t.Fatalf("expected second order 3, got %v", second)
	}
}
