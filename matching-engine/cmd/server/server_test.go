package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/divya-3005/matching-engine/internal/engine"
)

// newTestServer returns a Server backed by a fresh engine.
// Using capacity 1 mirrors the production configuration: the test helpers
// always call submitOrder (which does Submit + ProcessNext together), so the
// buffer never needs to hold more than one order.
func newTestServer(t *testing.T) *Server {
	t.Helper()
	return &Server{engine: engine.New(1)}
}

// post is a convenience helper: it calls submitOrder and returns the recorder.
func post(t *testing.T, s *Server, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(body))
	w := httptest.NewRecorder()
	s.submitOrder(w, req)
	return w
}

// mustPost submits an order and fails the test immediately if the handler
// does not return 200. Used to set up book state before the real assertion.
func mustPost(t *testing.T, s *Server, body string) {
	t.Helper()
	if w := post(t, s, body); w.Code != http.StatusOK {
		t.Fatalf("mustPost: expected 200, got %d: %s", w.Code, w.Body)
	}
}

// ---- GET /health ------------------------------------------------------------

func TestHealth(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	s.health(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["status"] != "ok" {
		t.Errorf("expected status=ok, got %q", resp["status"])
	}
}

// ---- POST /orders -----------------------------------------------------------

func TestSubmitOrder_ValidLimit(t *testing.T) {
	s := newTestServer(t)
	w := post(t, s, `{"id":1,"symbol":"AAPL","side":"BUY","type":"LIMIT","price":10000,"quantity":100}`)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body)
	}

	var resp struct {
		Order struct {
			ID        uint64 `json:"id"`
			Symbol    string `json:"symbol"`
			Side      string `json:"side"`
			Type      string `json:"type"`
			Remaining uint64 `json:"remaining"`
		} `json:"order"`
		Trades []interface{} `json:"trades"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if resp.Order.ID != 1 {
		t.Errorf("order.id: want 1, got %d", resp.Order.ID)
	}
	if resp.Order.Symbol != "AAPL" {
		t.Errorf("order.symbol: want AAPL, got %s", resp.Order.Symbol)
	}
	if resp.Order.Side != "BUY" {
		t.Errorf("order.side: want BUY, got %s", resp.Order.Side)
	}
	if resp.Order.Type != "LIMIT" {
		t.Errorf("order.type: want LIMIT, got %s", resp.Order.Type)
	}
	if resp.Order.Remaining != 100 {
		t.Errorf("order.remaining: want 100, got %d", resp.Order.Remaining)
	}
	// trades must be an empty array, not JSON null.
	if resp.Trades == nil {
		t.Error("trades: expected [], got null")
	}
	if len(resp.Trades) != 0 {
		t.Errorf("trades: want 0, got %d", len(resp.Trades))
	}
}

func TestSubmitOrder_TradeProduced(t *testing.T) {
	s := newTestServer(t)

	// Resting sell at 100 — will sit in the ask side.
	mustPost(t, s, `{"id":1,"symbol":"AAPL","side":"SELL","type":"LIMIT","price":100,"quantity":50}`)

	// Incoming buy at 100 — crosses the sell.
	w := post(t, s, `{"id":2,"symbol":"AAPL","side":"BUY","type":"LIMIT","price":100,"quantity":50}`)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body)
	}

	var resp struct {
		Trades []struct {
			BuyOrderID  uint64 `json:"buy_order_id"`
			SellOrderID uint64 `json:"sell_order_id"`
			Price       uint64 `json:"price"`
			Quantity    uint64 `json:"quantity"`
		} `json:"trades"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Trades) != 1 {
		t.Fatalf("trades: want 1, got %d", len(resp.Trades))
	}
	tr := resp.Trades[0]
	if tr.BuyOrderID != 2 {
		t.Errorf("trade.buy_order_id: want 2, got %d", tr.BuyOrderID)
	}
	if tr.SellOrderID != 1 {
		t.Errorf("trade.sell_order_id: want 1, got %d", tr.SellOrderID)
	}
	// Price is always the passive (resting) order's price.
	if tr.Price != 100 {
		t.Errorf("trade.price: want 100 (passive sell price), got %d", tr.Price)
	}
	if tr.Quantity != 50 {
		t.Errorf("trade.quantity: want 50, got %d", tr.Quantity)
	}
}

func TestSubmitOrder_MalformedJSON(t *testing.T) {
	s := newTestServer(t)
	w := post(t, s, "not valid json {{{")

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	assertErrorField(t, w)
}

func TestSubmitOrder_ValidationFailure(t *testing.T) {
	s := newTestServer(t)
	// quantity=0 is rejected by model.NewOrder.
	w := post(t, s, `{"id":1,"symbol":"AAPL","side":"BUY","type":"LIMIT","price":10000,"quantity":0}`)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	assertErrorField(t, w)
}

func TestSubmitOrder_BodyTooLarge(t *testing.T) {
	s := newTestServer(t)
	// Construct a body just over 1 MiB that is otherwise valid JSON.
	// Without http.MaxBytesReader the decoder would accept this body and the
	// handler would return 200; with it the handler must return 400.
	bigSymbol := strings.Repeat("A", 1<<20)
	body := `{"id":1,"symbol":"` + bigSymbol + `","side":"BUY","type":"LIMIT","price":10000,"quantity":1}`
	w := post(t, s, body)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for oversized body, got %d", w.Code)
	}
}

// ---- DELETE /orders/{id} ----------------------------------------------------

func TestCancelOrder_Exists(t *testing.T) {
	s := newTestServer(t)

	// Place a resting limit buy — it will not match (no asks).
	mustPost(t, s, `{"id":1,"symbol":"AAPL","side":"BUY","type":"LIMIT","price":10000,"quantity":100}`)

	req := httptest.NewRequest(http.MethodDelete, "/orders/1", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	s.cancelOrder(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body)
	}

	var resp struct {
		ID uint64 `json:"id"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.ID != 1 {
		t.Errorf("cancelled order id: want 1, got %d", resp.ID)
	}
}

func TestCancelOrder_NotFound(t *testing.T) {
	s := newTestServer(t)

	req := httptest.NewRequest(http.MethodDelete, "/orders/99", nil)
	req.SetPathValue("id", "99")
	w := httptest.NewRecorder()
	s.cancelOrder(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
	assertErrorField(t, w)
}

// ---- GET /book --------------------------------------------------------------

func TestGetBook_Empty(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/book", nil)
	w := httptest.NewRecorder()
	s.getBook(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp struct {
		Bids []interface{} `json:"bids"`
		Asks []interface{} `json:"asks"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Bids == nil {
		t.Error("bids: expected [], got null")
	}
	if resp.Asks == nil {
		t.Error("asks: expected [], got null")
	}
	if len(resp.Bids) != 0 || len(resp.Asks) != 0 {
		t.Errorf("expected empty book, got %d bids %d asks", len(resp.Bids), len(resp.Asks))
	}
}

func TestGetBook_WithRestingOrder(t *testing.T) {
	s := newTestServer(t)

	// Place a resting limit buy at 10000.
	mustPost(t, s, `{"id":1,"symbol":"AAPL","side":"BUY","type":"LIMIT","price":10000,"quantity":100}`)

	req := httptest.NewRequest(http.MethodGet, "/book", nil)
	w := httptest.NewRecorder()
	s.getBook(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp struct {
		Bids []struct {
			Price  uint64 `json:"price"`
			Orders []struct {
				ID        uint64 `json:"id"`
				Remaining uint64 `json:"remaining"`
			} `json:"orders"`
		} `json:"bids"`
		Asks []interface{} `json:"asks"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Bids) != 1 {
		t.Fatalf("bids: want 1 level, got %d", len(resp.Bids))
	}
	if resp.Bids[0].Price != 10000 {
		t.Errorf("bid price: want 10000, got %d", resp.Bids[0].Price)
	}
	if len(resp.Bids[0].Orders) != 1 {
		t.Fatalf("bid orders: want 1, got %d", len(resp.Bids[0].Orders))
	}
	if resp.Bids[0].Orders[0].ID != 1 {
		t.Errorf("bid order id: want 1, got %d", resp.Bids[0].Orders[0].ID)
	}
	if resp.Bids[0].Orders[0].Remaining != 100 {
		t.Errorf("bid order remaining: want 100, got %d", resp.Bids[0].Orders[0].Remaining)
	}
	if len(resp.Asks) != 0 {
		t.Errorf("asks: want 0, got %d", len(resp.Asks))
	}
}

// ---- helpers ----------------------------------------------------------------

// assertErrorField decodes the response body and checks that the "error" key
// is present, confirming the handler returned a structured error response.
func assertErrorField(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if _, ok := resp["error"]; !ok {
		t.Error("expected \"error\" field in response body")
	}
}
