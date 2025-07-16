package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"flashcards/models"
	"flashcards/services"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/invopop/jsonschema"
)

type ListNotesToolInput struct {}

// AgentTool interface that all tools must implement
type AgentTool interface {
	Name() string
	Description() string
	Call(ctx context.Context, input string) (string, error)
	GetAnthropicToolSpec() anthropic.ToolInputSchemaParam
}

type ListNotesTool struct {
	noteService *services.NoteService
}

func NewListNotesTool(noteService *services.NoteService) ListNotesTool {
	return ListNotesTool{noteService: noteService}
}

func (l ListNotesTool) Name() string {
	return "list_notes"
}

func (l ListNotesTool) Description() string {
	return "Lists all notes with preview information including ID, preview of content, and creation date"
}

func (l ListNotesTool) Call(ctx context.Context, input string) (string, error) {
	var params ListNotesToolInput
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("failed to parse list notes tool input: %v", err)
	}

	notes, err := l.noteService.GetAllNotes()
	if err != nil {
		return "", fmt.Errorf("failed to get notes: %v", err)
	}

	type NotePreview struct {
		ID         int    `json:"id"`
		Preview    string `json:"preview"`
		CreatedAt  string `json:"created_at"`
		TotalLines int    `json:"total_lines"`
	}

	var previews []NotePreview
	for _, note := range notes {
		preview := note.Content
		if len(preview) > 100 {
			preview = preview[:100] + "..."
		}
		totalLines := len(strings.Split(note.Content, "\n"))
		previews = append(previews, NotePreview{
			ID:         note.ID,
			Preview:    preview,
			CreatedAt:  note.CreatedAt.Format(time.RFC3339),
			TotalLines: totalLines,
		})
	}

	result, err := json.Marshal(previews)
	if err != nil {
		return "", fmt.Errorf("failed to marshal note previews: %v", err)
	}

	return string(result), nil
}

func generateAnthropicSchema[T any]() anthropic.ToolInputSchemaParam {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	schema := reflector.Reflect(v)

	return anthropic.ToolInputSchemaParam{
		Properties: schema.Properties,
	}
}

func (l ListNotesTool) GetAnthropicToolSpec() anthropic.ToolInputSchemaParam {
	return generateAnthropicSchema[ListNotesToolInput]()
}

type ReadNoteToolInput struct {
	NoteID          int `json:"note_id" jsonschema:"required,description=The ID of the note to read"`
	LineNumberStart int `json:"line_number_start,omitempty" jsonschema:"description=Starting line number (default: 1)"`
	LineNumberEnd   int `json:"line_number_end,omitempty" jsonschema:"description=Ending line number (default: end of file)"`
}

type ReadNoteTool struct {
	noteService *services.NoteService
}

func NewReadNoteTool(noteService *services.NoteService) ReadNoteTool {
	return ReadNoteTool{noteService: noteService}
}

func (r ReadNoteTool) Name() string {
	return "read_note"
}

func (r ReadNoteTool) Description() string {
	return "Reads the content of a specific note with optional line number range"
}

func (r ReadNoteTool) Call(ctx context.Context, input string) (string, error) {
	var params ReadNoteToolInput
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("failed to parse read note tool input: %v", err)
	}

	note, err := r.noteService.GetNoteByID(params.NoteID)
	if err != nil {
		return "", fmt.Errorf("failed to get note: %v", err)
	}

	lines := strings.Split(note.Content, "\n")
	totalLines := len(lines)

	start := params.LineNumberStart
	if start <= 0 {
		start = 1
	}

	end := params.LineNumberEnd
	if end <= 0 || end > totalLines {
		end = totalLines
	}

	if start > totalLines {
		return "", fmt.Errorf("start line %d exceeds total lines %d", start, totalLines)
	}

	if start > end {
		return "", fmt.Errorf("start line %d cannot be greater than end line %d", start, end)
	}

	selectedLines := lines[start-1 : end]
	result := strings.Join(selectedLines, "\n")

	type ReadNoteResult struct {
		NoteID      int    `json:"note_id"`
		Content     string `json:"content"`
		LineStart   int    `json:"line_start"`
		LineEnd     int    `json:"line_end"`
		TotalLines  int    `json:"total_lines"`
	}

	readResult := ReadNoteResult{
		NoteID:     params.NoteID,
		Content:    result,
		LineStart:  start,
		LineEnd:    end,
		TotalLines: totalLines,
	}

	resultJSON, err := json.Marshal(readResult)
	if err != nil {
		return "", fmt.Errorf("failed to marshal read note result: %v", err)
	}

	return string(resultJSON), nil
}

func (r ReadNoteTool) GetAnthropicToolSpec() anthropic.ToolInputSchemaParam {
	return generateAnthropicSchema[ReadNoteToolInput]()
}

type GetMemoryToolInput struct{}

type GetMemoryTool struct {
	memoryService *services.MemoryService
}

func NewGetMemoryTool(memoryService *services.MemoryService) GetMemoryTool {
	return GetMemoryTool{memoryService: memoryService}
}

func (g GetMemoryTool) Name() string {
	return "get_memory"
}

func (g GetMemoryTool) Description() string {
	return "Retrieves the agent's current memory content"
}

func (g GetMemoryTool) Call(ctx context.Context, input string) (string, error) {
	var params GetMemoryToolInput
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("failed to parse get memory tool input: %v", err)
	}

	memory, err := g.memoryService.GetMemory()
	if err != nil {
		return "", fmt.Errorf("failed to get memory: %v", err)
	}

	if memory.MemoryContent == "" {
		return "(empty)", nil
	}

	return memory.MemoryContent, nil
}

func (g GetMemoryTool) GetAnthropicToolSpec() anthropic.ToolInputSchemaParam {
	return generateAnthropicSchema[GetMemoryToolInput]()
}

type UpdateMemoryToolInput struct {
	Content string `json:"content" jsonschema:"required,description=The new memory content to store"`
}

type UpdateMemoryTool struct {
	memoryService *services.MemoryService
}

func NewUpdateMemoryTool(memoryService *services.MemoryService) UpdateMemoryTool {
	return UpdateMemoryTool{memoryService: memoryService}
}

func (u UpdateMemoryTool) Name() string {
	return "update_memory"
}

func (u UpdateMemoryTool) Description() string {
	return "Updates the agent's memory with new content"
}

func (u UpdateMemoryTool) Call(ctx context.Context, input string) (string, error) {
	var params UpdateMemoryToolInput
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("failed to parse update memory tool input: %v", err)
	}

	if err := u.memoryService.UpdateMemory(params.Content); err != nil {
		return "", fmt.Errorf("failed to update memory: %v", err)
	}

	return "Memory updated successfully", nil
}

func (u UpdateMemoryTool) GetAnthropicToolSpec() anthropic.ToolInputSchemaParam {
	return generateAnthropicSchema[UpdateMemoryToolInput]()
}

type GetCurrentTimeToolInput struct{}

type GetCurrentTimeTool struct{}

func NewGetCurrentTimeTool() GetCurrentTimeTool {
	return GetCurrentTimeTool{}
}

func (t GetCurrentTimeTool) Name() string {
	return "get_current_time"
}

func (t GetCurrentTimeTool) Description() string {
	return "Gets the current timestamp in ISO format"
}

func (t GetCurrentTimeTool) Call(ctx context.Context, input string) (string, error) {
	var params GetCurrentTimeToolInput
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("failed to parse get current time tool input: %v", err)
	}

	return time.Now().Format(time.RFC3339), nil
}

func (t GetCurrentTimeTool) GetAnthropicToolSpec() anthropic.ToolInputSchemaParam {
	return generateAnthropicSchema[GetCurrentTimeToolInput]()
}

type CreateEmptyKnowledgeCheckToolInput struct {
	NoteID          int    `json:"note_id" jsonschema:"required,description=The ID of the note to create knowledge check for"`
	LineNumberStart int    `json:"line_number_start" jsonschema:"required,description=Starting line number of the content section"`
	LineNumberEnd   int    `json:"line_number_end" jsonschema:"required,description=Ending line number of the content section"`
	TopicSummary    string `json:"topic_summary" jsonschema:"required,description=AI-generated summary of what this section covers"`
}

type CreateEmptyKnowledgeCheckTool struct {
	knowledgeCheckService *services.KnowledgeCheckService
}

func NewCreateEmptyKnowledgeCheckTool(knowledgeCheckService *services.KnowledgeCheckService) CreateEmptyKnowledgeCheckTool {
	return CreateEmptyKnowledgeCheckTool{knowledgeCheckService: knowledgeCheckService}
}

func (c CreateEmptyKnowledgeCheckTool) Name() string {
	return "create_empty_knowledge_check"
}

func (c CreateEmptyKnowledgeCheckTool) Description() string {
	return "Creates a new knowledge check in pending state for a specific section of a note"
}

func (c CreateEmptyKnowledgeCheckTool) Call(ctx context.Context, input string) (string, error) {
	var params CreateEmptyKnowledgeCheckToolInput
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("failed to parse create empty knowledge check tool input: %v", err)
	}

	req := &models.CreateKnowledgeCheckRequest{
		NoteID:          params.NoteID,
		LineNumberStart: params.LineNumberStart,
		LineNumberEnd:   params.LineNumberEnd,
		TopicSummary:    params.TopicSummary,
	}

	kc, err := c.knowledgeCheckService.CreateKnowledgeCheck(req)
	if err != nil {
		return "", fmt.Errorf("failed to create knowledge check: %v", err)
	}

	result, err := json.Marshal(kc)
	if err != nil {
		return "", fmt.Errorf("failed to marshal knowledge check: %v", err)
	}

	return string(result), nil
}

func (c CreateEmptyKnowledgeCheckTool) GetAnthropicToolSpec() anthropic.ToolInputSchemaParam {
	return generateAnthropicSchema[CreateEmptyKnowledgeCheckToolInput]()
}

type MarkKnowledgeCheckCompleteToolInput struct {
	KnowledgeCheckID     int    `json:"knowledge_check_id" jsonschema:"required,description=The ID of the knowledge check to mark as complete"`
	UserScore           int    `json:"user_score" jsonschema:"required,minimum=1,maximum=10,description=User's score from 1-10 on this knowledge check"`
	UserScoreExplanation string `json:"user_score_explanation" jsonschema:"required,description=Explanation of why the user received this score"`
}

type MarkKnowledgeCheckCompleteTool struct {
	knowledgeCheckService *services.KnowledgeCheckService
}

func NewMarkKnowledgeCheckCompleteTool(knowledgeCheckService *services.KnowledgeCheckService) MarkKnowledgeCheckCompleteTool {
	return MarkKnowledgeCheckCompleteTool{knowledgeCheckService: knowledgeCheckService}
}

func (m MarkKnowledgeCheckCompleteTool) Name() string {
	return "mark_knowledge_check_complete"
}

func (m MarkKnowledgeCheckCompleteTool) Description() string {
	return "Marks a knowledge check as completed with user score and explanation"
}

func (m MarkKnowledgeCheckCompleteTool) Call(ctx context.Context, input string) (string, error) {
	var params MarkKnowledgeCheckCompleteToolInput
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("failed to parse mark knowledge check complete tool input: %v", err)
	}

	state := "completed"
	req := &models.UpdateKnowledgeCheckRequest{
		State:                &state,
		UserScore:            &params.UserScore,
		UserScoreExplanation: &params.UserScoreExplanation,
	}

	kc, err := m.knowledgeCheckService.UpdateKnowledgeCheck(params.KnowledgeCheckID, req)
	if err != nil {
		return "", fmt.Errorf("failed to mark knowledge check as complete: %v", err)
	}

	result, err := json.Marshal(kc)
	if err != nil {
		return "", fmt.Errorf("failed to marshal knowledge check: %v", err)
	}

	return string(result), nil
}

func (m MarkKnowledgeCheckCompleteTool) GetAnthropicToolSpec() anthropic.ToolInputSchemaParam {
	return generateAnthropicSchema[MarkKnowledgeCheckCompleteToolInput]()
}

type GetKnowledgeCheckToolInput struct {
	KnowledgeCheckID int `json:"knowledge_check_id" jsonschema:"required,description=The ID of the knowledge check to retrieve"`
}

type GetKnowledgeCheckTool struct {
	knowledgeCheckService *services.KnowledgeCheckService
}

func NewGetKnowledgeCheckTool(knowledgeCheckService *services.KnowledgeCheckService) GetKnowledgeCheckTool {
	return GetKnowledgeCheckTool{knowledgeCheckService: knowledgeCheckService}
}

func (g GetKnowledgeCheckTool) Name() string {
	return "get_knowledge_check"
}

func (g GetKnowledgeCheckTool) Description() string {
	return "Retrieves a specific knowledge check by ID"
}

func (g GetKnowledgeCheckTool) Call(ctx context.Context, input string) (string, error) {
	var params GetKnowledgeCheckToolInput
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("failed to parse get knowledge check tool input: %v", err)
	}

	kc, err := g.knowledgeCheckService.GetKnowledgeCheckByID(params.KnowledgeCheckID)
	if err != nil {
		return "", fmt.Errorf("failed to get knowledge check: %v", err)
	}

	result, err := json.Marshal(kc)
	if err != nil {
		return "", fmt.Errorf("failed to marshal knowledge check: %v", err)
	}

	return string(result), nil
}

func (g GetKnowledgeCheckTool) GetAnthropicToolSpec() anthropic.ToolInputSchemaParam {
	return generateAnthropicSchema[GetKnowledgeCheckToolInput]()
}

type ListKnowledgeChecksToolInput struct {
	StartDate string `json:"start_date,omitempty" jsonschema:"description=Start date for filtering in ISO format (YYYY-MM-DD or YYYY-MM-DDTHH:MM:SS), optional"`
	EndDate   string `json:"end_date,omitempty" jsonschema:"description=End date for filtering in ISO format (YYYY-MM-DD or YYYY-MM-DDTHH:MM:SS), optional"`
}

type ListKnowledgeChecksTool struct {
	knowledgeCheckService *services.KnowledgeCheckService
}

func NewListKnowledgeChecksTool(knowledgeCheckService *services.KnowledgeCheckService) ListKnowledgeChecksTool {
	return ListKnowledgeChecksTool{knowledgeCheckService: knowledgeCheckService}
}

func (l ListKnowledgeChecksTool) Name() string {
	return "list_knowledge_checks"
}

func (l ListKnowledgeChecksTool) Description() string {
	return "Lists knowledge checks with optional date range filtering"
}

func (l ListKnowledgeChecksTool) Call(ctx context.Context, input string) (string, error) {
	var params ListKnowledgeChecksToolInput
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("failed to parse list knowledge checks tool input: %v", err)
	}

	var startDate, endDate *time.Time

	if params.StartDate != "" {
		parsedStart, err := l.parseDateTime(params.StartDate)
		if err != nil {
			return "", fmt.Errorf("invalid start date format: %v", err)
		}
		startDate = &parsedStart
	}

	if params.EndDate != "" {
		parsedEnd, err := l.parseDateTime(params.EndDate)
		if err != nil {
			return "", fmt.Errorf("invalid end date format: %v", err)
		}
		endDate = &parsedEnd
	}

	var knowledgeChecks []*models.KnowledgeCheck
	var err error

	if startDate != nil || endDate != nil {
		knowledgeChecks, err = l.knowledgeCheckService.GetKnowledgeChecksByDateRange(startDate, endDate)
	} else {
		knowledgeChecks, err = l.knowledgeCheckService.GetAllKnowledgeChecks()
	}

	if err != nil {
		return "", fmt.Errorf("failed to get knowledge checks: %v", err)
	}

	result, err := json.Marshal(knowledgeChecks)
	if err != nil {
		return "", fmt.Errorf("failed to marshal knowledge checks: %v", err)
	}

	return string(result), nil
}

func (l ListKnowledgeChecksTool) parseDateTime(dateStr string) (time.Time, error) {
	// Try parsing as full datetime first
	if t, err := time.Parse(time.RFC3339, dateStr); err == nil {
		return t, nil
	}

	// Try parsing as date only
	if t, err := time.Parse("2006-01-02", dateStr); err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("unsupported date format, use YYYY-MM-DD or YYYY-MM-DDTHH:MM:SS")
}

func (l ListKnowledgeChecksTool) GetAnthropicToolSpec() anthropic.ToolInputSchemaParam {
	return generateAnthropicSchema[ListKnowledgeChecksToolInput]()
}