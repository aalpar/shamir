// Copyright 2026 Aaron Alpar
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Package refresh implements proactive share refresh for secret sharing.
//
// Proactive refresh allows shareholders to generate new shares of the
// same secret, rendering previously compromised shares useless. Each
// participant generates a random "zero sharing" — shares of the zero
// polynomial — and adds sub-shares to their existing share.
//
// After refresh, all shares are algebraically different but still
// reconstruct the same secret. The security model shifts from "never
// compromise k shares" to "never compromise k shares within a single
// refresh period."
//
// Verification uses Feldman commitments: each participant publishes
// commitments for their zero polynomial, allowing others to verify
// sub-shares before applying them. The commitment C_0 = g^0 = 1 is
// a built-in correctness check (the "secret" of a zero sharing is zero).
//
// After refresh, the old Feldman commitments no longer verify against
// the new shares. Use UpdateCommitment to combine the original commitment
// with all zero-sharing commitments.
//
// See: BIBLIOGRAPHY.md — Herzberg et al. 1995 (proactive secret sharing).
package refresh

import (
	"fmt"
	"math/big"

	"github.com/aalpar/shamir/pkg/field"
	"github.com/aalpar/shamir/pkg/polynomial"
	"github.com/aalpar/shamir/pkg/sss"
	"github.com/aalpar/shamir/pkg/vss"
)

// SubShare is a delta value from a zero polynomial, intended for
// the participant at index X.
//
// In a refresh round, each participant generates n sub-shares (one per
// participant) and distributes them. Participant j collects all sub-shares
// where X = j and applies them to their existing share.
type SubShare struct {
	X     field.Element // participant index
	Delta field.Element // δ(X) where δ(0) = 0
}

// Contribution bundles sub-shares from a single participant's zero
// polynomial with the corresponding Feldman commitment for verification.
type Contribution struct {
	SubShares  []SubShare
	Commitment *vss.Commitment
}

// ZeroSharing generates n sub-shares from a random degree-(k-1) polynomial
// with f(0) = 0. Each sub-share is a delta to be added to the corresponding
// participant's existing share.
//
// Requirements: 2 <= k <= n.
func ZeroSharing(n, k int, f *field.Field) ([]SubShare, error) {
	if k < 2 {
		return nil, fmt.Errorf("refresh: threshold k=%d must be >= 2", k)
	}
	if k > n {
		return nil, fmt.Errorf("refresh: threshold k=%d exceeds share count n=%d", k, n)
	}

	poly, err := polynomial.Random(k-1, f.Zero(), f)
	if err != nil {
		return nil, fmt.Errorf("refresh: generating zero polynomial: %w", err)
	}

	points := poly.EvaluateAt(n, f)
	subs := make([]SubShare, n)
	for i, pt := range points {
		subs[i] = SubShare{X: pt.X, Delta: pt.Y}
	}
	return subs, nil
}

// Apply refreshes a share by adding all received sub-share deltas.
// Every delta must have the same X as the share (i.e., be intended
// for this participant).
//
// new_y = old_y + Σ delta_i
//
// Since each delta comes from a polynomial with f(0) = 0, the sum
// of all participants' deltas at x=0 is zero, so the reconstructed
// secret is unchanged.
func Apply(share sss.Share, deltas []SubShare) (sss.Share, error) {
	y := share.Y
	for i, d := range deltas {
		if !d.X.Equal(share.X) {
			return sss.Share{}, fmt.Errorf("refresh: sub-share %d has X=%s, want %s", i, d.X, share.X)
		}
		y = y.Add(d.Delta)
	}
	return sss.Share{X: share.X, Y: y}, nil
}

// VerifiableZeroSharing generates a zero sharing with Feldman commitments
// that allow recipients to verify sub-shares before applying them.
//
// This is equivalent to FeldmanDeal with secret = 0. The commitment
// C_0 = g^0 = 1 (mod p) serves as a built-in correctness check.
//
// Requires a Group (safe prime + generator).
func VerifiableZeroSharing(n, k int, grp *vss.Group) (*Contribution, error) {
	zero := grp.Field().Zero()
	shares, commitment, err := vss.FeldmanDeal(zero, n, k, grp)
	if err != nil {
		return nil, fmt.Errorf("refresh: verifiable zero sharing: %w", err)
	}

	subs := make([]SubShare, n)
	for i, s := range shares {
		subs[i] = SubShare{X: s.X, Delta: s.Y}
	}
	return &Contribution{SubShares: subs, Commitment: commitment}, nil
}

// VerifySubShare checks that a sub-share is consistent with the published
// Feldman commitment. Returns true if g^{delta} = ∏ C_j^{x^j} (mod p).
func VerifySubShare(sub SubShare, c *vss.Commitment, grp *vss.Group) bool {
	share := sss.Share{X: sub.X, Y: sub.Delta}
	return vss.FeldmanVerify(share, c, grp)
}

// UpdateCommitment returns new commitments valid for the refreshed shares.
// For each coefficient index j:
//
//	C'_j = C_j * ∏ δC_{i,j}  (mod p)
//
// where δC_{i,j} is the j-th commitment from participant i's zero sharing.
//
// All commitments must have the same length (same polynomial degree).
func UpdateCommitment(original *vss.Commitment, deltas []*vss.Commitment, grp *vss.Group) (*vss.Commitment, error) {
	if len(deltas) == 0 {
		// No refresh contributions — return a copy of the original.
		values := make([]*big.Int, len(original.Values))
		for j, v := range original.Values {
			values[j] = new(big.Int).Set(v)
		}
		return vss.NewCommitment(values, grp), nil
	}

	k := len(original.Values)
	for i, dc := range deltas {
		if len(dc.Values) != k {
			return nil, fmt.Errorf("refresh: delta commitment %d has %d values, want %d", i, len(dc.Values), k)
		}
	}

	p := grp.P()
	values := make([]*big.Int, k)
	for j := range k {
		prod := new(big.Int).Set(original.Values[j])
		for _, dc := range deltas {
			prod.Mul(prod, dc.Values[j])
			prod.Mod(prod, p)
		}
		values[j] = prod
	}
	return vss.NewCommitment(values, grp), nil
}
