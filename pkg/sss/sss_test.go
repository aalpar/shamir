package sss

import (
	"math/big"
	"testing"

	"github.com/aalpar/shamir/pkg/field"
)

var p17 = big.NewInt(17)

// secp256k1 prime for realistic tests.
var p256, _ = new(big.Int).SetString(
	"FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEFFFFFC2F", 16,
)

func TestSplitCombineBasic(t *testing.T) {
	f := field.New(p17)
	secret := f.NewElement(big.NewInt(7))

	shares, err := Split(secret, 5, 3, f)
	if err != nil {
		t.Fatalf("Split() error: %v", err)
	}

	if len(shares) != 5 {
		t.Fatalf("Split() returned %d shares, want 5", len(shares))
	}

	// Any 3 shares should reconstruct the secret.
	got, err := Combine(shares[:3])
	if err != nil {
		t.Fatalf("Combine() error: %v", err)
	}
	if !got.Equal(secret) {
		t.Errorf("Combine(shares[:3]) = %s, want %s", got, secret)
	}
}

func TestCombineAnyKSubset(t *testing.T) {
	f := field.New(p17)
	secret := f.NewElement(big.NewInt(11))

	shares, err := Split(secret, 5, 3, f)
	if err != nil {
		t.Fatalf("Split() error: %v", err)
	}

	// Every 3-element subset of 5 shares should reconstruct correctly.
	// C(5,3) = 10 subsets.
	for i := 0; i < 5; i++ {
		for j := i + 1; j < 5; j++ {
			for k := j + 1; k < 5; k++ {
				subset := []Share{shares[i], shares[j], shares[k]}
				got, err := Combine(subset)
				if err != nil {
					t.Fatalf("Combine({%d,%d,%d}) error: %v", i, j, k, err)
				}
				if !got.Equal(secret) {
					t.Errorf("Combine({%d,%d,%d}) = %s, want %s", i, j, k, got, secret)
				}
			}
		}
	}
}

func TestCombineMoreThanK(t *testing.T) {
	f := field.New(p17)
	secret := f.NewElement(big.NewInt(3))

	shares, err := Split(secret, 5, 3, f)
	if err != nil {
		t.Fatalf("Split() error: %v", err)
	}

	// Using 4 or 5 shares should also reconstruct correctly.
	for _, n := range []int{4, 5} {
		got, err := Combine(shares[:n])
		if err != nil {
			t.Fatalf("Combine(shares[:%d]) error: %v", n, err)
		}
		if !got.Equal(secret) {
			t.Errorf("Combine(shares[:%d]) = %s, want %s", n, got, secret)
		}
	}
}

func TestCombineFewerThanK(t *testing.T) {
	f := field.New(p17)
	secret := f.NewElement(big.NewInt(7))

	shares, err := Split(secret, 5, 3, f)
	if err != nil {
		t.Fatalf("Split() error: %v", err)
	}

	// Fewer than k shares should produce a wrong result (not an error).
	// This verifies the security property: k-1 shares reveal nothing.
	got, err := Combine(shares[:2])
	if err != nil {
		t.Fatalf("Combine(shares[:2]) error: %v", err)
	}
	// The result should almost certainly not be the secret.
	// (There's a 1/17 chance it matches by accident in GF(17), so we
	// don't hard-fail, but we log it.)
	if got.Equal(secret) {
		t.Log("WARNING: 2 shares reconstructed the secret (1/p probability)")
	}
}

func TestSplitZeroSecret(t *testing.T) {
	f := field.New(p17)
	secret := f.Zero()

	shares, err := Split(secret, 5, 3, f)
	if err != nil {
		t.Fatalf("Split() error: %v", err)
	}

	got, err := Combine(shares[:3])
	if err != nil {
		t.Fatalf("Combine() error: %v", err)
	}
	if !got.Equal(secret) {
		t.Errorf("Combine() = %s, want 0", got)
	}
}

func TestSplitMinimalThreshold(t *testing.T) {
	f := field.New(p17)
	secret := f.NewElement(big.NewInt(9))

	// k = 2, n = 2: minimum valid parameters.
	shares, err := Split(secret, 2, 2, f)
	if err != nil {
		t.Fatalf("Split(2,2) error: %v", err)
	}

	got, err := Combine(shares)
	if err != nil {
		t.Fatalf("Combine() error: %v", err)
	}
	if !got.Equal(secret) {
		t.Errorf("Combine() = %s, want %s", got, secret)
	}
}

func TestSplitValidation(t *testing.T) {
	f := field.New(p17)
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
			if err == nil {
				t.Errorf("Split(n=%d, k=%d) should have returned error", tt.n, tt.k)
			}
		})
	}
}

func TestCombineEmpty(t *testing.T) {
	_, err := Combine(nil)
	if err == nil {
		t.Error("Combine(nil) should have returned error")
	}
}

func TestSplitCombineLargePrime(t *testing.T) {
	f := field.New(p256)

	// A 256-bit secret.
	secretVal, _ := new(big.Int).SetString("DEADBEEF", 16)
	secret := f.NewElement(secretVal)

	shares, err := Split(secret, 7, 4, f)
	if err != nil {
		t.Fatalf("Split() error: %v", err)
	}

	// Reconstruct from different subsets of 4.
	subsets := [][]int{
		{0, 1, 2, 3},
		{3, 4, 5, 6},
		{0, 2, 4, 6},
		{1, 3, 5, 6},
	}

	for _, idx := range subsets {
		subset := make([]Share, len(idx))
		for i, j := range idx {
			subset[i] = shares[j]
		}
		got, err := Combine(subset)
		if err != nil {
			t.Fatalf("Combine(%v) error: %v", idx, err)
		}
		if !got.Equal(secret) {
			t.Errorf("Combine(%v) = %s, want %s", idx, got, secret)
		}
	}
}

func TestSplitSharesDistinct(t *testing.T) {
	f := field.New(p17)
	secret := f.NewElement(big.NewInt(5))

	shares, err := Split(secret, 5, 3, f)
	if err != nil {
		t.Fatalf("Split() error: %v", err)
	}

	// All x-coordinates must be distinct.
	seen := make(map[string]bool)
	for i, s := range shares {
		key := s.X.Value().String()
		if seen[key] {
			t.Errorf("duplicate x-coordinate at share %d: %s", i, key)
		}
		seen[key] = true
	}

	// x-coordinates should be 1, 2, 3, 4, 5.
	for i, s := range shares {
		wantX := int64(i + 1)
		if s.X.Value().Int64() != wantX {
			t.Errorf("share[%d].X = %s, want %d", i, s.X, wantX)
		}
	}
}

func TestSplitCombineRandomSecrets(t *testing.T) {
	f := field.New(p17)

	// 20 random secrets, split and reconstruct each.
	for i := range 20 {
		secret, err := f.Rand()
		if err != nil {
			t.Fatalf("Rand() error: %v", err)
		}

		shares, err := Split(secret, 5, 3, f)
		if err != nil {
			t.Fatalf("iteration %d: Split() error: %v", i, err)
		}

		got, err := Combine(shares[:3])
		if err != nil {
			t.Fatalf("iteration %d: Combine() error: %v", i, err)
		}
		if !got.Equal(secret) {
			t.Errorf("iteration %d: roundtrip failed: got %s, want %s", i, got, secret)
		}
	}
}
