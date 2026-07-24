// Package matching implements a low-level pairwise order matcher.
//
// NOTE: This package is not currently used by the engine package.
// It represents an early design iteration in which matching was
// implemented as a standalone operation on two orders with equal quantities.
//
// The engine instead implements its matching loop inline in ProcessNext,
// where it correctly handles partial fills (mismatched quantities) and
// sweeps across multiple price levels in a single call.
//
// This package is preserved as a standalone reference implementation.
package matching
