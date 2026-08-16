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
- Make required credentials fail Compose interpolation when absent; do not use known fallback passwords for production-facing services.

### GitHub Actions and VPS deployment pattern

When a project deploys to a VPS through GitHub Actions, derive names and paths from the local repository instead of hardcoding template leftovers:

- Use a Compose project name that matches the application or repository slug.
- Name containers and networks from that project name, for example `<project>-app`, `<project>-db`, and `<project>-net`.
- Use a GHCR image name based on `GITHUB_REPOSITORY`, with a safe fallback matching the current repository.
- Publish an immutable image tag such as the Git commit SHA alongside a convenience tag such as `latest`.
- Deploy the immutable tag selected by CI through an environment variable such as `IMAGE_TAG`; do not let a VPS deployment silently switch to a newer `latest` image.
- Keep app host/container port configurable through `.env` `PORT`; set a project-appropriate default.
- Inside the app container, bind to `0.0.0.0`, not `localhost`.
- Inside Compose, the app should connect to the database by service name and internal port.

```yaml
ports:
  - "${PORT:-8080}:${PORT:-8080}"
environment:
  ADDRESS: 0.0.0.0
  PORT: ${PORT:-8080}
  DB_HOST: db
  DB_PORT: 3306
```

- If host access to MariaDB/MySQL is needed, bind the DB port to VPS localhost only and map host-to-container explicitly:

```yaml
ports:
  - "127.0.0.1:<host-db-port>:3306"
```

- Do not reuse `.env` `DB_PORT` for host mapping unless the project intentionally exposes the same port. `DB_PORT` usually means the internal database port used by the app; host access may be a separate fixed mapping or a separate variable.
- In `deploy.sh`, set `APP_DIR` to the actual VPS project directory, then `cd "$APP_DIR"` before reading `.env`, pulling images, or running Compose.
- Support both Compose CLIs:

```bash
if docker compose version >/dev/null 2>&1; then
  COMPOSE="docker compose"
else
  COMPOSE="docker-compose"
fi
```

- GitHub Actions workflow files must live under `.github/workflows/`.
- A minimal SSH deploy workflow commonly needs repository secrets such as `CR_PAT`, `VPS_HOST`, `VPS_SSH_KEY`, and `VPS_USERNAME`; add `VPS_PORT` only when SSH does not use port 22.
- The workflow can build and push to GHCR with `GITHUB_TOKEN`, then SSH to the VPS, log in to GHCR with `CR_PAT`, export `GITHUB_REPOSITORY=${{ github.repository }}`, export the selected immutable `IMAGE_TAG`, export the project-specific `APP_DIR`, and run `bash deploy.sh`.
- Do not run `chmod +x deploy.sh` before a tracked script performs `git pull`: it can dirty the VPS working tree and block that pull. Either store the executable bit in Git or invoke the script explicitly with `bash`.
- After Compose starts, wait for the API container health status. On timeout or an exited container, print recent service logs and fail the deployment instead of reporting success.
- If deployment fails while pulling a public database image with `TLS handshake timeout`, treat it as VPS-to-registry network instability; retry or pre-pull the image on the VPS.
- For DBeaver or GUI database access, prefer an SSH tunnel to a DB port bound on `127.0.0.1` rather than exposing the database publicly.

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
5. Confirm the deployment pulls the CI-selected immutable image tag and does not fall back to a mutable tag unexpectedly.
6. Scan the image for known vulnerabilities and embedded secrets.
7. Run the same image in CI and production; change configuration, not the artifact.
