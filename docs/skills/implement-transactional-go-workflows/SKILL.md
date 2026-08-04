---
name: implement-transactional-go-workflows
description: Implement safe multi-step business workflows in Go with GORM transactions, state machines, row-level locks, idempotency, webhook verification, ledgers, and external-side-effect compensation. Use for payments, inventory allocation, custody handoffs, order claiming, balances, rewards, or any concurrent read-modify-write flow.
---

# Implement Transactional Go Workflows

Protect invariants first. Treat retries, duplicate delivery, concurrency, and partial remote failure as normal operating conditions.

## Define the workflow before coding

1. List the authoritative records and invariant for each one.
2. Define allowed states and transitions explicitly.
3. Identify commands that can arrive twice or out of order.
4. Choose an idempotency key for every retriable entry point.
5. Separate database-atomic work from remote side effects.
6. Define the recovery action for every failure boundary.

Represent transitions as an allowlist instead of scattered string assignments:

```text
pending -> accepted -> preparing -> ready -> picked_up -> delivered -> completed
       \-> cancelled
```

Reject transitions not present in the state machine with a conflict-style application error.

## Own the transaction in the service

Prefer a callback so rollback occurs automatically:

```go
err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
    order, err := orders.GetForUpdate(ctx, tx, orderID)
    if err != nil { return err }
    if order.Status == target { return nil } // idempotent replay
    if !allowed(order.Status, target) { return ErrInvalidTransition }
    return orders.UpdateStatus(ctx, tx, orderID, target, now)
})
```

If the codebase uses `Begin`, always check begin, commit, and rollback errors where they affect correctness. Never commit from a repository.

## Lock the minimum authoritative rows

- Lock the row whose current value controls the decision.
- Lock related balance, stock, token, or allocation rows before calculating their new values.
- Acquire multiple locks in a stable order to reduce deadlocks.
- Re-read and validate state after acquiring the lock.
- Keep network calls and expensive CPU work outside the locked transaction.
- Retry only known transient database errors and cap attempts with jitter.

## Make operations idempotent

- Put a unique constraint on provider event ID, order ID, idempotency key, or another durable key.
- Store the received event and processing result.
- On a duplicate, return the already-computed outcome without repeating ledger writes or side effects.
- Treat a terminal state replay as success only when the request describes the same outcome.
- Preserve raw webhook payloads only when retention and sensitive-data policies allow it.

## Integrate remote side effects safely

Avoid holding a database transaction open during an HTTP call. Use one of these patterns:

- Persist a pending operation, commit, call the provider with an idempotency key, then persist the provider result.
- Persist an outbox event in the same transaction and let a worker deliver it.
- For object uploads that must happen first, delete the uploaded object if later persistence fails.
- Add reconciliation for provider success followed by local update failure.

Do not promise exactly-once behavior across a database and remote API. Build at-least-once delivery plus idempotent consumers.

## Verify webhooks

1. Read the raw body when the provider's signature covers raw bytes.
2. Validate timestamp and signature with the provider-specific algorithm.
3. Use constant-time comparison for MAC values.
4. Reject stale events and unexpected merchant/account identifiers.
5. Record the provider event ID under a unique constraint.
6. Map provider statuses through one explicit function.
7. Return success for already-processed valid events so the provider stops retrying.

## Preserve an audit trail

- Append immutable ledger or custody records for value and possession changes.
- Store actor, source and destination, amount or item, previous state, new state, timestamp, and correlation ID.
- Derive hashes from a canonical encoding if tamper evidence is required.
- Never rely on a hash chain as a substitute for access control and database integrity.

## Test failure and concurrency

- Run two claims, redemptions, allocations, or webhook deliveries concurrently and assert only one effect occurs.
- Inject failure before each write, at commit, and after remote success.
- Verify rollback leaves every invariant intact.
- Verify duplicate and out-of-order events.
- Verify insufficient balance or stock under concurrent load.
- Run with the race detector and the production database dialect when locks matter.

