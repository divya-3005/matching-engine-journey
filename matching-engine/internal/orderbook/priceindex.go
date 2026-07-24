package orderbook

import "sort"

// PriceIndex maintains an ordered set of prices.
type PriceIndex struct {
	prices []uint64
	desc   bool // true = highest first (bids), false = lowest first (asks)
}

func NewPriceIndex(desc bool) *PriceIndex {
	return &PriceIndex{
		prices: make([]uint64, 0),
		desc:   desc,
	}
}

func (p *PriceIndex) Insert(price uint64) {
	for _, existing := range p.prices {
		if existing == price {
			return
		}
	}

	p.prices = append(p.prices, price)
	sort.Slice(p.prices, func(i, j int) bool {
		if p.desc {
			return p.prices[i] > p.prices[j]
		}
		return p.prices[i] < p.prices[j]
	})
}

func (p *PriceIndex) Remove(price uint64) {
	for i, existing := range p.prices {
		if existing == price {
			p.prices = append(p.prices[:i], p.prices[i+1:]...)
			return
		}
	}
}

func (p *PriceIndex) Best() (uint64, bool) {
	if len(p.prices) == 0 {
		return 0, false
	}

	return p.prices[0], true
}

func (p *PriceIndex) Len() int {
	return len(p.prices)
}
