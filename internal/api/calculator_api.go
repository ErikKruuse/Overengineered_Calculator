package api

import (
	"encoding/json"
	service "erikkruuse/calculator/internal/services"
	"net/http"
	"strconv"
	"strings"
)

type API struct {
	svc service.CalculatorService
}

func New(svc service.CalculatorService) *API {
	return &API{svc: svc}
}

func (a *API) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("GET /v1/history", a.getHistory)
	mux.HandleFunc("DELETE /v1/history", a.clearHistory)

	mux.HandleFunc("GET /v1/calculate", a.calculateQuery)
	mux.HandleFunc("POST /v1/add", a.binaryOp(func(a1, b1 float64) (float64, error) { return a.svc.Add(a1, b1), nil }))
	mux.HandleFunc("POST /v1/subtract", a.binaryOp(func(a1, b1 float64) (float64, error) { return a.svc.Subtract(a1, b1), nil }))
	mux.HandleFunc("POST /v1/multiply", a.binaryOp(func(a1, b1 float64) (float64, error) { return a.svc.Multiply(a1, b1), nil }))
	mux.HandleFunc("POST /v1/divide", a.binaryOp(a.svc.Divide))
}

func (a *API) getHistory(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if s := r.URL.Query().Get("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			limit = n
		}
	}
	items := a.svc.GetHistory(limit)
	WriteJSON(w, http.StatusOK, items)
}

func (a *API) clearHistory(w http.ResponseWriter, r *http.Request) {
	a.svc.ClearHistory()
	WriteJSON(w, http.StatusOK, map[string]string{"status": "cleaned"})
}

type calcRequest struct {
	A json.Number `json:"a"`
	B json.Number `json:"b"`
}

type calcResponse struct {
	Result float64 `json:"result,omitempty"`
	Error  string  `json:"error,omitempty"`
}

type binOp func(a, b float64) (float64, error)

func (a *API) binaryOp(op binOp) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req calcRequest
		if err := DecodeJSON(r, w, &req); err != nil {
			WriteProblem(w, http.StatusBadRequest, "invalid_json", err.Error())
			return
		}

		// Parse to float64 now (may yield +Inf/-Inf/NaN).
		av, errA := parseJSONNumber(req.A)
		bv, errB := parseJSONNumber(req.B)
		if errA != nil || errB != nil {
			WriteProblem(w, http.StatusBadRequest, "invalid_json", "a and b must be numbers")
			return
		}
		if !isFinite(av) || !isFinite(bv) {
			WriteProblem(w, http.StatusBadRequest, "invalid_input", "inputs must be finite numbers")
			return
		}

		res, err := op(av, bv)
		if err != nil {
			WriteProblem(w, http.StatusBadRequest, "calculation_error", err.Error())
			return
		}
		WriteJSON(w, http.StatusOK, calcResponse{Result: res})
	}
}

func (a *API) calculateQuery(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	op := strings.ToLower(q.Get("op"))
	aStr, bStr := q.Get("a"), q.Get("b")

	if op == "" || aStr == "" || bStr == "" {
		WriteProblem(w, http.StatusBadRequest, "missing_params", "op, a, and b are required")
		return
	}

	av, err1 := strconv.ParseFloat(aStr, 64)
	bv, err2 := strconv.ParseFloat(bStr, 64)
	if err1 != nil || err2 != nil || !isFinite(av) || !isFinite(bv) {
		WriteProblem(w, http.StatusBadRequest, "invalid_input", "a and b must be valid finite numbers")
		return
	}

	var (
		res float64
		err error
	)
	switch op {
	case "add", "+":
		res = a.svc.Add(av, bv)
	case "subtract", "-":
		res = a.svc.Subtract(av, bv)
	case "multiply", "*", "x":
		res = a.svc.Multiply(av, bv)
	case "divide", "/":
		res, err = a.svc.Divide(av, bv)
	default:
		WriteProblem(w, http.StatusBadRequest, "invalid_op", "use add|subtract|multiply|divide")
		return
	}
	if err != nil {
		WriteProblem(w, http.StatusBadRequest, "calculation_error", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, calcResponse{Result: res})
}
