---
name: containerize-go-api
description: Package and deploy a Go REST API with production containers, Docker Compose, health checks, and GitHub Actions CI/CD to a VPS. Use when creating or reviewing Dockerfiles, Compose stacks, deployment scripts, or GitHub Actions for Go services.
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

When the requested deployment target is a VPS, create `.github/workflows/deploy.yml` as part of the deliverable unless the user explicitly excludes CI/CD. Use the repository's current GitHub Actions workflow as the local formatting source when one exists; otherwise use the pattern below.

The workflow must:

- Run on pushes to `main` and support `workflow_dispatch`.
- Define `REGISTRY=ghcr.io` and `IMAGE_NAME=${{ github.repository }}` at workflow scope.
- Use separate `build-and-push` and `deploy` jobs; `deploy` must `needs: build-and-push`.
- Give the build job `contents: read` and `packages: write` permissions.
- Check out the repository, set up Go from `go.mod`, run `go test ./...`, log in to GHCR with `GITHUB_TOKEN`, set up Buildx, then push both `latest` and `${{ github.sha }}` tags through `docker/build-push-action` with GitHub Actions cache.
- Use `appleboy/ssh-action` for deployment. On the VPS, log in to GHCR with `CR_PAT`, export the lowercased `IMAGE_REPOSITORY` and immutable `IMAGE_TAG=${{ github.sha }}`, change into `${{ secrets.VPS_APP_DIR }}`, and invoke `bash deploy.sh`.
- Require and document these GitHub secrets: `VPS_HOST`, `VPS_USERNAME`, `VPS_SSH_KEY`, `VPS_APP_DIR`, and `CR_PAT`. `CR_PAT` needs `read:packages` to pull a private GHCR image.

Use this shape unless local project conventions require a compatible variation:

```yaml
name: Build and Deploy

on:
  push:
    branches: ["main"]
  workflow_dispatch:

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  build-and-push:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
      - run: go test ./...
      - uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - uses: docker/setup-buildx-action@v3
      - uses: docker/build-push-action@v6
        with:
          context: .
          push: true
          tags: |
            ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:latest
            ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:${{ github.sha }}
          cache-from: type=gha
          cache-to: type=gha,mode=max

  deploy:
    needs: build-and-push
    runs-on: ubuntu-latest
    steps:
      - uses: appleboy/ssh-action@v1.0.3
        with:
          host: ${{ secrets.VPS_HOST }}
          username: ${{ secrets.VPS_USERNAME }}
          key: ${{ secrets.VPS_SSH_KEY }}
          script: |
            set -eu
            echo "${{ secrets.CR_PAT }}" | docker login ghcr.io -u "${{ github.actor }}" --password-stdin
            export IMAGE_REPOSITORY="ghcr.io/$(printf '%s' '${{ github.repository }}' | tr '[:upper:]' '[:lower:]')"
            export IMAGE_TAG="${{ github.sha }}"
            cd "${{ secrets.VPS_APP_DIR }}"
            bash deploy.sh
```

Keep Compose and `deploy.sh` compatible with that workflow:

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
- Set Compose image names through `IMAGE_REPOSITORY` and `IMAGE_TAG`, with safe local defaults. If both are exported, `deploy.sh` must pull the app image and use `up -d --no-build`; otherwise it may use `up -d --build` for local deployment.
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
