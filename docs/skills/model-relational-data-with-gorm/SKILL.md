---
name: model-relational-data-with-gorm
description: Design relational domain entities, migrations, idempotent seeds, repositories, projections, joins, and row locks with Go and GORM. Use when adding or reviewing GORM models, MariaDB/MySQL schemas, aggregate queries, transaction-aware repositories, UUID relations, indexes, or database lifecycle code.
---

# Model Relational Data with GORM

Treat the database schema as a long-lived contract. Design entities for persistence, DTOs for use cases, and projections for reads; do not merge those concerns into one struct.

## Separate persistence, API, and read shapes

- **Entity:** table columns, database keys, indexes, constraints, and GORM associations only.
- **Request DTO:** transport input and binding validation. It must not be persisted directly.
- **Response DTO:** stable public output. It must not expose an entity merely because that is convenient.
- **Projection row:** a repository-private or model-level struct used for joins, aggregates, reports, or dashboards.

Rely on GORM's default table naming from the struct type. Add `TableName()` only when integrating with an existing schema whose table name cannot be changed; it is a compatibility exception, not a default entity requirement. Keep entity methods limited to persistence-neutral value behavior. Put request normalization, generated codes, hierarchy levels, authorization, state transitions, and cross-record rules in services or dedicated domain helpers instead.

## Define entities deliberately

Before writing tags, state the invariant, deletion semantics, and expected query patterns.

- Pick one primary-key strategy and use it consistently. Generate UUIDs in the application only when that is the chosen application convention.
- Use scalar ID fields for foreign keys and add an index when the relationship is filtered, joined, or traversed often.
- Use `*uuid.UUID`, `*time.Time`, pointers, or nullable database types when `NULL` differs from a zero value.
- Mark required columns `not null`; define default values only when the database should own the default.
- Define precision and scale for decimals. Prefer integer minor units or a decimal type for money; do not introduce binary floating point for new monetary values.
- Apply timestamps consistently. Store instants in UTC and convert only at the API or presentation boundary.
- Prefer portable string status columns with application constants unless a database enum is a deliberate operational choice.
- Add unique indexes for durable invariants, including composite uniqueness such as a name scoped to a parent.

Declare associations explicitly when GORM cannot infer them safely or when the relationship deserves documented semantics:

```go
type Membership struct {
    UserID  uuid.UUID `gorm:"type:char(36);primaryKey"`
    TeamID  uuid.UUID `gorm:"type:char(36);primaryKey"`
    Role    string    `gorm:"type:varchar(30);not null"`

    User User `json:"-" gorm:"foreignKey:UserID;references:UserID;belongsTo;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
    Team Team `json:"-" gorm:"foreignKey:TeamID;references:TeamID;belongsTo;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
```

### Use `belongsTo` precisely

Use `belongsTo` on the entity that **stores the foreign-key column**. The association does not create or validate that column; declare the scalar foreign key explicitly and choose its nullability from the domain rule.

```go
type Invoice struct {
    InvoiceID uuid.UUID  `gorm:"type:char(36);primaryKey"`
    AccountID *uuid.UUID `gorm:"type:char(36);index"`

    Account *Account `json:"-" gorm:"foreignKey:AccountID;references:AccountID;belongsTo;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
}
```

- Use a non-pointer FK and `RESTRICT` or `CASCADE` when the relationship is required. Use a pointer FK only when the relationship may truly be absent; `SET NULL` requires a nullable FK.
- Specify both `foreignKey` and `references` when names differ from GORM defaults, when there are multiple relations to the same table, or when a self-reference could be ambiguous. This prevents GORM from inferring the wrong association kind or key.
- The inverse collection is normally `has many`, not another `belongsTo`. For example, `Account.Invoices []Invoice` uses `foreignKey:AccountID` because `Invoice` owns the FK.
- `belongsTo` describes how records are joined; it does not decide authorization, whether a parent may be deleted, or whether a child may be reassigned. Express those lifecycle rules with constraints and service-level validation.
- Preload a belongs-to association only when the use case needs it. Prefer an explicit projection for lists or aggregate reads rather than serializing a fully loaded entity graph.

- Use `json:"-"` on persistence associations by default; return purpose-built response DTOs instead of serializing object graphs.
- Choose `CASCADE`, `RESTRICT`, or `SET NULL` from the domain lifecycle. Never use cascade only to make deletion easier.
- Model many-to-many relationships with an explicit join entity whenever the relation has metadata, lifecycle, audit fields, or its own constraints.
- For self-references, use a nullable parent FK plus explicit parent/children associations. Enforce cycle prevention, maximum depth, and subtree behavior in a service/repository workflow; foreign keys alone cannot do that.
- Keep generated sequences, migration-only helper structs, and one-off backfill types out of the public entity package unless they are part of the domain schema.

## Keep repositories transaction-aware

Use interfaces that describe use-case queries, not generic ORM verbs:

```go
type OrderRepository interface {
    GetForUpdate(tx *gorm.DB, id uuid.UUID) (*entity.Order, error)
    Create(tx *gorm.DB, order *entity.Order) error
}
```

- Pass either the base DB or an active transaction into every operation as the first argument.
- Choose one repository transaction convention and apply it consistently. A common GORM convention is `Method(tx *gorm.DB, ...)`, where callers pass either the base DB handle or an active transaction.
- Add `context.Context` only as a deliberate repo-wide refactor, not for one isolated method.
- Keep GORM clauses and raw SQL inside the repository.
- Return storage errors intact so the service can use `errors.Is`.
- Prefer explicit column updates for state changes over a broad `Save` when unintended fields could be overwritten.
- Add deterministic ordering to list queries and bounds to user-facing result sets.
- Do not open or finish transactions in repositories; services own `Begin`, `Rollback`, and `Commit`.

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

Expose locking through a repository method such as:

```go
func (r *RegistrationSessionRepository) GetRegistrationSessionForUpdate(tx *gorm.DB, tokenHash string) (*entity.RegistrationSession, error) {
    var session entity.RegistrationSession
    err := tx.
        Clauses(clause.Locking{Strength: "UPDATE"}).
        Where("session_token_hash = ?", tokenHash).
        First(&session).Error
    if err != nil {
        return nil, err
    }
    return &session, nil
}
```

## Manage migrations

- Register tables in foreign-key dependency order when the migration tool requires it.
- Review generated DDL before relying on `AutoMigrate` in production.
- Use explicit, versioned migrations for destructive changes, data backfills, renames, constraint changes, and zero-downtime deployments.
- Make additive migrations safe to run repeatedly; make data migrations observable and safe to resume.
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
