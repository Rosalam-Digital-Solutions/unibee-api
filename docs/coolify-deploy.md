# Coolify Deployment Notes

This repository now includes a root Dockerfile designed for Coolify.

## What Was Added

- Root `Dockerfile` for build and runtime.
- `manifest/docker/entrypoint.coolify.sh` to map common env vars to the app's existing CLI flags.
- `.dockerignore` to keep build context small.

Your existing `config.yaml` remains intact and is copied into the image. If you do not provide env vars, your current defaults continue to apply.

## Coolify Build Settings

- Build Pack: `Dockerfile`
- Dockerfile Location: `./Dockerfile`
- Port: `8088`
- Health Check Path: `/health`

## Editing Environment Variables After Deployment

You can edit environment variables at any time in Coolify.

1. Open your app in Coolify.
2. Go to **Environment Variables**.
3. Add or update variables.
4. Save changes.
5. Restart or Redeploy the service.

Important behavior:

- Variables are read at container startup (runtime), not baked into the image.
- You do not need to change `config.yaml` for normal production updates.
- If a variable is not set, the app keeps the value from `config.yaml`.
- If a variable is set, it overrides the matching config value at startup.

## Recommended Environment Variables

Set database/redis using one of these options:

- Option A (easiest): `MYSQL_URL` and `REDIS_URL`
- Option B: `DATABASE_LINK` and `REDIS_ADDRESS` style vars
- Option C: split `DB_*` and `REDIS_*` vars

For a full starting template, copy values from `.env.coolify.example` into your Coolify Environment Variables page.

### App

- `APP_ENV=prod`
- `MODE=stand-alone`
- `UNIBEE_API_URL=https://your-api-domain`
- `SERVER_ADDRESS=:8088`
- `SERVER_JWT_KEY=<strong-random-secret>`

### MySQL + Redis URL Style (Option A)

- `MYSQL_URL=mysql://USER:PASSWORD@HOST:3306/DBNAME`
- `REDIS_URL=redis://default:PASSWORD@HOST:6379/0`

Example from your current deployment values:

- `MYSQL_URL=mysql://mysql:***@gjh90y0omxsd0p5mgfhxupse:3306/default`
- `REDIS_URL=redis://default:***@gn38rbu6g7ac5lo4fhwgaxph:6379/0`

### MySQL (Option A: full link)

- `DATABASE_LINK=mysql:USER:PASSWORD@tcp(HOST:3306)/DBNAME?loc=UTC&parseTime=false`

### MySQL (Option B: split vars)

- `DB_HOST=<mysql-host>`
- `DB_PORT=3306`
- `DB_DATABASE=unib`
- `DB_USER=unibee`
- `DB_PASSWORD=<password>`
- `DB_QUERY=loc=UTC&parseTime=false`

### Redis

- `REDIS_HOST=<redis-host>`
- `REDIS_PORT=6379`
- `REDIS_PASSWORD=<password>`
- Optional: `REDIS_DATABASE=0`

## Behavior Guarantees

- Existing local `config.yaml` semantics are preserved.
- Env vars only override when explicitly provided.
- Health endpoint is `/health`.
