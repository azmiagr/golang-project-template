---
name: build-layered-go-api
description: Design or refactor a maintainable Go REST API with explicit handler, service, repository, entity, DTO, shared-package, and composition-root boundaries. Use when starting a Go backend, choosing package structure, untangling mixed HTTP/business/database code, reviewing dependency direction, or wiring Gin and GORM components.
---

# Build a Layered Go API

Build around dependency direction, not around framework calls. Keep the transport replaceable, business rules testable, and persistence isolated.

## Inspect before designing

1. Read `go.mod`, the entry point, route registration, constructors, and representative files from every existing layer.
2. Identify the current naming, error, response, transaction, configuration, and test conventions.
3. Preserve sound local conventions. Introduce a new abstraction only when it removes a concrete coupling or enables a needed test seam.
4. Record whether the application is a modular monolith, multiple binaries, or a service with background workers.

## Use this dependency flow

```text
HTTP request
  -> handler/transport
  -> service/use case
  -> repository port
  -> database adapter
```

Allow dependencies to point inward:

- Let handlers depend on service interfaces and transport helpers.
- Let services depend on repository and external-service interfaces.
- Let repositories depend on GORM and database entities.
- Keep entities and domain values independent of Gin.
- Keep service logic independent of `*gin.Context`.
- Never let repositories call handlers or services.

## Choose a portable layout

Adapt names to the repository instead of forcing this exact tree:

```text
cmd/api/main.go              composition root
internal/handler/rest/       HTTP binding, route params, responses
internal/service/            validation, policy, orchestration
internal/repository/         persistence interfaces and GORM adapters
entity/                      database entities and relations
model/                       request, response, query, and projection DTOs
pkg/                         genuinely reusable cross-cutting adapters
```

Keep project-specific helpers under `internal/` unless another module is expected to import them.

## Keep entity boundaries explicit

- Let entities describe database tables, keys, constraints, and ORM associations.
- Keep request/response DTOs and aggregate projections separate from entities.
- Place normalization, generated identifiers, hierarchy traversal, and multi-record invariants in services or dedicated domain helpers, not entity methods.
- Declare foreign keys and deletion behavior intentionally; use pointers where `NULL` and a zero value have different meanings.
- Do not serialize entities or their associations directly as HTTP responses by default.

## Assign responsibilities

### Handler

- Parse headers, route parameters, query strings, JSON, and multipart forms.
- Retrieve the authenticated principal through one typed helper.
- Call exactly the service operation represented by the endpoint.
- Convert known application errors into the standard response envelope.
- Avoid business decisions and direct database or SDK calls.

### Service

- Validate domain rules even if transport binding already validates shape.
- Normalize input, enforce ownership and state transitions, and coordinate repositories.
- Own transaction boundaries for multi-write use cases.
- Return transport-neutral errors and response-ready DTOs.
- Depend on injected clocks, hashers, storage, mail, payment, and token ports when behavior must be tested.

#### Manual transaction pattern

When the application uses manual GORM transactions in the service layer:

```go
tx := s.db.Begin()
defer tx.Rollback()

// Use tx for every repository read/write that belongs to the unit of work.

err := tx.Commit().Error
if err != nil {
    return nil, err
}
```

- Inject `db *gorm.DB` or a transaction factory into the service implementation; avoid hidden package globals in new designs.
- Use the base `s.db` only for non-transactional reads.
- For multi-write or read-modify-write flows, call repositories with the active `tx`.
- Do every transaction-scoped read before `tx.Commit()`. Never call a repository with `tx` after commit or rollback.
- Return before commit on validation or repository errors; the deferred rollback handles cleanup.
- Repositories must not begin, commit, or rollback transactions.

### Repository

- Own all GORM and SQL expressions.
- Accept the active `*gorm.DB` handle as the first argument so the service can supply either `s.db` or a transaction.
- Return database errors without choosing HTTP status codes.
- Use dedicated projection structs for aggregate reads rather than bloated entities.

## Wire at one composition root

Construct dependencies in this order:

1. Load and validate configuration.
2. Open infrastructure connections.
3. Run the selected migration policy.
4. Construct repositories.
5. Construct external adapters and services.
6. Construct middleware and handlers.
7. Mount routes and start the server.

Keep constructors explicit. Prefer a slightly verbose dependency list over hidden globals. If a constructor becomes unwieldy, group cohesive dependencies in a small struct; do not introduce a service locator.

## Standardize cross-cutting behavior

- Define one application error type with a stable category, safe client message, wrapped cause, and optional details.
- Define one JSON success/error envelope if the API contract requires it.
- Centralize authentication, authorization, CORS, request IDs, recovery, and logging as middleware.
- Keep secrets and environment lookup at the configuration edge; pass typed config inward.
- Carry `context.Context` from Gin into services, repositories, and external clients when introducing or maintaining a context-aware application boundary. Do not create a partial convention in only one method.

## Make architecture decisions deliberately

- Use a direct service-to-repository call for simple CRUD.
- Use a transaction-owning service for multi-aggregate changes.
- Use a projection/read repository for dashboards and join-heavy reads.
- Use an outbox or job boundary when a durable database change must trigger an unreliable remote side effect.
- Split a package only when it has more than one reason to change, not merely because it is long.

## Verify

1. Run formatting, static analysis, tests, and a build.
2. Confirm no handler imports GORM and no repository imports Gin.
3. Confirm every protected route composes authentication before authorization.
4. Confirm every new dependency is constructed at the composition root.
5. Confirm transactions, errors, and responses follow one convention across features.
