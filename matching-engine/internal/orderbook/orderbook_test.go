package orderbook

import (
	"testing"

	"github.com/divya-3005/matching-engine/internal/model"
)

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
