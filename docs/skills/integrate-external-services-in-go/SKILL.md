---
name: integrate-external-services-in-go
description: Integrate storage, payment, email, geocoding, or other third-party services into a Go backend through typed configuration, narrow interfaces, adapters, timeouts, retries, idempotency, validation, and cleanup. Use when adding an SDK or HTTP API without coupling handlers and business logic to a provider.
---

# Integrate External Services in Go

Place provider details at the edge and expose a small capability-oriented port to the service layer.

## Design the port from the use case

Avoid mirroring a vendor SDK. Define only what the application needs:

```go
type ObjectStore interface {
    PutImage(ctx context.Context, input ImageUpload) (StoredObject, error)
    Delete(ctx context.Context, key string) error
}
```

- Use application-owned input and output types.
- Return stable object keys or domain results, not SDK response structs.
- Accept `context.Context` for cancellation and deadlines.
- Inject the interface into the service constructor.
- Construct the concrete adapter only at the composition root.

## Load typed configuration once

- Read environment variables in the configuration package.
- Validate required URLs, credentials, bucket or merchant identifiers, environments, and timeouts at startup.
- Distinguish sandbox and production explicitly.
- Add safe placeholders to `.env.example`; never add real credentials.
- Avoid reading environment variables throughout adapter methods.
- Redact secrets from logs and error messages.

## Harden the client

- Set connection and request timeouts.
- Reuse HTTP transports and SDK clients.
- Retry only safe transient failures, honor `Retry-After`, and cap exponential backoff with jitter.
- Supply an idempotency key for retriable creates when the provider supports it.
- Translate provider errors into a small adapter error taxonomy while retaining the wrapped cause for logs.
- Instrument latency, result category, provider request ID, and retry count.
- Add a circuit breaker only when measured failure behavior warrants it.

## Handle object uploads

1. Limit request body and declared file size.
2. Decode or sniff content to verify the actual type; do not trust the extension alone.
3. Normalize or re-encode images when consistent output is required.
4. Generate collision-resistant object keys and preserve the key separately from the public URL.
5. Set explicit content type and cache policy.
6. Delete the new object if subsequent persistence fails.
7. Delete replaced objects only after the new database value commits.
8. Make cleanup safe when the object is already absent.

## Handle payments

- Persist a local pending payment and durable provider correlation ID.
- Call the provider outside long database locks.
- Store only provider fields required for reconciliation and support.
- Verify every webhook independently; never trust a browser redirect as proof of payment.
- Map provider statuses to application statuses in one function.
- Make notifications idempotent and preserve terminal-state rules.
- Implement reconciliation for missed or ambiguous callbacks.
- Represent money in integer minor units or a decimal type.

## Handle email and messaging

- Render templates separately from transport.
- Validate destination, subject, template data, and maximum size before sending.
- Queue non-interactive messages when request latency and delivery reliability matter.
- Record delivery intent and provider message ID without logging message secrets or OTP values.
- Rate-limit verification messages and make resend semantics explicit.

## Test without the real provider

- Use a fake port in service tests.
- Use an `httptest.Server` or SDK transport stub for adapter request/response tests.
- Test timeout, cancellation, throttling, malformed response, retry exhaustion, and duplicate create behavior.
- Test cleanup after database failure and no cleanup after commit.
- Keep a small opt-in sandbox integration suite outside normal unit tests.

## Review portability

Before finishing, confirm provider names do not leak into domain models, service interfaces, handler contracts, or unrelated packages. A provider replacement should mostly affect configuration, one adapter, and composition-root wiring.

