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
	"fmt"
	"math/big"

	"github.com/aalpar/shamir/pkg/field"
	"github.com/aalpar/shamir/pkg/polynomial"
	"github.com/aalpar/shamir/pkg/sss"
)

// FeldmanDeal splits a secret into n shares with threshold k and produces
// commitments that allow each shareholder to verify their share.
//
// The commitments C_j = g^{a_j} mod p are published by the dealer.
// Shares are identical to Shamir SSS shares and can be combined with
// sss.Combine.
//
// See: BIBLIOGRAPHY.md — Feldman 1987.
func FeldmanDeal(secret field.Element, n, k int, grp *Group) ([]sss.Share, *Commitment, error) {
	if k < 2 {
		return nil, nil, &Error{Op: "FeldmanDeal", Kind: ErrThreshold, Detail: fmt.Sprintf("k=%d must be >= 2", k)}
	}
	if k > n {
		return nil, nil, &Error{Op: "FeldmanDeal", Kind: ErrThreshold, Detail: fmt.Sprintf("k=%d exceeds n=%d", k, n)}
	}

	// Generate random polynomial f(x) with f(0) = secret.
	poly, err := polynomial.Random(k-1, secret, grp.f)
	if err != nil {
		return nil, nil, &Error{Op: "FeldmanDeal", Err: err}
	}

	// Evaluate at x = 1..n to produce shares.
	points := poly.EvaluateAt(n, grp.f)
	shares := make([]sss.Share, n)
	for i, pt := range points {
		shares[i] = sss.Share{X: pt.X, Y: pt.Y}
	}

	// Compute commitments: C_j = g^{a_j} mod p.
	values := make([]*big.Int, len(poly.Coefficients))
	for j, coeff := range poly.Coefficients {
		values[j] = grp.exp(grp.g, coeff.Value())
	}

	return shares, NewCommitment(values, grp), nil
}

// FeldmanVerify checks that a share is consistent with the published
// commitments. Returns true if g^{y_i} = ∏ C_j^{x_i^j} (mod p).
//
// This detects a dishonest dealer who distributed inconsistent shares.
// It does NOT detect a dealer who shared a different secret than claimed
// (the commitment C_0 = g^{secret} reveals nothing about the secret
// under the discrete log assumption).
func FeldmanVerify(share sss.Share, c *Commitment, grp *Group) bool {
	// LHS: g^{y_i} mod p
	lhs := grp.exp(grp.g, share.Y.Value())

	// RHS: ∏ C_j^{x^j} mod p
	rhs := big.NewInt(1)
	x := share.X.Value()
	xPow := big.NewInt(1) // x^0 = 1

	for _, cj := range c.Values {
		term := grp.exp(cj, xPow)
		rhs = grp.mul(rhs, term)
		xPow = new(big.Int).Mul(xPow, x)
		xPow.Mod(xPow, grp.q) // reduce exponent mod q (group order)
	}

	return lhs.Cmp(rhs) == 0
}
