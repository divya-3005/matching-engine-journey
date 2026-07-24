package engine

import (
    "github.com/divya-3005/matching-engine/internal/model"
    "github.com/divya-3005/matching-engine/internal/orderbook"
    "github.com/divya-3005/matching-engine/internal/ringbuffer"
)

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

func (e *Engine) Submit(order *model.Order) error {
    return e.queue.Enqueue(order)
}

func canMatch(incoming, resting *model.Order) bool {
    if incoming.Side == model.Buy {
        return incoming.Price >= resting.Price
    }
    return resting.Price >= incoming.Price
}

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

    if order.Remaining > 0 {
        e.book.Add(order)
    }

    if len(trades) == 0 {
        return nil, nil
    }
    return trades, nil
}
