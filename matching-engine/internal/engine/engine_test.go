package engine

import (
    "testing"

    "github.com/divya-3005/matching-engine/internal/model"
    "github.com/divya-3005/matching-engine/internal/orderbook"
)

func (e *Engine) OrderBook() *orderbook.OrderBook {
    return e.book
}

func TestSubmitAndProcess(t *testing.T) {
    e := New(2)

    order, err := model.NewOrder(
        1,
        "AAPL",
        model.Buy,
        model.Limit,
        15000,
        100,
    )
    if err != nil {
        t.Fatal(err)
    }

    if err := e.Submit(order); err != nil {
        t.Fatal(err)
    }

    _, err = e.ProcessNext()
    if err != nil {
        t.Fatal(err)
    }

    if len(e.OrderBook().Buys().Orders()) != 1 {
        t.Fatal("expected one buy order")
    }

    if len(e.OrderBook().Sells().Orders()) != 0 {
        t.Fatal("expected zero sell orders")
    }
}

func TestProcessEmptyQueue(t *testing.T) {
    e := New(1)

    if _, err := e.ProcessNext(); err == nil {
        t.Fatal("expected error")
    }
}

func TestTradeExecution(t *testing.T) {
    e := New(10)

    sell, err := model.NewOrder(
        1,
        "AAPL",
        model.Sell,
        model.Limit,
        100,
        10,
    )
    if err != nil {
        t.Fatal(err)
    }

    buy, err := model.NewOrder(
        2,
        "AAPL",
        model.Buy,
        model.Limit,
        100,
        10,
    )
    if err != nil {
        t.Fatal(err)
    }

    if err := e.Submit(sell); err != nil {
        t.Fatal(err)
    }

    if _, err := e.ProcessNext(); err != nil {
        t.Fatal(err)
    }

    if err := e.Submit(buy); err != nil {
        t.Fatal(err)
    }

    trades, err := e.ProcessNext()
    if err != nil {
        t.Fatal(err)
    }

    if len(trades) != 1 {
        t.Fatalf("expected one trade, got %d", len(trades))
    }

    trade := trades[0]
    if trade.BuyOrderID != 2 {
        t.Fatal("unexpected buy order")
    }

    if trade.SellOrderID != 1 {
        t.Fatal("unexpected sell order")
    }

    if trade.Price != 100 {
        t.Fatal("unexpected trade price")
    }
}

func TestProcessNextMatchesMultipleRestingOrders(t *testing.T) {
    e := New(10)

    sell1, err := model.NewOrder(1, "AAPL", model.Sell, model.Limit, 100, 30)
    if err != nil {
        t.Fatal(err)
    }
    sell2, err := model.NewOrder(2, "AAPL", model.Sell, model.Limit, 100, 40)
    if err != nil {
        t.Fatal(err)
    }
    sell3, err := model.NewOrder(3, "AAPL", model.Sell, model.Limit, 100, 50)
    if err != nil {
        t.Fatal(err)
    }

    if err := e.Submit(sell1); err != nil {
        t.Fatal(err)
    }
    if _, err := e.ProcessNext(); err != nil {
        t.Fatal(err)
    }

    if err := e.Submit(sell2); err != nil {
        t.Fatal(err)
    }
    if _, err := e.ProcessNext(); err != nil {
        t.Fatal(err)
    }

    if err := e.Submit(sell3); err != nil {
        t.Fatal(err)
    }
    if _, err := e.ProcessNext(); err != nil {
        t.Fatal(err)
    }

    buy, err := model.NewOrder(4, "AAPL", model.Buy, model.Limit, 100, 100)
    if err != nil {
        t.Fatal(err)
    }
    if err := e.Submit(buy); err != nil {
        t.Fatal(err)
    }

    trades, err := e.ProcessNext()
    if err != nil {
        t.Fatal(err)
    }

    if len(trades) != 3 {
        t.Fatalf("expected 3 trades, got %d", len(trades))
    }
    if trades[0].Quantity != 30 || trades[1].Quantity != 40 || trades[2].Quantity != 30 {
        t.Fatalf("expected quantities 30, 40, 30, got %d, %d, %d",
            trades[0].Quantity, trades[1].Quantity, trades[2].Quantity)
    }

    if buy.Remaining != 0 {
        t.Fatalf("expected buy order remaining 0, got %d", buy.Remaining)
    }

    level, ok := e.OrderBook().Sells().Level(100)
    if !ok {
        t.Fatal("expected remaining sell price level")
    }
    if level.Len() != 1 {
        t.Fatalf("expected one remaining sell order, got %d", level.Len())
    }
    if level.Front().Remaining != 20 {
        t.Fatalf("expected remaining quantity 20, got %d", level.Front().Remaining)
    }
}
