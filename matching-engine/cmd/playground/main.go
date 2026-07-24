package main

import (
	"fmt"
	"time"
)

func worker(id int) {
	for i := 1; i <= 5; i++ {
		fmt.Printf("Worker %d -> %d\n", id, i)
		time.Sleep(300 * time.Millisecond)
	}
}

func main() {

	go worker(1)
	go worker(2)
	go worker(3)

	time.Sleep(5 * time.Second)
}