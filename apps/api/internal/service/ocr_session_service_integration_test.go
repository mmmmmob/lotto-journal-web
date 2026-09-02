//go:build integration

package service_test

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"lotto-journal/api/internal/models"
	"lotto-journal/api/internal/repository"
	"lotto-journal/api/internal/service"
)

func TestOcrSessionService_FullLifecycle(t *testing.T) {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "postgres://lotto:password@localhost:5432/lotto_db?sslmode=disable"
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("Skipping OCR session integration test; cannot connect to DB: %v", err)
		return
	}

	tx := db.Begin()
	defer tx.Rollback()

	// Initialize repositories with the transaction tx
	userRepo := repository.NewUserRepository(tx)
	drawRepo := repository.NewDrawRepository(tx)
	ticketRepo := repository.NewTicketRepository(tx)
	fileRepo := repository.NewFileRepository(tx)
	ocrSessionRepo := repository.NewOcrSessionRepository(tx)

	// Create user
	user := &models.User{
		ID:         uuid.New(),
		LineUserID: "U1234567890",
		Language:   "en",
		Status:     "active",
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Create draw
	drawDate, _ := time.Parse("2006-01-02", "2026-07-01")
	_, err = drawRepo.FindOrCreate(drawDate)
	if err != nil {
		t.Fatalf("failed to resolve draw: %v", err)
	}

	// Create file
	fileRecord := &models.File{
		ID:               uuid.New(),
		OwnerID:          &user.ID,
		StorageKey:       "tickets/test.jpg",
		FileType:         ptrString("image/jpeg"),
		ExtractionStatus: ptrString("success"),
	}
	if err := fileRepo.Create(fileRecord); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	// Services
	drawSvc := service.NewDrawService(drawRepo, nil)
	ocrSessionSvc := service.NewOcrSessionService(ocrSessionRepo, ticketRepo, drawSvc)

	// 1. Create session
	inputTickets := []service.ParsedTicket{
		{Number: "123456", Quantity: 1, Type: "L6"},
		{Number: "789", Quantity: 2, Type: "N3"},
	}
	warnings := []string{"low_confidence"}

	session, err := ocrSessionSvc.CreateSession(user.ID, &fileRecord.ID, inputTickets, warnings)
	if err != nil {
		t.Fatalf("failed to create ocr session: %v", err)
	}

	if session.Status != "pending" {
		t.Errorf("expected pending status, got %s", session.Status)
	}

	// 2. Get pending session
	retrieved, err := ocrSessionSvc.GetPendingSession(user.ID)
	if err != nil {
		t.Fatalf("failed to get pending session: %v", err)
	}
	if retrieved == nil || retrieved.ID != session.ID {
		t.Fatalf("retrieved session does not match created session")
	}

	// 3. Start editing and Submit correction (edit 789 -> 123456x2 to trigger merge)
	if err := ocrSessionSvc.StartEditing(session.ID, "789"); err != nil {
		t.Fatalf("failed to start editing: %v", err)
	}

	session, err = ocrSessionSvc.SubmitCorrection(session.ID, "123456x2")
	if err != nil {
		t.Fatalf("failed to submit correction: %v", err)
	}

	// Verify merge result: "123456" quantity should be 1 + 2 = 3
	var updatedTickets []service.ParsedTicket
	if err := json.Unmarshal(session.Tickets, &updatedTickets); err != nil {
		t.Fatalf("failed to unmarshal updated tickets: %v", err)
	}

	if len(updatedTickets) != 1 {
		t.Errorf("expected 1 ticket after merge, got %d", len(updatedTickets))
	} else {
		ticket := updatedTickets[0]
		if ticket.Number != "123456" || ticket.Quantity != 3 {
			t.Errorf("expected merged ticket 123456x3, got %s x%d", ticket.Number, ticket.Quantity)
		}
	}

	// 4. Confirm session
	tickets, err := ocrSessionSvc.ConfirmSession(session.ID)
	if err != nil {
		t.Fatalf("failed to confirm session: %v", err)
	}

	if len(tickets) != 1 {
		t.Errorf("expected 1 ticket registered, got %d", len(tickets))
	} else {
		tkt := tickets[0]
		if tkt.Number != "123456" || tkt.Quantity != 3 || *tkt.TicketFileID != fileRecord.ID {
			t.Errorf("unexpected confirmed ticket: %+v", tkt)
		}
	}

	// Session status should be confirmed
	confirmedSession, err := ocrSessionRepo.FindByID(session.ID)
	if err != nil {
		t.Fatalf("failed to fetch session: %v", err)
	}
	if confirmedSession.Status != "confirmed" {
		t.Errorf("expected confirmed status, got %s", confirmedSession.Status)
	}

	// 5. Clean expired sessions (make a session expired first)
	expiredSession := &models.OcrSession{
		UserID:    user.ID,
		FileID:    &fileRecord.ID,
		Tickets:   session.Tickets,
		Warnings:  session.Warnings,
		Status:    "pending",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	if err := ocrSessionRepo.Create(expiredSession); err != nil {
		t.Fatalf("failed to create expired session: %v", err)
	}

	if err := ocrSessionSvc.CleanExpiredSessions(user.ID); err != nil {
		t.Fatalf("failed to clean expired sessions: %v", err)
	}

	// Verify expired session is deleted
	_, err = ocrSessionRepo.FindByID(expiredSession.ID)
	if err == nil {
		t.Errorf("expected expired session to be deleted, but it still exists")
	}
}

func ptrString(s string) *string {
	return &s
}
