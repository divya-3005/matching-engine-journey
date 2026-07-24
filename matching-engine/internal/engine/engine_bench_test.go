package engine

import (
	"fmt"
	"testing"

	"github.com/divya-3005/matching-engine/internal/model"
)

func BenchmarkProcessLimitOrders(b *testing.B) {
	e := New(1_000_000)
	b.ReportAllocs()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		order, err := model.NewOrder(
			uint64(i),
			"AAPL",
			model.Buy,
			model.Limit,
			100,
			100,
		)
		if err != nil {
			b.Fatal(err)
		}

		if err := e.Submit(order); err != nil {
			b.Fatal(err)
		}
		if _, err := e.ProcessNext(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSubmit(b *testing.B) {
	e := New(b.N + 1)
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		order, err := model.NewOrder(uint64(i), "AAPL", model.Buy, model.Limit, 100, 100)
		if err != nil {
			b.Fatal(err)
		}

		if err := e.Submit(order); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMatch(b *testing.B) {
	sizes := []int{100, 500, 1000}
	for _, size := range sizes {
		b.Run(fmt.Sprintf("resting-%d", size), func(b *testing.B) {
			b.ReportAllocs()

			restingOrders := make([]*model.Order, size)
			for j := 0; j < size; j++ {
				order, err := model.NewOrder(uint64(j+1), "AAPL", model.Sell, model.Limit, 100, 100)
				if err != nil {
					b.Fatal(err)
				}
				restingOrders[j] = order
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				e := New(size*2 + 1)
				for _, order := range restingOrders {
					if err := e.Submit(order); err != nil {
						b.Fatal(err)
					}
					if _, err := e.ProcessNext(); err != nil {
						b.Fatal(err)
					}
				}
				b.StartTimer()

				for j := 0; j < size; j++ {
					order, err := model.NewOrder(uint64(1000000+i*size+j), "AAPL", model.Buy, model.Limit, 100, 100)
					if err != nil {
						b.Fatal(err)
					}
					if err := e.Submit(order); err != nil {
						b.Fatal(err)
					}
					if _, err := e.ProcessNext(); err != nil {
						b.Fatal(err)
					}
				}
			}
		})
	}
}

func BenchmarkCancel(b *testing.B) {
	sizes := []int{100, 500, 1000}
	for _, size := range sizes {
		b.Run(fmt.Sprintf("orders-%d", size), func(b *testing.B) {
			b.ReportAllocs()

			orders := make([]*model.Order, size)
			for j := 0; j < size; j++ {
				order, err := model.NewOrder(uint64(j+1), "AAPL", model.Buy, model.Limit, 100, 100)
				if err != nil {
					b.Fatal(err)
				}
				orders[j] = order
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				e := New(size + 1)
				for _, order := range orders {
					if err := e.Submit(order); err != nil {
						b.Fatal(err)
					}
					if _, err := e.ProcessNext(); err != nil {
						b.Fatal(err)
					}
				}
				b.StartTimer()

				for j := 0; j < size; j++ {
					if _, err := e.OrderBook().Cancel(uint64(j + 1)); err != nil {
						b.Fatal(err)
					}
				}
			}
		})
	}
}
