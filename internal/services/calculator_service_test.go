package service

import (
	"errors"
	"sync"
	"testing"
	"time"
)

/* ------------------ basics & recording ------------------ */

func TestOperations_RecordHistory(t *testing.T) {
	svc := NewCalculatorService(WithMaxHistory(100))

	// Perform ops
	gotAdd := svc.Add(1, 2)          // 3
	gotSub := svc.Subtract(10, 4.5)  // 5.5
	gotMul := svc.Multiply(3, 7)     // 21
	gotDiv, err := svc.Divide(21, 7) // ~3.142857...

	if err != nil {
		t.Fatalf("Divide(21,7) unexpected error: %v", err)
	}
	if gotAdd != 3 || gotSub != 5.5 || gotMul != 21 || gotDiv != 3 {
		t.Fatalf("wrong results: add=%v sub=%v mul=%v div=%v", gotAdd, gotSub, gotMul, gotDiv)
	}

	// History newest-first. Expect: divide, multiply, subtract, add
	h := svc.GetHistory(10)
	if len(h) < 4 {
		t.Fatalf("history len=%d, want >=4", len(h))
	}

	want := []struct {
		op      string
		a, b    float64
		checkFn func(float64) bool
	}{
		{"divide", 21, 7, func(v float64) bool { return v == 3 }},
		{"multiply", 3, 7, func(v float64) bool { return v == 21 }},
		{"subtract", 10, 4.5, func(v float64) bool { return v == 5.5 }},
		{"add", 1, 2, func(v float64) bool { return v == 3 }},
	}

	for i := 0; i < 4; i++ {
		e := h[i]
		if e.Op != want[i].op || e.A != want[i].a || e.B != want[i].b {
			t.Fatalf("history[%d] = %+v; want op=%s a=%v b=%v", i, e, want[i].op, want[i].a, want[i].b)
		}
		if e.Error != "" {
			t.Fatalf("history[%d] has unexpected error: %q", i, e.Error)
		}
		if !want[i].checkFn(e.Result) {
			t.Fatalf("history[%d] result=%v failed check for op %s", i, e.Result, e.Op)
		}
		if e.Time.IsZero() {
			t.Fatalf("history[%d] time must be set", i)
		}
	}
}

func TestDivideByZero_RecordsError(t *testing.T) {
	svc := NewCalculatorService(WithMaxHistory(100))

	_, err := svc.Divide(5, 0)
	if err == nil {
		t.Fatalf("Divide(5,0) expected error, got nil")
	}
	if !errors.Is(err, err) && err.Error() != "division by zero is not allowed" {
		t.Fatalf("unexpected error: %v", err)
	}

	h := svc.GetHistory(1)
	if len(h) != 1 {
		t.Fatalf("history len=%d, want 1", len(h))
	}
	e := h[0]
	if e.Op != "divide" || e.A != 5 || e.B != 0 {
		t.Fatalf("history entry mismatch: %+v", e)
	}
	if e.Error == "" {
		t.Fatalf("expected error recorded in history")
	}
	if e.Result != 0 {
		t.Fatalf("expected result=0 when error, got %v", e.Result)
	}
}

/* ------------------ order, limit, clear ------------------ */

func TestGetHistory_LimitAndOrder(t *testing.T) {
	svc := NewCalculatorService(WithMaxHistory(100))

	// Create 5 entries: add(i, i)
	for i := 1; i <= 5; i++ {
		svc.Add(float64(i), float64(i)) // results 2,4,6,8,10
	}

	// Limit=3 → newest-first: i=5,4,3
	h := svc.GetHistory(3)
	if len(h) != 3 {
		t.Fatalf("len=%d; want 3", len(h))
	}
	if h[0].A != 5 || h[1].A != 4 || h[2].A != 3 {
		t.Fatalf("order wrong (newest first expected), got A values: %v, %v, %v", h[0].A, h[1].A, h[2].A)
	}
}

func TestClearHistory(t *testing.T) {
	svc := NewCalculatorService(WithMaxHistory(100))
	svc.Add(1, 2)
	svc.Multiply(3, 4)
	if n := len(svc.GetHistory(10)); n < 2 {
		t.Fatalf("precondition history len=%d; want >=2", n)
	}
	svc.ClearHistory()
	if n := len(svc.GetHistory(10)); n != 0 {
		t.Fatalf("after ClearHistory, len=%d; want 0", n)
	}
}

/* ------------------ cap / trimming behavior ------------------ */

func TestMaxHistoryCap_TrimsOldest(t *testing.T) {
	svc := NewCalculatorService(WithMaxHistory(3))

	// Produce 5 entries with distinct IDs and A values
	for i := 0; i < 5; i++ {
		svc.Add(float64(i), 0)
	}

	h := svc.GetHistory(10) // returns newest-first, but only last 3 kept
	if len(h) != 3 {
		t.Fatalf("len=%d; want 3 (maxHistory)", len(h))
	}

	// We created 5 entries (IDs 0..4). After trimming, remaining physical slice
	// is entries with IDs 2,3,4. GetHistory reverses to newest-first → 4,3,2.
	if h[0].ID != 4 || h[1].ID != 3 || h[2].ID != 2 {
		t.Fatalf("IDs wrong after trim (newest-first want 4,3,2): got %d,%d,%d", h[0].ID, h[1].ID, h[2].ID)
	}
	if h[0].A != 4 || h[1].A != 3 || h[2].A != 2 {
		t.Fatalf("payload mismatch after trim: A=%v,%v,%v", h[0].A, h[1].A, h[2].A)
	}
}

/* ------------------ basic concurrency smoke test ------------------ */

func TestService_Concurrency_IsThreadSafe(t *testing.T) {
	const (
		goroutines = 20
		perG       = 50
	)
	svc := NewCalculatorService(WithMaxHistory(goroutines*perG + 10))

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for g := 0; g < goroutines; g++ {
		go func(offset float64) {
			defer wg.Done()
			for i := 0; i < perG; i++ {
				// Mix operations a bit
				switch i % 4 {
				case 0:
					_ = svc.Add(offset, float64(i))
				case 1:
					_ = svc.Subtract(offset+float64(i), 1)
				case 2:
					_ = svc.Multiply(2, float64(i))
				case 3:
					_, _ = svc.Divide(float64(i)+1, 1) // avoid divide-by-zero here
				}
				time.Sleep(time.Microsecond) // tiny yield to interleave
			}
		}(float64(g))
	}
	wg.Wait()

	// Expect at least goroutines*perG history entries (no trimming)
	h := svc.GetHistory(goroutines*perG + 10)
	if len(h) != goroutines*perG {
		t.Fatalf("history len=%d; want %d", len(h), goroutines*perG)
	}
	// Spot-check latest entry fields make sense
	if h[0].Op == "" || h[0].Time.IsZero() {
		t.Fatalf("latest history entry looks invalid: %+v", h[0])
	}
}
