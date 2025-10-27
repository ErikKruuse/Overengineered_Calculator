package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

const maxBodyBytes = 1 << 20 // 1MB

// DecodeJSON strictly decodes JSON from the request body into v.
// Enforces Content-Type for write methods, caps body size, and disallows unknown fields.
func DecodeJSON(r *http.Request, w http.ResponseWriter, v any) error {
	switch r.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		ct := r.Header.Get("Content-Type")
		if !strings.HasPrefix(strings.ToLower(ct), "application/json") {
			return errors.New("Content-Type must be application/json")
		}
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	dec.UseNumber()

	if err := dec.Decode(v); err != nil {
		return err
	}
	// No trailing payload allowed
	if dec.More() {
		return errors.New("unexpected trailing data")
	}
	return nil
}

// WriteJSON writes a JSON response with the given status.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// Problem is a minimal error envelope (inspired by RFC7807).
type Problem struct {
	Type   string `json:"type,omitempty"`
	Title  string `json:"title,omitempty"`
	Status int    `json:"status,omitempty"`
	Detail string `json:"detail,omitempty"`
}

// WriteProblem writes a standardized error response.
func WriteProblem(w http.ResponseWriter, status int, title string, detail string) {
	WriteJSON(w, status, Problem{
		Type:   "", // could set a canonical URL per error class later
		Title:  title,
		Status: status,
		Detail: detail,
	})
}
