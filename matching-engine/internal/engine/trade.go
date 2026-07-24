package engine

import "github.com/divya-3005/matching-engine/internal/model"

func min(a, b uint64) uint64 {
    if a < b {
        return a
    }
    return b
}

func executeTrade(buy, sell *model.Order, qty uint64) *model.Trade {
    return &model.Trade{
        BuyOrderID:  buy.ID,
        SellOrderID: sell.ID,
        Symbol:      buy.Symbol,
        Price:       sell.Price,
        Quantity:    qty,
    }
}
