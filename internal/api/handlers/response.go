// Package handlers provides HTTP handlers for the AegisOpsAI REST API.
//
// Common response helpers are defined in this file and used across all handlers.
package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/AegisOpsAI/aegisops/internal/models"
)

// APIResponse is the standard JSON response envelope.
type APIResponse struct {
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

// JSON writes a JSON response with the given status code and data payload.
func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	resp := APIResponse{Data: data}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("failed to encode JSON response", slog.String("error", err.Error()))
	}
}

// JSONError writes a JSON error response with the given status code and message.
func JSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	resp := APIResponse{Error: message}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("failed to encode JSON error response", slog.String("error", err.Error()))
	}
}

// JSONRaw writes an arbitrary value as a JSON response without the envelope.
func JSONRaw(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to encode JSON response", slog.String("error", err.Error()))
	}
}

// ParsePagination extracts pagination parameters from query strings with defaults.
func ParsePagination(r *http.Request) models.PaginationParams {
	p := models.DefaultPagination()

	if v := r.URL.Query().Get("page"); v != "" {
		if page, err := strconv.Atoi(v); err == nil && page > 0 {
			p.Page = page
		}
	}
	if v := r.URL.Query().Get("pageSize"); v != "" {
		if size, err := strconv.Atoi(v); err == nil && size > 0 {
			p.PageSize = size
		}
	}

	return p
}

// DecodeJSON decodes a JSON request body into the given target.
// It returns false and writes an error response if decoding fails.
func DecodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	if r.Body == nil {
		JSONError(w, http.StatusBadRequest, "request body is required")
		return false
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		JSONError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return false
	}

	return true
}
