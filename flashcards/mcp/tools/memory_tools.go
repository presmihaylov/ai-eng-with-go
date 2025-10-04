package tools

import (
	"context"
	"fmt"

	"flashcards/models"
	"flashcards/services"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Input/Output types for get_memory
type GetMemoryInput struct{}

type GetMemoryOutput struct {
	Memory models.Memory `json:"memory" jsonschema:"the agent's memory"`
}

// Input/Output types for update_memory
type UpdateMemoryInput struct {
	Content string `json:"content" jsonschema:"the new memory content to store"`
}

type UpdateMemoryOutput struct {
	Success bool   `json:"success" jsonschema:"whether the update was successful"`
	Message string `json:"message" jsonschema:"status message"`
}

// GetMemory retrieves the agent's memory
func GetMemory(ctx context.Context, req *mcp.CallToolRequest, input GetMemoryInput, memoryService *services.MemoryService) (*mcp.CallToolResult, GetMemoryOutput, error) {
	memory, err := memoryService.GetMemory()
	if err != nil {
		return nil, GetMemoryOutput{}, fmt.Errorf("failed to get memory: %w", err)
	}

	output := GetMemoryOutput{
		Memory: *memory,
	}

	return nil, output, nil
}

// UpdateMemory updates the agent's memory content
func UpdateMemory(ctx context.Context, req *mcp.CallToolRequest, input UpdateMemoryInput, memoryService *services.MemoryService) (*mcp.CallToolResult, UpdateMemoryOutput, error) {
	err := memoryService.UpdateMemory(input.Content)
	if err != nil {
		return nil, UpdateMemoryOutput{
			Success: false,
			Message: fmt.Sprintf("Failed to update memory: %v", err),
		}, fmt.Errorf("failed to update memory: %w", err)
	}

	output := UpdateMemoryOutput{
		Success: true,
		Message: "Memory updated successfully",
	}

	return nil, output, nil
}
