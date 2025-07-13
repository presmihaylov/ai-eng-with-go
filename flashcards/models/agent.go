package models

type AgentMessage struct {
	Role        string       `json:"role"`
	Content     string       `json:"content"`
	ToolCalls   []ToolCall   `json:"tool_calls,omitempty"`
	ToolResults []ToolResult `json:"tool_results,omitempty"`
}

type ToolCall struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

type ToolResult struct {
	ToolCallID string `json:"tool_call_id"`
	Content    string `json:"content"`
}

type AgentRequest struct {
	Messages []AgentMessage `json:"messages"`
}

type AgentResponse struct {
	Messages []AgentMessage `json:"messages"`
}