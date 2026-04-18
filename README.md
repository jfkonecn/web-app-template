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

Build the frontend CSS bundle:

```bash
npm run build:css
```

Build the client-side TypeScript bundles:

```bash
npm run build:js
```

Build both frontend asset types:

```bash
npm run build:assets
```

Run the HTML screenshot tests:

```bash
npm run test:screenshots
```

Refresh the committed golden screenshots after an intentional UI change:

```bash
npm run test:screenshots:update
```

Environment files:

- Copy `.env.example` to `.env` for local development
- `.env` is ignored by git and is loaded automatically by the run and migration scripts
- `internal/config` no longer provides runtime defaults; startup exits if any required env var is missing or empty
- Logout is local to the app session; the dev Dex setup in this repo does not use a provider logout redirect

Server defaults:

- `APP_ENV=development`
- `APP_HOST=0.0.0.0`
- `APP_PORT=8080`
- `SESSION_SECRET` is required and must be set to a strong random value

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

Create or install the `migrate` binary into `./bin`:

```bash
./scripts/install-migrate.sh
```

According to the migrate docs (`migrate create -ext sql -dir <DIR> -seq <NAME>`), create a new migration with:

```bash
./scripts/create-migration.sh create_users_table
```

This writes sequential SQL migration files into `db/migrations`, for example:

- `db/migrations/000001_create_users_table.up.sql`
- `db/migrations/000001_create_users_table.down.sql`

Apply migrations to the Docker PostgreSQL database:

```bash
./scripts/apply-migrations.sh
```

`apply-migrations.sh` uses these defaults (override with env vars):

- `POSTGRES_HOST=localhost`
- `POSTGRES_PORT=5432`
- `POSTGRES_USER=postgres`
- `POSTGRES_PASSWORD=postgres`
- `POSTGRES_DB=web_app_template`
