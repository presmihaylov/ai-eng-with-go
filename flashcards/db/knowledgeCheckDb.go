package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"flashcards/models"

	_ "github.com/lib/pq"
)

type KnowledgeCheckRepository interface {
	CreateKnowledgeCheck(kc *models.KnowledgeCheck) error
	GetKnowledgeCheckByID(id int) (*models.KnowledgeCheck, error)
	GetAllKnowledgeChecks() ([]*models.KnowledgeCheck, error)
	GetKnowledgeChecksByDateRange(startDate *time.Time, endDate *time.Time) ([]*models.KnowledgeCheck, error)
	UpdateKnowledgeCheck(id int, req *models.UpdateKnowledgeCheckRequest) error
}

type PostgresKnowledgeCheckRepository struct {
	db *sql.DB
}

func NewPostgresKnowledgeCheckRepository(databaseURL string) (*PostgresKnowledgeCheckRepository, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresKnowledgeCheckRepository{db: db}, nil
}

func (r *PostgresKnowledgeCheckRepository) CreateKnowledgeCheck(kc *models.KnowledgeCheck) error {
	query := `
		INSERT INTO gocourse.knowledge_checks (note_id, line_number_start, line_number_end, state, user_score, user_score_explanation, topic_summary) 
		VALUES ($1, $2, $3, $4, $5, $6, $7) 
		RETURNING id, created_at, updated_at`

	row := r.db.QueryRow(query, kc.NoteID, kc.LineNumberStart, kc.LineNumberEnd, kc.State, kc.UserScore, kc.UserScoreExplanation, kc.TopicSummary)

	err := row.Scan(&kc.ID, &kc.CreatedAt, &kc.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create knowledge check: %w", err)
	}

	return nil
}

func (r *PostgresKnowledgeCheckRepository) GetKnowledgeCheckByID(id int) (*models.KnowledgeCheck, error) {
	query := `
		SELECT id, note_id, line_number_start, line_number_end, state, user_score, user_score_explanation, topic_summary, created_at, updated_at 
		FROM gocourse.knowledge_checks 
		WHERE id = $1`

	kc := &models.KnowledgeCheck{}
	row := r.db.QueryRow(query, id)

	err := row.Scan(&kc.ID, &kc.NoteID, &kc.LineNumberStart, &kc.LineNumberEnd, &kc.State, &kc.UserScore, &kc.UserScoreExplanation, &kc.TopicSummary, &kc.CreatedAt, &kc.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("knowledge check with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get knowledge check: %w", err)
	}

	return kc, nil
}

func (r *PostgresKnowledgeCheckRepository) GetAllKnowledgeChecks() ([]*models.KnowledgeCheck, error) {
	query := `
		SELECT id, note_id, line_number_start, line_number_end, state, user_score, user_score_explanation, topic_summary, created_at, updated_at 
		FROM gocourse.knowledge_checks 
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query knowledge checks: %w", err)
	}
	defer rows.Close()

	var knowledgeChecks []*models.KnowledgeCheck
	for rows.Next() {
		kc := &models.KnowledgeCheck{}
		err := rows.Scan(&kc.ID, &kc.NoteID, &kc.LineNumberStart, &kc.LineNumberEnd, &kc.State, &kc.UserScore, &kc.UserScoreExplanation, &kc.TopicSummary, &kc.CreatedAt, &kc.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan knowledge check: %w", err)
		}
		knowledgeChecks = append(knowledgeChecks, kc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over knowledge checks: %w", err)
	}

	return knowledgeChecks, nil
}

func (r *PostgresKnowledgeCheckRepository) GetKnowledgeChecksByDateRange(startDate *time.Time, endDate *time.Time) ([]*models.KnowledgeCheck, error) {
	query := `
		SELECT id, note_id, line_number_start, line_number_end, state, user_score, user_score_explanation, topic_summary, created_at, updated_at 
		FROM gocourse.knowledge_checks`

	var conditions []string
	var args []interface{}
	argIndex := 1

	if startDate != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIndex))
		args = append(args, *startDate)
		argIndex++
	}

	if endDate != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIndex))
		args = append(args, *endDate)
		argIndex++
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query knowledge checks by date range: %w", err)
	}
	defer rows.Close()

	var knowledgeChecks []*models.KnowledgeCheck
	for rows.Next() {
		kc := &models.KnowledgeCheck{}
		err := rows.Scan(&kc.ID, &kc.NoteID, &kc.LineNumberStart, &kc.LineNumberEnd, &kc.State, &kc.UserScore, &kc.UserScoreExplanation, &kc.TopicSummary, &kc.CreatedAt, &kc.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan knowledge check: %w", err)
		}
		knowledgeChecks = append(knowledgeChecks, kc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over knowledge checks: %w", err)
	}

	return knowledgeChecks, nil
}

func (r *PostgresKnowledgeCheckRepository) UpdateKnowledgeCheck(id int, req *models.UpdateKnowledgeCheckRequest) error {
	if req.State == nil && req.UserScore == nil && req.UserScoreExplanation == nil && req.TopicSummary == nil {
		return fmt.Errorf("no updates provided")
	}

	query := "UPDATE gocourse.knowledge_checks SET "
	var setParts []string
	var args []interface{}
	argIndex := 1

	if req.State != nil {
		setParts = append(setParts, fmt.Sprintf("state = $%d", argIndex))
		args = append(args, *req.State)
		argIndex++
	}

	if req.UserScore != nil {
		setParts = append(setParts, fmt.Sprintf("user_score = $%d", argIndex))
		args = append(args, *req.UserScore)
		argIndex++
	}

	if req.UserScoreExplanation != nil {
		setParts = append(setParts, fmt.Sprintf("user_score_explanation = $%d", argIndex))
		args = append(args, *req.UserScoreExplanation)
		argIndex++
	}

	if req.TopicSummary != nil {
		setParts = append(setParts, fmt.Sprintf("topic_summary = $%d", argIndex))
		args = append(args, *req.TopicSummary)
		argIndex++
	}

	query += strings.Join(setParts, ", ")
	query += fmt.Sprintf(", updated_at = NOW() WHERE id = $%d", argIndex)
	args = append(args, id)

	result, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update knowledge check: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("knowledge check with id %d not found", id)
	}

	return nil
}

func (r *PostgresKnowledgeCheckRepository) Close() error {
	return r.db.Close()
}