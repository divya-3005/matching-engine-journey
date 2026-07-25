package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/divya-3005/matching-engine/internal/engine"
)

// BenchmarkHealth measures the handler overhead for a request that does no
// engine work: JSON encoding of a static value.
func BenchmarkHealth(b *testing.B) {
	b.ReportAllocs()
	s := &Server{engine: engine.New(1)}

	// The health handler does not read r.Body or modify r, so the same
	// request object can safely be reused across iterations.
	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		s.health(w, req)
	}
}

// BenchmarkSubmitOrder_NoMatch measures the cost of submitting a limit order
// that finds no counterpart and rests in the book.
// The same order ID is reused deliberately: the engine does not enforce
// uniqueness, so each order simply appends to the price level.
func BenchmarkSubmitOrder_NoMatch(b *testing.B) {
	b.ReportAllocs()
	s := &Server{engine: engine.New(1)}
	const body = `{"id":1,"symbol":"AAPL","side":"BUY","type":"LIMIT","price":10000,"quantity":100}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(body))
		w := httptest.NewRecorder()
		s.submitOrder(w, req)
	}
}

// BenchmarkSubmitOrder_WithTrade measures the cost of submitting a limit order
// that immediately crosses a resting order, producing one trade per call.
//
// b.N resting sells are pre-loaded before the timed section so that every
// iteration in the loop finds exactly one counterpart and generates one trade.
// Pre-computing the buy bodies avoids fmt.Sprintf allocations inside the hot
// loop while keeping each request body unique.
func BenchmarkSubmitOrder_WithTrade(b *testing.B) {
	b.ReportAllocs()
	s := &Server{engine: engine.New(1)}

	// Pre-load b.N resting sells. IDs 1..b.N.
	for i := 0; i < b.N; i++ {
		body := `{"id":` + strconv.Itoa(i+1) + `,"symbol":"AAPL","side":"SELL","type":"LIMIT","price":10000,"quantity":100}`
		req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(body))
		s.submitOrder(httptest.NewRecorder(), req)
	}

	// Pre-compute buy bodies with IDs that do not collide with the sells.
	// IDs b.N+1..2*b.N.
	buyBodies := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		buyBodies[i] = `{"id":` + strconv.Itoa(i+b.N+1) + `,"symbol":"AAPL","side":"BUY","type":"LIMIT","price":10000,"quantity":100}`
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(buyBodies[i]))
		w := httptest.NewRecorder()
		s.submitOrder(w, req)
	}
}

// BenchmarkGetBook measures the cost of serialising the current resting book
// to JSON. Ten bids at distinct prices are pre-loaded to produce a realistic
// multi-level response; the book is read-only during the timed section.
func BenchmarkGetBook(b *testing.B) {
	b.ReportAllocs()
	s := &Server{engine: engine.New(1)}

	for i := 0; i < 10; i++ {
		body := fmt.Sprintf(
			`{"id":%d,"symbol":"AAPL","side":"BUY","type":"LIMIT","price":%d,"quantity":100}`,
			i+1, 9990+i,
		)
		req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(body))
		s.submitOrder(httptest.NewRecorder(), req)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/book", nil)
		w := httptest.NewRecorder()
		s.getBook(w, req)
	}
}

// BenchmarkCancelOrder measures the cost of cancelling a single resting order.
//
// One order is placed and immediately cancelled per iteration. The setup
// (Submit + ProcessNext via submitOrder) is excluded from the timed section
// with b.StopTimer/b.StartTimer so the benchmark captures only the cancel
// handler path. Each iteration operates on a book that contains exactly one
// order, giving constant and comparable conditions throughout the run.
func BenchmarkCancelOrder(b *testing.B) {
	b.ReportAllocs()
	s := &Server{engine: engine.New(1)}
	var orderID int

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Place one resting buy — excluded from the timed window.
		b.StopTimer()
		orderID++
		idStr := strconv.Itoa(orderID)
		setupBody := `{"id":` + idStr + `,"symbol":"AAPL","side":"BUY","type":"LIMIT","price":10000,"quantity":100}`
		setupReq := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(setupBody))
		s.submitOrder(httptest.NewRecorder(), setupReq)
		b.StartTimer()

		// Cancel it — this is the measured operation.
		req := httptest.NewRequest(http.MethodDelete, "/orders/"+idStr, nil)
		req.SetPathValue("id", idStr)
		w := httptest.NewRecorder()
		s.cancelOrder(w, req)
	}
}
