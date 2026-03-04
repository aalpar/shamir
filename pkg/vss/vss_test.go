// Copyright 2026 Aaron Alpar
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package vss

import (
	"math/big"
	"testing"

	"github.com/aalpar/shamir/pkg/sss"
)

// Safe prime p = 23: q = 11, both prime. p = 2*11 + 1.
// Generator g = 2: ord(2) = 11 in Z_23*.
// Generator h = 3: ord(3) = 11 in Z_23*.
var (
	testP = big.NewInt(23)
	testG = big.NewInt(2)
	testH = big.NewInt(3)
)

// Larger safe prime for realistic tests: p = 2579, q = 1289.
var (
	largeP = big.NewInt(2579)
	largeG = big.NewInt(2) // validated in test
)

func testGroup(t *testing.T) *Group {
	t.Helper()
	grp, err := NewGroup(testP, testG, testH)
	if err != nil {
		t.Fatalf("NewGroup() error: %v", err)
	}
	return grp
}

// --- Group tests ---

func TestNewGroup(t *testing.T) {
	grp, err := NewGroup(testP, testG, testH)
	if err != nil {
		t.Fatalf("NewGroup() error: %v", err)
	}
	if grp.Q().Cmp(big.NewInt(11)) != 0 {
		t.Errorf("Q() = %s, want 11", grp.Q())
	}
}

func TestNewGroupFeldmanOnly(t *testing.T) {
	// h = nil is valid for Feldman-only usage.
	grp, err := NewGroup(testP, testG, nil)
	if err != nil {
		t.Fatalf("NewGroup(h=nil) error: %v", err)
	}
	if grp.h != nil {
		t.Error("expected h to be nil")
	}
}

func TestNewGroupValidation(t *testing.T) {
	tests := []struct {
		name    string
		p, g, h *big.Int
	}{
		{"composite p", big.NewInt(21), testG, nil},
		{"non-safe prime", big.NewInt(13), testG, nil}, // q = 6, not prime
		{"g = 1", testP, big.NewInt(1), nil},
		{"g out of range", testP, big.NewInt(23), nil},
		{"g wrong order", testP, big.NewInt(22), nil}, // ord(22) = 2, not 11
		{"g = h", testP, testG, testG},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewGroup(tt.p, tt.g, tt.h)
			if err == nil {
				t.Error("expected error")
			}
		})
	}
}

// --- Feldman tests ---

func TestFeldmanDealVerify(t *testing.T) {
	grp := testGroup(t)
	secret := grp.Field().NewElement(big.NewInt(7))

	shares, commitment, err := FeldmanDeal(secret, 5, 3, grp)
	if err != nil {
		t.Fatalf("FeldmanDeal() error: %v", err)
	}

	if len(shares) != 5 {
		t.Fatalf("got %d shares, want 5", len(shares))
	}

	if len(commitment.Values) != 3 {
		t.Fatalf("got %d commitments, want 3 (degree k-1 = 2)", len(commitment.Values))
	}

	// Every share must verify.
	for i, share := range shares {
		if !FeldmanVerify(share, commitment, grp) {
			t.Errorf("share %d failed verification", i)
		}
	}
}

func TestFeldmanVerifyTampered(t *testing.T) {
	grp := testGroup(t)
	secret := grp.Field().NewElement(big.NewInt(5))

	shares, commitment, err := FeldmanDeal(secret, 5, 3, grp)
	if err != nil {
		t.Fatalf("FeldmanDeal() error: %v", err)
	}

	// Tamper with share 0: change Y value.
	tampered := sss.Share{
		X: shares[0].X,
		Y: shares[0].Y.Add(grp.Field().One()),
	}
	if FeldmanVerify(tampered, commitment, grp) {
		t.Error("tampered share passed verification")
	}
}

func TestFeldmanReconstruct(t *testing.T) {
	grp := testGroup(t)
	secret := grp.Field().NewElement(big.NewInt(9))

	shares, _, err := FeldmanDeal(secret, 5, 3, grp)
	if err != nil {
		t.Fatalf("FeldmanDeal() error: %v", err)
	}

	// Shares from Feldman are standard SSS shares — Combine should work.
	got, err := sss.Combine(shares[:3])
	if err != nil {
		t.Fatalf("Combine() error: %v", err)
	}
	if !got.Equal(secret) {
		t.Errorf("Combine() = %s, want %s", got, secret)
	}
}

func TestFeldmanAllSubsets(t *testing.T) {
	grp := testGroup(t)
	secret := grp.Field().NewElement(big.NewInt(4))

	shares, commitment, err := FeldmanDeal(secret, 4, 3, grp)
	if err != nil {
		t.Fatalf("FeldmanDeal() error: %v", err)
	}

	// All C(4,3) = 4 subsets must verify and reconstruct.
	for i := 0; i < 4; i++ {
		for j := i + 1; j < 4; j++ {
			for k := j + 1; k < 4; k++ {
				subset := []sss.Share{shares[i], shares[j], shares[k]}

				for _, s := range subset {
					if !FeldmanVerify(s, commitment, grp) {
						t.Errorf("subset {%d,%d,%d}: verification failed", i, j, k)
					}
				}

				got, err := sss.Combine(subset)
				if err != nil {
					t.Fatalf("subset {%d,%d,%d}: Combine error: %v", i, j, k, err)
				}
				if !got.Equal(secret) {
					t.Errorf("subset {%d,%d,%d}: got %s, want %s", i, j, k, got, secret)
				}
			}
		}
	}
}

func TestFeldmanValidation(t *testing.T) {
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
			if err == nil {
				t.Error("expected error")
			}
		})
	}
}

// --- Pedersen tests ---

func TestPedersenDealVerify(t *testing.T) {
	grp := testGroup(t)
	secret := grp.Field().NewElement(big.NewInt(7))

	shares, commitment, err := PedersenDeal(secret, 5, 3, grp)
	if err != nil {
		t.Fatalf("PedersenDeal() error: %v", err)
	}

	if len(shares) != 5 {
		t.Fatalf("got %d shares, want 5", len(shares))
	}

	// Every share must verify.
	for i, share := range shares {
		if !PedersenVerify(share, commitment, grp) {
			t.Errorf("share %d failed verification", i)
		}
	}
}

func TestPedersenVerifyTamperedY(t *testing.T) {
	grp := testGroup(t)
	secret := grp.Field().NewElement(big.NewInt(5))

	shares, commitment, err := PedersenDeal(secret, 5, 3, grp)
	if err != nil {
		t.Fatalf("PedersenDeal() error: %v", err)
	}

	// Tamper with Y component.
	tampered := PedersenShare{
		X: shares[0].X,
		Y: shares[0].Y.Add(grp.Field().One()),
		T: shares[0].T,
	}
	if PedersenVerify(tampered, commitment, grp) {
		t.Error("tampered Y passed verification")
	}
}

func TestPedersenVerifyTamperedT(t *testing.T) {
	grp := testGroup(t)
	secret := grp.Field().NewElement(big.NewInt(5))

	shares, commitment, err := PedersenDeal(secret, 5, 3, grp)
	if err != nil {
		t.Fatalf("PedersenDeal() error: %v", err)
	}

	// Tamper with T (blinding) component.
	tampered := PedersenShare{
		X: shares[0].X,
		Y: shares[0].Y,
		T: shares[0].T.Add(grp.Field().One()),
	}
	if PedersenVerify(tampered, commitment, grp) {
		t.Error("tampered T passed verification")
	}
}

func TestPedersenReconstruct(t *testing.T) {
	grp := testGroup(t)
	secret := grp.Field().NewElement(big.NewInt(3))

	shares, _, err := PedersenDeal(secret, 5, 3, grp)
	if err != nil {
		t.Fatalf("PedersenDeal() error: %v", err)
	}

	// Extract SSS shares and reconstruct.
	sssShares := PedersenToSSS(shares)
	got, err := sss.Combine(sssShares[:3])
	if err != nil {
		t.Fatalf("Combine() error: %v", err)
	}
	if !got.Equal(secret) {
		t.Errorf("Combine() = %s, want %s", got, secret)
	}
}

func TestPedersenAllSubsets(t *testing.T) {
	grp := testGroup(t)
	secret := grp.Field().NewElement(big.NewInt(10))

	shares, commitment, err := PedersenDeal(secret, 4, 3, grp)
	if err != nil {
		t.Fatalf("PedersenDeal() error: %v", err)
	}

	// All C(4,3) = 4 subsets must verify and reconstruct.
	for i := 0; i < 4; i++ {
		for j := i + 1; j < 4; j++ {
			for k := j + 1; k < 4; k++ {
				subset := []PedersenShare{shares[i], shares[j], shares[k]}

				for _, s := range subset {
					if !PedersenVerify(s, commitment, grp) {
						t.Errorf("subset {%d,%d,%d}: verification failed", i, j, k)
					}
				}

				sssSubset := PedersenToSSS(subset)
				got, err := sss.Combine(sssSubset)
				if err != nil {
					t.Fatalf("subset {%d,%d,%d}: Combine error: %v", i, j, k, err)
				}
				if !got.Equal(secret) {
					t.Errorf("subset {%d,%d,%d}: got %s, want %s", i, j, k, got, secret)
				}
			}
		}
	}
}

func TestPedersenRequiresH(t *testing.T) {
	grp, err := NewGroup(testP, testG, nil)
	if err != nil {
		t.Fatalf("NewGroup() error: %v", err)
	}

	secret := grp.Field().NewElement(big.NewInt(5))
	_, _, err = PedersenDeal(secret, 5, 3, grp)
	if err == nil {
		t.Error("PedersenDeal should require h")
	}
}

func TestPedersenValidation(t *testing.T) {
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
			_, _, err := PedersenDeal(secret, tt.n, tt.k, grp)
			if err == nil {
				t.Error("expected error")
			}
		})
	}
}

func TestPedersenToSSS(t *testing.T) {
	grp := testGroup(t)
	secret := grp.Field().NewElement(big.NewInt(8))

	shares, _, err := PedersenDeal(secret, 5, 3, grp)
	if err != nil {
		t.Fatalf("PedersenDeal() error: %v", err)
	}

	sssShares := PedersenToSSS(shares)
	if len(sssShares) != 5 {
		t.Fatalf("PedersenToSSS() returned %d, want 5", len(sssShares))
	}

	// X and Y must match.
	for i := range shares {
		if !sssShares[i].X.Equal(shares[i].X) || !sssShares[i].Y.Equal(shares[i].Y) {
			t.Errorf("share %d: SSS share doesn't match Pedersen share", i)
		}
	}
}

// --- Randomized property tests ---

func TestFeldmanRoundtripRandom(t *testing.T) {
	grp := testGroup(t)

	for range 20 {
		secret, err := grp.Field().Rand()
		if err != nil {
			t.Fatalf("Rand() error: %v", err)
		}

		shares, commitment, err := FeldmanDeal(secret, 5, 3, grp)
		if err != nil {
			t.Fatalf("FeldmanDeal() error: %v", err)
		}

		for i, s := range shares {
			if !FeldmanVerify(s, commitment, grp) {
				t.Errorf("share %d failed verification", i)
			}
		}

		got, err := sss.Combine(shares[:3])
		if err != nil {
			t.Fatalf("Combine() error: %v", err)
		}
		if !got.Equal(secret) {
			t.Errorf("roundtrip failed: got %s, want %s", got, secret)
		}
	}
}

func TestPedersenRoundtripRandom(t *testing.T) {
	grp := testGroup(t)

	for range 20 {
		secret, err := grp.Field().Rand()
		if err != nil {
			t.Fatalf("Rand() error: %v", err)
		}

		shares, commitment, err := PedersenDeal(secret, 5, 3, grp)
		if err != nil {
			t.Fatalf("PedersenDeal() error: %v", err)
		}

		for i, s := range shares {
			if !PedersenVerify(s, commitment, grp) {
				t.Errorf("share %d failed verification", i)
			}
		}

		sssShares := PedersenToSSS(shares[:3])
		got, err := sss.Combine(sssShares)
		if err != nil {
			t.Fatalf("Combine() error: %v", err)
		}
		if !got.Equal(secret) {
			t.Errorf("roundtrip failed: got %s, want %s", got, secret)
		}
	}
}
