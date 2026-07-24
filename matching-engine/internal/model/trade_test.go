package model

import (
	"testing"
	"time"
)

func TestTradeCreation(t *testing.T) {
	trade := Trade{
		ID:          1,
		BuyOrderID:  10,
		SellOrderID: 20,
		Symbol:      "AAPL",
		Price:       15000,
		Quantity:    100,
		Timestamp:   time.Now(),
	}

	if trade.BuyOrderID != 10 {
		t.Fatal("unexpected buy order ID")
	}

	if trade.SellOrderID != 20 {
		t.Fatal("unexpected sell order ID")
	}

	if trade.Price != 15000 {
		t.Fatal("unexpected price")
	}

	if trade.Quantity != 100 {
		t.Fatal("unexpected quantity")
	}
}