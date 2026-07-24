// Package engine coordinates the processing of incoming orders.
//
// The engine owns the ring buffer and is responsible for consuming
// orders and forwarding them to downstream components such as the
// order book and matching algorithm.
package engine