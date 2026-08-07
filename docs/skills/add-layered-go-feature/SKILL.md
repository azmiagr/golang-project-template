---
name: add-layered-go-feature
description: Implement an end-to-end feature in an existing layered Go API across entities, DTOs, repositories, services, handlers, dependency aggregators, and Gin routes. Use when adding an endpoint, CRUD operation, dashboard query, upload flow, or domain capability to a handler-service-repository codebase.
---

# Add a Layered Go Feature

Implement the smallest complete vertical slice that matches the repository's existing contracts.

## Discover the local pattern

1. Find the nearest analogous endpoint by transport type and domain complexity.
2. Read its route, handler, request/response models, service interface and implementation, repository interface and implementation, entity, constructor wiring, and tests.
3. Search for shared status constants, response helpers, application errors, user-context helpers, and transaction utilities.
4. List the exact files that must change before editing.

## Define the contract first

- Choose method, path, authentication, allowed roles, request encoding, status code, and response schema.
- Separate input DTOs, output DTOs, repository filters, and query projection rows.
- Use pointers for optional PATCH fields so omitted and zero values remain distinct.
- Add binding tags for transport shape, then repeat security- and domain-critical validation in the service.
- Avoid returning database entities directly when they contain secrets, internal fields, or unstable relations.

## Implement from the inside out

### 1. Entity and migration

- Change persistence entities only when storage changes.
- Add explicit primary keys, indexes, nullability, precision, timestamps, and relation rules.
- Register new entities in the migration mechanism in dependency order.
- Make seeds deterministic and idempotent.

### 2. Repository

- Add the narrowest operation the use case needs.
- Accept `tx *gorm.DB` as the first argument, matching this repository's caller-owned transaction convention.
- Keep filtering, joins, locking, preload choices, and SQL in the repository.
- Return `gorm.ErrRecordNotFound` or another storage error; map it in the service.
- Use a projection DTO for aggregates and multi-table reads.
- Never call `Begin`, `Commit`, or `Rollback` inside a repository.

### 3. Service

- Add the method to the service interface before implementing it.
- Validate the principal, normalized input, ownership, invariants, and current state.
- Map storage absence to not-found only when absence is semantically not-found.
- Open a transaction for related writes or read-modify-write operations using the local pattern below.
- Construct entities and response DTOs here, not in the repository.
- Inject external adapters instead of creating SDK clients inside the method.

Use this repository's manual transaction style:

```go
tx := s.db.Begin()
defer tx.Rollback()

record, err := s.someRepo.GetSomething(tx, param)
if err != nil {
    return nil, err
}

err = s.someRepo.UpdateSomething(tx, record)
if err != nil {
    return nil, err
}

err = tx.Commit().Error
if err != nil {
    return nil, err
}

return result, nil
```

- Use `s.db` directly only for reads that do not belong to a transaction.
- Use `errors.Is(err, gorm.ErrRecordNotFound)` in the service to distinguish absence from real database errors.
- All repository calls that depend on the transaction must happen before `tx.Commit()`.
- Do not use `tx` again after `Commit` or `Rollback`; use `s.db` for any post-commit read.
- Keep remote calls outside the transaction unless the current local pattern already intentionally keeps them inside and the risk is accepted.

### 4. Handler

- Retrieve the authenticated user with the shared typed helper.
- Bind JSON, form, multipart, path, and query inputs with the established conventions.
- Reject malformed transport data with a safe 400 response.
- Delegate domain errors to the shared error mapper.
- Return the documented status and response envelope.

### 5. Wiring and route

- Add the repository to its constructor or aggregate.
- Inject it into the service constructor or aggregate.
- Mount the route under the correct version and role group.
- Apply authentication before role authorization.
- Avoid constructing dependencies inside handlers.

## Handle special feature types

### Multipart upload

- Parse textual fields and files independently.
- Validate size and decoded content, not only filename extension.
- Delete uploaded objects when the database transaction fails.
- Keep storage URLs or object keys in persistence, never raw SDK responses.

### Dashboard read

- Build one or a few purpose-specific aggregate queries.
- Avoid loading large object graphs and aggregating them in Go.
- Scan into projection rows and map them to stable response DTOs.

### Concurrent state change

- Lock the authoritative row before checking its current state.
- Validate an explicit allowed transition.
- Make duplicate submissions idempotent when the operation can be retried.

## Add tests with the feature

- Test handler binding and status mapping.
- Test service validation, authorization, success, not-found, conflict, and dependency failure.
- Test repository filters, joins, locking, and uniqueness against the supported database where SQL behavior matters.
- Test transaction rollback and duplicate delivery for workflows.

## Completion checklist

- Search for every new interface method to ensure all implementations and fakes compile.
- Run `gofmt` on changed Go files.
- Run `go vet ./...`, `go test ./...`, and `go build ./...` or repository equivalents.
- Confirm the API contract, migration, configuration template, and route documentation are updated when applicable.
- Review the diff for accidental entity exposure, unbounded queries, leaked errors, missing role guards, and hidden global dependencies.
