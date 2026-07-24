// Package ringbuffer provides a generic fixed-size circular buffer.
//
// The ring buffer is designed for constant-time enqueue and dequeue
// operations while avoiding additional memory allocations after
// initialization.
//
// It will be used as the event queue between the API layer and the
// matching engine.
package ringbuffer