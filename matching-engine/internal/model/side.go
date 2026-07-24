package model

import "fmt"

// Side represents whether an order is a buy or sell.
type Side uint8

const (
	Buy Side = iota
	Sell
)

// String implements fmt.Stringer.
func (s Side) String() string {
	switch s {
	case Buy:
		return "BUY"
	case Sell:
		return "SELL"
	default:
		return "UNKNOWN"
	}
}

// IsValid reports whether the side is valid.
func (s Side) IsValid() bool {
	return s == Buy || s == Sell
}

// ParseSide converts a string into a Side.
func ParseSide(value string) (Side, error) {
	switch value {
	case "BUY":
		return Buy, nil
	case "SELL":
		return Sell, nil
	default:
		return Buy, fmt.Errorf("invalid side: %s", value)
	}
}