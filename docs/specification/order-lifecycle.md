# Order Lifecycle

## Goal

Describe the lifecycle of an order from submission to completion.

---

## Flow

Client

↓

API Validation

↓

Ring Buffer

↓

Matching Engine

↓

Order Book

↓

Trade Execution

↓

Response

---

## States

NEW

↓

VALIDATED

↓

QUEUED

↓

MATCHING

↓

PARTIALLY_FILLED

↓

FILLED

or

CANCELLED

or

REJECTED

---

## Description

### NEW

Order received from client.

### VALIDATED

Order passed validation.

### QUEUED

Order waiting inside the ring buffer.

### MATCHING

Matching engine is processing the order.

### PARTIALLY_FILLED

Part of the quantity has traded.

### FILLED

Entire quantity has traded.

### CANCELLED

Order removed before completion.

### REJECTED

Order failed validation or business rules.