package main

import "fmt"

func square(x int) int {
	y := x * x
	return y
}

func main() {
	result := square(5)
	fmt.Println(result)
}