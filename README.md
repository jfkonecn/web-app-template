# web-app-template

Conventional Go repository layout with a Gin web server.

## Layout

- `cmd/server`: application entrypoint
- `internal/config`: runtime configuration loading
- `internal/server`: router and middleware setup
- `internal/handlers`: HTTP handlers
- `pkg/response`: reusable response helpers
- `api`, `configs`, `scripts`, `build`, `deployments`: standard project directories

## Run

```bash
./scripts/run.sh
```

Server defaults:

- `APP_ENV=development`
- `APP_HOST=0.0.0.0`
- `APP_PORT=8080`

Health endpoint:

- `GET /healthz`

## Logging

- `APP_ENV=production` (or `prod`): JSON logs
- Any other `APP_ENV` value: text logs
- Every request is logged by middleware with method, path, status, latency, client IP, user agent, and response size

## PostgreSQL

Start PostgreSQL with Docker Compose:

```bash
./scripts/run-postgres.sh
```

Compose defaults:

- `POSTGRES_USER=postgres`
- `POSTGRES_PASSWORD=postgres`
- `POSTGRES_DB=web_app_template`
- `POSTGRES_PORT=5432`

Stop PostgreSQL:

```bash
docker compose -f docker-dev/docker-compose.yml down
```

## Migrate Binary

You can find releases/source here: https://github.com/golang-migrate/migrate/tree/master
