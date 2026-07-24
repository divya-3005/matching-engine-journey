package orderbook

import "github.com/divya-3005/matching-engine/internal/model"

// BookSide manages one side of the order book (buy or sell).
type BookSide struct {
	side   model.Side
	levels map[uint64]*PriceLevel
	index  *PriceIndex
}

func NewBookSide(side model.Side) *BookSide {
	return &BookSide{
		side:   side,
		levels: make(map[uint64]*PriceLevel),
		index:  NewPriceIndex(side == model.Buy),
	}
}

func (b *BookSide) Add(order *model.Order) {
	level, ok := b.levels[order.Price]
	if !ok {
		level = NewPriceLevel(order.Price)
		b.levels[order.Price] = level
		b.index.Insert(order.Price)
	}

	level.Add(order)
}

func (b *BookSide) Best() *PriceLevel {
	price, ok := b.index.Best()
	if !ok {
		return nil
	}

	return b.levels[price]
}

func (b *BookSide) Level(price uint64) (*PriceLevel, bool) {
	level, ok := b.levels[price]
	return level, ok
}

func (b *BookSide) Remove(price uint64) {
	delete(b.levels, price)
	b.index.Remove(price)
}

func (b *BookSide) Cancel(orderID uint64) (*model.Order, bool) {
	for price, level := range b.levels {
		order, ok := level.Remove(orderID)
		if !ok {
			continue
		}

		if level.Len() == 0 {
			b.Remove(price)
		}

		return order, true
	}

	return nil, false
}

func (b *BookSide) Len() int {
	return len(b.levels)
}

func (b *BookSide) Orders() []*model.Order {
	return flattenOrderLevels(b.levels)
}

func flattenOrderLevels(levels map[uint64]*PriceLevel) []*model.Order {
	orders := make([]*model.Order, 0)
	for _, level := range levels {
		orders = append(orders, level.orders...)
	}
	return orders
}
