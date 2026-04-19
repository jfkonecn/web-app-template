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

Build the Docker image:

```bash
docker build -t web-app-template .
```

Run the server container:

```bash
docker run --rm -p 8080:8080 \
  -e APP_ENV=production \
  -e APP_HOST=0.0.0.0 \
  -e APP_PORT=8080 \
  -e SESSION_SECRET=replace-me \
  -e POSTGRES_HOST=replace-me \
  -e POSTGRES_PORT=5432 \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=web_app_template \
  -e POSTGRES_SSLMODE=disable \
  -e OIDC_BASE_URL=replace-me \
  -e OIDC_CLIENT_ID=replace-me \
  -e OIDC_CLIENT_SECRET=replace-me \
  -e OIDC_CALLBACK_URL=http://localhost:8080/callback \
  web-app-template
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
- Logout is local to the app session; the dev Keycloak setup in this repo does not use a provider logout redirect

Server defaults:

- `APP_ENV=development`
- `APP_HOST=0.0.0.0`
- `APP_PORT=8080`
- `SESSION_SECRET` is required and must be set to a strong random value

Health endpoint:

- `GET /healthz`

## Keycloak

Start Keycloak with Docker Compose:

```bash
./scripts/run-keycloak.sh
```

Clear Keycloak realm configuration:

```bash
./scripts/clear-keycloak.sh
```

The bundled import file at `docker-dev/keycloak/realm-web-app-template.json` removes the manual setup from the Keycloak Docker guide. It creates:

- Realm: `web-app-template`
- Client ID: `web-app-template`
- Client secret: `replace-with-a-dev-only-client-secret`
- Realm role: `admin`
- Admin app user: `admin-user` / `password`
- Plain app user: `app-user` / `password`
- Admin console login: `admin` / `admin`

Keycloak runs on `http://localhost:8081`, and `.env.example` is preconfigured to use:

- `OIDC_BASE_URL=http://localhost:8081/realms/web-app-template`
- `OIDC_CLIENT_ID=web-app-template`
- `OIDC_CLIENT_SECRET=replace-with-a-dev-only-client-secret`

Open the admin console at `http://localhost:8081/admin/`.

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
