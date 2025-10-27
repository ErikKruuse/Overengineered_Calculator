package api

import (
	"encoding/json"
	"math"
	"testing"
)

func TestParseJSONNumber_NormalNumbers(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input json.Number
		want  float64
	}{
		{"int", json.Number("42"), 42},
		{"negative", json.Number("-7"), -7},
		{"decimal", json.Number("123.456"), 123.456},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseJSONNumber(tc.input)
			if err != nil {
				t.Fatalf("parseJSONNumber(%q) unexpected error: %v", tc.input, err)
			}
			if math.Abs(got-tc.want) > 1e-12 {
				t.Fatalf("parseJSONNumber(%q) = %v; want %v", tc.input, got, tc.want)
			}
			if !isFinite(got) {
				t.Fatalf("expected finite for %q, got non-finite", tc.input)
			}
		})
	}
}

func TestParseJSONNumber_OverflowIsAcceptedAsNonFinite(t *testing.T) {
	t.Parallel()

	// 1e309 overflows float64; strconv.ParseFloat returns +Inf with ErrRange.
	// parseJSONNumber should return +Inf and nil error (we treat ErrRange as acceptable).
	got, err := parseJSONNumber(json.Number("1e309"))
	if err != nil {
		t.Fatalf("parseJSONNumber(1e309) unexpected error: %v", err)
	}
	if !math.IsInf(got, +1) {
		t.Fatalf("parseJSONNumber(1e309) = %v; want +Inf", got)
	}
	// isFinite should flag it as non-finite (checked by handler in integration tests)
	if isFinite(got) {
		t.Fatalf("isFinite(+Inf) = true; want false")
	}
}

func TestParseJSONNumber_InvalidReturnsError(t *testing.T) {
	t.Parallel()

	// Not a number at all -> should return an error
	_, err := parseJSONNumber(json.Number("not-a-number"))
	if err == nil {
		t.Fatalf("parseJSONNumber('not-a-number') expected error, got nil")
	}
}

func TestIsFinite(t *testing.T) {
	t.Parallel()

	if !isFinite(123.0) {
		t.Fatalf("isFinite(123) = false; want true")
	}
	if isFinite(math.Inf(+1)) {
		t.Fatalf("isFinite(+Inf) = true; want false")
	}
	if isFinite(math.Inf(-1)) {
		t.Fatalf("isFinite(-Inf) = true; want false")
	}
	if isFinite(math.NaN()) {
		t.Fatalf("isFinite(NaN) = true; want false")
	}
}
