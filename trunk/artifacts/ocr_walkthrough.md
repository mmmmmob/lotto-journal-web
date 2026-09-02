# Walkthrough — R2 Storage and OpenAI Vision OCR Implementation

We have successfully implemented the OpenAI OCR and Cloudflare R2 image storage system for registering lottery tickets via photo uploads in the Lotto Journal LINE Bot, along with handler refactoring and configuration optimization.

---

## Changes Implemented

### 1. Database Migrations & Models
* **Migration**: Created [`000008_ocr_fields_and_renames.up.sql`](file:///Users/theppitak/Coding/Playground/web-playground/lotto-journal/apps/api/migrations/000008_ocr_fields_and_renames.up.sql) to add OCR metadata columns to the `files` table, rename ticket file association column to `ticket_file_id`, and construct the `ocr_sessions` table with a partial unique index ensuring a user can have at most one active `pending` session.
* **Models**:
  * Created [`file.go`](file:///Users/theppitak/Coding/Playground/web-playground/lotto-journal/apps/api/internal/models/file.go) to track R2 file uploads and extraction metadata.
  * Created [`ocr_session.go`](file:///Users/theppitak/Coding/Playground/web-playground/lotto-journal/apps/api/internal/models/ocr_session.go) representing the state machine of pending photo submissions.
  * Updated [`ticket.go`](file:///Users/theppitak/Coding/Playground/web-playground/lotto-journal/apps/api/internal/models/ticket.go) to reflect GORM field renames.

### 2. Configurations & Services
* **Config**: Added Cloudflare R2 and OpenAI API Key environment variables to [`config.go`](file:///Users/theppitak/Coding/Playground/web-playground/lotto-journal/apps/api/internal/config/config.go).
* **R2 Storage Service**: Implemented [`storage_service.go`](file:///Users/theppitak/Coding/Playground/web-playground/lotto-journal/apps/api/internal/service/storage_service.go) to downscale/resize images to max 1280px (maintaining aspect ratio), perform JPEG re-encoding, and upload to Cloudflare R2 using AWS SDK v2. Returns a base64 Data URL to pass to OpenAI.
* **OpenAI OCR Service**: Implemented [`ocr_service.go`](file:///Users/theppitak/Coding/Playground/web-playground/lotto-journal/apps/api/internal/service/ocr_service.go) utilizing the latest official `openai-go` (v1.x) SDK. Executes vision completions using `gpt-4o-mini` with Chat Completions Structured Outputs using a strict, hardcoded JSON Schema.
* **Session Service**: Implemented [`ocr_session_service.go`](file:///Users/theppitak/Coding/Playground/web-playground/lotto-journal/apps/api/internal/service/ocr_session_service.go) which manages active session state transitions (creating a pending session, editing, merging duplicates, confirming to permanent tickets, and deleting expired pending entries).

### 3. Database Connection & Logging
* **GORM Custom Logger**: Configured GORM's logger in [`db.go`](file:///Users/theppitak/Coding/Playground/web-playground/lotto-journal/apps/api/internal/database/db.go) with `IgnoreRecordNotFoundError: true` to suppress database logs about `record not found`. This silences noisy warning logs when checking if a user has a pending OCR session while keeping actual DB errors and slow queries (over 200ms) visible.

### 4. LINE Webhook & Refactoring
* **Localizer**: Added English and Thai translations for OCR confirmation headings, warnings (such as blurry image or multiple tickets), and interactive buttons to [`localizer.go`](file:///Users/theppitak/Coding/Playground/web-playground/lotto-journal/apps/api/internal/localization/localizer.go).
* **Refactoring**: Split the large 950-line `line_handler.go` file into four smaller files under the `handler` package to maintain clean organization and separation of concerns:
  1. [`line_handler.go`](file:///Users/theppitak/Coding/Playground/web-playground/lotto-journal/apps/api/internal/handler/line_handler.go): Core dependency struct, constructor, and HTTP route handling.
  2. [`line_handler_events.go`](file:///Users/theppitak/Coding/Playground/web-playground/lotto-journal/apps/api/internal/handler/line_handler_events.go): Webhook event routing and handlers for follow/unfollow/message/postback events, as well as image processing and session updates.
  3. [`line_handler_ui.go`](file:///Users/theppitak/Coding/Playground/web-playground/lotto-journal/apps/api/internal/handler/line_handler_ui.go): Reply builders, warning formats, and rich LINE Quick Reply message templates.
  4. [`line_handler_utils.go`](file:///Users/theppitak/Coding/Playground/web-playground/lotto-journal/apps/api/internal/handler/line_handler_utils.go): Command matchers, timezone-safe helpers, and URL query param parsers.

### 5. Repository Documentation & Tooling
* **Environment Configurations**: Synced all new environment variables (`R2_*` and `OPENAI_API_KEY`) to both the root [`env.example`](file:///Users/theppitak/Coding/Playground/web-playground/lotto-journal/.env.example) and package [`apps/api/.env.example`](file:///Users/theppitak/Coding/Playground/web-playground/lotto-journal/apps/api/.env.example), adding header notes distinguishing between the centralized monorepo run method and the isolated package run method.
* **Fly.io Production Env**: Updated [`fly.toml`](file:///Users/theppitak/Coding/Playground/web-playground/lotto-journal/fly.toml) with non-secret environment placeholders (`R2_ACCOUNT_ID`, `R2_BUCKET_NAME`, `R2_PUBLIC_URL_PREFIX`) and documented Fly secret locations for credentials.
* **DBDocs DBML**: Updated [`db_diagram.dbml`](file:///Users/theppitak/Coding/Playground/web-playground/lotto-journal/trunk/db_diagram.dbml) with table-level notes and detailed column descriptions for the newly added `ocr_sessions` table and modified `files`/`tickets` relation columns.
* **Documentation**: Updated the migration version history tables and environment setup guides in both the root [`README.md`](file:///Users/theppitak/Coding/Playground/web-playground/lotto-journal/README.md) and [`apps/api/README.md`](file:///Users/theppitak/Coding/Playground/web-playground/lotto-journal/apps/api/README.md).

---

## Verification Results

### Integration Tests
We wrote an integration test [`ocr_session_service_integration_test.go`](file:///Users/theppitak/Coding/Playground/web-playground/lotto-journal/apps/api/internal/service/ocr_session_service_integration_test.go) executing the full session state machine lifecycle:
* Creating a session
* Modifying ticket quantities/numbers
* Collapsing and merging duplicate numbers
* Confirming the session and generating permanent tickets in the database
* Clearing expired pending records

The tests passed successfully:
```bash
$ go test -v -tags=integration -run=TestOcrSessionService_FullLifecycle ./internal/service/...
=== RUN   TestOcrSessionService_FullLifecycle
--- PASS: TestOcrSessionService_FullLifecycle (0.08s)
PASS
ok  	lotto-journal/api/internal/service	0.728s
```

All handlers and connection logic compile and pass all tests:
```bash
$ go test -tags=integration ./...
ok  	lotto-journal/api/internal/handler	0.773s
ok  	lotto-journal/api/internal/service	0.709s
```
