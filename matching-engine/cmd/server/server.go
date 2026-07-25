package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"sync"

	"github.com/divya-3005/matching-engine/internal/engine"
	"github.com/divya-3005/matching-engine/internal/model"
	"github.com/divya-3005/matching-engine/internal/orderbook"
)

// Server wraps the matching engine and exposes it over HTTP.
// A mutex serialises all engine access — the engine is not goroutine-safe.
type Server struct {
	mu     sync.Mutex
	engine *engine.Engine
}

// ---- request / response types -----------------------------------------------

type submitRequest struct {
	ID       uint64 `json:"id"`
	Symbol   string `json:"symbol"`
	Side     string `json:"side"`
	Type     string `json:"type"`
	Price    uint64 `json:"price"`
	Quantity uint64 `json:"quantity"`
}

type orderResponse struct {
	ID        uint64 `json:"id"`
	Symbol    string `json:"symbol"`
	Side      string `json:"side"`
	Type      string `json:"type"`
	Price     uint64 `json:"price"`
	Quantity  uint64 `json:"quantity"`
	Remaining uint64 `json:"remaining"`
}

type tradeResponse struct {
	BuyOrderID  uint64 `json:"buy_order_id"`
	SellOrderID uint64 `json:"sell_order_id"`
	Symbol      string `json:"symbol"`
	Price       uint64 `json:"price"`
	Quantity    uint64 `json:"quantity"`
}

type levelResponse struct {
	Price  uint64          `json:"price"`
	Orders []orderResponse `json:"orders"`
}

// ---- handlers ---------------------------------------------------------------

// GET /health
func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// POST /orders
// Accepts a JSON order, submits it to the engine, runs the matching loop,
// and returns the accepted order plus any trades generated.
func (s *Server) submitOrder(w http.ResponseWriter, r *http.Request) {
	// Reject bodies larger than 1 MiB before the JSON decoder reads them.
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req submitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	side, err := model.ParseSide(req.Side)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	orderType, err := model.ParseOrderType(req.Type)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	order, err := model.NewOrder(req.ID, req.Symbol, side, orderType, req.Price, req.Quantity)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Lock only for engine access; release before writing the HTTP response.
	s.mu.Lock()
	submitErr := s.engine.Submit(order)
	var trades []*model.Trade
	var processErr error
	if submitErr == nil {
		trades, processErr = s.engine.ProcessNext()
	}
	s.mu.Unlock()

	if submitErr != nil {
		writeError(w, http.StatusServiceUnavailable, "engine queue full")
		return
	}
	if processErr != nil {
		writeError(w, http.StatusInternalServerError, processErr.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"order":  toOrderResponse(order),
		"trades": toTradeResponses(trades),
	})
}

// DELETE /orders/{id}
// Cancels a resting order. Returns 404 if the order is not in the book.
func (s *Server) cancelOrder(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id must be a positive integer")
		return
	}

	s.mu.Lock()
	order, err := s.engine.OrderBook().Cancel(id)
	s.mu.Unlock()

	if err != nil {
		if errors.Is(err, orderbook.ErrOrderNotFound) {
			writeError(w, http.StatusNotFound, "order not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, toOrderResponse(order))
}

// GET /book
// Returns the current resting bids and asks, ordered by price priority.
func (s *Server) getBook(w http.ResponseWriter, r *http.Request) {
	// Hold the lock for the full duration of level iteration: level.Orders()
	// reads the underlying slice which the matching loop may concurrently modify.
	s.mu.Lock()
	bids := toLevelResponses(s.engine.OrderBook().BuyLevels())
	asks := toLevelResponses(s.engine.OrderBook().SellLevels())
	s.mu.Unlock()

	writeJSON(w, http.StatusOK, map[string]any{
		"bids": bids,
		"asks": asks,
	})
}

// ---- conversion helpers -----------------------------------------------------

func toOrderResponse(o *model.Order) orderResponse {
	return orderResponse{
		ID:        o.ID,
		Symbol:    o.Symbol,
		Side:      o.Side.String(),
		Type:      o.Type.String(),
		Price:     o.Price,
		Quantity:  o.Quantity,
		Remaining: o.Remaining,
	}
}

// toTradeResponses returns an empty slice (not nil) when trades is empty,
// so the JSON response always encodes as [] rather than null.
func toTradeResponses(trades []*model.Trade) []tradeResponse {
	result := make([]tradeResponse, len(trades))
	for i, t := range trades {
		result[i] = tradeResponse{
			BuyOrderID:  t.BuyOrderID,
			SellOrderID: t.SellOrderID,
			Symbol:      t.Symbol,
			Price:       t.Price,
			Quantity:    t.Quantity,
		}
	}
	return result
}

// toLevelResponses copies all order data out of the book while the lock is
// held, producing plain value types that are safe to use after unlocking.
func toLevelResponses(levels []*orderbook.PriceLevel) []levelResponse {
	result := make([]levelResponse, len(levels))
	for i, level := range levels {
		orders := level.Orders() // returns a defensive copy
		orderResps := make([]orderResponse, len(orders))
		for j, o := range orders {
			orderResps[j] = toOrderResponse(o)
		}
		result[i] = levelResponse{
			Price:  level.Price(),
			Orders: orderResps,
		}
	}
	return result
}

// ---- HTTP helpers -----------------------------------------------------------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
