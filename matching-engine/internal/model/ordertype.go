package model

import "fmt"

// OrderType represents how an order should be executed.
type OrderType uint8

const (
	Limit OrderType = iota
	Market
)

// String implements fmt.Stringer.
func (o OrderType) String() string {
	switch o {
	case Limit:
		return "LIMIT"
	case Market:
		return "MARKET"
	default:
		return "UNKNOWN"
	}
}

// IsValid reports whether the order type is valid.
func (o OrderType) IsValid() bool {
	return o == Limit || o == Market
}

// ParseOrderType converts a string into an OrderType.
func ParseOrderType(value string) (OrderType, error) {
	switch value {
	case "LIMIT":
		return Limit, nil
	case "MARKET":
		return Market, nil
	default:
		return Limit, fmt.Errorf("invalid order type: %s", value)
	}
}