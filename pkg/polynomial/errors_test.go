package polynomial

import (
	"errors"
	"math/big"
	"testing"

	"github.com/aalpar/shamir/pkg/field"
)

func TestRandomNegativeDegreeErrorType(t *testing.T) {
	f := field.New(big.NewInt(17))
	_, err := Random(-1, f.Zero(), f)
	if !errors.Is(err, ErrNegativeDegree) {
		t.Errorf("errors.Is(err, ErrNegativeDegree) = false; err = %v", err)
	}
	var pe *Error
	if !errors.As(err, &pe) {
		t.Fatal("errors.As failed")
	}
	if pe.Op != "Random" {
		t.Errorf("Op = %q, want %q", pe.Op, "Random")
	}
}

func TestLagrangeZeroNoPointsErrorType(t *testing.T) {
	_, err := LagrangeZero(nil)
	if !errors.Is(err, ErrNoPoints) {
		t.Errorf("errors.Is(err, ErrNoPoints) = false; err = %v", err)
	}
}

func TestLagrangeZeroDuplicateXErrorType(t *testing.T) {
	f := field.New(big.NewInt(17))
	x := f.NewElement(big.NewInt(3))
	points := []Point{
		{X: x, Y: f.NewElement(big.NewInt(5))},
		{X: x, Y: f.NewElement(big.NewInt(7))},
	}
	_, err := LagrangeZero(points)
	if !errors.Is(err, ErrDuplicateX) {
		t.Errorf("errors.Is(err, ErrDuplicateX) = false; err = %v", err)
	}
}

func TestErrorUnwrapChain(t *testing.T) {
	inner := errors.New("source failed")
	err := &Error{Op: "Random", Kind: ErrNegativeDegree, Err: inner}

	if !errors.Is(err, ErrNegativeDegree) {
		t.Error("errors.Is for Kind failed")
	}
	if !errors.Is(err, inner) {
		t.Error("errors.Is for Err failed")
	}
}
