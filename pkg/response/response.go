// pkg/response/response.go
package response

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Data    any         `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// Meta represents response metadata
type Meta struct {
	Page       int `json:"page,omitempty"`
	Limit      int `json:"limit,omitempty"`
	Total      int `json:"total,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

// JSON sends a JSON response
func JSON(w http.ResponseWriter, status int, data any) {
	response := Response{
		Success: status < 400,
		Data:    data,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("failed to encode JSON response", "error", err)
	}
}

// Error sends an error response
func Error(w http.ResponseWriter, status int, message string) {
	response := Response{
		Success: false,
		Error:   message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("failed to encode error response", "error", err)
	}
}

// Success sends a success response
func Success(w http.ResponseWriter, data any) {
	JSON(w, http.StatusOK, data)
}

// Created sends a created response
func Created(w http.ResponseWriter, data any) {
	JSON(w, http.StatusCreated, data)
}

// ValidationError sends validation error response
func ValidationError(w http.ResponseWriter, errors map[string]string) {
	JSON(w, http.StatusUnprocessableEntity, map[string]interface{}{
		"validation_errors": errors,
	})
}
