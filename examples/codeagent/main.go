package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

const (
	ColorReset  = "\033[0m"
	ColorBlue   = "\033[34m"
	ColorYellow = "\033[33m"
	ColorGreen  = "\033[32m"
)

func colorize(color, text string) string {
	return color + text + ColorReset
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

type Agent struct {
	client         *anthropic.Client
	getUserMessage func() (string, bool)
	tools          []ToolDefinition
}

func NewAgent(apiKey string) *Agent {
	client := anthropic.NewClient(option.WithAPIKey(apiKey))

	agent := &Agent{
		client: &client,
		getUserMessage: func() (string, bool) {
			fmt.Print(colorize(ColorBlue, "You") + ": ")
			scanner := bufio.NewScanner(os.Stdin)
			if scanner.Scan() {
				text := strings.TrimSpace(scanner.Text())
				if text == "exit" || text == "quit" {
					return "", false
				}
				return text, true
			}
			return "", false
		},
	}

	agent.setupTools()
	return agent
}

func (a *Agent) Run() {
	fmt.Println("Code Editing Agent started. Type 'exit' or 'quit' to stop.")

	messages := []anthropic.MessageParam{}

	for {
		userInput, ok := a.getUserMessage()
		if !ok {
			break
		}

		messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(userInput)))

		for {
			tools := make([]anthropic.ToolUnionParam, len(a.tools))
			for i, tool := range a.tools {
				tools[i] = anthropic.ToolUnionParam{
					OfTool: &anthropic.ToolParam{
						Name:        tool.Name,
						Description: anthropic.String(tool.Description),
						InputSchema: tool.InputSchema,
					},
				}
			}

			response, err := a.client.Messages.New(context.Background(), anthropic.MessageNewParams{
				Model:     anthropic.ModelClaude4Sonnet20250514,
				MaxTokens: 4096,
				Messages:  messages,
				Tools:     tools,
			})

			if err != nil {
				log.Printf("Error calling Claude: %v", err)
				break
			}

			messages = append(messages, response.ToParam())

			toolUses := []anthropic.ToolUseBlock{}
			assistantResponse := strings.Builder{}

			for _, block := range response.Content {
				switch block := block.AsAny().(type) {
				case anthropic.TextBlock:
					assistantResponse.WriteString(block.Text)
				case anthropic.ToolUseBlock:
					toolUses = append(toolUses, block)
				}
			}

			if assistantResponse.Len() > 0 {
				fmt.Printf("%s: %s\n", colorize(ColorYellow, "Assistant"), assistantResponse.String())
			}

			if len(toolUses) == 0 {
				break
			}

			toolResults := []anthropic.ContentBlockParamUnion{}
			for _, toolUse := range toolUses {
				var tool *ToolDefinition
				for _, t := range a.tools {
					if t.Name == toolUse.Name {
						tool = &t
						break
					}
				}

				if tool == nil {
					log.Printf("Unknown tool: %s", toolUse.Name)
					continue
				}

				inputJSON, _ := json.Marshal(toolUse.Input)
				
				// Show tool call with green color
				inputStr := truncateString(string(inputJSON), 200)
				fmt.Printf("%s: %s(%s)\n", colorize(ColorGreen, "Tool Call"), toolUse.Name, inputStr)
				
				result, err := tool.Function(inputJSON)
				if err != nil {
					log.Printf("Error executing tool %s: %v", toolUse.Name, err)
					result = fmt.Sprintf("Error: %v", err)
				}

				// Show tool result with green color
				resultStr := truncateString(result, 300)
				fmt.Printf("%s: %s\n", colorize(ColorGreen, "Tool Result"), resultStr)

				toolResults = append(toolResults, anthropic.ContentBlockParamUnion{
					OfToolResult: &anthropic.ToolResultBlockParam{
						ToolUseID: toolUse.ID,
						Content: []anthropic.ToolResultBlockParamContentUnion{
							{
								OfText: &anthropic.TextBlockParam{Text: result},
							},
						},
					},
				})
			}

			if len(toolResults) > 0 {
				messages = append(messages, anthropic.NewUserMessage(toolResults...))
			}
		}
	}

	fmt.Println("Goodbye!")
}

func main() {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Fatal("ANTHROPIC_API_KEY environment variable is required")
	}

	agent := NewAgent(apiKey)
	agent.Run()
}