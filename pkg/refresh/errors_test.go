package refresh

import (
	"errors"
	"math/big"
	"testing"

	"github.com/aalpar/shamir/pkg/field"
	"github.com/aalpar/shamir/pkg/sss"
)

func TestZeroSharingThreshold(t *testing.T) {
	f := field.New(big.NewInt(17))

	tests := []struct {
		name string
		n, k int
	}{
		{"k < 2", 5, 1},
		{"k > n", 3, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ZeroSharing(tt.n, tt.k, f)
			if !errors.Is(err, ErrThreshold) {
				t.Errorf("errors.Is(err, ErrThreshold) = false; err = %v", err)
			}
			var re *Error
			if !errors.As(err, &re) {
				t.Fatal("errors.As failed")
			}
			if re.Op != "ZeroSharing" {
				t.Errorf("Op = %q, want %q", re.Op, "ZeroSharing")
			}
		})
	}
}

func TestApplyXMismatch(t *testing.T) {
	f := field.New(big.NewInt(17))
	share := sss.Share{
		X: f.NewElement(big.NewInt(1)),
		Y: f.NewElement(big.NewInt(5)),
	}
	deltas := []SubShare{
		{X: f.NewElement(big.NewInt(2)), Delta: f.NewElement(big.NewInt(3))},
	}

	_, err := Apply(share, deltas)
	if !errors.Is(err, ErrXMismatch) {
		t.Errorf("errors.Is(err, ErrXMismatch) = false; err = %v", err)
	}
	var re *Error
	if !errors.As(err, &re) {
		t.Fatal("errors.As failed")
	}
	if re.Op != "Apply" {
		t.Errorf("Op = %q, want %q", re.Op, "Apply")
	}
}

func TestErrorUnwrapChain(t *testing.T) {
	inner := errors.New("polynomial failed")
	err := &Error{Op: "ZeroSharing", Kind: ErrThreshold, Err: inner}

	if !errors.Is(err, ErrThreshold) {
		t.Error("errors.Is for Kind failed")
	}
	if !errors.Is(err, inner) {
		t.Error("errors.Is for Err failed")
	}

	var re *Error
	if !errors.As(err, &re) {
		t.Fatal("errors.As failed")
	}
}
