// Package modes - Unified Anthropic API types
package modes

// APIRequest represents a request to the Anthropic API
type APIRequest struct {
	Model       string    `json:"model"`
	MaxTokens   int       `json:"max_tokens"`
	System      string    `json:"system,omitempty"`
	Messages    []Message `json:"messages"`
	Tools       []any     `json:"tools,omitempty"` // Can be Tool or ServerTool
	ToolChoice  any       `json:"tool_choice,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
}

// Message represents a conversation message
type Message struct {
	Role    string `json:"role"`
	Content any    `json:"content"` // string or []ContentBlock
}

// ContentBlock represents a block in message content
type ContentBlock struct {
	Type      string         `json:"type"`
	Text      string         `json:"text,omitempty"`
	ID        string         `json:"id,omitempty"`
	Name      string         `json:"name,omitempty"`
	Input     map[string]any `json:"input,omitempty"`
	ToolUseID string         `json:"tool_use_id,omitempty"`
	Content   any            `json:"content,omitempty"` // string for tool_result, array for web_search_tool_result
	IsError   bool           `json:"is_error,omitempty"`
}

// Tool represents a regular tool definition for the API
type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	InputSchema any    `json:"input_schema"`
}

// ServerTool represents a server-side tool (e.g., web_search)
type ServerTool struct {
	Type         string        `json:"type"`                    // e.g., "web_search_20250305"
	Name         string        `json:"name"`                    // e.g., "web_search"
	MaxUses      int           `json:"max_uses,omitempty"`      // Optional: limit searches per request
	UserLocation *UserLocation `json:"user_location,omitempty"` // Optional: localization
}

// UserLocation for localizing search results
type UserLocation struct {
	Type     string `json:"type"` // "approximate"
	City     string `json:"city,omitempty"`
	Region   string `json:"region,omitempty"`
	Country  string `json:"country,omitempty"`
	Timezone string `json:"timezone,omitempty"`
}

// APIResponse represents the API response
type APIResponse struct {
	Content    []ResponseBlock `json:"content"`
	StopReason string          `json:"stop_reason"`
	Usage      Usage           `json:"usage"`
}

// ResponseBlock represents a content block in the response
type ResponseBlock struct {
	Type      string         `json:"type"`
	Text      string         `json:"text,omitempty"`
	ID        string         `json:"id,omitempty"`
	Name      string         `json:"name,omitempty"`
	Input     map[string]any `json:"input,omitempty"`
	ToolUseID string         `json:"tool_use_id,omitempty"` // For web_search_tool_result
	Content   any            `json:"content,omitempty"`     // For web_search_tool_result (array of results)
	Citations []Citation     `json:"citations,omitempty"`   // For text blocks with citations
}

// WebSearchResult represents a search result in web_search_tool_result
type WebSearchResult struct {
	Type             string `json:"type"`
	URL              string `json:"url"`
	Title            string `json:"title"`
	EncryptedContent string `json:"encrypted_content,omitempty"`
	PageAge          string `json:"page_age,omitempty"`
}

// Usage tracks token usage
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// NewTextMessage creates a text message
func NewTextMessage(role, content string) Message {
	return Message{Role: role, Content: content}
}

// NewBlockMessage creates a message with content blocks
func NewBlockMessage(role string, blocks []ContentBlock) Message {
	return Message{Role: role, Content: blocks}
}

// TextBlock creates a text content block
func TextBlock(text string) ContentBlock {
	return ContentBlock{Type: "text", Text: text}
}

// ToolUseBlock creates a tool use content block
func ToolUseBlock(id, name string, input map[string]any) ContentBlock {
	return ContentBlock{
		Type:  "tool_use",
		ID:    id,
		Name:  name,
		Input: input,
	}
}

// ToolResultBlock creates a tool result content block
func ToolResultBlock(toolUseID, content string, isError bool) ContentBlock {
	return ContentBlock{
		Type:      "tool_result",
		ToolUseID: toolUseID,
		Content:   content,
		IsError:   isError,
	}
}
