package models

import "time"

type KnowledgeCheck struct {
	ID                   int       `json:"id" db:"id"`
	NoteID               int       `json:"note_id" db:"note_id"`
	LineNumberStart      int       `json:"line_number_start" db:"line_number_start"`
	LineNumberEnd        int       `json:"line_number_end" db:"line_number_end"`
	State                string    `json:"state" db:"state"`
	UserScore            *int      `json:"user_score,omitempty" db:"user_score"`
	UserScoreExplanation *string   `json:"user_score_explanation,omitempty" db:"user_score_explanation"`
	TopicSummary         string    `json:"topic_summary" db:"topic_summary"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time `json:"updated_at" db:"updated_at"`
}

type CreateKnowledgeCheckRequest struct {
	NoteID          int    `json:"note_id"`
	LineNumberStart int    `json:"line_number_start"`
	LineNumberEnd   int    `json:"line_number_end"`
	TopicSummary    string `json:"topic_summary"`
}

type UpdateKnowledgeCheckRequest struct {
	State                *string `json:"state,omitempty"`
	UserScore            *int    `json:"user_score,omitempty"`
	UserScoreExplanation *string `json:"user_score_explanation,omitempty"`
	TopicSummary         *string `json:"topic_summary,omitempty"`
}