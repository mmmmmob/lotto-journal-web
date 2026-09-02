DROP INDEX IF EXISTS "uidx_ocr_sessions_user_pending";
DROP TABLE IF EXISTS "ocr_sessions";

ALTER TABLE "tickets" RENAME COLUMN "ticket_file_id" TO "tickets_file_id";

ALTER TABLE "files" DROP COLUMN IF EXISTS "extraction_model";
ALTER TABLE "files" DROP COLUMN IF EXISTS "extracted_response";
ALTER TABLE "files" DROP COLUMN IF EXISTS "extraction_status";
ALTER TABLE "files" RENAME COLUMN "storage_key" TO "file_path";
