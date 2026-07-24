package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

var counter int64

func increment(wg *sync.WaitGroup) {
	defer wg.Done()

	for i := 0; i < 1000; i++ {
		atomic.AddInt64(&counter, 1)
	}
}

func main() {
	var wg sync.WaitGroup

	wg.Add(2)

	go increment(&wg)
	go increment(&wg)

	wg.Wait()

	fmt.Println("Counter:", counter)
}