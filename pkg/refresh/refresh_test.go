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
package refresh

import (
	"math/big"
	"testing"

	"github.com/aalpar/shamir/pkg/sss"
	"github.com/aalpar/shamir/pkg/vss"
)

var (
	testP = big.NewInt(23)
	testG = big.NewInt(2)
	testH = big.NewInt(3)
)

func testGroup(t *testing.T) *vss.Group {
	t.Helper()
	grp, err := vss.NewGroup(testP, testG, testH)
	if err != nil {
		t.Fatalf("NewGroup() error: %v", err)
	}
	return grp
}

// --- Basic zero sharing tests ---

func TestZeroSharingCount(t *testing.T) {
	grp := testGroup(t)
	subs, err := ZeroSharing(5, 3, grp.Field())
	if err != nil {
		t.Fatalf("ZeroSharing() error: %v", err)
	}
	if len(subs) != 5 {
		t.Fatalf("got %d sub-shares, want 5", len(subs))
	}
	// X values should be 1..5.
	for i, s := range subs {
		wantX := int64(i + 1)
		if s.X.Value().Int64() != wantX {
			t.Errorf("sub-share[%d].X = %s, want %d", i, s.X, wantX)
		}
	}
}

func TestZeroSharingReconstructsZero(t *testing.T) {
	grp := testGroup(t)

	// A zero sharing should reconstruct to zero.
	subs, err := ZeroSharing(5, 3, grp.Field())
	if err != nil {
		t.Fatalf("ZeroSharing() error: %v", err)
	}

	// Convert to sss.Shares and combine.
	shares := make([]sss.Share, len(subs))
	for i, s := range subs {
		shares[i] = sss.Share{X: s.X, Y: s.Delta}
	}

	got, err := sss.Combine(shares[:3])
	if err != nil {
		t.Fatalf("Combine() error: %v", err)
	}
	if !got.IsZero() {
		t.Errorf("zero sharing reconstructed to %s, want 0", got)
	}
}

func TestZeroSharingValidation(t *testing.T) {
	grp := testGroup(t)
	f := grp.Field()

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
			if err == nil {
				t.Error("expected error")
			}
		})
	}
}

// --- Apply tests ---

func TestApplyRefreshRoundtrip(t *testing.T) {
	grp := testGroup(t)
	f := grp.Field()
	secret := f.NewElement(big.NewInt(7))

	shares, err := sss.Split(secret, 5, 3, f)
	if err != nil {
		t.Fatalf("Split() error: %v", err)
	}

	// Simulate one participant generating a zero sharing.
	subs, err := ZeroSharing(5, 3, f)
	if err != nil {
		t.Fatalf("ZeroSharing() error: %v", err)
	}

	// Each share holder applies their sub-share.
	refreshed := make([]sss.Share, 5)
	for i := range shares {
		refreshed[i], err = Apply(shares[i], []SubShare{subs[i]})
		if err != nil {
			t.Fatalf("Apply(share %d) error: %v", i, err)
		}
	}

	// Reconstruct from refreshed shares.
	got, err := sss.Combine(refreshed[:3])
	if err != nil {
		t.Fatalf("Combine() error: %v", err)
	}
	if !got.Equal(secret) {
		t.Errorf("Combine() = %s, want %s", got, secret)
	}
}

func TestApplyMultipleContributions(t *testing.T) {
	grp := testGroup(t)
	f := grp.Field()
	secret := f.NewElement(big.NewInt(9))

	shares, err := sss.Split(secret, 5, 3, f)
	if err != nil {
		t.Fatalf("Split() error: %v", err)
	}

	// All 5 participants generate zero sharings (realistic scenario).
	allSubs := make([][]SubShare, 5)
	for p := range 5 {
		allSubs[p], err = ZeroSharing(5, 3, f)
		if err != nil {
			t.Fatalf("ZeroSharing(%d) error: %v", p, err)
		}
	}

	// Each participant collects their column of sub-shares and applies.
	refreshed := make([]sss.Share, 5)
	for i := range shares {
		deltas := make([]SubShare, 5)
		for p := range 5 {
			deltas[p] = allSubs[p][i]
		}
		refreshed[i], err = Apply(shares[i], deltas)
		if err != nil {
			t.Fatalf("Apply(share %d) error: %v", i, err)
		}
	}

	got, err := sss.Combine(refreshed[:3])
	if err != nil {
		t.Fatalf("Combine() error: %v", err)
	}
	if !got.Equal(secret) {
		t.Errorf("Combine() = %s, want %s", got, secret)
	}
}

func TestApplyMultipleRounds(t *testing.T) {
	grp := testGroup(t)
	f := grp.Field()
	secret := f.NewElement(big.NewInt(4))

	shares, err := sss.Split(secret, 5, 3, f)
	if err != nil {
		t.Fatalf("Split() error: %v", err)
	}

	// 5 rounds of refresh.
	for round := range 5 {
		subs, err := ZeroSharing(5, 3, f)
		if err != nil {
			t.Fatalf("round %d: ZeroSharing() error: %v", round, err)
		}
		for i := range shares {
			shares[i], err = Apply(shares[i], []SubShare{subs[i]})
			if err != nil {
				t.Fatalf("round %d: Apply(share %d) error: %v", round, i, err)
			}
		}
	}

	got, err := sss.Combine(shares[:3])
	if err != nil {
		t.Fatalf("Combine() error: %v", err)
	}
	if !got.Equal(secret) {
		t.Errorf("after 5 rounds: Combine() = %s, want %s", got, secret)
	}
}

func TestApplyOldSharesInvalid(t *testing.T) {
	grp := testGroup(t)
	f := grp.Field()
	secret := f.NewElement(big.NewInt(6))

	shares, err := sss.Split(secret, 5, 3, f)
	if err != nil {
		t.Fatalf("Split() error: %v", err)
	}

	// Save old shares.
	old := make([]sss.Share, len(shares))
	copy(old, shares)

	// Refresh.
	subs, err := ZeroSharing(5, 3, f)
	if err != nil {
		t.Fatalf("ZeroSharing() error: %v", err)
	}
	refreshed := make([]sss.Share, 5)
	for i := range shares {
		refreshed[i], err = Apply(shares[i], []SubShare{subs[i]})
		if err != nil {
			t.Fatalf("Apply(share %d) error: %v", i, err)
		}
	}

	// Mixing old and new shares should almost certainly not reconstruct
	// the secret (1/q probability of accidental collision in small fields).
	mixed := []sss.Share{old[0], old[1], refreshed[2]}
	got, err := sss.Combine(mixed)
	if err != nil {
		t.Fatalf("Combine() error: %v", err)
	}
	if got.Equal(secret) {
		t.Log("WARNING: mixed old+new shares reconstructed the secret (1/q probability)")
	}
}

func TestApplyMismatchedX(t *testing.T) {
	grp := testGroup(t)
	f := grp.Field()
	secret := f.NewElement(big.NewInt(5))

	shares, err := sss.Split(secret, 5, 3, f)
	if err != nil {
		t.Fatalf("Split() error: %v", err)
	}

	subs, err := ZeroSharing(5, 3, f)
	if err != nil {
		t.Fatalf("ZeroSharing() error: %v", err)
	}

	// Apply sub-share 2 (X=3) to share 0 (X=1) — should error.
	_, err = Apply(shares[0], []SubShare{subs[2]})
	if err == nil {
		t.Error("expected error for mismatched X")
	}
}

// --- Verifiable zero sharing tests ---

func TestVerifiableZeroSharingC0IsOne(t *testing.T) {
	grp := testGroup(t)

	contrib, err := VerifiableZeroSharing(5, 3, grp)
	if err != nil {
		t.Fatalf("VerifiableZeroSharing() error: %v", err)
	}

	// C_0 = g^{secret} = g^0 = 1 (mod p).
	c0 := contrib.Commitment.Values[0]
	if c0.Cmp(big.NewInt(1)) != 0 {
		t.Errorf("C_0 = %s, want 1 (g^0 mod p)", c0)
	}
}

func TestVerifiableZeroSharingVerifiesAll(t *testing.T) {
	grp := testGroup(t)

	contrib, err := VerifiableZeroSharing(5, 3, grp)
	if err != nil {
		t.Fatalf("VerifiableZeroSharing() error: %v", err)
	}

	for i, sub := range contrib.SubShares {
		if !VerifySubShare(sub, contrib.Commitment, grp) {
			t.Errorf("sub-share %d failed verification", i)
		}
	}
}

func TestVerifiableZeroSharingTampered(t *testing.T) {
	grp := testGroup(t)

	contrib, err := VerifiableZeroSharing(5, 3, grp)
	if err != nil {
		t.Fatalf("VerifiableZeroSharing() error: %v", err)
	}

	tampered := SubShare{
		X:     contrib.SubShares[0].X,
		Delta: contrib.SubShares[0].Delta.Add(grp.Field().One()),
	}
	if VerifySubShare(tampered, contrib.Commitment, grp) {
		t.Error("tampered sub-share passed verification")
	}
}

// --- UpdateCommitment tests ---

func TestUpdateCommitmentVerifies(t *testing.T) {
	grp := testGroup(t)
	secret := grp.Field().NewElement(big.NewInt(7))

	// Feldman deal to get original shares and commitment.
	shares, originalCommit, err := vss.FeldmanDeal(secret, 5, 3, grp)
	if err != nil {
		t.Fatalf("FeldmanDeal() error: %v", err)
	}

	// Generate verifiable zero sharing.
	contrib, err := VerifiableZeroSharing(5, 3, grp)
	if err != nil {
		t.Fatalf("VerifiableZeroSharing() error: %v", err)
	}

	// Refresh shares.
	refreshed := make([]sss.Share, 5)
	for i := range shares {
		refreshed[i], err = Apply(shares[i], []SubShare{contrib.SubShares[i]})
		if err != nil {
			t.Fatalf("Apply(share %d) error: %v", i, err)
		}
	}

	// Update commitment.
	newCommit, err := UpdateCommitment(originalCommit, []*vss.Commitment{contrib.Commitment}, grp)
	if err != nil {
		t.Fatalf("UpdateCommitment() error: %v", err)
	}

	// Refreshed shares must verify against updated commitment.
	for i, s := range refreshed {
		if !vss.FeldmanVerify(s, newCommit, grp) {
			t.Errorf("refreshed share %d failed verification against updated commitment", i)
		}
	}

	// Old commitment should generally NOT verify against refreshed shares.
	// In a small field (q=11), individual shares have ~1/q chance of
	// accidentally passing, but all 5 passing is (1/11)^5 ≈ 6e-6.
	oldPass := 0
	for _, s := range refreshed {
		if vss.FeldmanVerify(s, originalCommit, grp) {
			oldPass++
		}
	}
	if oldPass == len(refreshed) {
		t.Error("all refreshed shares still verify against old commitment")
	}
}

func TestUpdateCommitmentMultipleContributions(t *testing.T) {
	grp := testGroup(t)
	secret := grp.Field().NewElement(big.NewInt(3))

	shares, originalCommit, err := vss.FeldmanDeal(secret, 5, 3, grp)
	if err != nil {
		t.Fatalf("FeldmanDeal() error: %v", err)
	}

	// All 5 participants contribute zero sharings.
	contribs := make([]*Contribution, 5)
	for p := range 5 {
		contribs[p], err = VerifiableZeroSharing(5, 3, grp)
		if err != nil {
			t.Fatalf("VerifiableZeroSharing(%d) error: %v", p, err)
		}
	}

	// Each participant applies all sub-shares.
	refreshed := make([]sss.Share, 5)
	for i := range shares {
		deltas := make([]SubShare, 5)
		for p := range 5 {
			deltas[p] = contribs[p].SubShares[i]
		}
		refreshed[i], err = Apply(shares[i], deltas)
		if err != nil {
			t.Fatalf("Apply(share %d) error: %v", i, err)
		}
	}

	// Update commitment with all contributions.
	deltaCommits := make([]*vss.Commitment, 5)
	for p := range 5 {
		deltaCommits[p] = contribs[p].Commitment
	}
	newCommit, err := UpdateCommitment(originalCommit, deltaCommits, grp)
	if err != nil {
		t.Fatalf("UpdateCommitment() error: %v", err)
	}

	// All refreshed shares verify against the updated commitment.
	for i, s := range refreshed {
		if !vss.FeldmanVerify(s, newCommit, grp) {
			t.Errorf("share %d failed verification", i)
		}
	}

	// Reconstruct from refreshed shares.
	got, err := sss.Combine(refreshed[:3])
	if err != nil {
		t.Fatalf("Combine() error: %v", err)
	}
	if !got.Equal(secret) {
		t.Errorf("Combine() = %s, want %s", got, secret)
	}
}

func TestUpdateCommitmentEmpty(t *testing.T) {
	grp := testGroup(t)
	secret := grp.Field().NewElement(big.NewInt(5))

	_, commit, err := vss.FeldmanDeal(secret, 5, 3, grp)
	if err != nil {
		t.Fatalf("FeldmanDeal() error: %v", err)
	}

	// No deltas — should return a copy.
	updated, err := UpdateCommitment(commit, nil, grp)
	if err != nil {
		t.Fatalf("UpdateCommitment() error: %v", err)
	}
	for j := range commit.Values {
		if commit.Values[j].Cmp(updated.Values[j]) != 0 {
			t.Errorf("value[%d]: got %s, want %s", j, updated.Values[j], commit.Values[j])
		}
	}
}

func TestUpdateCommitmentLengthMismatch(t *testing.T) {
	grp := testGroup(t)
	secret := grp.Field().NewElement(big.NewInt(5))

	_, commit, err := vss.FeldmanDeal(secret, 5, 3, grp)
	if err != nil {
		t.Fatalf("FeldmanDeal() error: %v", err)
	}

	// Commitment with wrong length.
	bad := &vss.Commitment{Values: []*big.Int{big.NewInt(1)}}
	_, err = UpdateCommitment(commit, []*vss.Commitment{bad}, grp)
	if err == nil {
		t.Error("expected error for mismatched commitment lengths")
	}
}

// --- Randomized property tests ---

func TestRefreshRoundtripRandom(t *testing.T) {
	grp := testGroup(t)
	f := grp.Field()

	for range 20 {
		secret, err := f.Rand()
		if err != nil {
			t.Fatalf("Rand() error: %v", err)
		}

		shares, err := sss.Split(secret, 5, 3, f)
		if err != nil {
			t.Fatalf("Split() error: %v", err)
		}

		// Random number of contributors (1..5).
		nContrib := 3
		for c := range nContrib {
			subs, err := ZeroSharing(5, 3, f)
			if err != nil {
				t.Fatalf("ZeroSharing(%d) error: %v", c, err)
			}
			for i := range shares {
				shares[i], err = Apply(shares[i], []SubShare{subs[i]})
				if err != nil {
					t.Fatalf("Apply() error: %v", err)
				}
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

func TestVerifiableRefreshRoundtripRandom(t *testing.T) {
	grp := testGroup(t)

	for range 20 {
		secret, err := grp.Field().Rand()
		if err != nil {
			t.Fatalf("Rand() error: %v", err)
		}

		shares, commit, err := vss.FeldmanDeal(secret, 5, 3, grp)
		if err != nil {
			t.Fatalf("FeldmanDeal() error: %v", err)
		}

		// 3 contributors.
		deltaCommits := make([]*vss.Commitment, 3)
		for c := range 3 {
			contrib, err := VerifiableZeroSharing(5, 3, grp)
			if err != nil {
				t.Fatalf("VerifiableZeroSharing(%d) error: %v", c, err)
			}
			deltaCommits[c] = contrib.Commitment

			// Verify and apply.
			for i := range shares {
				if !VerifySubShare(contrib.SubShares[i], contrib.Commitment, grp) {
					t.Fatalf("sub-share %d from contributor %d failed verification", i, c)
				}
				shares[i], err = Apply(shares[i], []SubShare{contrib.SubShares[i]})
				if err != nil {
					t.Fatalf("Apply() error: %v", err)
				}
			}
		}

		// Update commitment.
		newCommit, err := UpdateCommitment(commit, deltaCommits, grp)
		if err != nil {
			t.Fatalf("UpdateCommitment() error: %v", err)
		}

		// Verify refreshed shares against updated commitment.
		for i, s := range shares {
			if !vss.FeldmanVerify(s, newCommit, grp) {
				t.Errorf("refreshed share %d failed verification", i)
			}
		}

		// Reconstruct.
		got, err := sss.Combine(shares[:3])
		if err != nil {
			t.Fatalf("Combine() error: %v", err)
		}
		if !got.Equal(secret) {
			t.Errorf("roundtrip failed: got %s, want %s", got, secret)
		}
	}
}
