package ringbuffer

import "errors"

var (
	ErrBufferFull  = errors.New("ring buffer is full")
	ErrBufferEmpty = errors.New("ring buffer is empty")
)

type RingBuffer[T any] struct {
	buffer     []T
	readIndex  int
	writeIndex int
	size       int
	capacity   int
}

func New[T any](capacity int) *RingBuffer[T] {
	if capacity <= 0 {
		panic("capacity must be greater than zero")
	}

	return &RingBuffer[T]{
		buffer:   make([]T, capacity),
		capacity: capacity,
	}
}

func (r *RingBuffer[T]) Enqueue(value T) error {
	if r.IsFull() {
		return ErrBufferFull
	}

	r.buffer[r.writeIndex] = value
	r.writeIndex = (r.writeIndex + 1) % r.capacity
	r.size++

	return nil
}

func (r *RingBuffer[T]) Dequeue() (T, error) {
	var zero T

	if r.IsEmpty() {
		return zero, ErrBufferEmpty
	}

	value := r.buffer[r.readIndex]
	r.readIndex = (r.readIndex + 1) % r.capacity
	r.size--

	return value, nil
}

func (r *RingBuffer[T]) IsEmpty() bool {
	return r.size == 0
}

func (r *RingBuffer[T]) IsFull() bool {
	return r.size == r.capacity
}

func (r *RingBuffer[T]) Size() int {
	return r.size
}

func (r *RingBuffer[T]) Capacity() int {
	return r.capacity
}