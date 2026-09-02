# Lotto Journal

A LINE chatbot service that lets Thai lottery players record their ticket numbers and get automatically notified via LINE when any of their tickets win a prize.

Try it out: <https://lin.ee/oZPTXQW> or add `@249lytsb` on LINE

**Stack:** Go + Fiber · PostgreSQL · LINE Messaging API

---

## Bot Commands

All interaction happens inside the LINE chat with the bot.

| Message sent     | What happens                                                        |
| ---------------- | ------------------------------------------------------------------- |
| `123456`         | Records one L6 ticket for the upcoming draw                         |
| `456`            | Records one N3 ticket for the upcoming draw                         |
| `123456 x3`      | Records 3 copies of the same L6 ticket                              |
| `123456, 789012` | Records two tickets in one message (comma or space separated)       |
| `โพย`            | Lists all tickets you have registered for the current upcoming draw |
| Follow bot       | Creates your account; sends welcome message                         |
| Unfollow bot     | Marks your account inactive; ticket history is preserved            |

---

## Prerequisites

- [Docker Desktop](https://www.docker.com/products/docker-desktop/)
- [golang-migrate CLI](https://github.com/golang-migrate/migrate)

```shell
brew install golang-migrate
```

---

## Local Setup

> Run all commands from the **repo root** unless noted otherwise.

### 1. Environment variables

Copy the example env file and fill in your values:

```shell
cp .env.example .env.local
```

Key variables:

| Variable                    | Used by        | Example value                                                                                                                                                                                                              |
| --------------------------- | -------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `DB_USERNAME`               | docker-compose | `postgres`                                                                                                                                                                                                                 |
| `DB_PASSWORD`               | docker-compose | `yourpassword`                                                                                                                                                                                                             |
| `DB_NAME`                   | docker-compose | `lotto_journal`                                                                                                                                                                                                            |
| `DB_DSN`                    | Go app         | `postgres://postgres:yourpassword@localhost:5432/lotto_journal?sslmode=disable`                                                                                                                                            |
| `DB_MAX_IDLE_CONNS`         | Go app         | Maximum idle connections in pool (default: `0`). _Keep at 0 for serverless databases (Neon) to close idle links immediately._                                                                                              |
| `DB_MAX_OPEN_CONNS`         | Go app         | Maximum open connections in pool (default: `5`).                                                                                                                                                                           |
| `DB_CONN_MAX_LIFETIME`      | Go app         | Maximum lifetime duration of a connection (default: `"3m"`).                                                                                                                                                               |
| `DB_CONN_MAX_IDLE_TIME`     | Go app         | Maximum idle duration of a connection before closure (default: `"1m"`). _Highly recommended for scale-to-zero databases._                                                                                                  |
| `PORT`                      | Go app         | `:3000`                                                                                                                                                                                                                    |
| `LINE_CHANNEL_SECRET`       | Go app         | from LINE Developers console → Basic Settings                                                                                                                                                                              |
| `LINE_CHANNEL_ACCESS_TOKEN` | Go app         | from LINE Developers console → Messaging API                                                                                                                                                                               |
| `APP_ENV`                   | Go app         | App environment: `development`, `staging`, `production` (default: `development`)                                                                                                                                           |
| `CRON_SECRET`               | Go app         | Secret token used to authenticate request triggers for scheduled jobs. Default is empty (disabled) in application config, but is set to `"local-cron-secret-change-me"` in `.env.example` templates for local development. |
| `R2_ACCOUNT_ID`             | Go app         | Cloudflare Account ID for ticket image storage                                                                                                                                                                             |
| `R2_ACCESS_KEY_ID`          | Go app         | Cloudflare R2 Access Key ID                                                                                                                                                                                                |
| `R2_SECRET_ACCESS_KEY`      | Go app         | Cloudflare R2 Secret Access Key                                                                                                                                                                                            |
| `R2_BUCKET_NAME`            | Go app         | Cloudflare R2 Bucket Name                                                                                                                                                                                                  |
| `R2_PUBLIC_URL_PREFIX`      | Go app         | Public domain prefix for bucket assets (e.g. `https://pub-yourdomain.r2.dev`)                                                                                                                                              |
| `OPENAI_API_KEY`            | Go app         | API Key for gpt-4o-mini Vision OCR Structured Outputs                                                                                                                                                                      |

> [!TIP]
> **Environment File Resolution:**
>
> - **Centralized Configuration**: The root `.env.local` is the primary environment config file. Tooling (e.g. `pnpm dev`, `docker-compose`) and package Makefiles automatically load and export this file when run from the monorepo root.
> - **Isolated Configuration**: If running/debugging the Go API directly from inside `apps/api/` (outside the Makefile or root tooling context), the app looks for `apps/api/.env`. You can copy `apps/api/.env.example` to `apps/api/.env` for this setup.

### 2. Start the database (Optional)

> [!NOTE]
> Running `pnpm dev` automatically spins up the database container and waits for it to be healthy. You only need to run the command below manually if you want to run database commands/migrations without starting the API server.

```shell
pnpm db:start
```

### 3. Run migrations

Ensure the database is running (either via `pnpm db:start` or by starting the API) before executing migrations:

```shell
pnpm migrate:up
```

### 4. Start the API & Database (with hot reload)

This will spin up the database container, wait for it to be fully healthy, and boot up the hot-reloading API server:

```shell
pnpm dev
```

When you stop the server by pressing `Ctrl+C` once, the database container will automatically shut down as well.

---

## Production Deployment (Fly.io + Neon)

> Run all commands from the **repo root**.

### 1. Prerequisites

- Install [flyctl](https://fly.io/docs/flyctl/install/)
- Login to Fly:

```shell
flyctl auth login
```

- Ensure production infra is ready:
  - Fly app name: `lotto-journal-api`
  - Production LINE channel created (separate from dev)
  - Neon production database created

### 2. Set production secrets on Fly

Set runtime secrets (never commit these to git):

```shell
flyctl secrets set \
  DB_DSN="postgres://...sslmode=require" \
  LINE_CHANNEL_SECRET="..." \
  LINE_CHANNEL_ACCESS_TOKEN="..." \
  CRON_SECRET="..." \
  -a lotto-journal-api
```

### 3. Deploy to Fly

```shell
flyctl deploy -a lotto-journal-api
```

### 4. Run migrations against Neon

Use the Neon connection string for production:

```shell
migrate -path apps/api/migrations -database "postgres://...sslmode=require" up
```

### 5. Verify production health

```shell
flyctl status -a lotto-journal-api
curl -i https://lotto-journal-api.fly.dev/health
```

> [!NOTE]
>
> - Since we want the VM to scale-to-zero (sleep) when idle, active periodic health checks are commented out in `fly.toml`.
> - The first request (like the `curl` above) will trigger a VM cold start, booting the VM back up (which takes 2–4 seconds). Subsequent requests will respond instantly.

### 6. Configure LINE production webhook

In LINE Developers Console (production channel):

- Webhook URL: `https://lotto-journal-api.fly.dev/webhook`
- Enable webhook delivery

### 7. Smoke test

- Add the production bot account as friend
- Send a ticket message (example: `123456 x2`)
- Verify ticket rows in Neon DB

### Runtime env map (production)

| Key                         | Where it lives        | Notes                                                                |
| --------------------------- | --------------------- | -------------------------------------------------------------------- |
| `DB_DSN`                    | Fly secrets           | Neon connection string (`sslmode=require`)                           |
| `LINE_CHANNEL_SECRET`       | Fly secrets           | Production channel secret                                            |
| `LINE_CHANNEL_ACCESS_TOKEN` | Fly secrets           | Production channel access token                                      |
| `CRON_SECRET`               | Fly secrets           | Secret token used to authorize GitHub Actions cron trigger requests  |
| `APP_ENV`                   | `fly.toml` `[env]`    | Non-secret (`production`)                                            |
| `PORT`                      | `fly.toml` `[env]`    | Non-secret (`:8080`)                                                 |
| `CRON_SYNC_SCHEDULE`        | `fly.toml` `[env]`    | Optional: Cron spec for draw sync (default: `"0 3 * * *"`)           |
| `CRON_VERIFY_SCHEDULE`      | `fly.toml` `[env]`    | Optional: Cron spec for results check (default: `"*/5 16-23 * * *"`) |
| `FLY_API_TOKEN`             | GitHub Actions secret | Used only by CI/CD deploy workflow (not app runtime)                 |

> Current production config supports **autoscaling to zero (sleep)** in primary region `sin` when there is no traffic. When active, it manages connection pooling aggressively to avoid compute leaks in Neon DB.

---

## Versioning and Release

For release process details, use:

- `doc/04-way-of-work/versioning-policy.md` — SemVer policy (`MAJOR.MINOR.PATCH`)
- `doc/04-way-of-work/release-checklist.md` — step-by-step release checklist

---

## API Testing (Bruno)

The collection lives in `trunk/bruno/`. Open it in [Bruno](https://www.usebruno.com/) by choosing **Open Collection** and selecting that folder.

### Requests

| Folder | Request            | What it tests                                                                                                    |
| ------ | ------------------ | ---------------------------------------------------------------------------------------------------------------- |
| REST   | Webhook - Follow   | User adds the bot as a friend → creates user record, sends welcome reply                                         |
| REST   | Webhook - Message  | User sends ticket numbers (e.g. `123456 x2, 789`) → parses and saves tickets                                     |
| REST   | Webhook - Unfollow | User removes the bot → marks user `inactive`                                                                     |
| REST   | Health             | `GET /health` — liveness + DB readiness; `200` ok / `503` degraded                                               |
| GLO    | Check Result       | Calls the Thai Government Lottery API directly — useful for inspecting the result payload                        |
| GLO    | Get Draw Schedule  | Calls the Thai Government Lottery API directly — useful for inspecting the draw schedule/periods payload by year |

### One-time setup

**1. Set the environment**

In Bruno, select the **dev** environment (top-right dropdown). Then open **Configure** and set the secret variable:

| Variable              | Value                                                             |
| --------------------- | ----------------------------------------------------------------- |
| `line_channel_secret` | Your channel secret from LINE Developers Console → Basic Settings |

This value is stored locally by Bruno and never committed to git.

**2. Make sure the API is running**

```shell
pnpm dev
```

The `endpoint_url` in the dev environment points to `http://localhost:3000` by default. Adjust if your `PORT` is different.

### How the signature works

Every webhook request has a pre-request script that automatically computes the `X-Line-Signature` header before sending:

```js
const crypto = require("crypto");
const secret = bru.getEnvVar("line_channel_secret");
const body = JSON.stringify(req.body);
const signature = crypto
  .createHmac("sha256", secret)
  .update(body)
  .digest("base64");
req.setHeader("X-Line-Signature", signature);
```

You don't need to compute the signature manually — just send the request.

### Idempotency caveat

Each request body contains a `webhookEventId`. On first send, this ID is written to the `webhook_events` table. Sending the **same request a second time** will be silently skipped (the deduplication is working correctly).

To re-trigger processing, change the `webhookEventId` to any unique value before sending again.

### `userId` and `destination` fields

| Field                    | What it is                                        | Does our code use it?                            |
| ------------------------ | ------------------------------------------------- | ------------------------------------------------ |
| `events[].source.userId` | The LINE user ID of the person who sent the event | **Yes** — used to find or create the user record |
| `destination`            | Your bot's own LINE user ID                       | **No** — ignored by the handler                  |

The sample bodies use a fake `userId` (`U1234567890abcdef…`). Because `FindOrCreate` is idempotent, re-running the same request always operates on the same test user record in your local DB.

---

## Scripts reference

All `pnpm` commands run from the **repo root**.
All `make` commands run from `apps/api/` — the `pnpm` shortcuts above call these automatically.

### App

| pnpm (root)    | make (apps/api) | What it does                              |
| -------------- | --------------- | ----------------------------------------- |
| `pnpm dev`     | `make run`      | Start API with hot reload (air)           |
| `pnpm build`   | `make build`    | Build binary to `dist/`                   |
| —              | `make clean`    | Remove `tmp/` and `dist/`                 |
| `pnpm swagger` | `make swagger`  | Generate Swagger spec files under `docs/` |
| `pnpm mock`    | `make mock`     | Generate Interface mocks under `mocks/`   |

### Database

| pnpm (root)     | make (apps/api) | What it does                                                           |
| --------------- | --------------- | ---------------------------------------------------------------------- |
| `pnpm db:start` | `make db-start` | Start PostgreSQL container                                             |
| `pnpm db:stop`  | `make db-stop`  | Stop PostgreSQL container (pause — data kept, restart with `db:start`) |

### Migrations

| pnpm (root)            | make (apps/api)              | What it does                                 |
| ---------------------- | ---------------------------- | -------------------------------------------- |
| `pnpm migrate:up`      | `make migrate-up`            | Apply all pending migrations                 |
| —                      | `make migrate-up-one`        | Apply the next 1 migration only              |
| `pnpm migrate:down`    | `make migrate-down`          | Roll back the last migration                 |
| —                      | `make migrate-down-all`      | Roll back all migrations                     |
| `pnpm migrate:version` | `make migrate-version`       | Show current schema version                  |
| —                      | `make migrate-force N=<ver>` | Force-set version (recover from dirty state) |

### Migration history

| Version | File                            | Description                                                                                             |
| ------- | ------------------------------- | ------------------------------------------------------------------------------------------------------- |
| 000001  | `000001_init_schema`            | Initial schema — all tables, enums, indexes                                                             |
| 000002  | `000002_line_identity`          | LINE identity redesign — replace email/password with `line_user_id`; rename `N6→L6`, `n6_*→l6_*`        |
| 000003  | `000003_webhook_events`         | Idempotency table — store processed LINE `webhookEventId` values (ON CONFLICT DO NOTHING)               |
| 000004  | `000004_widen_winning_number`   | Widen `draw_results.winning_number` to `varchar(12)` for N3 Jackpot                                     |
| 000005  | `000005_notification_logs`      | Notification logs — table for auditing outgoing push/reply messages                                     |
| 000006  | `000006_user_language`          | Add language setting preference column (defaults to `'en'`) to the `users` table                        |
| 000007  | `000007_add_notification_types` | Add new notification type enum values (`language_changed`, `help_add`, `help_notify`) to audit log enum |
| 000008  | `000008_ocr_fields_and_renames` | Add OCR sessions table, pending file associations, and R2 metadata fields                               |
