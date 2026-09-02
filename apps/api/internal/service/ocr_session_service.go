package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"lotto-journal/api/internal/models"
	"lotto-journal/api/internal/repository"
)



type OcrSessionService struct {
	ocrSessionRepo *repository.OcrSessionRepository
	ticketRepo     *repository.TicketRepository
	drawSvc        *DrawService
}

func NewOcrSessionService(
	ocrSessionRepo *repository.OcrSessionRepository,
	ticketRepo *repository.TicketRepository,
	drawSvc *DrawService,
) *OcrSessionService {
	return &OcrSessionService{
		ocrSessionRepo: ocrSessionRepo,
		ticketRepo:     ticketRepo,
		drawSvc:        drawSvc,
	}
}

func (s *OcrSessionService) CreateSession(userID uuid.UUID, fileID *uuid.UUID, tickets []ParsedTicket, warnings []string) (*models.OcrSession, error) {
	// Clean up any stale sessions first
	if err := s.ocrSessionRepo.DeleteExpiredPendingByUserID(userID); err != nil {
		return nil, fmt.Errorf("failed to clean expired pending sessions: %w", err)
	}

	ticketsJSON, err := json.Marshal(tickets)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tickets: %w", err)
	}

	warningsJSON, err := json.Marshal(warnings)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal warnings: %w", err)
	}

	session := &models.OcrSession{
		UserID:    userID,
		FileID:    fileID,
		Tickets:   ticketsJSON,
		Warnings:  warningsJSON,
		Status:    "pending",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := s.ocrSessionRepo.Create(session); err != nil {
		return nil, fmt.Errorf("failed to create ocr session: %w", err)
	}

	return session, nil
}

func (s *OcrSessionService) GetPendingSession(userID uuid.UUID) (*models.OcrSession, error) {
	return s.ocrSessionRepo.FindPendingByUserID(userID)
}

func (s *OcrSessionService) ConfirmSession(sessionID uuid.UUID) ([]models.Ticket, error) {
	session, err := s.ocrSessionRepo.FindByID(sessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if session.Status != "pending" {
		return nil, fmt.Errorf("session is not pending (status: %s)", session.Status)
	}

	if time.Now().After(session.ExpiresAt) {
		session.Status = "expired"
		_ = s.ocrSessionRepo.Update(session)
		return nil, fmt.Errorf("session has expired")
	}

	var parsedTickets []ParsedTicket
	if err := json.Unmarshal(session.Tickets, &parsedTickets); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tickets: %w", err)
	}

	draw, err := s.drawSvc.FindOrCreateUpcoming()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve upcoming draw: %w", err)
	}

	var savedTickets []models.Ticket
	for _, pt := range parsedTickets {
		ticket := models.Ticket{
			OwnerID:      session.UserID,
			DrawID:       draw.ID,
			Type:         pt.Type,
			Number:       pt.Number,
			Quantity:     pt.Quantity,
			TicketFileID: session.FileID,
		}
		if err := s.ticketRepo.Create(&ticket); err != nil {
			return nil, fmt.Errorf("failed to save ticket %s: %w", pt.Number, err)
		}
		savedTickets = append(savedTickets, ticket)
	}

	session.Status = "confirmed"
	if err := s.ocrSessionRepo.Update(session); err != nil {
		return nil, fmt.Errorf("failed to update ocr session: %w", err)
	}

	return savedTickets, nil
}

func (s *OcrSessionService) CancelSession(sessionID uuid.UUID) error {
	session, err := s.ocrSessionRepo.FindByID(sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	session.Status = "cancelled"
	return s.ocrSessionRepo.Update(session)
}

func (s *OcrSessionService) StartEditing(sessionID uuid.UUID, number string) error {
	session, err := s.ocrSessionRepo.FindByID(sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	session.CurrentEditNumber = &number
	return s.ocrSessionRepo.Update(session)
}

func (s *OcrSessionService) SubmitCorrection(sessionID uuid.UUID, text string) (*models.OcrSession, error) {
	session, err := s.ocrSessionRepo.FindByID(sessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if session.CurrentEditNumber == nil || *session.CurrentEditNumber == "" {
		return nil, fmt.Errorf("session is not in an active editing state")
	}

	parsed, invalid := ParseTicketInput(text)
	if len(parsed) == 0 {
		var invalidErr string
		if len(invalid) > 0 {
			invalidErr = fmt.Sprintf(" (invalid format: %s)", invalid[0])
		}
		return nil, fmt.Errorf("please send a valid 3-digit or 6-digit ticket number and optional quantity, e.g. 123456x2 or 123456%s", invalidErr)
	}
	newTicket := parsed[0]

	var tickets []ParsedTicket
	if err := json.Unmarshal(session.Tickets, &tickets); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tickets: %w", err)
	}

	targetIndex := -1
	for i, t := range tickets {
		if t.Number == *session.CurrentEditNumber {
			targetIndex = i
			break
		}
	}

	if targetIndex == -1 {
		return nil, fmt.Errorf("ticket to edit %s was not found in the session", *session.CurrentEditNumber)
	}

	// Update the ticket at targetIndex
	tickets[targetIndex] = newTicket

	// Collapse duplicate numbers by merging quantities
	var merged []ParsedTicket
	seen := make(map[string]int)
	for _, t := range tickets {
		if idx, ok := seen[t.Number]; ok {
			merged[idx].Quantity += t.Quantity
		} else {
			seen[t.Number] = len(merged)
			merged = append(merged, t)
		}
	}

	ticketsJSON, err := json.Marshal(merged)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal updated tickets: %w", err)
	}

	session.Tickets = ticketsJSON
	session.CurrentEditNumber = nil // reset editing state

	if err := s.ocrSessionRepo.Update(session); err != nil {
		return nil, fmt.Errorf("failed to update ocr session: %w", err)
	}

	return session, nil
}

func (s *OcrSessionService) CleanExpiredSessions(userID uuid.UUID) error {
	return s.ocrSessionRepo.DeleteExpiredPendingByUserID(userID)
}
