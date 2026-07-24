// simulator generates synthetic orders, feeds them into the matching engine,
// and prints a concise performance summary.
//
// Usage:
//
//	go run ./cmd/simulator [flags]
//
// Flags:
//
//	-orders        total number of orders to generate (default 10000)
//	-buy-ratio     fraction of orders that are buys, 0.0–1.0 (default 0.5)
//	-market-ratio  fraction of orders that are market orders, 0.0–1.0 (default 0.10)
//	-start-price   reference mid price in integer units, e.g. cents (default 10000)
//	-spread        maximum price deviation from start price in either direction (default 50)
//	-seed          random seed; omit for a non-deterministic run
package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/divya-3005/matching-engine/internal/engine"
	"github.com/divya-3005/matching-engine/internal/model"
)

func main() {
	numOrders := flag.Int("orders", 10_000,
		"total number of orders to generate")
	buyRatio := flag.Float64("buy-ratio", 0.5,
		"fraction of orders that are buys (0.0–1.0)")
	marketRatio := flag.Float64("market-ratio", 0.10,
		"fraction of orders that are market orders (0.0–1.0)")
	startPrice := flag.Int("start-price", 10_000,
		"reference mid price in integer units (e.g. cents)")
	spread := flag.Int("spread", 50,
		"maximum price deviation from start-price in either direction")
	seed := flag.Int64("seed", time.Now().UnixNano(),
		"random seed; set explicitly for deterministic runs")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: simulator [flags]\n\nFlags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	// Validate flags.
	switch {
	case *numOrders <= 0:
		fatal("orders must be greater than 0")
	case *buyRatio < 0 || *buyRatio > 1:
		fatal("buy-ratio must be between 0.0 and 1.0")
	case *marketRatio < 0 || *marketRatio > 1:
		fatal("market-ratio must be between 0.0 and 1.0")
	case *startPrice <= 0:
		fatal("start-price must be greater than 0")
	case *spread < 0:
		fatal("spread must be >= 0")
	case *spread >= *startPrice:
		fatal("spread must be less than start-price to avoid non-positive prices")
	}

	rng := rand.New(rand.NewSource(*seed))

	// Capacity 1: we submit one order then immediately call ProcessNext, so the
	// ring buffer never needs to hold more than one order at a time.
	e := engine.New(1)

	var (
		ordersSubmitted int
		tradesExecuted  int
		volumeTraded    uint64
	)

	start := time.Now()

	for i := 0; i < *numOrders; i++ {
		side := model.Buy
		if rng.Float64() >= *buyRatio {
			side = model.Sell
		}

		var (
			orderType model.OrderType
			price     uint64
		)

		if rng.Float64() < *marketRatio {
			// Market orders have no price; they execute against whatever is resting.
			orderType = model.Market
		} else {
			// Limit orders: price varies uniformly within ±spread of the reference price.
			// Both buys and sells draw from the same range, so the book naturally
			// produces a mix of crossing orders (which trade) and non-crossing orders
			// (which rest).
			orderType = model.Limit
			dev := 0
			if *spread > 0 {
				dev = rng.Intn(2*(*spread)+1) - *spread
			}
			p := int64(*startPrice) + int64(dev)
			if p < 1 {
				p = 1
			}
			price = uint64(p)
		}

		qty := uint64(rng.Intn(100) + 1)

		order, err := model.NewOrder(uint64(i+1), "SIM", side, orderType, price, qty)
		if err != nil {
			// Unreachable given the validation above, but guard defensively.
			continue
		}

		if err := e.Submit(order); err != nil {
			// ErrBufferFull: cannot happen with capacity=1 and ProcessNext below.
			continue
		}
		ordersSubmitted++

		trades, err := e.ProcessNext()
		if err != nil {
			continue
		}

		tradesExecuted += len(trades)
		for _, t := range trades {
			volumeTraded += t.Quantity
		}
	}

	elapsed := time.Since(start)

	restingBuys := len(e.OrderBook().BuyOrders())
	restingSells := len(e.OrderBook().SellOrders())
	restingTotal := restingBuys + restingSells

	printSummary(
		*numOrders, *buyRatio, *marketRatio, *startPrice, *spread, *seed,
		ordersSubmitted, tradesExecuted, volumeTraded,
		restingTotal, restingBuys, restingSells,
		elapsed,
	)
}

func printSummary(
	numOrders int, buyRatio, marketRatio float64, startPrice, spread int, seed int64,
	ordersSubmitted, tradesExecuted int, volumeTraded uint64,
	restingTotal, restingBuys, restingSells int,
	elapsed time.Duration,
) {
	const col = 26 // left-column width for labels
	sep := strings.Repeat("─", 48)

	fmt.Println()
	fmt.Println("  Matching Engine Simulator")
	fmt.Println("  " + sep)

	fmt.Println()
	fmt.Println("  Configuration")
	fmt.Printf("    %-*s %d\n", col, "Orders:", numOrders)
	fmt.Printf("    %-*s %.0f%%\n", col, "Buy ratio:", buyRatio*100)
	fmt.Printf("    %-*s %.0f%%\n", col, "Market ratio:", marketRatio*100)
	fmt.Printf("    %-*s %d\n", col, "Start price:", startPrice)
	fmt.Printf("    %-*s ±%d\n", col, "Price spread:", spread)
	fmt.Printf("    %-*s %d\n", col, "Seed:", seed)

	fmt.Println()
	fmt.Println("  Results")
	fmt.Printf("    %-*s %d\n", col, "Orders submitted:", ordersSubmitted)
	fmt.Printf("    %-*s %d\n", col, "Trades executed:", tradesExecuted)
	fmt.Printf("    %-*s %d\n", col, "Volume traded:", volumeTraded)
	fmt.Printf("    %-*s %d  (%d bids · %d asks)\n", col,
		"Resting orders:", restingTotal, restingBuys, restingSells)

	fmt.Println()
	fmt.Println("  Performance")
	fmt.Printf("    %-*s %s\n", col, "Elapsed:", fmtDuration(elapsed))

	if s := elapsed.Seconds(); s > 0 {
		fmt.Printf("    %-*s %s orders/s\n", col, "Throughput:", fmtRate(float64(ordersSubmitted)/s))
		fmt.Printf("    %-*s %s trades/s\n", col, "Trade rate:", fmtRate(float64(tradesExecuted)/s))
	}

	fmt.Println()
}

// fatal writes msg to stderr, prints usage, then exits.
func fatal(msg string) {
	fmt.Fprintf(os.Stderr, "simulator: %s\n\n", msg)
	flag.Usage()
	os.Exit(1)
}

// fmtDuration formats a duration with an appropriate unit.
func fmtDuration(d time.Duration) string {
	switch {
	case d >= time.Second:
		return fmt.Sprintf("%.3f s", d.Seconds())
	case d >= time.Millisecond:
		return fmt.Sprintf("%.2f ms", float64(d)/float64(time.Millisecond))
	default:
		return fmt.Sprintf("%.2f µs", float64(d)/float64(time.Microsecond))
	}
}

// fmtRate formats a rate with M or K suffix when the value warrants it.
func fmtRate(f float64) string {
	switch {
	case f >= 1_000_000:
		return fmt.Sprintf("%.2fM", f/1_000_000)
	case f >= 1_000:
		return fmt.Sprintf("%.2fK", f/1_000)
	default:
		return fmt.Sprintf("%.0f", f)
	}
}
