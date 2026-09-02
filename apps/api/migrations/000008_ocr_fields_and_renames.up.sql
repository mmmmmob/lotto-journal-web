ALTER TABLE "files" RENAME COLUMN "file_path" TO "storage_key";
ALTER TABLE "files" ADD COLUMN "extraction_status" varchar(20);
ALTER TABLE "files" ADD COLUMN "extracted_response" jsonb;
ALTER TABLE "files" ADD COLUMN "extraction_model" varchar(50);

ALTER TABLE "tickets" RENAME COLUMN "tickets_file_id" TO "ticket_file_id";

CREATE TABLE "ocr_sessions" (
    "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    "user_id" uuid NOT NULL REFERENCES "users" ("id") ON DELETE CASCADE,
    "file_id" uuid REFERENCES "files" ("id") ON DELETE SET NULL,
    "tickets" jsonb NOT NULL,
    "warnings" jsonb NOT NULL DEFAULT '[]'::jsonb,
    "status" varchar(20) NOT NULL DEFAULT 'pending',
    "current_edit_number" varchar(6) DEFAULT NULL,
    "expires_at" timestamp NOT NULL,
    "created_at" timestamp DEFAULT now(),
    "updated_at" timestamp DEFAULT now()
);

CREATE UNIQUE INDEX "uidx_ocr_sessions_user_pending" ON "ocr_sessions" ("user_id") WHERE "status" = 'pending';
