package wal

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/divya-3005/matching-engine/internal/engine"
	"github.com/divya-3005/matching-engine/internal/model"
	"github.com/divya-3005/matching-engine/internal/orderbook"
)

// WAL is an append-only write-ahead log that records every order event.
// Each write is fsynced immediately to guarantee durability on crash.
type WAL struct {
	file *os.File
}

// New opens or creates a WAL at the given path.
func New(path string) (*WAL, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}

	return &WAL{file: file}, nil
}

func (w *WAL) writeLine(format string, args ...any) error {
	if w == nil || w.file == nil {
		return fmt.Errorf("wal is closed")
	}

	line := fmt.Sprintf(format+"\n", args...)
	if _, err := w.file.WriteString(line); err != nil {
		return err
	}

	return w.file.Sync()
}

// LogSubmit records an order submission event.
// Format: SUBMIT,<id>,<symbol>,<side>,<type>,<price>,<quantity>
func (w *WAL) LogSubmit(order *model.Order) error {
	return w.writeLine(
		"SUBMIT,%d,%s,%s,%s,%d,%d",
		order.ID,
		order.Symbol,
		order.Side.String(),
		order.Type.String(),
		order.Price,
		order.Quantity,
	)
}

// LogTrade records a trade execution event.
// Format: TRADE,<id>,<buyOrderID>,<sellOrderID>,<symbol>,<price>,<quantity>
func (w *WAL) LogTrade(trade *model.Trade) error {
	return w.writeLine(
		"TRADE,%d,%d,%d,%s,%d,%d",
		trade.ID,
		trade.BuyOrderID,
		trade.SellOrderID,
		trade.Symbol,
		trade.Price,
		trade.Quantity,
	)
}

// LogCancel records an order cancellation event.
// Format: CANCEL,<orderID>
func (w *WAL) LogCancel(orderID uint64) error {
	return w.writeLine("CANCEL,%d", orderID)
}

// Close flushes and closes the underlying file.
func (w *WAL) Close() error {
	if w == nil || w.file == nil {
		return nil
	}

	err := w.file.Close()
	w.file = nil
	return err
}

// Replay reads the WAL from the beginning and reconstructs engine state.
//
// SUBMIT records are replayed by re-submitting the order to the engine and
// calling ProcessNext, which re-derives all trades deterministically.
//
// TRADE records are output events produced by the engine during the original
// run. They are skipped during replay because replaying SUBMIT records will
// naturally reproduce the same trades.
//
// CANCEL records are replayed by cancelling the order from the book.
// If the order is not found (because it was filled before the cancel could
// execute), the cancel is silently ignored — this is valid during replay.
func (w *WAL) Replay(engine *engine.Engine) error {
	if w == nil || w.file == nil {
		return fmt.Errorf("wal is closed")
	}

	file, err := os.Open(w.file.Name())
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Split(line, ",")
		switch parts[0] {
		case "SUBMIT":
			if len(parts) != 7 {
				return fmt.Errorf("invalid SUBMIT record at line %d: %s", lineNumber, line)
			}

			id, err := strconv.ParseUint(parts[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid order id at line %d: %w", lineNumber, err)
			}

			side, err := model.ParseSide(parts[3])
			if err != nil {
				return fmt.Errorf("invalid side at line %d: %w", lineNumber, err)
			}

			orderType, err := model.ParseOrderType(parts[4])
			if err != nil {
				return fmt.Errorf("invalid order type at line %d: %w", lineNumber, err)
			}

			price, err := strconv.ParseUint(parts[5], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid price at line %d: %w", lineNumber, err)
			}

			quantity, err := strconv.ParseUint(parts[6], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid quantity at line %d: %w", lineNumber, err)
			}

			order, err := model.NewOrder(id, parts[2], side, orderType, price, quantity)
			if err != nil {
				return fmt.Errorf("invalid order at line %d: %w", lineNumber, err)
			}

			if err := engine.Submit(order); err != nil {
				return fmt.Errorf("engine submit failed at line %d: %w", lineNumber, err)
			}

			if _, err := engine.ProcessNext(); err != nil {
				return fmt.Errorf("engine process failed at line %d: %w", lineNumber, err)
			}

		case "TRADE":
			// TRADE records are output events; the engine reproduces them
			// deterministically when SUBMIT records are replayed.
			continue

		case "CANCEL":
			if len(parts) != 2 {
				return fmt.Errorf("invalid CANCEL record at line %d: %s", lineNumber, line)
			}

			orderID, err := strconv.ParseUint(parts[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid order id at line %d: %w", lineNumber, err)
			}

			_, err = engine.OrderBook().Cancel(orderID)
			if err != nil && !errors.Is(err, orderbook.ErrOrderNotFound) {
				// ErrOrderNotFound is expected when the order was filled before
				// the cancel was processed; any other error is a real failure.
				return fmt.Errorf("cancel failed at line %d: %w", lineNumber, err)
			}

		default:
			return fmt.Errorf("unknown record type at line %d: %s", lineNumber, parts[0])
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
