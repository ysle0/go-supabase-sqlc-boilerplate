package ws_example

// Response represents a generic WebSocket response
type Response struct {
	Type    string      `json:"type"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Success bool        `json:"success"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Type    string `json:"type"`
	Error   string `json:"error"`
	Success bool   `json:"success"`
}

// NewErrorResponse creates a new error response
func NewErrorResponse(packetType, errorMsg string) ErrorResponse {
	return ErrorResponse{
		Type:    packetType,
		Error:   errorMsg,
		Success: false,
	}
}

// NewSuccessResponse creates a new success response
func NewSuccessResponse(packetType string, data interface{}) Response {
	return Response{
		Type:    packetType,
		Data:    data,
		Success: true,
	}
}
