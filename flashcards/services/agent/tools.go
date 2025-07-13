package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/invopop/jsonschema"
)

type WeatherToolInput struct {
	City string `json:"city" jsonschema:"required,description=The city to get weather for"`
}

// AgentTool interface that all tools must implement
type AgentTool interface {
	Name() string
	Description() string
	Call(ctx context.Context, input string) (string, error)
	GetAnthropicToolSpec() anthropic.ToolInputSchemaParam
}

type WeatherTool struct{}

func (w WeatherTool) Name() string {
	return "get_weather"
}

func (w WeatherTool) Description() string {
	return "Get current weather for a given city"
}

func (w WeatherTool) Call(ctx context.Context, input string) (string, error) {
	var params WeatherToolInput
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("failed to parse weather tool input: %v", err)
	}

	return fmt.Sprintf("The weather in %s is 20 degrees", params.City), nil
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

func (w WeatherTool) GetAnthropicToolSpec() anthropic.ToolInputSchemaParam {
	return generateAnthropicSchema[WeatherToolInput]()
}

// Example of how easy it is to add more tools with this pattern
type TimeToolInput struct {
	Timezone string `json:"timezone" jsonschema:"required,description=The timezone to get time for (e.g. America/New_York)"`
}

type TimeTool struct{}

func (t TimeTool) Name() string {
	return "get_time"
}

func (t TimeTool) Description() string {
	return "Get current time for a given timezone"
}

func (t TimeTool) Call(ctx context.Context, input string) (string, error) {
	var params TimeToolInput
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("failed to parse time tool input: %v", err)
	}

	return fmt.Sprintf("The time in %s is 12:00 PM", params.Timezone), nil
}

func (t TimeTool) GetAnthropicToolSpec() anthropic.ToolInputSchemaParam {
	return generateAnthropicSchema[TimeToolInput]()
}