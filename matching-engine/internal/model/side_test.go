package model

import "testing"

func TestSideString(t *testing.T) {
	tests := []struct {
		side Side
		want string
	}{
		{Buy, "BUY"},
		{Sell, "SELL"},
		{Side(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		if got := tt.side.String(); got != tt.want {
			t.Fatalf("expected %q, got %q", tt.want, got)
		}
	}
}

func TestSideIsValid(t *testing.T) {
	if !Buy.IsValid() {
		t.Fatal("Buy should be valid")
	}

	if !Sell.IsValid() {
		t.Fatal("Sell should be valid")
	}

	if Side(99).IsValid() {
		t.Fatal("expected invalid side")
	}
}

func TestParseSide(t *testing.T) {
	side, err := ParseSide("BUY")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if side != Buy {
		t.Fatalf("expected Buy")
	}

	_, err = ParseSide("INVALID")
	if err == nil {
		t.Fatal("expected error")
	}
}