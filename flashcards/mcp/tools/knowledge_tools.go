package tools

import (
	"context"
	"fmt"
	"time"

	"flashcards/models"
	"flashcards/services"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Input/Output types for list_knowledge_checks
type ListKnowledgeChecksInput struct {
	StartDate *string `json:"start_date,omitempty" jsonschema:"optional start date for filtering (RFC3339 format)"`
	EndDate   *string `json:"end_date,omitempty" jsonschema:"optional end date for filtering (RFC3339 format)"`
}

type ListKnowledgeChecksOutput struct {
	KnowledgeChecks []models.KnowledgeCheck `json:"knowledge_checks" jsonschema:"list of knowledge checks"`
	Count           int                     `json:"count" jsonschema:"total number of knowledge checks"`
}

// Input/Output types for get_knowledge_check
type GetKnowledgeCheckInput struct {
	ID int `json:"id" jsonschema:"the ID of the knowledge check to retrieve"`
}

type GetKnowledgeCheckOutput struct {
	KnowledgeCheck models.KnowledgeCheck `json:"knowledge_check" jsonschema:"the requested knowledge check"`
}

// Input/Output types for create_knowledge_check
type CreateKnowledgeCheckInput struct {
	NoteID          int    `json:"note_id" jsonschema:"the ID of the note this check belongs to"`
	LineNumberStart int    `json:"line_number_start" jsonschema:"starting line number of the content section"`
	LineNumberEnd   int    `json:"line_number_end" jsonschema:"ending line number of the content section"`
	TopicSummary    string `json:"topic_summary" jsonschema:"summary of the topic being tested"`
}

type CreateKnowledgeCheckOutput struct {
	KnowledgeCheck models.KnowledgeCheck `json:"knowledge_check" jsonschema:"the created knowledge check"`
}

// Input/Output types for update_knowledge_check
type UpdateKnowledgeCheckInput struct {
	ID                   int     `json:"id" jsonschema:"the ID of the knowledge check to update"`
	State                *string `json:"state,omitempty" jsonschema:"new state (pending or completed)"`
	UserScore            *int    `json:"user_score,omitempty" jsonschema:"user's score (1-10)"`
	UserScoreExplanation *string `json:"user_score_explanation,omitempty" jsonschema:"explanation for the score"`
	TopicSummary         *string `json:"topic_summary,omitempty" jsonschema:"updated topic summary"`
}

type UpdateKnowledgeCheckOutput struct {
	KnowledgeCheck models.KnowledgeCheck `json:"knowledge_check" jsonschema:"the updated knowledge check"`
}

// ListKnowledgeChecks retrieves all knowledge checks, optionally filtered by date range
func ListKnowledgeChecks(ctx context.Context, req *mcp.CallToolRequest, input ListKnowledgeChecksInput, service *services.KnowledgeCheckService) (*mcp.CallToolResult, ListKnowledgeChecksOutput, error) {
	var checks []*models.KnowledgeCheck
	var err error

	// If date range is provided, use date-filtered query
	if input.StartDate != nil || input.EndDate != nil {
		var startDate, endDate *time.Time

		if input.StartDate != nil {
			t, parseErr := time.Parse(time.RFC3339, *input.StartDate)
			if parseErr != nil {
				return nil, ListKnowledgeChecksOutput{}, fmt.Errorf("invalid start_date format: %w", parseErr)
			}
			startDate = &t
		}

		if input.EndDate != nil {
			t, parseErr := time.Parse(time.RFC3339, *input.EndDate)
			if parseErr != nil {
				return nil, ListKnowledgeChecksOutput{}, fmt.Errorf("invalid end_date format: %w", parseErr)
			}
			endDate = &t
		}

		checks, err = service.GetKnowledgeChecksByDateRange(startDate, endDate)
		if err != nil {
			return nil, ListKnowledgeChecksOutput{}, fmt.Errorf("failed to get knowledge checks by date range: %w", err)
		}
	} else {
		checks, err = service.GetAllKnowledgeChecks()
		if err != nil {
			return nil, ListKnowledgeChecksOutput{}, fmt.Errorf("failed to get knowledge checks: %w", err)
		}
	}

	// Convert []*models.KnowledgeCheck to []models.KnowledgeCheck
	checksList := make([]models.KnowledgeCheck, len(checks))
	for i, check := range checks {
		checksList[i] = *check
	}

	output := ListKnowledgeChecksOutput{
		KnowledgeChecks: checksList,
		Count:           len(checksList),
	}

	return nil, output, nil
}

// GetKnowledgeCheck retrieves a specific knowledge check by ID
func GetKnowledgeCheck(ctx context.Context, req *mcp.CallToolRequest, input GetKnowledgeCheckInput, service *services.KnowledgeCheckService) (*mcp.CallToolResult, GetKnowledgeCheckOutput, error) {
	check, err := service.GetKnowledgeCheckByID(input.ID)
	if err != nil {
		return nil, GetKnowledgeCheckOutput{}, fmt.Errorf("failed to get knowledge check: %w", err)
	}

	output := GetKnowledgeCheckOutput{
		KnowledgeCheck: *check,
	}

	return nil, output, nil
}

// CreateKnowledgeCheck creates a new knowledge check
func CreateKnowledgeCheck(ctx context.Context, req *mcp.CallToolRequest, input CreateKnowledgeCheckInput, service *services.KnowledgeCheckService) (*mcp.CallToolResult, CreateKnowledgeCheckOutput, error) {
	createReq := &models.CreateKnowledgeCheckRequest{
		NoteID:          input.NoteID,
		LineNumberStart: input.LineNumberStart,
		LineNumberEnd:   input.LineNumberEnd,
		TopicSummary:    input.TopicSummary,
	}

	check, err := service.CreateKnowledgeCheck(createReq)
	if err != nil {
		return nil, CreateKnowledgeCheckOutput{}, fmt.Errorf("failed to create knowledge check: %w", err)
	}

	output := CreateKnowledgeCheckOutput{
		KnowledgeCheck: *check,
	}

	return nil, output, nil
}

// UpdateKnowledgeCheck updates an existing knowledge check
func UpdateKnowledgeCheck(ctx context.Context, req *mcp.CallToolRequest, input UpdateKnowledgeCheckInput, service *services.KnowledgeCheckService) (*mcp.CallToolResult, UpdateKnowledgeCheckOutput, error) {
	updateReq := &models.UpdateKnowledgeCheckRequest{
		State:                input.State,
		UserScore:            input.UserScore,
		UserScoreExplanation: input.UserScoreExplanation,
		TopicSummary:         input.TopicSummary,
	}

	check, err := service.UpdateKnowledgeCheck(input.ID, updateReq)
	if err != nil {
		return nil, UpdateKnowledgeCheckOutput{}, fmt.Errorf("failed to update knowledge check: %w", err)
	}

	output := UpdateKnowledgeCheckOutput{
		KnowledgeCheck: *check,
	}

	return nil, output, nil
}
