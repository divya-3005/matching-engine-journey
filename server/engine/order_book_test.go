package engine

import (
	"errors"
	"sort"
	"sync"
	"testing"
	"time"
)

func TestFullMatch(t *testing.T) {
	ob := NewOrderBook("AAPL")

	// Alice wants to sell 10 shares at $100
	aliceSell := &Order{
		ID:        1,
		Symbol:    "AAPL",
		Side:      Sell,
		Price:     100,
		Amount:    10,
		Timestamp: time.Now().UnixMilli(),
	}

	// Bob wants to buy 10 shares at $100
	bobBuy := &Order{
		ID:        2,
		Symbol:    "AAPL",
		Side:      Buy,
		Price:     100,
		Amount:    10,
		Timestamp: time.Now().UnixMilli(),
	}

	// Alice's order rests on the book (no trades yet)
	trades1, err := ob.ProcessOrder(aliceSell)
	if err != nil {
		t.Fatalf("unexpected error processing aliceSell: %v", err)
	}
	if len(trades1) != 0 {
		t.Fatalf("expected 0 trades, got %d", len(trades1))
	}

	// Bob's order matches with Alice's
	trades2, err := ob.ProcessOrder(bobBuy)
	if err != nil {
		t.Fatalf("unexpected error processing bobBuy: %v", err)
	}
	if len(trades2) != 1 {
		t.Fatalf("expected 1 trade, got %d", len(trades2))
	}

	trade := trades2[0]
	if trade.MakerOrderID != 1 || trade.TakerOrderID != 2 {
		t.Errorf("wrong maker/taker IDs: got maker %d, taker %d", trade.MakerOrderID, trade.TakerOrderID)
	}
	if trade.Amount != 10 || trade.Price != 100 {
		t.Errorf("wrong trade amount or price: got %d @ %d", trade.Amount, trade.Price)
	}

	// Book should now have 0 orders
	if len(ob.Orders) != 0 {
		t.Errorf("expected 0 resting orders, got %d", len(ob.Orders))
	}
}

func TestPartialFillAndFIFO(t *testing.T) {
	ob := NewOrderBook("AAPL")

	// Alice sells 10 @ $100
	if _, err := ob.ProcessOrder(&Order{ID: 1, Symbol: "AAPL", Side: Sell, Price: 100, Amount: 10, Timestamp: 1}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Bob sells 10 @ $100 (arrived later)
	if _, err := ob.ProcessOrder(&Order{ID: 2, Symbol: "AAPL", Side: Sell, Price: 100, Amount: 10, Timestamp: 2}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Charlie buys 15 @ $100
	trades, err := ob.ProcessOrder(&Order{ID: 3, Symbol: "AAPL", Side: Buy, Price: 100, Amount: 15, Timestamp: 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(trades) != 2 {
		t.Fatalf("expected 2 trades, got %d", len(trades))
	}

	// Trade 1: Alice (10 shares)
	if trades[0].MakerOrderID != 1 || trades[0].Amount != 10 {
		t.Errorf("expected trade 1 with Alice for 10 shares, got maker %d amount %d", trades[0].MakerOrderID, trades[0].Amount)
	}

	// Trade 2: Bob (5 shares)
	if trades[1].MakerOrderID != 2 || trades[1].Amount != 5 {
		t.Errorf("expected trade 2 with Bob for 5 shares, got maker %d amount %d", trades[1].MakerOrderID, trades[1].Amount)
	}

	// Bob should still have 5 shares left on the book
	remainingOrder, exists := ob.Orders[2]
	if !exists {
		t.Fatalf("expected Bob's order to still exist on the book")
	}
	if remainingOrder.Amount != 5 {
		t.Errorf("expected Bob to have 5 shares remaining, got %d", remainingOrder.Amount)
	}
}

func TestCancelOrder(t *testing.T) {
	ob := NewOrderBook("AAPL")

	order := &Order{
		ID:        1,
		Symbol:    "AAPL",
		Side:      Buy,
		Price:     100,
		Amount:    10,
		Timestamp: time.Now().UnixMilli(),
	}

	// Place the order
	if _, err := ob.ProcessOrder(order); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ob.Orders) != 1 {
		t.Fatalf("expected 1 order on book, got %d", len(ob.Orders))
	}

	// Cancel it
	if err := ob.CancelOrder(1); err != nil {
		t.Fatalf("expected cancel to succeed, got %v", err)
	}

	// Verify it's gone
	if len(ob.Orders) != 0 {
		t.Errorf("expected 0 orders in map, got %d", len(ob.Orders))
	}
	if len(ob.Bids) != 0 {
		t.Errorf("expected 0 bid levels, got %d", len(ob.Bids))
	}

	// Cancelling again should return ErrOrderNotFound
	if err := ob.CancelOrder(1); !errors.Is(err, ErrOrderNotFound) {
		t.Errorf("expected second cancel to return ErrOrderNotFound, got %v", err)
	}
}

func TestDuplicateOrderID(t *testing.T) {
	ob := NewOrderBook("AAPL")

	order1 := &Order{ID: 42, Symbol: "AAPL", Side: Buy, Type: Limit, Price: 100, Amount: 10}
	_, err := ob.ProcessOrder(order1)
	if err != nil {
		t.Fatalf("unexpected error on first order: %v", err)
	}

	// Attempting to place another order with the same ID must be rejected
	order2 := &Order{ID: 42, Symbol: "AAPL", Side: Buy, Type: Limit, Price: 105, Amount: 5}
	_, err = ob.ProcessOrder(order2)
	if err == nil {
		t.Fatalf("expected error for duplicate order ID, got nil")
	}

	// Original resting order must remain intact and uncorrupted
	resting, exists := ob.Orders[42]
	if !exists || resting.Price != 100 || resting.Amount != 10 {
		t.Fatalf("original order was corrupted or overwritten: %+v", resting)
	}
}

func BenchmarkProcessOrder(b *testing.B) {
	ob := NewOrderBook("AAPL")

	nextID := uint64(1)
	for i := 0; i < 1000; i++ {
		_, _ = ob.ProcessOrder(&Order{
			ID:        nextID,
			Symbol:    "AAPL",
			Side:      Sell,
			Price:     uint64(100 + (i % 10)),
			Amount:    10,
			Timestamp: int64(i),
		})
		nextID++
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		trades, err := ob.ProcessOrder(&Order{
			ID: nextID, Symbol: "AAPL", Side: Buy, Type: Limit,
			Price: 105, Amount: 1, Timestamp: int64(i),
		})
		nextID++
		if err != nil || len(trades) != 1 {
			b.Fatalf("iteration %d: buy must match exactly one resting ask (err=%v, trades=%d)", i, err, len(trades))
		}

		// Put the consumed unit back at the exact price it traded, so the book shape never
		// changes and every iteration exercises the matching path.
		if _, err := ob.ProcessOrder(&Order{
			ID: nextID, Symbol: "AAPL", Side: Sell, Type: Limit,
			Price: trades[0].Price, Amount: 1, Timestamp: int64(i),
		}); err != nil {
			b.Fatal(err)
		}
		nextID++
	}
}

func BenchmarkProcessOrderWithWAL(b *testing.B) {
	dir := b.TempDir()
	wal, err := OpenWAL(dir + "/bench_wal.log")
	if err != nil {
		b.Fatalf("failed to open WAL: %v", err)
	}
	defer wal.Close()

	ob := NewOrderBook("AAPL")
	ob.SetWAL(wal)
	nextID := uint64(1)
	for i := 0; i < 1000; i++ {
		_, _ = ob.ProcessOrder(&Order{
			ID: nextID, Symbol: "AAPL", Side: Sell,
			Price: uint64(100 + (i % 10)), Amount: 10, Timestamp: int64(i),
		})
		nextID++
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		trades, err := ob.ProcessOrder(&Order{
			ID: nextID, Symbol: "AAPL", Side: Buy, Type: Limit,
			Price: 105, Amount: 1, Timestamp: int64(i),
		})
		nextID++
		if err != nil || len(trades) != 1 {
			b.Fatalf("iteration %d: buy must match exactly one resting ask (err=%v, trades=%d)", i, err, len(trades))
		}

		if _, err := ob.ProcessOrder(&Order{
			ID: nextID, Symbol: "AAPL", Side: Sell, Type: Limit,
			Price: trades[0].Price, Amount: 1, Timestamp: int64(i),
		}); err != nil {
			b.Fatal(err)
		}
		nextID++
	}
}

func TestMarketOrder(t *testing.T) {
	ob := NewOrderBook("AAPL")

	// Alice sells 5 @ $100
	if _, err := ob.ProcessOrder(&Order{ID: 1, Symbol: "AAPL", Side: Sell, Type: Limit, Price: 100, Amount: 5}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Bob sells 10 @ $105
	if _, err := ob.ProcessOrder(&Order{ID: 2, Symbol: "AAPL", Side: Sell, Type: Limit, Price: 105, Amount: 10}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Charlie places a Market Buy for 12 shares
	marketBuy := &Order{
		ID:     3,
		Symbol: "AAPL",
		Side:   Buy,
		Type:   Market,
		Amount: 12,
	}

	trades, err := ob.ProcessOrder(marketBuy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(trades) != 2 {
		t.Fatalf("expected 2 trades, got %d", len(trades))
	}

	// Trade 1: Alice (5 @ $100)
	if trades[0].MakerOrderID != 1 || trades[0].Amount != 5 || trades[0].Price != 100 {
		t.Errorf("trade 1 mismatch: got maker %d, amount %d, price %d", trades[0].MakerOrderID, trades[0].Amount, trades[0].Price)
	}

	// Trade 2: Bob (7 @ $105)
	if trades[1].MakerOrderID != 2 || trades[1].Amount != 7 || trades[1].Price != 105 {
		t.Errorf("trade 2 mismatch: got maker %d, amount %d, price %d", trades[1].MakerOrderID, trades[1].Amount, trades[1].Price)
	}

	// Bob should have 3 shares remaining on the book
	remainingBob, exists := ob.Orders[2]
	if !exists || remainingBob.Amount != 3 {
		t.Errorf("expected Bob to have 3 shares remaining, got %v", remainingBob)
	}

	// Market order must NOT be on the book
	if _, exists := ob.Orders[3]; exists {
		t.Errorf("market order should not rest on the book")
	}
}

func TestConcurrentOrders(t *testing.T) {
	ob := NewOrderBook("AAPL")
	const numGoroutines = 16
	const ordersPerGoroutine = 200

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for g := 0; g < numGoroutines; g++ {
		go func(routineID int) {
			defer wg.Done()
			for i := 0; i < ordersPerGoroutine; i++ {
				orderID := uint64(routineID*100000 + i + 1)
				side := Buy
				if i%2 == 0 {
					side = Sell
				}
				price := uint64(100 + (i % 20))

				// Submit Limit order
				_, _ = ob.ProcessOrder(&Order{
					ID:        orderID,
					Symbol:    "AAPL",
					Side:      side,
					Type:      Limit,
					Price:     price,
					Amount:    5,
					Timestamp: time.Now().UnixMilli(),
				})

				// Concurrently read snapshot and top-of-book
				if i%10 == 0 {
					ob.GetSnapshot()
					ob.TopOfBook()
				}

				// Concurrently cancel some orders
				if i%4 == 0 {
					ob.CancelOrder(orderID)
				}
			}
		}(g)
	}

	wg.Wait()
}

func TestRejectsOversizedAmount(t *testing.T) {
	ob := NewOrderBook("AAPL")
	_, err := ob.ProcessOrder(&Order{ID: 1, Symbol: "AAPL", Side: Buy, Type: Limit, Price: 100, Amount: MaxOrderAmount + 1})
	if !errors.Is(err, ErrInvalidOrder) {
		t.Fatalf("expected ErrInvalidOrder, got %v", err)
	}
}

func TestRemoveOrderUnlinkedSafe(t *testing.T) {
	pl := NewPriceLevel(100)
	order1 := &Order{ID: 1, Symbol: "AAPL", Side: Buy, Type: Limit, Price: 100, Amount: 10}
	order2 := &Order{ID: 2, Symbol: "AAPL", Side: Buy, Type: Limit, Price: 100, Amount: 20}
	pl.AddOrder(order1)
	pl.AddOrder(order2)

	unlinked := &Order{ID: 3, Symbol: "AAPL", Side: Buy, Type: Limit, Price: 100, Amount: 50}
	pl.RemoveOrder(unlinked)

	if pl.Head != order1 || pl.Tail != order2 {
		t.Fatalf("unlinked order corrupted price level queue: head=%v tail=%v", pl.Head, pl.Tail)
	}
	if pl.TotalVolume != 30 {
		t.Fatalf("expected total volume 30, got %d", pl.TotalVolume)
	}
}

func TestOrderBookRejectsMismatchedSymbol(t *testing.T) {
	ob := NewOrderBook("AAPL")
	_, err := ob.ProcessOrder(&Order{ID: 1, Symbol: "TSLA", Side: Buy, Type: Limit, Price: 100, Amount: 10})
	if !errors.Is(err, ErrInvalidOrder) {
		t.Fatalf("expected ErrInvalidOrder for symbol mismatch, got %v", err)
	}
}

func TestRemoveOrderPriceMismatch(t *testing.T) {
	pl := NewPriceLevel(100)
	order1 := &Order{ID: 1, Symbol: "AAPL", Side: Buy, Type: Limit, Price: 100, Amount: 10}
	pl.AddOrder(order1)

	// Order with different price
	wrongPriceOrder := &Order{ID: 2, Symbol: "AAPL", Side: Buy, Type: Limit, Price: 200, Amount: 10}
	pl.RemoveOrder(wrongPriceOrder)

	if pl.TotalVolume != 10 {
		t.Fatalf("expected TotalVolume 10, got %d", pl.TotalVolume)
	}
	if pl.Head != order1 || pl.Tail != order1 {
		t.Fatalf("queue corrupted by removing mismatched price order")
	}
}

func TestDuplicateOrderIDAfterExecution(t *testing.T) {
	ob := NewOrderBook("AAPL")

	// 1. Sell order rests on book
	_, err := ob.ProcessOrder(&Order{ID: 100, Symbol: "AAPL", Side: Sell, Type: Limit, Price: 15000, Amount: 10})
	if err != nil {
		t.Fatalf("unexpected error placing sell: %v", err)
	}

	// 2. Buy order matches fully
	_, err = ob.ProcessOrder(&Order{ID: 101, Symbol: "AAPL", Side: Buy, Type: Limit, Price: 15000, Amount: 10})
	if err != nil {
		t.Fatalf("unexpected error placing buy: %v", err)
	}

	// Neither order is in ob.Orders anymore
	if _, ok := ob.Orders[100]; ok {
		t.Fatalf("order 100 should have been removed after full match")
	}
	if _, ok := ob.Orders[101]; ok {
		t.Fatalf("order 101 should not be in ob.Orders after full match")
	}

	// 3. Re-submitting order 100 must be rejected
	_, err = ob.ProcessOrder(&Order{ID: 100, Symbol: "AAPL", Side: Buy, Type: Limit, Price: 14000, Amount: 5})
	if !errors.Is(err, ErrDuplicateOrderID) {
		t.Fatalf("expected ErrDuplicateOrderID when reusing executed ID 100, got %v", err)
	}

	// 4. Re-submitting order 101 must also be rejected
	_, err = ob.ProcessOrder(&Order{ID: 101, Symbol: "AAPL", Side: Sell, Type: Limit, Price: 16000, Amount: 5})
	if !errors.Is(err, ErrDuplicateOrderID) {
		t.Fatalf("expected ErrDuplicateOrderID when reusing executed ID 101, got %v", err)
	}
}

func TestDuplicateOrderIDAfterCancellation(t *testing.T) {
	ob := NewOrderBook("AAPL")

	_, err := ob.ProcessOrder(&Order{ID: 200, Symbol: "AAPL", Side: Buy, Type: Limit, Price: 14000, Amount: 10})
	if err != nil {
		t.Fatalf("unexpected error placing buy: %v", err)
	}

	if err := ob.CancelOrder(200); err != nil {
		t.Fatalf("unexpected failure cancelling order: %v", err)
	}

	if _, exists := ob.Orders[200]; exists {
		t.Fatalf("order 200 should have been removed from ob.Orders after cancellation")
	}

	// Reusing cancelled ID 200 must be rejected
	_, err = ob.ProcessOrder(&Order{ID: 200, Symbol: "AAPL", Side: Buy, Type: Limit, Price: 14000, Amount: 5})
	if !errors.Is(err, ErrDuplicateOrderID) {
		t.Fatalf("expected ErrDuplicateOrderID when reusing cancelled ID 200, got %v", err)
	}
}

func TestGetOpenOrders(t *testing.T) {
	eng := NewEngine()
	eng.RegisterSymbol("AAPL")
	ob, _ := eng.GetOrderBook("AAPL")

	ob.ProcessOrder(&Order{ID: 1, Symbol: "AAPL", Side: Buy, Type: Limit, Price: 14000, Amount: 5, Timestamp: 1000})
	ob.ProcessOrder(&Order{ID: 2, Symbol: "AAPL", Side: Buy, Type: Limit, Price: 14500, Amount: 10, Timestamp: 2000})
	ob.ProcessOrder(&Order{ID: 3, Symbol: "AAPL", Side: Sell, Type: Limit, Price: 15500, Amount: 8, Timestamp: 3000})

	openOrders := ob.GetOpenOrders()
	if len(openOrders) != 3 {
		t.Fatalf("expected 3 open orders, got %d", len(openOrders))
	}

	ordersFromEng, ok := eng.GetOpenOrders("AAPL")
	if !ok || len(ordersFromEng) != 3 {
		t.Fatalf("expected 3 open orders from engine, got ok=%v, len=%d", ok, len(ordersFromEng))
	}

	// Verify unknown symbol
	_, ok = eng.GetOpenOrders("UNKNOWN")
	if ok {
		t.Fatalf("expected ok=false for unknown symbol open orders")
	}
}

func TestSelfTradePrevention(t *testing.T) {
	ob := NewOrderBook("AAPL")

	// 1. Trader 42 places a resting sell of 10 shares @ $100
	aliceSell := &Order{
		ID:        1,
		TraderID:  42,
		Symbol:    "AAPL",
		Side:      Sell,
		Type:      Limit,
		Price:     100,
		Amount:    10,
		Timestamp: time.Now().UnixMilli(),
	}
	if _, err := ob.ProcessOrder(aliceSell); err != nil {
		t.Fatalf("failed placing aliceSell: %v", err)
	}

	// 2. Different participant (Trader 99) buys 4 shares @ $100 -> must match
	bobBuy := &Order{
		ID:        2,
		TraderID:  99,
		Symbol:    "AAPL",
		Side:      Buy,
		Type:      Limit,
		Price:     100,
		Amount:    4,
		Timestamp: time.Now().UnixMilli(),
	}
	trades, err := ob.ProcessOrder(bobBuy)
	if err != nil || len(trades) != 1 || trades[0].Amount != 4 {
		t.Fatalf("expected Bob's trade to execute 4 shares, got trades=%v err=%v", trades, err)
	}

	// Alice should have 6 shares remaining
	if ob.Orders[1].Amount != 6 {
		t.Fatalf("expected Alice to have 6 shares left, got %d", ob.Orders[1].Amount)
	}

	// 3. Same participant (Trader 42) submits an aggressive buy @ $100 -> Self-Trade Prevention must trigger!
	aliceBuy := &Order{
		ID:        3,
		TraderID:  42, // Same trader ID
		Symbol:    "AAPL",
		Side:      Buy,
		Type:      Limit,
		Price:     100,
		Amount:    6,
		Timestamp: time.Now().UnixMilli(),
	}
	stpTrades, err := ob.ProcessOrder(aliceBuy)
	if err != nil {
		t.Fatalf("unexpected error on STP order: %v", err)
	}
	if len(stpTrades) != 0 {
		t.Fatalf("STP violation: order matched against same TraderID! Generated trades: %v", stpTrades)
	}

	// Alice's resting sell order must still have 6 shares intact
	if ob.Orders[1].Amount != 6 {
		t.Fatalf("resting order volume should remain intact after STP, got %d", ob.Orders[1].Amount)
	}

	// Alice's buy order should NOT rest on the book (Cancel-Taker policy)
	if _, exists := ob.Orders[3]; exists {
		t.Fatalf("STP taker order should not be resting on the book")
	}
}

// TestOrderLatencyDistribution samples 100,000 matching cycles to calculate the true
// latency distribution (p50, p90, p99, p99.9). Financial systems demand predictable tail
// latency, not just mean averages that hide GC or lock spikes.
func TestOrderLatencyDistribution(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping latency distribution test in short mode")
	}

	ob := NewOrderBook("AAPL")
	nextID := uint64(1)

	// Pre-seed liquidity
	for i := 0; i < 500; i++ {
		_, _ = ob.ProcessOrder(&Order{
			ID: nextID, Symbol: "AAPL", Side: Sell,
			Price: uint64(100 + (i % 20)), Amount: 50, Timestamp: int64(i),
		})
		nextID++
	}

	const samples = 100_000
	latencies := make([]int64, samples)

	for i := 0; i < samples; i++ {
		buyOrder := &Order{
			ID: nextID, Symbol: "AAPL", Side: Buy, Type: Limit,
			Price: 110, Amount: 1, Timestamp: int64(i),
		}
		nextID++

		start := time.Now()
		trades, err := ob.ProcessOrder(buyOrder)
		elapsed := time.Since(start).Nanoseconds()

		if err != nil || len(trades) != 1 {
			t.Fatalf("unexpected matching failure: %v", err)
		}
		latencies[i] = elapsed

		// Replenish
		_, _ = ob.ProcessOrder(&Order{
			ID: nextID, Symbol: "AAPL", Side: Sell, Type: Limit,
			Price: trades[0].Price, Amount: 1, Timestamp: int64(i),
		})
		nextID++
	}

	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })

	p50 := latencies[int(float64(samples)*0.50)]
	p90 := latencies[int(float64(samples)*0.90)]
	p99 := latencies[int(float64(samples)*0.99)]
	p999 := latencies[int(float64(samples)*0.999)]
	maxLat := latencies[samples-1]

	t.Logf("=== Matching Latency Distribution (N=%d) ===", samples)
	t.Logf("  p50   : %d ns", p50)
	t.Logf("  p90   : %d ns", p90)
	t.Logf("  p99   : %d ns", p99)
	t.Logf("  p99.9 : %d ns", p999)
	t.Logf("  max   : %d ns", maxLat)

	if p99 > 10_000 { // 10 microseconds threshold
		t.Errorf("p99 latency exceeded 10us SLA: got %d ns", p99)
	}
}

// FuzzOrderMatchingInvariants runs property-based fuzz testing across randomized order streams,
// rigorously validating four financial and structural invariants after every transaction:
// 1. Cross-Book Invariant: BestBid.Price < BestAsk.Price (resting books never cross).
// 2. Volume Conservation: PriceLevel.TotalVolume == sum of resting orders at that level.
// 3. Intrusive DLL Symmetry: For every order, order.Next.Prev == order and order.Prev.Next == order.
// 4. Index Consistency: Every order in ob.Orders must exist with positive amount and consistent ID.
func FuzzOrderMatchingInvariants(f *testing.F) {
	f.Add(uint8(0), uint8(0), uint64(100), uint64(10))
	f.Add(uint8(1), uint8(0), uint64(105), uint64(15))
	f.Add(uint8(0), uint8(1), uint64(0), uint64(5))
	f.Add(uint8(1), uint8(1), uint64(0), uint64(20))

	ob := NewOrderBook("AAPL")
	nextID := uint64(1)

	f.Fuzz(func(t *testing.T, rawSide uint8, rawType uint8, rawPrice uint64, rawAmount uint64) {
		side := Side(rawSide % 2)
		orderType := OrderType(rawType % 2)
		price := (rawPrice % 1000) + 1
		amount := (rawAmount % 500) + 1

		order := &Order{
			ID:        nextID,
			Symbol:    "AAPL",
			Side:      side,
			Type:      orderType,
			Price:     price,
			Amount:    amount,
			Timestamp: time.Now().UnixMilli(),
		}
		nextID++

		_, _ = ob.ProcessOrder(order)

		// Invariant 1: Resting books must never cross (bids strictly below asks)
		bid, hasBid, ask, hasAsk := ob.TopOfBook()
		if hasBid && hasAsk && bid >= ask {
			t.Fatalf("INVARIANT VIOLATION (Crossed Book): best bid %d >= best ask %d", bid, ask)
		}

		// Invariant 2 & 3: Level Volume & Pointer Symmetry
		for _, level := range ob.Bids {
			var vol uint64
			curr := level.Head
			var prev *Order
			for curr != nil {
				vol += curr.Amount
				if curr.Prev != prev {
					t.Fatalf("INVARIANT VIOLATION: Bid DLL prev pointer broken at order %d", curr.ID)
				}
				prev = curr
				curr = curr.Next
			}
			if level.Tail != prev {
				t.Fatalf("INVARIANT VIOLATION: Bid DLL tail mismatch")
			}
			if vol != level.TotalVolume {
				t.Fatalf("INVARIANT VIOLATION: Bid volume mismatch: computed %d, level has %d", vol, level.TotalVolume)
			}
		}

		for _, level := range ob.Asks {
			var vol uint64
			curr := level.Head
			var prev *Order
			for curr != nil {
				vol += curr.Amount
				if curr.Prev != prev {
					t.Fatalf("INVARIANT VIOLATION: Ask DLL prev pointer broken at order %d", curr.ID)
				}
				prev = curr
				curr = curr.Next
			}
			if level.Tail != prev {
				t.Fatalf("INVARIANT VIOLATION: Ask DLL tail mismatch")
			}
			if vol != level.TotalVolume {
				t.Fatalf("INVARIANT VIOLATION: Ask volume mismatch: computed %d, level has %d", vol, level.TotalVolume)
			}
		}

		// Invariant 4: Map consistency
		for id, ord := range ob.Orders {
			if ord.ID != id {
				t.Fatalf("INVARIANT VIOLATION: Map ID mismatch: key %d != order %d", id, ord.ID)
			}
			if ord.Amount == 0 {
				t.Fatalf("INVARIANT VIOLATION: Resting order %d has 0 amount", id)
			}
		}
	})
}
