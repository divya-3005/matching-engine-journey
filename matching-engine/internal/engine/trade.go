package engine

import (
	"time"

	"github.com/divya-3005/matching-engine/internal/model"
)

// executeTrade creates a Trade record for a matched buy and sell order.
// The trade price is always the resting (passive) order's price.
// qty is the number of units exchanged, which may be less than either
// order's original quantity in the case of a partial fill.
func executeTrade(buy, sell *model.Order, qty uint64) *model.Trade {
	return &model.Trade{
		BuyOrderID:  buy.ID,
		SellOrderID: sell.ID,
		Symbol:      buy.Symbol,
		Price:       sell.Price,
		Quantity:    qty,
		Timestamp:   time.Now(),
	}
}
