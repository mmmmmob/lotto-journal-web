# apps/api — Go API Service

The backend service for Lotto Journal. Built with Go + [Fiber v3](https://github.com/gofiber/fiber).

> **Setup and running instructions are in the [root README](../../README.md).**
> This document covers API-specific development reference.

---

## Make targets

All commands run from `apps/api/`.

### App

| Command      | What it does                               |
| ------------ | ------------------------------------------ |
| `make run`   | Start API with hot reload (air)            |
| `make build` | Build binary to `dist/`                    |
| `make clean` | Remove `tmp/` and `dist/`                  |
| `make swagger`| Generate Swagger spec files under `docs/` |
| `make mock`  | Generate Interface mocks under `mocks/`    |

### Database

| Command         | What it does                                                                |
| --------------- | --------------------------------------------------------------------------- |
| `make db-start` | Start PostgreSQL container                                                  |
| `make db-stop`  | Stop PostgreSQL container (pause — data kept, restart with `make db-start`) |

### Migrations

| Command                      | What it does                                        |
| ---------------------------- | --------------------------------------------------- |
| `make migrate-up`            | Apply all pending migrations                        |
| `make migrate-up-one`        | Apply the next 1 migration only                     |
| `make migrate-down`          | Roll back the last migration                        |
| `make migrate-down-all`      | Roll back all migrations                            |
| `make migrate-version`       | Show current schema version                         |
| `make migrate-force N=<ver>` | Force-set version (use to recover from dirty state) |

---

## Routes

| Method | Path                    | Handler        | Timeout | Description                                                 |
| ------ | ----------------------- | -------------- | ------- | ----------------------------------------------------------- |
| `POST` | `/webhook`              | `LineHandler`  | 25 s    | LINE webhook receiver — all chatbot events                  |
| `GET`  | `/health`               | `HealthHandler`| none    | Liveness + DB readiness; `200` ok / `503` degraded          |
| `GET`  | `/swagger/*`            | Swaggo Handler | none    | Swagger UI documentation (available in dev/staging only)    |
| `POST` | `/jobs/sync-schedule`   | `JobHandler`   | none    | Triggers draw schedule sync (Secured with `CRON_SECRET`)    |
| `POST` | `/jobs/verify-results`  | `JobHandler`   | none    | Triggers result checking/verification (Secured with `CRON_SECRET`) |

---

## Middleware stack

Global middleware runs on every request in this order:

| Order | Middleware  | Scope           | Behaviour                                                                     |
| ----- | ----------- | --------------- | ----------------------------------------------------------------------------- |
| 1     | `recoverer` | Global          | Catches panics, returns 500; prints stack trace to stdout                     |
| 2     | `requestid` | Global          | Generates a `X-Request-ID` UUID; accessible via `requestid.FromContext(c)`    |
| 3     | `Logging`   | Global          | Logs `[METHOD] /path - STATUS - duration - req_id: UUID` after every response |
| —     | `timeout`   | `/webhook` only | Returns 408 if the handler does not respond within **25 s**                   |

#### Why does `/webhook` need a timeout?

When the LINE Platform delivers a webhook and does **not** receive a `2xx` response (including
no response at all), it treats the delivery as failed and **redelivers** the event — potentially
multiple times. Without a timeout, a hanging handler (e.g. DB stall, downstream API stuck)
produces two compounding problems:

1. **Goroutine leak** — Fiber spawns a goroutine per request; a hung handler holds it open indefinitely.
2. **Retry storm** — LINE keeps redelivering; each retry spawns another goroutine against the already-stalled server.

The 25 s timeout guarantees the server always emits a response code (`200` on success, `408` on
overrun). LINE receives the `408`, logs it, and does not add to the backlog. The `webhookEventId`
idempotency table (migration 000003) prevents a redelivered event from being processed twice if
the first attempt succeeded before the timeout fired.

> **Note:** LINE's redelivery count and interval are not publicly disclosed and may change.
> See [LINE docs — Redeliver a webhook that failed to be received](https://developers.line.biz/en/docs/messaging-api/receiving-messages/#redeliver-a-webhook-that-failed-to-be-received).

Log line format:

```
[POST] /webhook - 200 - 1.234ms - req_id: 550e8400-e29b-41d4-a716-446655440000
```

---

## Project structure

```
apps/api/
├── app/
│   └── main.go              # Entry point
├── docs/                    # Swagger spec files
├── internal/
│   ├── client/              # GLO API client
│   ├── config/              # Env config loader
│   ├── database/            # DB connection
│   ├── handler/             # HTTP handlers (split into line_handler*.go for modularity)
│   ├── mocks/               # Generated mockery files
│   ├── models/              # GORM models
│   ├── repository/          # DB access layer
│   └── service/             # Business logic
├── migrations/              # SQL migration files
├── middlewares/             # Fiber middlewares
├── .mockery.yml             # Mockery v3 config
├── Makefile
└── go.mod
```

---

## Adding a new migration

1. Create two files in `migrations/` following the naming convention:
   ```
   000004_<description>.up.sql
   000004_<description>.down.sql
   ```
2. Write the `up` SQL (schema change) and the `down` SQL (full reversal).
3. Apply with `make migrate-up-one` and verify with `make migrate-version`.
4. Always test the `down` migration too: `make migrate-down` then `make migrate-up-one`.

### Migration history

| Version | File                            | Description                                                                                      |
| ------- | ------------------------------- | ------------------------------------------------------------------------------------------------ |
| 000001  | `000001_init_schema`            | Initial schema — all tables, enums, indexes                                                      |
| 000002  | `000002_line_identity`          | LINE identity redesign — replace email/password with `line_user_id`; rename `N6→L6`, `n6_*→l6_*` |
| 000003  | `000003_webhook_events`         | Idempotency table — store processed LINE `webhookEventId` values (ON CONFLICT DO NOTHING)        |
| 000004  | `000004_widen_winning_number`   | Widen `draw_results.winning_number` to `varchar(12)` for N3 Jackpot                              |
| 000005  | `000005_notification_logs`      | Notification logs — table for auditing outgoing push/reply messages                              |
| 000006  | `000006_user_language`          | Add language setting preference column (defaults to `'en'`) to the `users` table                 |
| 000007  | `000007_add_notification_types` | Add new notification type enum values (`language_changed`, `help_add`, `help_notify`) to audit log enum |
| 000008  | `000008_ocr_fields_and_renames` | Add OCR sessions table, pending file associations, and R2 metadata fields                        |
