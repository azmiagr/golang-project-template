---
name: secure-gin-api-with-jwt-rbac
description: Secure a Gin REST API with strict Bearer parsing, JWT validation, account loading, token revocation, typed request context, role-based route guards, password hashing, and safe error responses. Use when implementing login/logout, protected routes, account status checks, multi-role authorization, or reviewing Go API authentication security.
---

# Secure a Gin API with JWT and RBAC

Compose security as authentication, principal resolution, account policy, then authorization.

## Build the authentication pipeline

1. Require exactly `Authorization: Bearer <token>` and reject malformed schemes.
2. Parse the JWT with an explicit allowed signing algorithm.
3. Validate signature, expiration, not-before, issuer, audience, and token type as required by the deployment.
4. Require a cryptographically random secret of sufficient length for HMAC, or use asymmetric keys with controlled rotation.
5. Extract only a stable subject or user ID from claims.
6. Check a revocation record or session version when immediate logout is required.
7. Load the current user and role from trusted persistence.
8. Reject disabled, unverified, or otherwise ineligible accounts.
9. Put a minimal typed principal in the Gin context and call `Next`.

Abort immediately after every authentication failure. Do not allow handlers to parse tokens independently.

## Keep claims minimal

- Include subject, issued-at, expiration, issuer, audience, and a unique token ID when useful.
- Treat role claims as advisory if roles can change before token expiry; reload authorization state from persistence or use short-lived tokens with a session version.
- Never place password hashes, personal identifiers, API keys, or sensitive profiles in a token.
- Use access-token lifetimes appropriate to risk and a separate refresh-token design if long sessions are needed.

## Implement logout and revocation

- Hash stored raw tokens or store the JWT ID instead of persisting bearer secrets.
- Store the expiration time and remove expired revocation records.
- Make repeated logout safe.
- Revoke all sessions through a user session version or session table when password reset or compromise demands it.

## Authorize by route group

Create one generic role guard and small semantic wrappers only when they improve readability:

```go
admin := api.Group("/admin")
admin.Use(authenticate, onlyRoles(RoleAdmin))
```

- Run authentication before authorization.
- Return 401 when identity is absent or invalid.
- Return 403 when identity is valid but lacks permission.
- Use centralized role constants; do not accept aliases unless they are a documented migration policy.
- Enforce resource ownership and object-level permissions again in the service.
- Never rely on hidden UI controls as authorization.

## Access the principal safely

Use a single helper that checks key existence and type assertion. Return a controlled error instead of panicking. Prefer a small principal struct over exposing the full database entity to every handler.

## Protect credentials and responses

- Hash passwords with bcrypt or Argon2id using a maintained cost policy.
- Compare hashes through the library API and use a generic login failure message.
- Rate-limit login, OTP, password reset, and verification endpoints.
- Store OTPs hashed, expire them, limit attempts, and invalidate them after use.
- Never return raw internal errors, token validation details, password hashes, or secret configuration.
- Log security events with request and user IDs but without credentials or bearer tokens.

## Test the matrix

- Missing header, wrong scheme, malformed token, wrong algorithm, bad signature, expired token, and wrong issuer/audience.
- Revoked token, nonexistent user, inactive user, and role change.
- Every protected group with each allowed and denied role.
- Object ownership denial even for an otherwise allowed role.
- Repeated logout and expired revocation cleanup.
- Confirm a middleware error aborts the chain and the handler is never called.

