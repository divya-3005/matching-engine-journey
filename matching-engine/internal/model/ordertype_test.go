package model

import "testing"

func TestOrderTypeString(t *testing.T) {
	tests := []struct {
		orderType OrderType
		want      string
	}{
		{Limit, "LIMIT"},
		{Market, "MARKET"},
		{OrderType(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		if got := tt.orderType.String(); got != tt.want {
			t.Fatalf("expected %q, got %q", tt.want, got)
		}
	}
}

func TestOrderTypeIsValid(t *testing.T) {
	if !Limit.IsValid() {
		t.Fatal("Limit should be valid")
	}

	if !Market.IsValid() {
		t.Fatal("Market should be valid")
	}

	if OrderType(99).IsValid() {
		t.Fatal("expected invalid order type")
	}
}

func TestParseOrderType(t *testing.T) {
	orderType, err := ParseOrderType("LIMIT")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if orderType != Limit {
		t.Fatal("expected Limit")
	}

	_, err = ParseOrderType("INVALID")
	if err == nil {
		t.Fatal("expected error")
	}
}