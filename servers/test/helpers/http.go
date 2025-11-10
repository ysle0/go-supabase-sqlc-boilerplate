package helpers

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
)

// StandardResponse is a generic response wrapper matching the API response format
type (
	StandardResponse[T any] struct {
		Message string `json:"message"`
		Data    T      `json:"data"`
	}

	// ErrorResponse represents the error response structure returned by httputil.ErrWithMsg
	ErrorResponse struct {
		Status  string `json:"status"`
		Message string `json:"msg"`
	}
)

// CreateTestLogger creates a logger configured for testing with error-level output
func CreateTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

// AddLoggerToRequest adds a test logger to the request context
func AddLoggerToRequest(req *http.Request) *http.Request {
	logger := CreateTestLogger()
	ctx := context.WithValue(req.Context(), "logger", logger)
	return req.WithContext(ctx)
}

// CreateJSONRequest creates an HTTP test request with JSON body and logger in context
func CreateJSONRequest(method, path string, body interface{}) (*http.Request, error) {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = AddLoggerToRequest(req)

	return req, nil
}

// MustCreateJSONRequest is like CreateJSONRequest but panics on error (useful with testify's Require)
func MustCreateJSONRequest(method, path string, body interface{}) *http.Request {
	req, err := CreateJSONRequest(method, path, body)
	if err != nil {
		panic(err)
	}
	return req
}

// DecodeJSONResponse decodes a JSON response into the provided struct
// The response parameter should be a pointer to the target struct
func DecodeJSONResponse(w *httptest.ResponseRecorder, response interface{}) error {
	return json.NewDecoder(w.Body).Decode(response)
}

// DecodeStandardResponse decodes a standard API response with typed data
func DecodeStandardResponse[T any](w *httptest.ResponseRecorder) (*StandardResponse[T], error) {
	var response StandardResponse[T]
	err := json.NewDecoder(w.Body).Decode(&response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// DecodeErrorResponse decodes an error response returned by httputil.ErrWithMsg
func DecodeErrorResponse(w *httptest.ResponseRecorder) (*ErrorResponse, error) {
	var response ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}
