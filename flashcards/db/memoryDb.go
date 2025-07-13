package db

import (
	"database/sql"
	"fmt"

	"flashcards/models"

	_ "github.com/lib/pq"
)

const AgentMemoryID = "agent"

type MemoryRepository interface {
	GetMemory() (*models.Memory, error)
	UpdateMemory(content string) error
}

type PostgresMemoryRepository struct {
	db *sql.DB
}

func NewPostgresMemoryRepository(databaseURL string) (*PostgresMemoryRepository, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresMemoryRepository{db: db}, nil
}

func (r *PostgresMemoryRepository) GetMemory() (*models.Memory, error) {
	query := `
		SELECT id, memory_content, created_at, updated_at 
		FROM gocourse.agent_memory 
		WHERE id = $1`

	memory := &models.Memory{}
	row := r.db.QueryRow(query, AgentMemoryID)

	err := row.Scan(&memory.ID, &memory.MemoryContent, &memory.CreatedAt, &memory.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			// Create the memory record if it doesn't exist
			return r.createMemory()
		}
		return nil, fmt.Errorf("failed to get memory: %w", err)
	}

	return memory, nil
}

func (r *PostgresMemoryRepository) createMemory() (*models.Memory, error) {
	query := `
		INSERT INTO gocourse.agent_memory (id, memory_content) 
		VALUES ($1, '') 
		RETURNING id, memory_content, created_at, updated_at`

	memory := &models.Memory{}
	row := r.db.QueryRow(query, AgentMemoryID)

	err := row.Scan(&memory.ID, &memory.MemoryContent, &memory.CreatedAt, &memory.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create memory: %w", err)
	}

	return memory, nil
}

func (r *PostgresMemoryRepository) UpdateMemory(content string) error {
	query := `
		UPDATE gocourse.agent_memory 
		SET memory_content = $1, updated_at = NOW() 
		WHERE id = $2`

	result, err := r.db.Exec(query, content, AgentMemoryID)
	if err != nil {
		return fmt.Errorf("failed to update memory: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("memory record not found")
	}

	return nil
}

func (r *PostgresMemoryRepository) Close() error {
	return r.db.Close()
}