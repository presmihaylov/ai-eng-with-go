package services

import (
	"fmt"
	"log"
	"strings"

	"flashcards/db"
	"flashcards/models"
)

type MemoryService struct {
	repo db.MemoryRepository
}

func NewMemoryService(repo db.MemoryRepository) *MemoryService {
	return &MemoryService{repo: repo}
}

func (s *MemoryService) GetMemory() (*models.Memory, error) {
	log.Printf("[INFO] Starting get memory")

	memory, err := s.repo.GetMemory()
	if err != nil {
		log.Printf("[ERROR] Failed to get memory: %v", err)
		return nil, fmt.Errorf("failed to get memory: %w", err)
	}

	log.Printf("[INFO] Successfully retrieved memory with %d characters", len(memory.MemoryContent))
	return memory, nil
}

func (s *MemoryService) UpdateMemory(content string) error {
	log.Printf("[INFO] Starting update memory with %d characters", len(content))

	content = strings.TrimSpace(content)

	if err := s.repo.UpdateMemory(content); err != nil {
		log.Printf("[ERROR] Failed to update memory: %v", err)
		return fmt.Errorf("failed to update memory: %w", err)
	}

	log.Printf("[INFO] Successfully updated memory")
	return nil
}