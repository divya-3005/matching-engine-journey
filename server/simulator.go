package main

import (
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/divya-3005/OME/server/engine"
)

type MarketSimulator struct {
	eng *engine.Engine
	hub *Hub

	mu      sync.Mutex
	running bool
	stop    chan struct{} // closed to stop the current run
	done    chan struct{} // closed by the current run's goroutine when it exits
}

func NewMarketSimulator(eng *engine.Engine, hub *Hub) *MarketSimulator {
	return &MarketSimulator{eng: eng, hub: hub}
}

func referenceMid(symbol string) (uint64, bool) {
	switch symbol {
	case "AAPL":
		return 15000, true // $150.00
	case "TSLA":
		return 24000, true // $240.00
	case "BTC-USD":
		return 6400000, true // $64,000.00
	}
	return 0, false
}

func tickSize(symbol string) uint64 {
	switch symbol {
	case "BTC-USD":
		return 1000 // $10.00 per level
	default:
		return 10 // $0.10 per level
	}
}

// placeLadder rests 10 limit orders stepping away from anchor by tickSize per level
// (below anchor for buys, above anchor for sells).
func (sim *MarketSimulator) placeLadder(symbol string, side engine.Side, anchor uint64) {
	step := tickSize(symbol)
	for i := 1; i <= 10; i++ {
		diff := uint64(i) * step
		var price uint64
		if side == engine.Buy {
			if anchor <= diff {
				break // never place a zero or wrapped-around price
			}
			price = anchor - diff
		} else {
			price = anchor + diff
		}
		order := &engine.Order{
			ID:        sim.eng.NextOrderID(),
			Symbol:    symbol,
			Side:      side,
			Type:      engine.Limit,
			Price:     price,
			Amount:    uint64(5 + rand.Intn(25)),
			Timestamp: time.Now().UnixMilli(),
		}
		trades, err := sim.eng.ProcessOrder(order)
		if err != nil {
			log.Printf("simulator: failed to place ladder order %d for %s: %v", order.ID, symbol, err)
		} else {
			publishOrderEvents(sim.hub, symbol, trades, order.Type == engine.Limit && order.Amount > 0)
		}
	}
}

// SeedSymbol pre-populates 10 bids and 10 asks around the symbol's reference mid price.
func (sim *MarketSimulator) SeedSymbol(symbol string) {
	mid, ok := referenceMid(symbol)
	if !ok {
		return
	}
	sim.placeLadder(symbol, engine.Buy, mid)
	sim.placeLadder(symbol, engine.Sell, mid)
}

// SeedMarket pre-populates realistic bids and asks for all symbols.
func (sim *MarketSimulator) SeedMarket() {
	for _, sym := range supportedSymbols {
		sim.SeedSymbol(sym)
	}
}

// EnsureLiquidity re-seeds whichever side of the book is empty so the simulator never stalls.
func (sim *MarketSimulator) EnsureLiquidity(symbol string) {
	ob, ok := sim.eng.GetOrderBook(symbol)
	if !ok {
		return
	}
	bid, hasBid, ask, hasAsk := ob.TopOfBook()
	switch {
	case !hasBid && !hasAsk:
		sim.SeedSymbol(symbol)
	case !hasAsk:
		sim.placeLadder(symbol, engine.Sell, bid)
	case !hasBid:
		step := tickSize(symbol)
		if ask <= step {
			sim.SeedSymbol(symbol)
		} else {
			sim.placeLadder(symbol, engine.Buy, ask)
		}
	}
}

// Start begins periodic order placement. Calling it while running is a no-op.
func (sim *MarketSimulator) Start() {
	sim.mu.Lock()
	defer sim.mu.Unlock()
	sim.startLocked()
}

func (sim *MarketSimulator) startLocked() {
	if sim.running {
		return
	}
	stop, done := make(chan struct{}), make(chan struct{})
	sim.stop, sim.done, sim.running = stop, done, true
	go sim.loop(stop, done)
}

// Stop halts the simulator and waits until its goroutine has fully exited.
func (sim *MarketSimulator) Stop() {
	sim.mu.Lock()
	if !sim.running {
		sim.mu.Unlock()
		return
	}
	close(sim.stop)
	done := sim.done
	sim.running = false
	sim.mu.Unlock()
	<-done
}

// Toggle toggles the simulator on or off and returns the new running state.
func (sim *MarketSimulator) Toggle() bool {
	sim.mu.Lock()
	if sim.running {
		close(sim.stop)
		done := sim.done
		sim.running = false
		sim.mu.Unlock()
		<-done
		return false
	}
	sim.startLocked()
	sim.mu.Unlock()
	return true
}

func (sim *MarketSimulator) IsRunning() bool {
	sim.mu.Lock()
	defer sim.mu.Unlock()
	return sim.running
}

// loop only ever reads its own stop channel (passed by value), never the sim.stop field.
func (sim *MarketSimulator) loop(stop <-chan struct{}, done chan<- struct{}) {
	defer close(done)
	ticker := time.NewTicker(1200 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			sim.step()
		}
	}
}

func (sim *MarketSimulator) step() {
	symbol := supportedSymbols[rand.Intn(len(supportedSymbols))]
	ob, exists := sim.eng.GetOrderBook(symbol)
	if !exists {
		return
	}

	bestBid, hasBid, bestAsk, hasAsk := ob.TopOfBook()
	if !hasBid || !hasAsk {
		sim.EnsureLiquidity(symbol)
		return
	}

	step := tickSize(symbol)
	side := engine.Side(rand.Intn(2))
	qty := uint64(1 + rand.Intn(6))

	var orderType engine.OrderType
	var price uint64

	// 1. If spread is wider than 2 ticks, tighten it with a limit order inside the spread
	if bestAsk > bestBid+2*step {
		orderType = engine.Limit
		if side == engine.Buy {
			price = bestBid + step
		} else {
			price = bestAsk - step
		}
	} else if rand.Float32() < 0.45 {
		// 2. 45% Taker orders: cross the spread to trigger an immediate FIFO match
		orderType = engine.Market
		if side == engine.Buy {
			price = bestAsk
		} else {
			price = bestBid
		}
	} else {
		// 3. 55% Maker orders: replenish resting liquidity 0 to 2 ticks from the spread
		orderType = engine.Limit
		offset := uint64(rand.Intn(3)) * step
		if side == engine.Buy {
			if bestBid > offset {
				price = bestBid - offset
			} else {
				price = 1
			}
		} else {
			price = bestAsk + offset
		}
	}

	order := &engine.Order{
		ID:        sim.eng.NextOrderID(),
		Symbol:    symbol,
		Side:      side,
		Type:      orderType,
		Price:     price,
		Amount:    qty,
		Timestamp: time.Now().UnixMilli(),
	}
	trades, err := sim.eng.ProcessOrder(order)
	if err != nil {
		log.Printf("simulator: order %d rejected: %v", order.ID, err)
		return
	}
	publishOrderEvents(sim.hub, symbol, trades, order.Type == engine.Limit && order.Amount > 0)
}
