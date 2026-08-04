---
name: model-relational-data-with-gorm
description: Design relational domain entities, migrations, idempotent seeds, repositories, projections, joins, and row locks with Go and GORM. Use when adding or reviewing GORM models, MariaDB/MySQL schemas, aggregate queries, transaction-aware repositories, UUID relations, indexes, or database lifecycle code.
---

# Model Relational Data with GORM

Treat the database schema as a long-lived contract. Optimize entities for persistence and DTOs for use cases.

## Model three distinct shapes

- Use an entity for table columns, keys, constraints, and relationships.
- Use request/response DTOs for public API stability and validation.
- Use projection rows for joins, aggregates, reports, and dashboards.

Do not reuse one struct for all three concerns.

## Design the schema

1. Write the invariant and access patterns before writing tags.
2. Select primary keys consistently; generate UUIDs in the application when the project follows that convention.
3. Mark foreign keys and frequently filtered or sorted columns with indexes.
4. Define decimal precision explicitly. Prefer integer minor units or a decimal type for money; avoid binary floating-point for new financial data.
5. Use pointers or nullable types when absence differs from a zero value.
6. Define created and updated timestamps consistently and store operational timestamps in UTC.
7. Decide cascade, restrict, or set-null behavior for every relationship.
8. Prefer portable varchar status columns plus Go constants when database enum migrations would make evolution fragile.

## Keep repositories transaction-aware

Use interfaces that describe use-case queries, not generic ORM verbs:

```go
type OrderRepository interface {
    GetForUpdate(ctx context.Context, tx *gorm.DB, id uuid.UUID) (*entity.Order, error)
    Create(ctx context.Context, tx *gorm.DB, order *entity.Order) error
}
```

- Pass either the base DB or an active transaction into every operation.
- Call `WithContext(ctx)` before the query.
- Keep GORM clauses and raw SQL inside the repository.
- Return storage errors intact so the service can use `errors.Is`.
- Prefer explicit column updates for state changes over a broad `Save` when unintended fields could be overwritten.
- Add deterministic ordering to list queries and bounds to user-facing result sets.

## Build aggregate reads intentionally

- Start from the table that determines result cardinality.
- Aggregate one-to-many relations in subqueries before joining them to avoid multiplication errors.
- Use `COALESCE` only where the response contract defines a zero or empty fallback.
- Alias every selected expression to the matching projection field.
- Use `COUNT(DISTINCT ...)` when joins can duplicate the counted entity.
- Inspect generated SQL and query plans for high-volume paths.
- Avoid `Preload` for dashboard-style reads when a projection query is clearer and cheaper.

## Apply row locking only inside transactions

Use `clause.Locking{Strength: "UPDATE"}` for authoritative read-modify-write flows. Lock rows in a consistent order across code paths and keep the transaction short. Never treat a lock outside a transaction as concurrency protection.

## Manage migrations

- Register referenced tables before tables with foreign keys.
- Review generated DDL before relying on `AutoMigrate` in production.
- Use versioned migrations for destructive changes, data backfills, renames, and zero-downtime deployments.
- Separate schema migration from normal server startup when multiple replicas can race or startup latency matters.
- Test migration from the previous released schema, not only against an empty database.

## Write idempotent seeds

- Use a stable natural key or deterministic UUID to find an existing record.
- Return success when the desired row already exists.
- Distinguish record-not-found from real database failures with `errors.Is`.
- Wrap related seed rows in a transaction.
- Never ship default production credentials in a seed; require secure bootstrap configuration.

## Verify

1. Test uniqueness, foreign keys, nullability, precision, and deletion behavior.
2. Test repository behavior against the same database family used in production when using dialect-specific SQL or locks.
3. Run concurrent tests for allocation, claim, balance, inventory, or state-transition paths.
4. Confirm query results remain correct with multiple children on every joined relationship.
5. Confirm no API response serializes password hashes, secret tokens, or unrelated associations from an entity.

