package model

import (
	"errors"
	"time"
)

// Order represents a single order submitted to the exchange.
type Order struct {
	ID        uint64
	Symbol    string
	Side      Side
	Type      OrderType
	Price     uint64
	Quantity  uint64
	Remaining uint64
	Timestamp time.Time
}

// NewOrder creates a validated order.
func NewOrder(
	id uint64,
	symbol string,
	side Side,
	orderType OrderType,
	price uint64,
	quantity uint64,
) (*Order, error) {

	if symbol == "" {
		return nil, errors.New("symbol cannot be empty")
	}

	if !side.IsValid() {
		return nil, errors.New("invalid side")
	}

	if !orderType.IsValid() {
		return nil, errors.New("invalid order type")
	}

	if quantity == 0 {
		return nil, errors.New("quantity must be greater than zero")
	}

	if orderType == Limit && price == 0 {
		return nil, errors.New("limit order price must be greater than zero")
	}

	return &Order{
		ID:        id,
		Symbol:    symbol,
		Side:      side,
		Type:      orderType,
		Price:     price,
		Quantity:  quantity,
		Remaining: quantity,
		Timestamp: time.Now(),
	}, nil
}

// Fill reduces the remaining quantity.
func (o *Order) Fill(quantity uint64) error {
	if quantity == 0 {
		return errors.New("fill quantity must be greater than zero")
	}

	if quantity > o.Remaining {
		return errors.New("fill quantity exceeds remaining")
	}

	o.Remaining -= quantity
	return nil
}

func (o *Order) Filled() bool {
	return o.Remaining == 0
}
