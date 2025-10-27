package service

import (
	"erikkruuse/calculator/calculator"
	"errors"
	"sync"
	"time"
)

type HistoryEntry struct {
	ID     int64     `json:"id"`
	Time   time.Time `json:"time"`
	Op     string    `json:"op"`
	A      float64   `json:"a"`
	B      float64   `json:"b"`
	Result float64   `json:"result,omitempty"`
	Error  string    `json:"error,omitempty"`
}

type CalculatorService interface {
	Add(a, b float64) float64
	Subtract(a, b float64) float64
	Multiply(a, b float64) float64
	Divide(a, b float64) (float64, error)

	GetHistory(limit int) []HistoryEntry
	ClearHistory()
}

func NewCalculatorService(opts ...Option) CalculatorService {
	cfg := config{maxHistory: 1000}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &calcSvc{
		maxHistory: cfg.maxHistory,
	}
}

type config struct {
	maxHistory int
}

type Option func(*config)

func WithMaxHistory(n int) Option {
	return func(c *config) {
		if n > 0 {
			c.maxHistory = n
		}
	}
}

type calcSvc struct {
	mu         sync.Mutex
	history    []HistoryEntry
	nextID     int64
	maxHistory int
}

func (s *calcSvc) record(op string, a, b, result float64, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry := HistoryEntry{
		ID:     s.nextID,
		Time:   time.Now(),
		Op:     op,
		A:      a,
		B:      b,
		Result: result,
	}
	s.nextID++

	if err != nil {
		entry.Error = err.Error()
	}

	s.history = append(s.history, entry)
	if len(s.history) > s.maxHistory {
		drop := len(s.history) - s.maxHistory
		s.history = s.history[drop:]
	}
}

func (s *calcSvc) Add(a, b float64) float64 {
	res := calculator.Add(a, b)
	s.record("add", a, b, res, nil)
	return res
}

func (s *calcSvc) Subtract(a, b float64) float64 {
	res := calculator.Subtract(a, b)
	s.record("subtract", a, b, res, nil)
	return res
}

func (s *calcSvc) Multiply(a, b float64) float64 {
	res := calculator.Multiply(a, b)
	s.record("multiply", a, b, res, nil)
	return res
}

func (s *calcSvc) Divide(a, b float64) (float64, error) {
	if b == 0 {
		err := errors.New("division by zero is not allowed")
		s.record("divide", a, b, 0, err)
		return 0, err
	}
	res, err := calculator.Divide(a, b)
	s.record("divide", a, b, res, err)
	return res, err
}

func (s *calcSvc) GetHistory(limit int) []HistoryEntry {
	s.mu.Lock()
	defer s.mu.Unlock()

	if limit <= 0 || limit > len(s.history) {
		limit = len(s.history)
	}

	out := make([]HistoryEntry, 0, limit)
	for i := 0; i < limit; i++ {
		index := len(s.history) - 1 - i
		if index < 0 {
			break
		}
		out = append(out, s.history[index])
	}
	return out
}

func (s *calcSvc) ClearHistory() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.history = nil
}
