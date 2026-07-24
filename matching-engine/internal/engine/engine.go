package engine

import (
	"github.com/divya-3005/matching-engine/internal/model"
	"github.com/divya-3005/matching-engine/internal/orderbook"
	"github.com/divya-3005/matching-engine/internal/ringbuffer"
)

// Engine coordinates the processing of incoming orders.
// It owns the ring buffer (ingestion queue) and the order book (resting state).
// All operations are single-threaded; the caller is responsible for sequencing.
type Engine struct {
	queue *ringbuffer.RingBuffer[*model.Order]
	book  *orderbook.OrderBook
}

func New(capacity int) *Engine {
	return &Engine{
		queue: ringbuffer.New[*model.Order](capacity),
		book:  orderbook.New(),
	}
}

// Submit enqueues an order for processing. Returns ErrBufferFull if the ring
// buffer is at capacity.
func (e *Engine) Submit(order *model.Order) error {
	return e.queue.Enqueue(order)
}

// canMatch reports whether an incoming order can trade against a resting order
// on the opposite side of the book.
func canMatch(incoming, resting *model.Order) bool {
	if incoming.Type == model.Market {
		return true
	}

	if incoming.Side == model.Buy {
		return incoming.Price >= resting.Price
	}
	return resting.Price >= incoming.Price
}

// ProcessNext dequeues one order and runs the matching loop against the
// opposite side of the book. Limit orders with remaining quantity are parked
// in the book. Market orders with remaining quantity are discarded.
func (e *Engine) ProcessNext() ([]*model.Trade, error) {
	order, err := e.queue.Dequeue()
	if err != nil {
		return nil, err
	}

	trades := make([]*model.Trade, 0)
	var opposite *orderbook.BookSide

	if order.Side == model.Buy {
		opposite = e.book.Sells()
	} else {
		opposite = e.book.Buys()
	}

	for order.Remaining > 0 {
		best := opposite.Best()
		if best == nil {
			break
		}

		resting := best.Front()
		if resting == nil || !canMatch(order, resting) {
			break
		}

		qty := min(order.Remaining, resting.Remaining)
		var trade *model.Trade

		if order.Side == model.Buy {
			trade = executeTrade(order, resting, qty)
		} else {
			trade = executeTrade(resting, order, qty)
		}

		_ = order.Fill(qty)
		_ = resting.Fill(qty)

		trades = append(trades, trade)

		if resting.Filled() {
			best.Pop()
			if best.Len() == 0 {
				opposite.Remove(best.Price())
			}
		}
	}

	if order.Remaining > 0 && order.Type == model.Limit {
		e.book.Add(order)
	}

	return trades, nil
}

// OrderBook returns the engine's order book. Callers may use this to cancel
// orders or query resting state directly.
func (e *Engine) OrderBook() *orderbook.OrderBook {
	return e.book
}
