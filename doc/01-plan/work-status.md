<!-- AI-CONTEXT
src: v0.3
phase: M4
direction: Multi-language & Localization
focus: []
done: [T-000, T-001, T-005, T-008, T-004, T-007, T-006, T-002, T-010, T-011, T-012, T-013, T-014, T-016, T-018, T-015, T-017, T-019, T-003, T-023, T-022, T-021, T-024]
future: [T-020 Photo OCR+R2 — post-MVP, T-009 LIFF — post-MVP]
blocked: none
next: none
risk: none
adr: ADR-001, ADR-002
read_more:
  prd: doc/00-source/versions/v0.3/01-prd.md
  migration_design: doc/06-extensions/T-004-migration-002-design.md
  architecture: doc/07-decisions/README.md
  entities: doc/07-decisions/entity-register.md
  source_current: doc/00-source/versions/v0.3/
  walkthrough: apps/api/walkthrough.md
updated: 2026-07-24
-->

---

# Project Status — Lotto Journal

Last updated: 2026-07-24

## Source References

- `doc/00-source/versions/v0.3/01-prd.md` — current PRD (LINE-based with Localization)
- `doc/07-decisions/ADR-001-line-messaging-pivot.md` — Accepted

---

## Phase and Direction

**Current phase:** M4 — Hardening / Post-MVP (T-024 Complete)

With **T-024** complete, the API database connection pool is optimized for scale-to-zero (Neon DB), the internal scheduler is migrated to scheduled GitHub Actions workflows, and job endpoints are secured with precomputed SHA-256 token authorization checks. Additionally, the custom domain `lotto-journal.theppitak.work` has been successfully set up on Fly.io via Cloudflare, with active Edge WAF rules (unauthorized jobs blocker) and Rate Limiting active and verified in production to prevent VM wake-up abuse.

Historically, **T-021** (Multi-language & Localization support) and **T-022** (LINE Win Notification) are completed, allowing users to receive localized win/loss notifications shortly after a draw is verified in `ResultService`, with full transaction logs audited in the `notification_logs` table.

---

## Active Tasks

None. All PRD v0.3 localization features are fully implemented and verified!

---

## Completed Tasks

- `T-024` — Neon DB connection leak and Fly.io sleep optimization — done (2026-07-24)
- `T-021` — Multi-language & Localization support (EN/TH) — done (2026-06-29)
- `T-022` — Implement win notification via LINE push message — done (2026-06-28)
- `T-023` — Swagger Documentation and Mockery Setup — done (2026-06-27)
- `T-003` — Design and implement cronjob: lottery result fetch + comparison flow — done (2026-06-27)
- `T-000` — Documentation setup (2026-04-30)
- `T-001` — Architecture pivot decided: Option B (LINE Messaging API) — ADR-001 Accepted (2026-04-30)
- `T-005` — Formal source docs written: PRD v0.2 created (2026-04-30)
- `T-008` — `trunk/glo_result.json` committed by owner (2026-04-30)
- `T-004` — User identity schema designed; DBML updated; owner approved (2026-04-30)
- `T-002` — LINE webhook handler implemented; build passes (2026-05-07)
- `T-010` — Middleware: recover + requestid + enhanced logger + webhook timeout — done (2026-05-08)
- `T-011` — GET /health implemented; DB ping; 200/503 JSON response (2026-05-08)
- `T-019` — UX: loading indicator + personalized follow welcome [FOUND-IN-PASSING] — done (2026-05-11)
- `T-017` — Improvement: atomic draws upsert via GORM clause.OnConflict — done (2026-05-11)
- `T-015` — GitHub Actions CI/CD pipeline implemented and verified green — done (2026-05-11)
- `T-018` — Improve list command parsing for spaced/Unicode input [FOUND-IN-PASSING] — done (2026-05-11)
- `T-016` — Bug: ticket parsing around `x` and whitespace fixed; tests added — done (2026-05-11)
- `T-014` — First production deploy to Fly.io + Neon wiring — done (2026-05-11)
- `T-013` — Infra prep: Dockerfile + fly.toml + env secrets mapping — done (2026-05-09)
- `T-012` — Feature: list upcoming draw tickets — done (2026-05-08)
- `T-007` — Migration 000002 written (up + down); Go model + code updated; build passes (2026-04-30)
- `T-006` — `apps/web` deleted; `turbo.json` + `pnpm-workspace.yaml` cleaned up (2026-04-30)

---

## Blocked Tasks

None currently.

---

## Next Steps

1. Deploy the new features (migration `000005_notification_logs`, `NotificationService`, GORM logging logic) to Fly.io production.
2. Monitor production logs for database schema updates and outbox messaging delivery.

---

## Future Direction

- **T-020 — Photo OCR capture (post-MVP):** Add image-based ticket intake in LINE chat.
  Flow: user sends 1 image → upload to Cloudflare R2 → OpenAI extracts candidate number JSON →
  user confirms quantity (or corrects with `numberxquantity`) → save via existing ticket flow.
  See `doc/06-extensions/T-020-photo-ocr-openai-r2-proposal.md`.
- **T-009 — LIFF app (post-MVP):** A LIFF (LINE Front-end Framework) web app is planned to
  complement the chatbot. Lives in `apps/liff`. Monorepo structure (Turbo, pnpm workspaces)
  intentionally kept for this purpose. Task: design when post-MVP phase begins.
- **T-021 — Multi-language & Localization (post-MVP):** Support EN/TH localization dynamically using LINE Profile language settings combined with setting override commands.


---

## Risks and Notes

- **T-003 completed (session 15):** Implemented GLO schedule synchronization and lottery result checking. To protect the server against IP blocking by GLO, we designed a database-first schedule caching strategy where the GLO API is only queried once a day in the background. The database `draws` table serves as the single source of truth for upcoming draw dates at runtime. Migration `000004_widen_winning_number` was created and applied to accommodate 12-digit N3 Jackpot winning numbers.
- **LINE channel separation + production go-live (session 10):** Dedicated dev and production LINE channels are now both in use. Production webhook points to Fly.io app URL; end-to-end test (LINE message → DB insert in Neon) passed. Keep credentials isolated per channel and never mix them.
- **T-016 bugfix applied (session 11):** Ticket parser now correctly handles `144333 x2` and `122222 x 3`. Root cause was regex replacement tokenization (`$1x$2` in Go replacement syntax); fixed to `${1}x${2}`. Parser now also normalizes Unicode whitespace and common non-ASCII x variants (`×`, `ｘ`, `Ｘ`, `✕`). Unit tests added in `internal/service/ticket_service_test.go`.
- **T-018 [FOUND-IN-PASSING] (session 11):** `isTicketListCmd` now normalizes internal/Unicode spaces and zero-width characters so command variants like `โ พย`, `โ\u00A0พย`, and `โ\u200Bพย` correctly map to `โพย`. Unit tests added in `internal/handler/line_handler_test.go`.
- **T-015 completed (session 13):** `.github/workflows/deploy.yml` is active. Owner added repository secret `FLY_API_TOKEN` and confirmed first successful GitHub Actions run (PR checks + deploy on `main`).
- **T-017 completed (session 14):** `DrawRepository.FindOrCreate` now uses atomic `INSERT ... ON CONFLICT` via GORM `Clauses(clause.OnConflict...)` instead of `FirstOrCreate`, removing the SELECT+INSERT race window under concurrent submissions.
- **T-019 [FOUND-IN-PASSING] (session 14):** Improved chat UX by adding LINE loading indicator (`ShowLoadingAnimation`) while processing text messages and personalized follow welcome using LINE profile display name (`GetProfile`).
- **JS toolchain removed (session 7):** `.husky/`, `eslint.config.mjs`, `lint-staged.config.mjs`, `tsconfig.base.json` deleted. 8 dead devDeps removed from `package.json`. `turbo.json` trimmed to `dev`+`build` only. `.npmrc` Prisma line removed. `prettier` and `turbo` kept. Turbo updated `2.6.1`→`2.9.10`. 150 packages removed; lockfile resynced. CI/CD (T-015) will use Go toolchain directly.
- **Fiber v3 (session 6):** Upgraded from v2.52.9 → v3.2.0. All handler signatures updated (`*fiber.Ctx` → `fiber.Ctx`). `go mod tidy` removed v2 entirely. No v2 references remain.
- **Migration 000002 notes (for reference):**
  - `account_status` was drop+recreated (PostgreSQL has no `DROP VALUE`)
  - `user_winnings.user_id` was missing from 000001 SQL — added in 000002 [FOUND-IN-PASSING]
  - Enum renames: `N6`→`L6`, `n6_*`→`l6_*` (9 prize_type values)
  - `AUTH` code removed: `auth_handler.go`, `auth_service.go`, `auth_dto.go` deleted
