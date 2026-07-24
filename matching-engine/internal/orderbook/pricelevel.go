package orderbook

import "github.com/divya-3005/matching-engine/internal/model"

// PriceLevel stores all orders at a single price in FIFO order.
type PriceLevel struct {
	price  uint64
	orders []*model.Order
}

func NewPriceLevel(price uint64) *PriceLevel {
	return &PriceLevel{
		price:  price,
		orders: make([]*model.Order, 0),
	}
}

func (p *PriceLevel) Price() uint64 {
	return p.price
}

func (p *PriceLevel) Add(order *model.Order) {
	p.orders = append(p.orders, order)
}

func (p *PriceLevel) Front() *model.Order {
	if len(p.orders) == 0 {
		return nil
	}

	return p.orders[0]
}

func (p *PriceLevel) Pop() *model.Order {
	if len(p.orders) == 0 {
		return nil
	}

	order := p.orders[0]
	p.orders = p.orders[1:]
	return order
}

func (p *PriceLevel) Len() int {
	return len(p.orders)
}

func (p *PriceLevel) Orders() []*model.Order {
	orders := make([]*model.Order, len(p.orders))
	copy(orders, p.orders)
	return orders
}

func (p *PriceLevel) Remove(orderID uint64) (*model.Order, bool) {
	for i, order := range p.orders {
		if order.ID == orderID {
			p.orders = append(p.orders[:i], p.orders[i+1:]...)
			return order, true
		}
	}

	return nil, false
}
