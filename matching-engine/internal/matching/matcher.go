package matching

import (
	"errors"

	"github.com/divya-3005/matching-engine/internal/model"
)

var ErrQuantityMismatch = errors.New("orders have different quantities")

type Matcher struct{}

func New() *Matcher {
	return &Matcher{}
}

func (m *Matcher) Match(
	buy *model.Order,
	sell *model.Order,
) (*model.Trade, error) {

	if buy.Side != model.Buy {
		return nil, errors.New("first order must be buy")
	}

	if sell.Side != model.Sell {
		return nil, errors.New("second order must be sell")
	}

	if buy.Price < sell.Price {
		return nil, nil
	}

	if buy.Quantity != sell.Quantity {
		return nil, ErrQuantityMismatch
	}

	trade := &model.Trade{
		BuyOrderID:  buy.ID,
		SellOrderID: sell.ID,
		Symbol:      buy.Symbol,
		Price:       sell.Price,
		Quantity:    buy.Quantity,
	}

	return trade, nil
}
