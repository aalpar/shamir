package vss

import (
	"errors"
	"math/big"
	"testing"

	"github.com/aalpar/shamir/pkg/field"
)

func TestNewGroupNotPrime(t *testing.T) {
	_, err := NewGroup(big.NewInt(15), big.NewInt(4), nil)
	if !errors.Is(err, ErrNotPrime) {
		t.Errorf("errors.Is(err, ErrNotPrime) = false; err = %v", err)
	}
}

func TestNewGroupNotSafePrime(t *testing.T) {
	// 13 is prime but not a safe prime: (13-1)/2 = 6, not prime.
	_, err := NewGroup(big.NewInt(13), big.NewInt(4), nil)
	if !errors.Is(err, ErrNotSafePrime) {
		t.Errorf("errors.Is(err, ErrNotSafePrime) = false; err = %v", err)
	}
}

func TestNewGroupGeneratorRange(t *testing.T) {
	_, err := NewGroup(testP, big.NewInt(0), nil)
	if !errors.Is(err, ErrGeneratorRange) {
		t.Errorf("errors.Is(err, ErrGeneratorRange) = false; err = %v", err)
	}
}

func TestNewGroupGeneratorOrder(t *testing.T) {
	// 2 has order 11? 2^11 mod 23 = 2048 mod 23 = 1. Actually yes.
	// Use 3: 3^11 mod 23 = 177147 mod 23 = 177147/23 = 7702*23 +1 → 1. Also order 11.
	// Use 22: 22^11 mod 23. 22 ≡ -1 mod 23. (-1)^11 = -1 ≡ 22. Not 1, so not order q.
	_, err := NewGroup(testP, big.NewInt(22), nil)
	if !errors.Is(err, ErrGeneratorOrder) {
		t.Errorf("errors.Is(err, ErrGeneratorOrder) = false; err = %v", err)
	}
}

func TestNewGroupGeneratorsEqual(t *testing.T) {
	_, err := NewGroup(testP, testG, testG)
	if !errors.Is(err, ErrGeneratorsEqual) {
		t.Errorf("errors.Is(err, ErrGeneratorsEqual) = false; err = %v", err)
	}
}

func TestFeldmanDealThreshold(t *testing.T) {
	grp := testGroup(t)
	secret := grp.Field().NewElement(big.NewInt(5))

	tests := []struct {
		name string
		n, k int
	}{
		{"k < 2", 5, 1},
		{"k > n", 3, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := FeldmanDeal(secret, tt.n, tt.k, grp)
			if !errors.Is(err, ErrThreshold) {
				t.Errorf("errors.Is(err, ErrThreshold) = false; err = %v", err)
			}
		})
	}
}

func TestPedersenDealNoSecondGenerator(t *testing.T) {
	grp, err := NewGroup(testP, testG, nil) // no h
	if err != nil {
		t.Fatalf("NewGroup: %v", err)
	}
	secret := grp.Field().NewElement(big.NewInt(5))

	_, _, err = PedersenDeal(secret, 5, 3, grp)
	if !errors.Is(err, ErrNoSecondGenerator) {
		t.Errorf("errors.Is(err, ErrNoSecondGenerator) = false; err = %v", err)
	}
}

func TestPedersenDealThreshold(t *testing.T) {
	grp := testGroup(t) // has both g and h
	secret := grp.Field().NewElement(big.NewInt(5))

	_, _, err := PedersenDeal(secret, 5, 1, grp)
	if !errors.Is(err, ErrThreshold) {
		t.Errorf("errors.Is(err, ErrThreshold) = false; err = %v", err)
	}
}

func TestErrorAs(t *testing.T) {
	grp := testGroup(t)
	secret := grp.Field().NewElement(big.NewInt(5))

	_, _, err := FeldmanDeal(secret, 5, 1, grp)
	var ve *Error
	if !errors.As(err, &ve) {
		t.Fatal("errors.As failed")
	}
	if ve.Op != "FeldmanDeal" {
		t.Errorf("Op = %q, want %q", ve.Op, "FeldmanDeal")
	}
}

func TestUnmarshalErrors(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{"bad json", `{bad`},
		{"bad prime", `{"p":"zz","g":"4","values":[]}`},
		{"bad generator", `{"p":"17","g":"zz","values":[]}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var c Commitment
			err := c.UnmarshalJSON([]byte(tt.json))
			if !errors.Is(err, ErrUnmarshal) {
				t.Errorf("errors.Is(err, ErrUnmarshal) = false; err = %v", err)
			}
		})
	}
}

func TestNoGroupError(t *testing.T) {
	c := &Commitment{Values: []*big.Int{big.NewInt(1)}}
	_, err := c.MarshalJSON()
	if !errors.Is(err, ErrNoGroup) {
		t.Errorf("errors.Is(err, ErrNoGroup) = false; err = %v", err)
	}
}

func TestPedersenShareUnmarshalError(t *testing.T) {
	var s PedersenShare
	err := s.UnmarshalJSON([]byte(`{"q":"zz","x":"1","y":"2","t":"3"}`))
	if !errors.Is(err, ErrUnmarshal) {
		t.Errorf("errors.Is(err, ErrUnmarshal) = false; err = %v", err)
	}
}

func TestFieldRandErrorPropagates(t *testing.T) {
	// Verify that a wrapped field error is accessible through the chain.
	inner := &field.Error{Op: "Rand", Kind: field.ErrRandom}
	err := &Error{Op: "FeldmanDeal", Err: inner}

	if !errors.Is(err, field.ErrRandom) {
		t.Error("field.ErrRandom not reachable through vss.Error chain")
	}
}
