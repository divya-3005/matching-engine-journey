package wal

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/divya-3005/matching-engine/internal/engine"
	"github.com/divya-3005/matching-engine/internal/model"
)

type WAL struct {
	file *os.File
}

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

func (w *WAL) LogCancel(orderID uint64) error {
	return w.writeLine("CANCEL,%d", orderID)
}

func (w *WAL) Close() error {
	if w == nil || w.file == nil {
		return nil
	}

	err := w.file.Close()
	w.file = nil
	return err
}

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
			return fmt.Errorf("replay not implemented for TRADE records")
		case "CANCEL":
			return fmt.Errorf("replay not implemented for CANCEL records")
		default:
			return fmt.Errorf("unknown record type: %s", parts[0])
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
