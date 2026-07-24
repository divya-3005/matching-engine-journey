package model

import "time"

// Trade represents an execution between a buy order and a sell order.
type Trade struct {
	ID          uint64
	BuyOrderID  uint64
	SellOrderID uint64

	Symbol string

	Price    uint64
	Quantity uint64

	Timestamp time.Time
}