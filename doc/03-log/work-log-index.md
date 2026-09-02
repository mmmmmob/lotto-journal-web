<!-- AI-CONTEXT
last_session: 2026-07-24 (session 19)
tool: Antigravity
completed: [T-024]
in_progress: []
checkpoint: Neon DB connection leak and Fly.io VM sleep optimization (scale-to-zero) complete
next_from_last: none
notes: Optimized database connection pooling and Fly.io VM scaling settings to allow full scale-to-zero (sleep) for the serverless Neon Postgres database and Fly compute. Migrated cron schedules from in-process timers to secure HTTP endpoints triggered via GitHub Actions.
deep_context: doc/07-decisions/ADR-002-serverless-db-scale-to-zero.md
-->

---

# Work Log Index — Lotto Journal

Last updated: 2026-07-24 (session 19)

---

## Milestone Summary

_(Updated when milestones close — never archived)_

- **M0 complete (2026-04-30):** ADR-001 accepted (Option B — LINE Messaging API).
  PRD v0.2 written. Entity register updated. doc/ structure established.

---

### 2026-07-24 — Session 19 — [Antigravity]

- **Session summary:** T-024 completed. Optimised database connection pooling and Fly.io VM scaling settings to allow full scale-to-zero (sleep) for the serverless Neon Postgres database and Fly compute. Migrated cron schedules from in-process timers to secure HTTP endpoints triggered via GitHub Actions. Configured custom domain and Cloudflare edge WAF / Rate Limiter rules to prevent unauthorized VM wake-ups.
- **Work done:**
  - Modified `apps/api/internal/database/db.go` to set `MaxIdleConns(0)`, `MaxOpenConns(5)`, and `ConnMaxLifetime(3 * time.Minute)`.
  - Added `CRON_SECRET` variable in `config.go` and documented in `.env.example` (both root and api).
  - Created `job_handler.go` and registered trigger endpoints `/jobs/sync-schedule` and `/jobs/verify-results` in `main.go`.
  - Secured endpoints using `Authorization: Bearer <CRON_SECRET>` token checks.
  - Deleted the in-process `CronScheduler` (`cron_scheduler.go`) and removed the `github.com/robfig/cron/v3` dependency completely from `go.mod` (ran `go mod tidy`).
  - Updated `main.go` to run a one-off startup schedule sync in development instead of a running background scheduler thread.
  - Modified `fly.toml` to set `min_machines_running = 0` and comment out the periodic health checks.
  - Created two GitHub Actions workflows (`cron-sync-schedule.yml` and `cron-verify-results.yml`) with secure environment secrets passing to trigger endpoints on Bangkok time (ICT).
  - Added unit tests in `job_handler_test.go` and `health_handler_test.go` verifying authorization and health status codes/errors, all tests passing.
  - Configured custom domain `lotto-journal.theppitak.work` on Fly.io using Cloudflare DNS.
  - Deployed Cloudflare Custom WAF rule targeting `lotto-journal.theppitak.work/jobs/*` to block requests not carrying the matching `Bearer` token before they can trigger Fly.io VM wake-ups.
  - Deployed Cloudflare Rate Limiting rule targeting `lotto-journal.theppitak.work/health` (5 requests / 10 seconds per IP, blocked for 10 seconds) to prevent scraper/pinger spam from waking the VM.
- **Validation evidence:**
  - `go test ./...` runs and passes successfully.
  * Tested WAF rule: unauthenticated requests to `/jobs/*` return `403 Forbidden` from Cloudflare without waking the VM.
  * Tested Rate Limiting: sending 6 requests to `/health` within 10 seconds triggers the limit, returning `429 Too Many Requests` at the edge.
- **Tasks changed:**
  - T-024: done
- **Next priority:** none (All implementation, deployment, and edge protection configurations are fully verified)

---

### 2026-06-29 — Session 18 — [Antigravity]

- **Session summary:** T-021 completed. Implemented Multi-language & Localization (EN/TH) preference persistence, automatic profile language detection on follow, manual toggling commands, and dynamic Quick Replies navigation.
- **Work done:**
  - Created migration `000006_user_language` to provision `language` column on `users` table, and `000007_add_notification_types` to register new audit notification types (`language_changed`, `help_add`, `help_notify`) in the DB enum.
  - Updated GORM models `User` and `DrawTicketWithOwner` repository struct to map user language preferences.
  - Implemented `UpdateLanguage` on `UserRepository` and `UserService` to persist language settings.
  - Modified `ListTickets` signature on `TicketServiceInterface` and `TicketService` to return the draw date for localized headers.
  - Regenerated mockery mocks using `pnpm mock`.
  - Created `localization/localizer.go` providing dynamic bilingual (EN/TH) templates and LINE Quick Replies.
  - Updated `LineHandler` to query profile language on follow, switch settings on toggles, return bilingual error replies, format list output with draw date, handle quick replies instructions, and attach Quick Replies to text messages.
  - Localised win/loss notifications inside `NotificationService`.
  - Updated unit tests in `line_handler_test.go` to cover localized greetings.
- **Validation evidence:**
  - `pnpm test:api` passes successfully (with DB integration transaction verification)
  - `pnpm build` passes successfully
- **Tasks changed:**
  - T-021: done
- **Next priority:** none (Awaiting next lottery period verification or proceeding to T-020 OCR OCR+R2)

---

### 2026-06-28 — Session 17 — [Antigravity]

- **Session summary:** T-022 completed. Implemented win/loss notifications via LINE push messaging and a robust outbound logging/audit pipeline. Removed inline GORM queries from services, isolated integration tests via Go build tags, and resolved local test database state pollution.
- **Work done:**
  - Created migration `000005_notification_logs` to provision the `notification_logs` table, its types, and GORM model `NotificationLog`.
  - Updated `ticket_repository.go` to implement `FindDrawTicketsWithOwners` and `user_winning_repository.go` to implement `FindDrawWinnings` for clean encapsulation of repository join queries.
  - Implemented `FindSpecialResultByDrawID` on `DrawResultRepository` to retrieve N3 special jackpot results.
  - Built `NotificationService` to group user winnings/tickets, perform N3 jackpot checks, push formatted Thai notifications via LINE Push API with backoff retries, and write logs to `notification_logs` DB table.
  - Updated `ResultService.VerifyDrawResults` to trigger notifications asynchronously after database draws commit.
  - Updated `LineHandler` to capture resolved `draw_id` and audit all outbound reply texts (welcome, ticket submissions, ticket lists) to `notification_logs`.
  - Renamed integration test files to suffix with `_integration_test.go` and configured `//go:build integration` tag to separate fast unit tests.
  - Configured root `package.json` to execute integration tests via tag, and added `.vscode/settings.json` to configure gopls to read build tags.
  - Appended `.vscode/` folder to gitignore.
- **Validation evidence:**
  - `pnpm test:api` (and uncached go tests with `-tags=integration`) execute and pass successfully.
  - `pnpm build` passes successfully.
- **Tasks changed:**
  - T-022: done
- **Next priority:** none (PRD v0.2 MVP features fully implemented)

---

### 2026-06-27 — Session 16 — [Antigravity]

- **Session summary:** T-023 completed. Mockery v3 and Swagger documentation setups are fully integrated. Bypassed Turborepo in `dev` to resolve signal interception issues, and automated database startup/shutdown on `pnpm dev` with health-checks. Refactored `CronScheduler` to use `robfig/cron/v3` with environmental overrides for cron schedules.
- **Work done:**
  - Extracted repository and service interfaces (`interfaces.go`).
  - Configured Mockery v3 with `.mockery.yml` and generated type mocks.
  - Added Swagger specs and Fiber v3 swaggo middleware, restricted to non-production env.
  - Automated `swag init` execution in `.air.toml` during hot-reload builds, excluding `docs` and `mocks` to prevent build loops.
  - Integrated `pnpm swagger` and `pnpm mock` script bindings.
  - Configured parent shell traps in root `package.json` to handle clean database shutdowns on `Ctrl+C`.
  - Refactored `CronScheduler` to use `robfig/cron/v3` and added support for `CRON_SYNC_SCHEDULE` and `CRON_VERIFY_SCHEDULE` env variables.
  - Updated root and package READMEs to document environment variables, new targets, and dev lifecycle.
- **Validation evidence:**
  - `pnpm test:api` passes successfully
  - `pnpm build` passes successfully
  - Graceful start/stop tested and verified with single `Ctrl+C`
- **Tasks changed:**
  - T-023: done
- **Next priority:** T-022 (LINE win push notifications)

---

### 2026-06-27 — Session 15 — [Antigravity]

- **Session summary:** T-003 completed. The cronjob scheduler, GLO result checker, schedule caching database-first resolver, and ticket checking/win comparison engine are fully implemented.
- **Work done:**
  - Created migration `000004_widen_winning_number` to expand `draw_results.winning_number` from `varchar(6)` to `varchar(12)` for N3 Jackpot.
  - Implemented `LotteryClient` in `apps/api/internal/client/lottery_client.go` to connect to GLO API endpoints with retries and duplicate date filtering.
  - Updated `DrawService` in `apps/api/internal/service/draw_service.go` to resolve draw dates database-first, caching GLO schedule dates in the database and keeping a mathematical fallback.
  - Created `ResultService` in `apps/api/internal/service/result_service.go` for checking L6 and N3 ticket numbers and creating `user_winnings` in a transaction.
  - Created `CronScheduler` in `apps/api/internal/service/cron_scheduler.go` to run background checking (at 16:00 draw days) and schedule syncing (daily at 3 AM + startup) in Bangkok timezone.
  - Wired all dependencies and started the scheduler goroutine in `apps/api/app/main.go`.
  - Created unit tests in `apps/api/internal/service/result_service_test.go`.
- **Validation evidence:**
  - `pnpm test:api` passes successfully (with DB integration transaction verification)
  - `pnpm build` passes successfully
- **Tasks changed:**
  - T-003: done
- **Next priority:** T-022 (LINE win push notifications)

---

### 2026-05-11 — Session 14 — [GPT-5.3-Codex]

- **Session summary:** T-017 completed. Draw creation path is now atomic and race-safe under concurrent ticket submissions. Additional UX improvements were also delivered in this same session (T-019 [FOUND-IN-PASSING]).
- **Root cause:** `FirstOrCreate` issues `SELECT` then `INSERT`; with two simultaneous requests and no existing draw row, both can observe "not found" and both attempt insert. One loses on unique `draw_date`, causing an avoidable error path.
- **Work done:**
  - `internal/repository/draw_repository.go`:
    - Replaced `Where(...).FirstOrCreate(&draw)` with atomic upsert:
      - `db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "draw_date"}}, DoUpdates: clause.AssignmentColumns([]string{"draw_date"})}).Create(&draw)`
    - Added inline note explaining no-op update purpose (forces PostgreSQL `RETURNING` on conflict so GORM populates model fields)
  - `internal/handler/line_handler.go` [FOUND-IN-PASSING]:
    - Added LINE loading indicator via `ShowLoadingAnimation` while handling text messages
    - Added follow-event personalization: fetch LINE profile via `GetProfile` and include display name in welcome message
    - Added guard logic for loading duration (min 5s, max 60s, 5-second increments)
  - `internal/handler/line_handler_test.go`:
    - Added tests for `buildWelcomeMessage` (with/without display name)
- **Validation evidence:**
  - `pnpm test:api` passed
  - `pnpm build` passed
- **Tasks changed:**
  - T-017: done
  - T-019 [FOUND-IN-PASSING]: done
- **Next priority:** T-003 (cronjob design)
- **Daily Log:** _(local only — not committed)_

---

### 2026-05-11 — Session 13 — [GPT-5.3-Codex]

- **Session summary:** T-015 completed. GitHub Actions CI/CD is now active and verified.
- **Owner-confirmed completion evidence:**
  - Repository secret `FLY_API_TOKEN` added in GitHub Actions settings
  - Workflow run confirmed green end-to-end (PR checks + deploy on `main`)
- **Tasks changed:**
  - T-015: `review` → `done`
- **Next priority:** T-003 (cronjob design)
- **Daily Log:** _(local only — not committed)_

---

### 2026-05-11 — Session 12 — [GPT-5.3-Codex]

- **Session summary:** T-015 implementation added (CI/CD workflow), moved to `review` pending first GitHub run.
- **Work done:**
  - Added `.github/workflows/deploy.yml`:
    - Trigger: `pull_request` to `main` and `push` to `main`
    - CI job in `apps/api`: `go mod download`, `go vet ./...`, `go test ./...`, `go build`
    - Deploy job (push to main only): `flyctl deploy --remote-only --config fly.toml -a lotto-journal-api`
    - Deploy auth via `secrets.FLY_API_TOKEN`
  - Updated task tracking docs to reflect T-015 `review` status and owner follow-up requirement.
- **Validation evidence:**
  - `pnpm test:api` passed
  - `pnpm build` passed
- **Tasks changed:**
  - T-015: `todo` → `review`
- **Awaiting owner action:** Add repository secret `FLY_API_TOKEN` and verify first successful workflow run (PR checks and `main` deploy).
- **Daily Log:** _(local only — not committed)_

---

### 2026-05-11 — Session 11 — [GPT-5.3-Codex]

- **Session summary:** T-016 completed. Fixed ticket parsing bug for quantity syntax with spaces around `x` and added parser tests. Also fixed list-command recognition for spaced/Unicode `โพย` input as T-018 [FOUND-IN-PASSING].
- **Root cause:** In Go regex replacement strings, `$1x$2` is parsed as `${1x}${2}` (invalid first capture reference), which dropped the main number and left only the quantity digit as a standalone token (e.g. `2`, `3`) marked invalid.
- **Work done:**
  - `internal/service/ticket_service.go`:
    - Updated `spaceXRe` to allow optional whitespace after `x`: `(?i)(\d+)\s+x\s*(\d+)`
    - Fixed merge replacement string from `$1x$2` to `${1}x${2}`
    - Added `normalizeTicketText()` to normalize:
      - commas to spaces
      - Unicode whitespace to ASCII space
      - common non-ASCII x characters (`×`, `✕`, `ｘ`, `Ｘ`) to `x`
  - Added `internal/service/ticket_service_test.go` with unit tests covering:
    - `144333 x2`
    - `122222 x 3`
    - Unicode variants such as `123456×2`, `123456\u00A0×\u00A02`, and `456ｘ4`
  - `internal/handler/line_handler.go` [FOUND-IN-PASSING]: updated `isTicketListCmd` to normalize internal/Unicode whitespace and zero-width characters before matching `โพย`
  - Added `internal/handler/line_handler_test.go` to cover command variants:
    - `โพย`, `  โพย  `, `โ พย`, `โ\u00A0พย`, `โ\u200Bพย`
    - and negative cases (`ขอโพย`, `โพยครับ`)
- **Validation evidence:**
  - `pnpm --filter @lotto/api exec go test ./...` passed
  - `pnpm build` passed
- **Tasks changed:**
  - T-016: done
  - T-018 [FOUND-IN-PASSING]: done
- **Awaiting owner action:** Proceed with T-015 (GitHub Actions CI/CD pipeline)
- **Daily Log:** _(local only — not committed)_

---

### 2026-05-11 — Session 10 — [GPT-5.3-Codex]

- **Session summary:** T-014 (first production deploy to Fly.io + Neon wiring) completed by owner. App is live and verified end-to-end.
- **Work done (owner-confirmed):**
  - Deployed production app to Fly.io (`lotto-journal-api`)
  - Running footprint set to **1 machine** in `sin` (cost-aware MVP setup)
  - Production Neon database connected via `DB_DSN`
  - Applied schema migrations to Neon successfully
  - Configured production LINE webhook to Fly app URL
  - Smoke-tested by sending ticket numbers in LINE; rows appeared in Neon DB explorer
- **Operational notes captured this session:**
  - Production deployment becomes the base target for upcoming T-015 CI/CD (`main` branch deploy)
  - T-015 is now unblocked and promoted to next priority
- **Tasks changed:**
  - T-014: done
  - T-015: unblocked / ready
- **Awaiting owner action:** Start T-015 implementation (`.github/workflows/deploy.yml` + `FLY_API_TOKEN` secret)
- **Daily Log:** _(local only — not committed)_

---

### 2026-05-09 — Session 9 — [GPT-5.3-Codex]

- **Session summary:** T-013 (infra prep) completed. Production deployment scaffolding is now in-repo and T-014 is unblocked.
- **Work done:**
  - Added root `Dockerfile` (multi-stage): builds `apps/api/app/main.go` to a static binary and runs it in a minimal Alpine runtime image as non-root user
  - Added root `fly.toml`: app config with `primary_region = "sin"`, Docker build source, `APP_ENV=production`, `PORT=:8080`, `internal_port = 8080`, and `/health` HTTP service check
  - Added root `.dockerignore` to reduce build context size and exclude secrets/local artifacts
  - Updated `doc/02-task/task-board.md`:
    - T-013 marked `done`
    - T-014 steps updated to use `DB_DSN` secret key (instead of `DATABASE_URL`)
    - Env Map changed to `DB_DSN`
    - Blocked table updated (T-014 unblocked)
    - Added T-013 completion evidence row
  - Updated `doc/01-plan/work-status.md` to reflect T-013 completion and T-014 as infra priority
- **Decisions resolved this session:**
  - Keep application config unchanged for DB connection key (`DB_DSN`) as requested by owner
  - Fly non-secret runtime env values committed in `fly.toml`; secrets remain external (`fly secrets set`)
- **Validation evidence:**
  - `pnpm build` passed successfully (`turbo run build` → `@lotto/api make build`)
- **Tasks changed:**
  - T-013: done
  - T-014: unblocked
- **Awaiting owner action:** Execute T-014 deployment steps on Fly.io + Neon with production LINE channel credentials
- **Daily Log:** _(local only — not committed)_

---

### 2026-05-08 — Session 8 — [Claude (Sonnet 4.6)]

- **Session summary:** T-012 (list upcoming draw tickets) implemented in a guided learning session — owner wrote all the code. Build passes.
- **Work done:**
  - `internal/repository/ticket_repository.go`: added `List(drawId, ownerID uuid.UUID) ([]*models.Ticket, error)` — queries tickets by `owner_id` AND `draw_id` via GORM `Where` + `Find`
  - `internal/service/ticket_service.go`: added `ListTickets(ownerID uuid.UUID) ([]*models.Ticket, error)` — resolves upcoming draw via `FindOrCreate`, delegates to `ticketRepo.List`
  - `internal/handler/line_handler.go`: added `isTicketListCmd(text string) bool` helper (keyword: "โพย"); added `buildTicketListReply(tickets []*models.Ticket) string` with early-return empty state; wired both into `handleMessage` if/else routing
- **Decisions resolved this session:**
  - Ticket list lives in `TicketRepository`, not `DrawRepository` — single responsibility
  - Service resolves draw internally (same pattern as `SubmitTickets`) — caller only needs `ownerID`
  - `FindOrCreate` used instead of `FindByDate` — handles case where no draw exists yet (returns empty list, not error)
  - Keyword detection is a small helper, not inline — cleaner `handleMessage`
  - Separate `buildTicketListReply` instead of extending `buildReply` — different input shape (`[]*models.Ticket` vs `[]ParsedTicket`)
  - Empty ticket list: early return before building `lines` slice — clean, no header shown
- **Tasks changed:**
  - T-012: done (build passes)
- **Awaiting owner action:** ~~Test via LINE bot with keyword "โพย"~~ — tested and passed
- **Post-session owner actions:**
  - Created dedicated dev LINE channel (separate from future production channel); updated local `.env` with new `LINE_CHANNEL_SECRET` and `LINE_CHANNEL_ACCESS_TOKEN` — dev/prod channel separation in place ahead of T-014
- **Daily Log:** _(local only — not committed)_

---

### 2026-05-08 — Session 7 — [Claude (Sonnet 4.6)]

- **Session summary:** Planning and housekeeping session. No feature code written. Added 5 new tasks to the board (T-012 to T-016). Removed the entire dead JS toolchain left over from the pre-pivot web app era. Upgraded turbo.
- **Work done:**
  - `doc/02-task/task-board.md`: added T-012 (ticket summary feature), T-013 (Dockerfile + fly.toml + env map), T-014 (first Fly.io deploy), T-015 (GitHub Actions CI/CD), T-016 (ticket parsing bug); added Env Map section documenting which secrets go to Fly.io vs GitHub Actions
  - `doc/01-plan/work-status.md`: added T-012 to T-016 to active tasks and next steps
  - Deleted `.husky/` — pre-commit hook + all husky internals (no JS/TS to lint)
  - Deleted `eslint.config.mjs` — was targeting deleted `apps/web` + non-existent JS files in Go API
  - Deleted `lint-staged.config.mjs` — no `*.{js,jsx,ts,tsx}` files exist
  - Deleted `tsconfig.base.json` — no TypeScript anywhere in the project
  - `package.json`: removed 8 devDeps (`@eslint/eslintrc`, `@eslint/js`, `@typescript-eslint/eslint-plugin`, `@typescript-eslint/parser`, `eslint`, `globals`, `husky`, `lint-staged`); removed scripts `lint`, `typecheck`, `lint-staged`; kept `prettier`, `turbo`, `format`, `format:check`
  - `turbo.json`: removed `lint` and `typecheck` tasks; kept `dev` and `build`
  - `.prettierignore`: removed `apps/web` and `.next` / `out` references
  - `.npmrc`: removed `package-build-deps=@prisma/client,@prisma/engines,prisma` (Prisma left with apps/web); kept `save-workspace-protocol=true`
  - `turbo` updated `2.6.1` → `2.9.10`; `pnpm install` run — 150 packages removed, lockfile resynced
- **Decisions resolved this session:**
  - Dead JS toolchain removed — no JS/TS source files remain; CI/CD (T-015) will use Go tools directly (`go build`, `go vet`, `go test`)
  - `prettier` kept — useful for doc/yaml/markdown formatting; `format:check` can be added to CI cheaply
  - `turbo` + `pnpm-workspace.yaml` kept — monorepo shell intentionally preserved for future LIFF app (T-009)
  - `save-workspace-protocol=true` kept in `.npmrc` — correct pnpm default for workspaces, relevant when LIFF lands
  - Env Map added to task board as a permanent reference (not a task): Fly.io secrets hold runtime vars; GitHub Actions holds only `FLY_API_TOKEN`; Neon itself holds no secrets
- **Tasks changed:**
  - T-012: added (todo)
  - T-013: added (todo)
  - T-014: added (todo, blocked by T-013)
  - T-015: added (todo, blocked by T-014)
  - T-016: added (todo — two broken parsing cases documented with screenshot evidence)
- **Awaiting owner action:** None
- **Daily Log:** _(local only — not committed)_

---

### 2026-05-08 — Session 6 — [Claude (Sonnet 4.6)]

- **Session summary:** T-010 (middleware hardening) implemented, then immediately upgraded to Fiber v3 after the deprecated `timeout.New` warning surfaced. Fetched official Fiber v3 docs, migrated the full codebase from v2 → v3 (v3.2.0). T-011 (GET /health) implemented. Updated `apps/api/README.md` and `trunk/webhook-flow.md` to reflect new middleware stack and corrected a false LINE retry-interval claim. Build passes.
- **Work done:**
  - `internal/handler/health_handler.go`: created; `GET /health`; calls `db.DB().Ping()`; returns `{"status":"ok","db":"ok"}` (200) or `{"status":"degraded","db":"<err>"}` (503)
  - `app/main.go`: wired `healthHandler`; registered `GET /health` (no timeout wrapper)
  - `middlewares/log.go`: upgraded to log status code + request ID; `c.Locals("requestid")` → `requestid.FromContext(c)` (v3 API); handler sig `*fiber.Ctx` → `fiber.Ctx`
  - `app/main.go`: `recoverer.New(EnableStackTrace: true)` globally; `requestid.New()` globally; `/webhook` wrapped with `timeout.New(handler, timeout.Config{Timeout: 25s})` (v3 race-free timeout); `log.Fatal(app.Listen(...))` (v3 always returns error); `recover` import aliased as `recoverer` (v3 idiom)
  - `internal/handler/line_handler.go`: import `v2` → `v3`; handler sig `*fiber.Ctx` → `fiber.Ctx`
  - `go.mod`: added `github.com/gofiber/fiber/v3 v3.2.0`; `go mod tidy` removed v2 entirely
  - `apps/api/README.md`: added **Middleware stack** section (table + log format + timeout rationale with LINE redelivery reference)
  - `trunk/webhook-flow.md`: updated Big Picture diagram (middleware chain shown); Step 1 (middleware registrations + timeout-wrapped route); Step 2 (handler sig); Step 5 (removed false "within 30 seconds" claim, linked LINE docs); Files Involved table
- **Decisions resolved this session:**
  - Fiber v3 replaces v2 — `timeout.New` in v3 is race-free via Abandon mechanism; no `NewWithContext` needed
  - `requestid.FromContext(c)` is the v3 accessor — v3 drops the `ContextKey` config field
  - `recover` middleware import aliased as `recoverer` to avoid shadowing the Go built-in
  - LINE's webhook retry count/interval is not publicly disclosed — removed the false "30 seconds" claim from webhook-flow.md
- **Tasks changed:**
  - T-010: done (build passes, all middleware wired, Fiber v3)
  - T-011: done (build passes)
- **Awaiting owner action:** None
- **Daily Log:** _(local only — not committed)_

---

### 2026-05-07 — Session 5 — [Claude (Sonnet 4.6)]

- **Session summary:** T-002 (LINE webhook handler) fully designed and implemented. LINE Bot SDK v8 added. All layers built from scratch (models, repos, services, handler). Migration 000003 added for idempotency. Build passes.
- **Work done:**
  - Added `github.com/line/line-bot-sdk-go/v8` (v8.20.0) to go.mod
  - Created `migrations/000003_webhook_events.up/down.sql` — idempotency table
  - Created `models/draw.go`, `models/ticket.go`, `models/webhook_event.go`
  - Updated `repository/user_repository.go` — added `FindByLineUserID`, `FindOrCreate`, `UpdateStatus`
  - Created `repository/draw_repository.go` — `FindByDate`, `FindOrCreate`
  - Created `repository/ticket_repository.go` — `Create`
  - Created `repository/webhook_event_repository.go` — atomic `MarkProcessed` (ON CONFLICT DO NOTHING)
  - Created `service/user_service.go` — `FindOrCreate`, `Deactivate`
  - Created `service/draw_service.go` — `NextDrawDate` (Bangkok time, 1st/16th logic), `FindOrCreateUpcoming`
  - Created `service/ticket_service.go` — `ParseTicketInput` (commas+spaces, xN quantity, 3/6-digit validation), `SubmitTickets`
  - Created `handler/line_handler.go` — Fiber→SDK bridge (synthetic `*http.Request`), `follow`/`unfollow`/`message` routing, reply builder
  - Updated `config/config.go` — added `LINE_CHANNEL_SECRET`, `LINE_CHANNEL_ACCESS_TOKEN`
  - Updated `app/main.go` — wired all layers; registered `POST /webhook`
- **Decisions resolved this session:**
  - LINE SDK bridge pattern: build synthetic `*http.Request` from Fiber context to pass to `webhook.ParseRequest()`
  - Idempotency: `webhook_events` table with `ON CONFLICT DO NOTHING` insert + `RowsAffected` check
  - `NextDrawDate`: Bangkok time, candidates = [1st, 16th of current month, 1st of next month], first >= today
  - `ParseTicketInput`: normalise commas→spaces, merge `\d+ x\d+` pattern, validate 3/6-digit only
  - `UserSource` type assertion is value type (`webhook.UserSource`, not pointer)
  - Microcopy: Thai language placeholder (PRD marks it TBD)
- **Tasks changed:**
  - T-002: done (build passes, all events handled)
- **Awaiting owner action:** Run `make migrate-up` against the DB to apply migration 000003 before testing
- **Daily Log:** _(local only — not committed)_

---

### 2026-04-30 — Session 4 — [Claude (Sonnet 4.6)]

- **Session summary:** T-004 design completed. T-007 (migration 000002) written. T-006 (remove apps/web) done. LIFF planned as T-009 post-MVP. T-008 closed (glo_result.json committed by owner).
- **Work done:**
  - Analysed migration 000002 scope; produced design doc `doc/06-extensions/T-004-migration-002-design.md`
  - Updated `trunk/db_diagram.dbml` to post-000002 target state
  - Written `000002_line_identity.up/down.sql`; updated `models/user.go`; removed auth code; build passes
  - Deleted `apps/web`; cleaned `turbo.json` (removed generate/prisma tasks); cleaned `pnpm-workspace.yaml`
  - Extended `Makefile` and `package.json` scripts (db:start, db:stop, migrate:\* targets)
  - Updated README with pnpm-first setup guide
  - Added T-009 (LIFF planning) to task board as post-MVP future task
  - Added LIFF note to PRD v0.2 §8 (Out of Scope) — intentionally deferred, not abandoned
- **Decisions resolved this session:**
  - `account_status` enum: DROP `pending`, ADD `inactive`, KEEP `active`+`suspended`
  - `user_winnings.user_id` [FOUND-IN-PASSING]: added in migration 000002
  - Monorepo structure kept intentionally — LIFF app will use `apps/liff` in future
- **Tasks changed:**
  - T-008: done (committed by owner)
  - T-004: done (design approved)
  - T-007: done (migration written, build passes)
  - T-006: done (apps/web deleted, configs cleaned)
  - T-009: added (future / post-MVP)
- **Awaiting owner action:** None — ready to start T-002 or T-003
- **Daily Log:** _(local only — not committed)_

---

- **Session summary:** Owner filled in all `<NEEDS_CLARIFICATION>` placeholders in PRD v0.2
  by direct file edit. PRD status effectively complete (only microcopy deferred).
- **Decisions resolved:**
  - Message input format: plain number(s), comma/space separated, optional `xN` quantity suffix
  - Ticket type terminology: `L6` (not `N6`) for 6-digit; `N3` unchanged
  - GLO API endpoint: `POST https://www.glo.or.th/api/lottery/getLatestLottery`
  - GLO API response format: documented in `trunk/glo_result.json` (to be committed)
  - Draw schedule: 1st & 16th monthly, **configurable** to handle holiday exceptions
  - Cronjob trigger time: 16:00 Bangkok time (GMT+0700)
  - Retry on API failure: 5 times
  - Non-win notification: YES — send push message even for non-winning tickets
  - `unfollow` event: mark user `status = inactive` (soft delete)
  - `user_profiles` table: DROP in migration 000002
  - `files` table: KEEP — reserved for future photo upload MVP
  - On-demand status query (§3.4): post-MVP
  - Scalability: ≤ 100 users for MVP
  - Monitoring: stdout logs only for MVP
- **New issues flagged by AI review:**
  1. **Enum rename** — PRD now uses `L6` / `l6_*` but migration 000001 created `N6` / `n6_*`.
     T-007 scope expanded to include `ALTER TYPE` rename statements.
  2. **`trunk/glo_result.json` missing** — T-008 added to track committing this file.
- **Tasks added:** `T-008` (commit glo_result.json)
- **Daily Log:** _(local only — not committed)_

---

### 2026-04-30 — Session 2 — [Claude (Sonnet 4.6)]

- **Session summary:** Accepted ADR-001 Option B (LINE Messaging API pivot). Updated ADR
  status to Accepted, filled in rationale and consequences. Updated entity register
  (LINE Messaging API active, Next.js deprecated). Wrote PRD v0.2
  (`doc/00-source/versions/v0.2/01-prd.md`). Updated all planning docs to reference v0.2.
  Advanced project phase from Pre-M0 → M1.
- **Tasks completed:** `T-001` (ADR-001 accepted), `T-005` (PRD v0.2 written)
- **Key decisions recorded in PRD v0.2:**
  - User identity: `line_user_id` replaces email/password
  - Tables to drop: `user_auth_methods`, `user_verifications`
  - Enums to drop: `provider_service`, `verification_type`
  - `users` table redesign → migration 000002 (T-007)
  - `apps/web` to be removed (T-006)
- **Open placeholders in PRD v0.2 that need resolution before implementation:**
  - Thai Gov Lottery API endpoint and response format (blocks T-003)
  - LINE message input format for ticket submission (blocks T-002 implementation)
  - Whether `user_profiles` and `files` tables should be kept or dropped
  - Whether to send "no win" notifications or only win notifications
  - Draw day schedule (1st & 16th assumed — needs confirmation)
- **Daily Log:** _(local only — not committed)_

---

### 2026-04-30 — Session 1 — [Claude (Sonnet 4.6)]

- **Session summary:** Initial documentation setup. Read all 19 core templates (00–18).
  Explored the existing codebase to understand current state before creating any docs.
- **Key findings from code exploration:**
  - `apps/api`: Go + Fiber setup with partial `SignUp` handler (hardcoded password — not production-ready)
  - `apps/web`: Next.js skeleton only (boilerplate pages, no business logic)
  - `trunk/db_diagram.dbml` + migration: Full DB schema exists — users, tickets, draws,
    draw_results, user_winnings, files, enums (account_status, lottery_type, prize_type, etc.)
  - Current `users` table is built for email/password + OAuth — does NOT have `line_user_id`
  - Lottery data model looks solid and likely survives the pivot
- **Tasks completed:** `T-000` (doc setup)
- **ADR created:** ADR-001 (Proposed) — architecture pivot web → LINE Messaging API
- **Validation:** Bootstrap checklist verified before closing session
- **Daily Log:** _(local only — not committed)_
