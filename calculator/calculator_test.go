package calculator

import (
	"testing"
)

func TestAdd(t *testing.T) {
	tests := []struct {
		a, b float64
		want float64
	}{
		{0, 0, 0},
		{1, 2, 3},
		{-2, 1, -1},
		{1.5, 3.75, 5.25},
	}
	for _, test := range tests {
		res := Add(test.a, test.b)
		if res != test.want {
			t.Errorf("Add(%v, %v) = %v, want %v", test.a, test.b, res, test.want)
		}
	}
}

func TestSubtract(t *testing.T) {
	tests := []struct {
		a, b float64
		want float64
	}{
		{0, 0, 0},
		{1, 2, -1},
		{-2, 1, -3},
		{1.5, 3.75, -2.25},
	}
	for _, test := range tests {
		res := Subtract(test.a, test.b)
		if res != test.want {
			t.Errorf("Subtract(%v, %v) = %v, want %v", test.a, test.b, res, test.want)
		}
	}
}

func TestMultiply(t *testing.T) {
	tests := []struct {
		a, b float64
		want float64
	}{
		{0, 0, 0},
		{1, 2, 2},
		{-2, 2, -4},
		{-2, -2, 4},
		{1.5, 2, 3},
	}
	for _, test := range tests {
		res := Multiply(test.a, test.b)
		if res != test.want {
			t.Errorf("Multiply(%v, %v) = %v, want %v", test.a, test.b, res, test.want)
		}
	}
}

func TestDivide(t *testing.T) {
	type want struct {
		val float64
		err bool
	}
	tests := []struct {
		a, b float64
		want want
	}{
		{0, 0, want{0, true}},
		{1, 2, want{0.5, false}},
		{-4, 2, want{-2, false}},
		{-2, -2, want{1, false}},
		{1.5, 2, want{0.75, false}},
		{4, 0, want{0, true}},
	}
	for _, test := range tests {
		res, err := Divide(test.a, test.b)
		if test.want.err && err == nil {
			t.Errorf("Divide(%v, %v) expected error, got nil", test.a, test.b)
		}
		if !test.want.err && err != nil {
			t.Errorf("Divide(%v, %v) unexpected error, got %v", test.a, test.b, err)
		}
		if !test.want.err && test.want.val != res {
			t.Errorf("Divide(%v, %v) = %v, want %v", test.a, test.b, res, test.want.val)
		}
	}
}
