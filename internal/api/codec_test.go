package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDecodeJSON_OK(t *testing.T) {
	body := []byte(`{"a":1,"b":2}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/add", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	var got struct{ A, B float64 }
	if err := DecodeJSON(req, rr, &got); err != nil {
		t.Fatalf("DecodeJSON() error = %v", err)
	}
	if got.A != 1 || got.B != 2 {
		t.Fatalf("decoded wrong values: %+v", got)
	}
}

func TestDecodeJSON_RejectsUnknownFields(t *testing.T) {
	body := []byte(`{"a":1,"b":2,"extra":"nope"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/add", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	var got struct{ A, B float64 }
	if err := DecodeJSON(req, rr, &got); err == nil {
		t.Fatalf("expected error for unknown field, got nil")
	}
}

func TestDecodeJSON_RequiresJSONCT(t *testing.T) {
	body := []byte(`{"a":1,"b":2}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/add", bytes.NewReader(body))
	// no content-type set
	rr := httptest.NewRecorder()

	var got struct{ A, B float64 }
	if err := DecodeJSON(req, rr, &got); err == nil {
		t.Fatalf("expected content-type error, got nil")
	}
}

func TestWriteProblem(t *testing.T) {
	rr := httptest.NewRecorder()
	WriteProblem(rr, http.StatusBadRequest, "invalid_input", "a and b must be numbers")

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d; want 400", rr.Code)
	}
	ct := rr.Header().Get("Content-Type")
	if ct == "" || ct[:16] != "application/json" {
		t.Fatalf("content-type = %q; want application/json", ct)
	}
}
