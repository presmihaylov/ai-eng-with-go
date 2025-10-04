package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"flashcards/services"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// NoteMetadata represents note without full content (for listing)
type NoteMetadata struct {
	ID        int       `json:"id" jsonschema:"note ID"`
	Preview   string    `json:"preview" jsonschema:"first 200 characters of note content"`
	CreatedAt time.Time `json:"createdAt" jsonschema:"creation timestamp"`
	UpdatedAt time.Time `json:"updatedAt" jsonschema:"last update timestamp"`
}

// Input/Output types for list_notes
type ListNotesInput struct{}

type ListNotesOutput struct {
	Notes []NoteMetadata `json:"notes" jsonschema:"list of note metadata (without content)"`
	Count int            `json:"count" jsonschema:"total number of notes"`
}

// Input/Output types for get_note
type GetNoteInput struct {
	ID     int  `json:"id" jsonschema:"the ID of the note to retrieve"`
	Offset *int `json:"offset,omitempty" jsonschema:"optional line number to start reading from (1-based)"`
	Limit  *int `json:"limit,omitempty" jsonschema:"optional number of lines to return from offset"`
}

type GetNoteOutput struct {
	ID           int       `json:"id" jsonschema:"note ID"`
	Content      string    `json:"content" jsonschema:"note content (full or partial based on offset/limit)"`
	TotalLines   int       `json:"total_lines" jsonschema:"total number of lines in the note"`
	LinesReturned int      `json:"lines_returned" jsonschema:"number of lines returned in this response"`
	OffsetUsed   int       `json:"offset_used" jsonschema:"starting line number used (1-based)"`
	CreatedAt    time.Time `json:"createdAt" jsonschema:"creation timestamp"`
	UpdatedAt    time.Time `json:"updatedAt" jsonschema:"last update timestamp"`
}

// Input/Output types for search_notes
type SearchNotesInput struct {
	SearchTerms []string `json:"search_terms" jsonschema:"list of keywords to search for in note content"`
}

type SearchNotesOutput struct {
	Notes []NoteMetadata `json:"notes" jsonschema:"list of matching note metadata (without content)"`
	Count int            `json:"count" jsonschema:"number of matching notes"`
}

// ListNotes retrieves all notes metadata from the database with preview
func ListNotes(ctx context.Context, req *mcp.CallToolRequest, input ListNotesInput, noteService *services.NoteService) (*mcp.CallToolResult, ListNotesOutput, error) {
	notes, err := noteService.GetAllNotes()
	if err != nil {
		return nil, ListNotesOutput{}, fmt.Errorf("failed to get notes: %w", err)
	}

	// Convert to metadata with preview
	notesList := make([]NoteMetadata, len(notes))
	for i, note := range notes {
		preview := note.Content
		if len(preview) > 200 {
			preview = preview[:200]
		}

		notesList[i] = NoteMetadata{
			ID:        note.ID,
			Preview:   preview,
			CreatedAt: note.CreatedAt,
			UpdatedAt: note.UpdatedAt,
		}
	}

	output := ListNotesOutput{
		Notes: notesList,
		Count: len(notesList),
	}

	return nil, output, nil
}

// GetNote retrieves a specific note by ID with optional line-based pagination
func GetNote(ctx context.Context, req *mcp.CallToolRequest, input GetNoteInput, noteService *services.NoteService) (*mcp.CallToolResult, GetNoteOutput, error) {
	note, err := noteService.GetNoteByID(input.ID)
	if err != nil {
		return nil, GetNoteOutput{}, fmt.Errorf("failed to get note: %w", err)
	}

	// Split content into lines
	lines := strings.Split(note.Content, "\n")
	totalLines := len(lines)

	// Default offset and limit
	offset := 1
	if input.Offset != nil && *input.Offset > 0 {
		offset = *input.Offset
	}

	// Calculate which lines to return
	startIdx := offset - 1 // Convert to 0-based index
	if startIdx >= totalLines {
		return nil, GetNoteOutput{}, fmt.Errorf("offset %d exceeds total lines %d", offset, totalLines)
	}

	endIdx := totalLines
	if input.Limit != nil && *input.Limit > 0 {
		endIdx = startIdx + *input.Limit
		if endIdx > totalLines {
			endIdx = totalLines
		}
	}

	// Extract the requested lines
	contentLines := lines[startIdx:endIdx]
	content := strings.Join(contentLines, "\n")
	linesReturned := len(contentLines)

	output := GetNoteOutput{
		ID:           note.ID,
		Content:      content,
		TotalLines:   totalLines,
		LinesReturned: linesReturned,
		OffsetUsed:   offset,
		CreatedAt:    note.CreatedAt,
		UpdatedAt:    note.UpdatedAt,
	}

	return nil, output, nil
}

// SearchNotes searches for notes by content keywords (returns metadata with preview)
func SearchNotes(ctx context.Context, req *mcp.CallToolRequest, input SearchNotesInput, noteService *services.NoteService) (*mcp.CallToolResult, SearchNotesOutput, error) {
	notes, err := noteService.SearchNotesByContent(input.SearchTerms)
	if err != nil {
		return nil, SearchNotesOutput{}, fmt.Errorf("failed to search notes: %w", err)
	}

	// Convert to metadata with preview
	notesList := make([]NoteMetadata, len(notes))
	for i, note := range notes {
		preview := note.Content
		if len(preview) > 200 {
			preview = preview[:200]
		}

		notesList[i] = NoteMetadata{
			ID:        note.ID,
			Preview:   preview,
			CreatedAt: note.CreatedAt,
			UpdatedAt: note.UpdatedAt,
		}
	}

	output := SearchNotesOutput{
		Notes: notesList,
		Count: len(notesList),
	}

	return nil, output, nil
}
