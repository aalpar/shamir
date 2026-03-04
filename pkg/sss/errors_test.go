package sss

import (
	"errors"
	"math/big"
	"testing"

	"github.com/aalpar/shamir/pkg/field"
)

func TestSplitThresholdErrors(t *testing.T) {
	f := field.New(big.NewInt(17))
	secret := f.NewElement(big.NewInt(5))

	tests := []struct {
		name string
		n, k int
	}{
		{"k < 2", 5, 1},
		{"k = 0", 5, 0},
		{"k > n", 3, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Split(secret, tt.n, tt.k, f)
			if !errors.Is(err, ErrThreshold) {
				t.Errorf("errors.Is(err, ErrThreshold) = false; err = %v", err)
			}
			var se *Error
			if !errors.As(err, &se) {
				t.Fatal("errors.As failed")
			}
			if se.Op != "Split" {
				t.Errorf("Op = %q, want %q", se.Op, "Split")
			}
		})
	}
}

func TestCombineNoSharesError(t *testing.T) {
	_, err := Combine(nil)
	if !errors.Is(err, ErrNoShares) {
		t.Errorf("errors.Is(err, ErrNoShares) = false; err = %v", err)
	}
	var se *Error
	if !errors.As(err, &se) {
		t.Fatal("errors.As failed")
	}
	if se.Op != "Combine" {
		t.Errorf("Op = %q, want %q", se.Op, "Combine")
	}
}

func TestUnmarshalError(t *testing.T) {
	var s Share
	err := s.UnmarshalJSON([]byte(`{"p":"zz","x":"1","y":"2"}`))
	if !errors.Is(err, ErrUnmarshal) {
		t.Errorf("errors.Is(err, ErrUnmarshal) = false; err = %v", err)
	}
}

func TestSplitWrapsPolynomialError(t *testing.T) {
	// When Split wraps a lower-layer error, errors.As still works.
	f := field.New(big.NewInt(17))
	secret := f.NewElement(big.NewInt(5))

	// Valid threshold, so any error would come from polynomial layer.
	// Just verify that success produces no error (regression).
	_, err := Split(secret, 5, 3, f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
