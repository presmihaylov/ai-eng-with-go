package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"flashcards/models"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

type Service struct {
	client *anthropic.Client
	tools  []AgentTool
}

func NewService(anthropicAPIKey string) (*Service, error) {
	client := anthropic.NewClient(option.WithAPIKey(anthropicAPIKey))

	tools := []AgentTool{
		WeatherTool{},
		// TimeTool{}, // Example: easily add more tools here
	}

	return &Service{
		client: &client,
		tools:  tools,
	}, nil
}

func (s *Service) ProcessMessage(messages []models.AgentMessage) (*models.AgentResponse, error) {
	log.Printf("[INFO] Starting agent message processing with %d messages", len(messages))

	ctx := context.Background()

	// Convert our messages to Anthropic format
	anthropicMessages := s.convertToAnthropicMessages(messages)

	// Build tool specs for Anthropic
	toolSpecs := s.buildAnthropicToolSpecs()

	s.logAnthropicRequest("Initial request", anthropicMessages, toolSpecs)

	// Call Anthropic API
	response, err := s.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaude4Sonnet20250514,
		MaxTokens: 4096,
		Messages:  anthropicMessages,
		Tools:     toolSpecs,
	})
	if err != nil {
		log.Printf("[ERROR] Failed to call Anthropic API: %v", err)
		return nil, fmt.Errorf("failed to call Anthropic API: %v", err)
	}

	s.logAnthropicResponse("Initial response", response)

	// Start with input messages
	updatedMessages := make([]models.AgentMessage, len(messages))
	copy(updatedMessages, messages)

	// Process assistant response
	toolUses := []anthropic.ToolUseBlock{}
	assistantContent := ""

	// Extract content and tool uses from response
	for _, block := range response.Content {
		switch block := block.AsAny().(type) {
		case anthropic.TextBlock:
			assistantContent += block.Text
		case anthropic.ToolUseBlock:
			toolUses = append(toolUses, block)
		}
	}

	// Create assistant message
	assistantMsg := models.AgentMessage{
		Role:    "assistant",
		Content: assistantContent,
	}

	// Add tool calls to assistant message if present
	for _, toolUse := range toolUses {
		// Convert input to map for our model
		inputJSON, _ := json.Marshal(toolUse.Input)
		var inputMap map[string]interface{}
		json.Unmarshal(inputJSON, &inputMap)
		
		toolCall := models.ToolCall{
			ID:        toolUse.ID,
			Name:      toolUse.Name,
			Arguments: inputMap,
		}
		assistantMsg.ToolCalls = append(assistantMsg.ToolCalls, toolCall)
	}

	// Add assistant message
	updatedMessages = append(updatedMessages, assistantMsg)

	// Execute tools if present
	if len(toolUses) > 0 {
		for _, toolUse := range toolUses {
			log.Printf("[INFO] Executing tool: %s with arguments: %v", toolUse.Name, toolUse.Input)
			
			// Convert input to JSON for tool execution
			inputJSON, _ := json.Marshal(toolUse.Input)
			
			result, err := s.executeTool(ctx, toolUse.Name, string(inputJSON))
			if err != nil {
				log.Printf("[ERROR] Tool execution failed: %v", err)
				result = fmt.Sprintf("Error: %v", err)
			} else {
				log.Printf("[INFO] Tool execution result: %s", result)
			}

			// Add tool result message
			updatedMessages = append(updatedMessages, models.AgentMessage{
				Role: "tool",
				ToolResults: []models.ToolResult{
					{
						ToolCallID: toolUse.ID,
						Content:    result,
					},
				},
			})
		}
	}

	log.Printf("[INFO] Agent message processing completed successfully")

	return &models.AgentResponse{
		Messages: updatedMessages,
	}, nil
}

func (s *Service) convertToAnthropicMessages(messages []models.AgentMessage) []anthropic.MessageParam {
	var anthropicMessages []anthropic.MessageParam

	for _, msg := range messages {
		switch msg.Role {
		case "user":
			anthropicMessages = append(anthropicMessages, anthropic.NewUserMessage(anthropic.NewTextBlock(msg.Content)))
		case "assistant":
			// Build content blocks for assistant message
			contentBlocks := []anthropic.ContentBlockParamUnion{}
			
			// Add text content if present
			if msg.Content != "" {
				contentBlocks = append(contentBlocks, anthropic.ContentBlockParamUnion{
					OfText: &anthropic.TextBlockParam{Text: msg.Content},
				})
			}
			
			// Add tool use blocks if present
			for _, toolCall := range msg.ToolCalls {
				contentBlocks = append(contentBlocks, anthropic.ContentBlockParamUnion{
					OfToolUse: &anthropic.ToolUseBlockParam{
						ID:    toolCall.ID,
						Name:  toolCall.Name,
						Input: toolCall.Arguments,
					},
				})
			}
			
			anthropicMessages = append(anthropicMessages, anthropic.NewAssistantMessage(contentBlocks...))
		case "tool":
			// Convert tool results to user message with tool result blocks
			toolResultBlocks := []anthropic.ContentBlockParamUnion{}
			for _, result := range msg.ToolResults {
				toolResultBlocks = append(toolResultBlocks, anthropic.ContentBlockParamUnion{
					OfToolResult: &anthropic.ToolResultBlockParam{
						ToolUseID: result.ToolCallID,
						Content: []anthropic.ToolResultBlockParamContentUnion{
							{OfText: &anthropic.TextBlockParam{Text: result.Content}},
						},
					},
				})
			}
			anthropicMessages = append(anthropicMessages, anthropic.NewUserMessage(toolResultBlocks...))
		}
	}

	return anthropicMessages
}

func (s *Service) buildAnthropicToolSpecs() []anthropic.ToolUnionParam {
	var toolSpecs []anthropic.ToolUnionParam

	for _, tool := range s.tools {
		toolSpecs = append(toolSpecs, anthropic.ToolUnionParam{
			OfTool: &anthropic.ToolParam{
				Name:        tool.Name(),
				Description: anthropic.String(tool.Description()),
				InputSchema: tool.GetAnthropicToolSpec(),
			},
		})
	}

	return toolSpecs
}

func (s *Service) executeTool(ctx context.Context, toolName, arguments string) (string, error) {
	for _, tool := range s.tools {
		if tool.Name() == toolName {
			return tool.Call(ctx, arguments)
		}
	}
	return "", fmt.Errorf("tool %s not found", toolName)
}

func (s *Service) logAnthropicRequest(stage string, messages []anthropic.MessageParam, tools []anthropic.ToolUnionParam) {
	log.Printf("[INFO] ========== Anthropic Request (%s) ==========", stage)
	
	// Log messages
	log.Printf("[INFO] Messages (%d total):", len(messages))
	for i, msg := range messages {
		log.Printf("[INFO]   [%d] Role: %s", i, msg.Role)
	}
	
	// Log tools if present
	if tools != nil && len(tools) > 0 {
		log.Printf("[INFO] Available Tools (%d total):", len(tools))
		for i, tool := range tools {
			if tool.OfTool != nil {
				log.Printf("[INFO]   [%d] Name: %s", i, tool.OfTool.Name)
			}
		}
	} else {
		log.Printf("[INFO] No tools provided")
	}
	
	log.Printf("[INFO] ================================================")
}

func (s *Service) logAnthropicResponse(stage string, response *anthropic.Message) {
	log.Printf("[INFO] ========== Anthropic Response (%s) ==========", stage)
	
	log.Printf("[INFO] Model: %s", response.Model)
	log.Printf("[INFO] StopReason: %s", response.StopReason)
	log.Printf("[INFO] Content blocks (%d total):", len(response.Content))
	
	toolCallCount := 0
	for i, block := range response.Content {
		switch block := block.AsAny().(type) {
		case anthropic.TextBlock:
			log.Printf("[INFO]   [%d] Text: %s", i, block.Text)
		case anthropic.ToolUseBlock:
			toolCallCount++
			log.Printf("[INFO]   [%d] Tool Use: ID=%s, Name=%s, Input=%v", i, block.ID, block.Name, block.Input)
		}
	}
	
	if toolCallCount > 0 {
		log.Printf("[INFO] Total tool calls: %d", toolCallCount)
	} else {
		log.Printf("[INFO] No tool calls made")
	}
	
	log.Printf("[INFO] =================================================")
}