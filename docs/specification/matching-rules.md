# Matching Rules

## Matching Priority

The engine uses Price-Time Priority.

---

## Buy Orders

Highest price executes first.

If prices are equal:

Oldest order executes first.

---

## Sell Orders

Lowest price executes first.

If prices are equal:

Oldest order executes first.

---

## Partial Fills

Partial executions are allowed.

Remaining quantity stays on the book.

---

## Market Orders

Execute immediately against the best available prices.

Any remaining quantity follows the configured policy (for example, cancel if not fully filled).

---

## Limit Orders

Execute only at the specified price or better.

Otherwise remain in the order book.