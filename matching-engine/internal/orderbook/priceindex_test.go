package orderbook

import "testing"

func TestBidPriceIndex(t *testing.T) {
	idx := NewPriceIndex(true)

	idx.Insert(100)
	idx.Insert(105)
	idx.Insert(103)

	best, ok := idx.Best()
	if !ok {
		t.Fatal("expected best price")
	}

	if best != 105 {
		t.Fatalf("expected 105, got %d", best)
	}
}

func TestAskPriceIndex(t *testing.T) {
	idx := NewPriceIndex(false)

	idx.Insert(105)
	idx.Insert(100)
	idx.Insert(103)

	best, ok := idx.Best()
	if !ok {
		t.Fatal("expected best price")
	}

	if best != 100 {
		t.Fatalf("expected 100, got %d", best)
	}
}

func TestRemove(t *testing.T) {
	idx := NewPriceIndex(true)

	idx.Insert(100)
	idx.Insert(105)

	idx.Remove(105)

	best, ok := idx.Best()
	if !ok {
		t.Fatal("expected best price")
	}

	if best != 100 {
		t.Fatal("expected remaining price to be 100")
	}
}
