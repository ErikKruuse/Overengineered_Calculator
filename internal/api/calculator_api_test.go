package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	service "erikkruuse/calculator/internal/services"
)

/* ---------- helpers ---------- */

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	svc := service.NewCalculatorService(service.WithMaxHistory(100))
	a := New(svc)
	mux := http.NewServeMux()
	a.RegisterRoutes(mux)
	return httptest.NewServer(mux)
}

func postJSON(t *testing.T, url string, body any) (*http.Response, []byte) {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode json: %v", err)
		}
	}
	req, _ := http.NewRequest(http.MethodPost, url, &buf)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST %s failed: %v", url, err)
	}
	b := readAll(t, resp)
	return resp, b
}

func postRaw(t *testing.T, url string, raw string, contentType string) (*http.Response, []byte) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodPost, url, strings.NewReader(raw))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST %s failed: %v", url, err)
	}
	b := readAll(t, resp)
	return resp, b
}

func get(t *testing.T, url string) (*http.Response, []byte) {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET %s failed: %v", url, err)
	}
	b := readAll(t, resp)
	return resp, b
}

func del(t *testing.T, url string) (*http.Response, []byte) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodDelete, url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE %s failed: %v", url, err)
	}
	b := readAll(t, resp)
	return resp, b
}

func readAll(t *testing.T, resp *http.Response) []byte {
	t.Helper()
	defer resp.Body.Close()
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		t.Fatalf("read body: %v", err)
	}
	return buf.Bytes()
}

/* ---------- tests ---------- */

func TestHealth(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	resp, body := get(t, ts.URL+"/health")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d body=%s", resp.StatusCode, string(body))
	}
	var m map[string]string
	_ = json.Unmarshal(body, &m)
	if m["status"] != "ok" {
		t.Fatalf("want status=ok, got %v", m)
	}
}

func TestAdd_Success(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()
	want := 3.75

	resp, body := postJSON(t, ts.URL+"/v1/add", map[string]any{"a": 1.5, "b": 2.25})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d body=%s", resp.StatusCode, string(body))
	}
	var got struct {
		Result float64 `json:"result"`
	}
	json.Unmarshal(body, &got)
	if got.Result != want {
		t.Fatalf("want %v, got %v", want, got.Result)
	}
}

func TestSubtract_Success(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()
	want := -0.75

	resp, body := postJSON(t, ts.URL+"/v1/subtract", map[string]any{"a": 1.5, "b": 2.25})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d body=%s", resp.StatusCode, string(body))
	}
	var got struct {
		Result float64 `json:"result"`
	}
	json.Unmarshal(body, &got)
	if got.Result != -0.75 {
		t.Fatalf("want %v, got %v", want, got.Result)
	}
}

func TestDivide_Success(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()
	want := 5.0

	resp, body := postJSON(t, ts.URL+"/v1/divide", map[string]any{"a": 10, "b": 2})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d body=%s", resp.StatusCode, string(body))
	}
	var got struct {
		Result float64 `json:"result"`
	}
	json.Unmarshal(body, &got)
	if got.Result != want {
		t.Fatalf("want %v, got %v", want, got.Result)
	}
}

func TestMultiply_Success(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()
	want := 20.0

	resp, body := postJSON(t, ts.URL+"/v1/multiply", map[string]any{"a": 10, "b": 2})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d body=%s", resp.StatusCode, string(body))
	}
	var got struct {
		Result float64 `json:"result"`
	}
	json.Unmarshal(body, &got)
	if got.Result != want {
		t.Fatalf("want %v, got %v", want, got.Result)
	}
}

func TestDivideByZero_ReturnsProblem(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	resp, body := postJSON(t, ts.URL+"/v1/divide", map[string]any{"a": 5, "b": 0})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", resp.StatusCode, string(body))
	}
	var p Problem
	json.Unmarshal(body, &p)
	if p.Title != "calculation_error" {
		t.Fatalf("want title=calculation_error, got %q", p.Title)
	}
	if p.Status != 400 {
		t.Fatalf("want status=400, got %d", p.Status)
	}
	if !strings.Contains(strings.ToLower(p.Detail), "division by zero") {
		t.Fatalf("detail should mention division by zero; got %q", p.Detail)
	}
}

func TestCalculate_InvalidOp(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	resp, body := get(t, ts.URL+"/v1/calculate?op=pow&a=2&b=3")
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", resp.StatusCode, string(body))
	}
	var p Problem
	json.Unmarshal(body, &p)
	if p.Title != "invalid_op" || p.Status != 400 {
		t.Fatalf("want invalid_op/400, got %+v", p)
	}
}

func TestCalculate_MissingParams(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	// missing b
	resp, body := get(t, ts.URL+"/v1/calculate?op=add&a=2")
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", resp.StatusCode, string(body))
	}
	var p Problem
	json.Unmarshal(body, &p)
	if p.Title != "missing_params" || p.Status != 400 {
		t.Fatalf("want missing_params/400, got %+v", p)
	}
}

func TestInvalidJSON_UnknownField(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	// extra field triggers DisallowUnknownFields
	resp, body := postJSON(t, ts.URL+"/v1/add", map[string]any{"a": 1, "b": 2, "extra": "fish"})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", resp.StatusCode, string(body))
	}
	var p Problem
	json.Unmarshal(body, &p)
	if p.Title != "invalid_json" {
		t.Fatalf("want title=invalid_json, got %q", p.Title)
	}
}

func TestBinaryOp_InvalidJSON_NonNumericValues(t *testing.T) {
	t.Parallel()

	a := New(service.NewCalculatorService())
	h := a.binaryOp(func(a, b float64) (float64, error) {
		t.Fatalf("op should not be invoked when parsing fails")
		return 0, nil
	})

	body := `{"a": "not-a-number", "b": 5}`
	req := httptest.NewRequest(http.MethodPost, "/v1/add", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s; want 400", rr.Code, rr.Body.String())
	}

	var p Problem
	if err := json.Unmarshal(rr.Body.Bytes(), &p); err != nil {
		t.Fatalf("unmarshal problem: %v; body=%s", err, rr.Body.String())
	}
	if p.Title != "invalid_json" || p.Status != http.StatusBadRequest {
		t.Fatalf("want title=invalid_json/status=400, got %+v", p)
	}

	// Optional: sanity check that a detail exists (don’t hardcode exact text)
	if p.Detail == "" {
		t.Fatalf("expected non-empty detail for invalid_json")
	}
}

func TestBinaryOp_InvalidInput_NonFinite_FromJSON(t *testing.T) {
	t.Parallel()

	// Build handler for POST /v1/add path through binaryOp
	a := New(service.NewCalculatorService())
	h := a.binaryOp(func(a, b float64) (float64, error) {
		t.Fatalf("op must not be called when inputs are non-finite")
		return 0, nil
	})

	// Use a float64 overflow that decodes as json.Number -> Float64() -> +Inf
	body := `{"a": 1e309, "b": 2}`
	req := httptest.NewRequest(http.MethodPost, "/v1/add", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s; want 400", rr.Code, rr.Body.String())
	}
	var p Problem
	if err := json.Unmarshal(rr.Body.Bytes(), &p); err != nil {
		t.Fatalf("unmarshal problem: %v; body=%s", err, rr.Body.String())
	}
	if p.Title != "invalid_input" {
		t.Fatalf("want title=invalid_input, got %q (status=%d, detail=%q)", p.Title, p.Status, p.Detail)
	}
	if p.Status != http.StatusBadRequest {
		t.Fatalf("want status=400, got %d", p.Status)
	}
	if !strings.Contains(strings.ToLower(p.Detail), "finite") {
		t.Fatalf("expected detail to mention finite numbers, got %q", p.Detail)
	}
}

func TestInvalidJSON_MissingContentType(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	// POST without Content-Type: application/json
	resp, body := postRaw(t, ts.URL+"/v1/add", `{"a":1,"b":2}`, "")
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", resp.StatusCode, string(body))
	}
	var p Problem
	json.Unmarshal(body, &p)
	if p.Title != "invalid_json" {
		t.Fatalf("want title=invalid_json, got %q", p.Title)
	}
}

func TestHistory_Flow(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	// perform a few operations (some success + one error)
	_, _ = postJSON(t, ts.URL+"/v1/add", map[string]any{"a": 2, "b": 3})
	_, _ = postJSON(t, ts.URL+"/v1/multiply", map[string]any{"a": 3, "b": 7})
	_, _ = postJSON(t, ts.URL+"/v1/divide", map[string]any{"a": 5, "b": 0}) // error, still recorded

	// history (no limit)
	resp, body := get(t, ts.URL+"/v1/history")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d body=%s", resp.StatusCode, string(body))
	}
	var items []service.HistoryEntry
	json.Unmarshal(body, &items)
	if len(items) < 3 {
		t.Fatalf("expected >= 3 history items, got %d", len(items))
	}

	// limit=2
	resp, body = get(t, ts.URL+"/v1/history?limit=2")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d body=%s", resp.StatusCode, string(body))
	}
	items = nil
	json.Unmarshal(body, &items)
	if len(items) > 2 {
		t.Fatalf("expected <= 2 history items, got %d", len(items))
	}

	// clear
	resp, body = del(t, ts.URL+"/v1/history")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d body=%s", resp.StatusCode, string(body))
	}

	// verify empty
	resp, body = get(t, ts.URL+"/v1/history")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d body=%s", resp.StatusCode, string(body))
	}
	items = nil
	json.Unmarshal(body, &items)
	if len(items) != 0 {
		t.Fatalf("expected empty history after clear, got %d", len(items))
	}
}
