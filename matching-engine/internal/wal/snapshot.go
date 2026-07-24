package wal

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/divya-3005/matching-engine/internal/model"
	"github.com/divya-3005/matching-engine/internal/orderbook"
)

type Snapshot struct {
	Buys  []SnapshotLevel `json:"buys"`
	Sells []SnapshotLevel `json:"sells"`
}

type SnapshotLevel struct {
	Price  uint64          `json:"price"`
	Orders []SnapshotOrder `json:"orders"`
}

type SnapshotOrder struct {
	ID        uint64 `json:"id"`
	Symbol    string `json:"symbol"`
	Side      string `json:"side"`
	Type      string `json:"type"`
	Price     uint64 `json:"price"`
	Remaining uint64 `json:"remaining"`
}

func SaveSnapshot(path string, book *orderbook.OrderBook) error {
	snapshot := Snapshot{
		Buys:  buildSnapshotLevels(book.BuyLevels()),
		Sells: buildSnapshotLevels(book.SellLevels()),
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(snapshot); err != nil {
		return err
	}

	return file.Sync()
}

func LoadSnapshot(path string) (*orderbook.OrderBook, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var snapshot Snapshot
	if err := json.NewDecoder(file).Decode(&snapshot); err != nil {
		return nil, err
	}

	book := orderbook.New()
	if err := restoreSnapshotLevels(book, snapshot.Buys, model.Buy); err != nil {
		return nil, err
	}
	if err := restoreSnapshotLevels(book, snapshot.Sells, model.Sell); err != nil {
		return nil, err
	}

	return book, nil
}

func buildSnapshotLevels(levels []*orderbook.PriceLevel) []SnapshotLevel {
	snapshotLevels := make([]SnapshotLevel, 0, len(levels))
	for _, level := range levels {
		orders := make([]SnapshotOrder, 0, len(level.Orders()))
		for _, order := range level.Orders() {
			orders = append(orders, SnapshotOrder{
				ID:        order.ID,
				Symbol:    order.Symbol,
				Side:      order.Side.String(),
				Type:      order.Type.String(),
				Price:     order.Price,
				Remaining: order.Remaining,
			})
		}

		snapshotLevels = append(snapshotLevels, SnapshotLevel{
			Price:  level.Price(),
			Orders: orders,
		})
	}

	return snapshotLevels
}

func restoreSnapshotLevels(book *orderbook.OrderBook, levels []SnapshotLevel, side model.Side) error {
	for _, level := range levels {
		for _, dto := range level.Orders {
			if dto.Price != level.Price {
				return fmt.Errorf("snapshot price mismatch for order %d", dto.ID)
			}

			orderType, err := model.ParseOrderType(dto.Type)
			if err != nil {
				return err
			}

			sideValue, err := model.ParseSide(dto.Side)
			if err != nil {
				return err
			}
			if sideValue != side {
				return fmt.Errorf("snapshot side mismatch for order %d", dto.ID)
			}

			order, err := model.NewOrder(dto.ID, dto.Symbol, sideValue, orderType, dto.Price, dto.Remaining)
			if err != nil {
				return err
			}

			book.Add(order)
		}
	}

	return nil
}
