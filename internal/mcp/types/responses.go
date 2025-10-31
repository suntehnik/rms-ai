package types

import (
	"encoding/json"
)

// ToolResponse represents the response from a tool call
type ToolResponse struct {
	Content []ContentItem `json:"content"`
}

// ContentItem represents a single content item in a tool response
type ContentItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// CreateToolResponse creates a standard tool response with message and optional data
func CreateToolResponse(message string, data interface{}) *ToolResponse {
	content := []ContentItem{
		{
			Type: "text",
			Text: message,
		},
	}

	if data != nil {
		if jsonData, err := json.MarshalIndent(data, "", "  "); err == nil {
			content = append(content, ContentItem{
				Type: "text",
				Text: string(jsonData),
			})
		}
	}

	return &ToolResponse{Content: content}
}

// CreateSuccessResponse creates a success response with formatted message
func CreateSuccessResponse(message string) *ToolResponse {
	return &ToolResponse{
		Content: []ContentItem{
			{
				Type: "text",
				Text: message,
			},
		},
	}
}

// CreateDataResponse creates a response with both message and structured data
func CreateDataResponse(message string, data interface{}) *ToolResponse {
	return CreateToolResponse(message, data)
}
