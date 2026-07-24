package orderbook

import "github.com/divya-3005/matching-engine/internal/model"

type OrderBook struct {
	buys  *BookSide
	sells *BookSide
}

func New() *OrderBook {
	return &OrderBook{
		buys:  NewBookSide(model.Buy),
		sells: NewBookSide(model.Sell),
	}
}

func (ob *OrderBook) Add(order *model.Order) {
	if order.Side == model.Buy {
		ob.buys.Add(order)
		return
	}

	ob.sells.Add(order)
}

func (ob *OrderBook) Buys() *BookSide {
	return ob.buys
}

func (ob *OrderBook) Sells() *BookSide {
	return ob.sells
}

func (ob *OrderBook) BuyOrders() []*model.Order {
	return ob.buys.Orders()
}

func (ob *OrderBook) SellOrders() []*model.Order {
	return ob.sells.Orders()
}
