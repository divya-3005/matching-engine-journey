package model

import "testing"

func TestNewOrder(t *testing.T) {
	order, err := NewOrder(
		1,
		"AAPL",
		Buy,
		Limit,
		15000,
		100,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if order.ID != 1 {
		t.Fatal("unexpected ID")
	}

	if order.Symbol != "AAPL" {
		t.Fatal("unexpected symbol")
	}

	if order.Remaining != 100 {
		t.Fatal("remaining should equal quantity")
	}
}

func TestNewOrderInvalidQuantity(t *testing.T) {
	_, err := NewOrder(
		1,
		"AAPL",
		Buy,
		Limit,
		15000,
		0,
	)

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewOrderInvalidPrice(t *testing.T) {
	_, err := NewOrder(
		1,
		"AAPL",
		Buy,
		Limit,
		0,
		100,
	)

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewMarketOrder(t *testing.T) {
	order, err := NewOrder(
		1,
		"AAPL",
		Sell,
		Market,
		0,
		100,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if order.Type != Market {
		t.Fatal("expected market order")
	}
}
func TestOrderFill(t *testing.T) {
	order, err := NewOrder(
		1,
		"AAPL",
		Buy,
		Limit,
		100,
		100,
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := order.Fill(40); err != nil {
		t.Fatal(err)
	}

	if order.Remaining != 60 {
		t.Fatalf("expected remaining 60, got %d", order.Remaining)
	}

	if order.Filled() {
		t.Fatal("order should not be filled")
	}

	if err := order.Fill(60); err != nil {
		t.Fatal(err)
	}

	if !order.Filled() {
		t.Fatal("expected order to be filled")
	}
}
