---
name: containerize-go-api
description: Package a Go REST API into a small, secure production container with multi-stage builds, CGO-aware dependencies, runtime configuration, Docker Compose services, database health checks, and reproducible verification. Use when creating or reviewing a Dockerfile, Compose stack, container startup, or deployment packaging for Go services.
---

# Containerize a Go API

Build once in a dedicated stage and copy only the binary and required runtime assets into the final image.

## Inspect build requirements

1. Read `go.mod` and imports for CGO-backed packages or native libraries.
2. Identify templates, certificates, timezone data, migrations, and other runtime files.
3. Confirm the binary package and listening address.
4. Decide whether the target supports a fully static binary or needs shared libraries.
5. Match build and runtime architectures to deployment.

## Build reproducibly

- Pin Go and runtime base image versions.
- Copy `go.mod` and `go.sum` before source code to preserve dependency cache layers.
- Download modules with checksums enabled.
- Copy source and build the intended `./cmd/...` package.
- Set `-trimpath`; inject version and commit metadata with `-ldflags` when the service exposes them.
- Use BuildKit cache mounts when supported by the deployment pipeline.
- Add a `.dockerignore` for `.git`, local secrets, build output, coverage, and editor files.

For a pure-Go binary, prefer `CGO_ENABLED=0`. For image codecs, SQLite, or other native dependencies, install the builder headers and copy or install only the matching runtime libraries in the final stage.

## Harden the runtime image

- Include CA certificates for outbound TLS and timezone data only if the application requires local zone conversion.
- Run as a dedicated non-root user.
- Use a read-only root filesystem where feasible and mount explicit writable paths.
- Do not copy source, package managers, compilers, `.env`, or credentials into the final image.
- Use exec-form `ENTRYPOINT` or `CMD` so signals reach the Go process.
- Implement graceful shutdown in the application and set an adequate orchestrator stop timeout.
- Expose the documented port for discoverability, but bind the server to `0.0.0.0` inside a container.

## Configure at runtime

- Inject configuration through environment variables or mounted secrets.
- Fail fast on missing required production configuration.
- Keep `.env.example` safe and descriptive.
- Never bake environment-specific endpoints or keys into the image.
- Emit structured startup logs without secrets.

## Compose local dependencies

- Put the API and database on a private named network.
- Use a named volume for database persistence.
- Add a real database health check and gate application startup on it for local development.
- Bind database ports to localhost if host access is needed; omit the host port otherwise.
- Use service DNS names such as `db` inside Compose, not `localhost`.
- Keep production credentials out of Compose defaults.

## Choose a migration policy

- Use a separate migration job for replicated or production deployments.
- Allow startup migrations only when they are backward compatible, concurrency safe, and intentionally controlled.
- Keep idempotent development seeds separate from production bootstrap data.
- Back up and test rollback strategy before destructive schema changes.

## Add health endpoints

- Make liveness report whether the process can serve.
- Make readiness verify required dependencies with a short timeout.
- Avoid expensive checks or schema mutation in probes.
- Return failure while graceful shutdown is draining traffic.

## Verify the artifact

1. Build the image without using local untracked dependencies.
2. Inspect the final image size, layers, user, architecture, and native library linkage.
3. Start the Compose stack from an empty volume and wait for health.
4. Exercise liveness, readiness, one database-backed endpoint, and graceful stop.
5. Scan the image for known vulnerabilities and embedded secrets.
6. Run the same image in CI and production; change configuration, not the artifact.

