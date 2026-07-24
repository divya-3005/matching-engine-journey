package ringbuffer

type RingBuffer[T any] struct {
	buffer     []T
	readIndex  int
	writeIndex int
	capacity   int
}