package services

import (
	"fmt"
	"log"
	"strings"
	"time"

	"flashcards/db"
	"flashcards/models"
)

type KnowledgeCheckService struct {
	repo db.KnowledgeCheckRepository
}

func NewKnowledgeCheckService(repo db.KnowledgeCheckRepository) *KnowledgeCheckService {
	return &KnowledgeCheckService{repo: repo}
}

func (s *KnowledgeCheckService) CreateKnowledgeCheck(req *models.CreateKnowledgeCheckRequest) (*models.KnowledgeCheck, error) {
	log.Printf("[INFO] Starting knowledge check creation for note ID %d", req.NoteID)

	if err := s.validateCreateRequest(req); err != nil {
		log.Printf("[ERROR] Knowledge check creation validation failed: %v", err)
		return nil, err
	}

	kc := &models.KnowledgeCheck{
		NoteID:               req.NoteID,
		LineNumberStart:      req.LineNumberStart,
		LineNumberEnd:        req.LineNumberEnd,
		State:                "pending",
		UserScore:            nil,
		UserScoreExplanation: nil,
		TopicSummary:         strings.TrimSpace(req.TopicSummary),
	}

	if err := s.repo.CreateKnowledgeCheck(kc); err != nil {
		log.Printf("[ERROR] Failed to create knowledge check in repository: %v", err)
		return nil, fmt.Errorf("failed to create knowledge check: %w", err)
	}

	log.Printf("[INFO] Successfully created knowledge check with ID %d", kc.ID)
	return kc, nil
}

func (s *KnowledgeCheckService) GetKnowledgeCheckByID(id int) (*models.KnowledgeCheck, error) {
	log.Printf("[INFO] Starting get knowledge check by ID %d", id)

	if id <= 0 {
		log.Printf("[ERROR] Invalid knowledge check ID provided: %d", id)
		return nil, fmt.Errorf("invalid knowledge check ID: %d", id)
	}

	kc, err := s.repo.GetKnowledgeCheckByID(id)
	if err != nil {
		log.Printf("[ERROR] Failed to get knowledge check by ID %d: %v", id, err)
		return nil, err
	}

	log.Printf("[INFO] Successfully retrieved knowledge check with ID %d", id)
	return kc, nil
}

func (s *KnowledgeCheckService) GetAllKnowledgeChecks() ([]*models.KnowledgeCheck, error) {
	log.Printf("[INFO] Starting get all knowledge checks")

	knowledgeChecks, err := s.repo.GetAllKnowledgeChecks()
	if err != nil {
		log.Printf("[ERROR] Failed to get all knowledge checks: %v", err)
		return nil, fmt.Errorf("failed to get knowledge checks: %w", err)
	}

	log.Printf("[INFO] Successfully retrieved %d knowledge checks", len(knowledgeChecks))
	return knowledgeChecks, nil
}

func (s *KnowledgeCheckService) GetKnowledgeChecksByDateRange(startDate *time.Time, endDate *time.Time) ([]*models.KnowledgeCheck, error) {
	log.Printf("[INFO] Starting get knowledge checks by date range")

	knowledgeChecks, err := s.repo.GetKnowledgeChecksByDateRange(startDate, endDate)
	if err != nil {
		log.Printf("[ERROR] Failed to get knowledge checks by date range: %v", err)
		return nil, fmt.Errorf("failed to get knowledge checks by date range: %w", err)
	}

	log.Printf("[INFO] Successfully retrieved %d knowledge checks for date range", len(knowledgeChecks))
	return knowledgeChecks, nil
}

func (s *KnowledgeCheckService) UpdateKnowledgeCheck(id int, req *models.UpdateKnowledgeCheckRequest) (*models.KnowledgeCheck, error) {
	log.Printf("[INFO] Starting update knowledge check with ID %d", id)

	if id <= 0 {
		log.Printf("[ERROR] Invalid knowledge check ID provided for update: %d", id)
		return nil, fmt.Errorf("invalid knowledge check ID: %d", id)
	}

	// Check if knowledge check exists and get its current state
	existingKC, err := s.repo.GetKnowledgeCheckByID(id)
	if err != nil {
		log.Printf("[ERROR] Failed to get existing knowledge check ID %d: %v", id, err)
		return nil, err
	}

	// Prevent updates to completed knowledge checks
	if existingKC.State == "completed" {
		log.Printf("[ERROR] Attempted to update completed knowledge check ID %d", id)
		return nil, fmt.Errorf("cannot update knowledge check with ID %d: knowledge check is already completed and immutable", id)
	}

	if err := s.validateUpdateRequest(req); err != nil {
		log.Printf("[ERROR] Knowledge check update validation failed for ID %d: %v", id, err)
		return nil, err
	}

	if err := s.repo.UpdateKnowledgeCheck(id, req); err != nil {
		log.Printf("[ERROR] Failed to update knowledge check ID %d in repository: %v", id, err)
		return nil, err
	}

	log.Printf("[INFO] Successfully updated knowledge check with ID %d", id)
	return s.repo.GetKnowledgeCheckByID(id)
}

func (s *KnowledgeCheckService) validateCreateRequest(req *models.CreateKnowledgeCheckRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if req.NoteID <= 0 {
		return fmt.Errorf("note ID must be positive")
	}

	if req.LineNumberStart <= 0 {
		return fmt.Errorf("line number start must be positive")
	}

	if req.LineNumberEnd <= 0 {
		return fmt.Errorf("line number end must be positive")
	}

	if req.LineNumberStart > req.LineNumberEnd {
		return fmt.Errorf("line number start cannot be greater than line number end")
	}

	topicSummary := strings.TrimSpace(req.TopicSummary)
	if topicSummary == "" {
		return fmt.Errorf("topic summary is required")
	}

	return nil
}

func (s *KnowledgeCheckService) validateUpdateRequest(req *models.UpdateKnowledgeCheckRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if req.State == nil && req.UserScore == nil && req.UserScoreExplanation == nil && req.TopicSummary == nil {
		return fmt.Errorf("at least one field must be provided for update")
	}

	if req.State != nil {
		if *req.State != "pending" && *req.State != "completed" {
			return fmt.Errorf("state must be either 'pending' or 'completed'")
		}
	}

	if req.UserScore != nil {
		if *req.UserScore < 1 || *req.UserScore > 10 {
			return fmt.Errorf("user score must be between 1 and 10")
		}
	}

	if req.TopicSummary != nil {
		topicSummary := strings.TrimSpace(*req.TopicSummary)
		if topicSummary == "" {
			return fmt.Errorf("topic summary cannot be empty")
		}
	}

	return nil
}