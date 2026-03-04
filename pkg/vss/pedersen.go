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
package vss

import (
	"fmt"
	"math/big"

	"github.com/aalpar/shamir/pkg/field"
	"github.com/aalpar/shamir/pkg/polynomial"
	"github.com/aalpar/shamir/pkg/sss"
)

// PedersenDeal splits a secret into n shares with threshold k, using a
// second random "blinding" polynomial to produce information-theoretically
// hiding commitments.
//
// Each share has two components: Y = f(i) (the secret share) and
// T = r(i) (the blinding share). Commitments C_j = g^{a_j} * h^{b_j}
// hide the polynomial coefficients even against unbounded adversaries.
//
// Reconstruction uses only the Y components via sss.Combine — the T
// values are solely for verification.
//
// Requires a Group with two generators (h != nil).
//
// See: BIBLIOGRAPHY.md — Pedersen 1991.
func PedersenDeal(secret field.Element, n, k int, grp *Group) ([]PedersenShare, *Commitment, error) {
	if grp.h == nil {
		return nil, nil, fmt.Errorf("vss: Pedersen requires a second generator h")
	}
	if k < 2 {
		return nil, nil, fmt.Errorf("vss: threshold k=%d must be >= 2", k)
	}
	if k > n {
		return nil, nil, fmt.Errorf("vss: threshold k=%d exceeds share count n=%d", k, n)
	}

	// Generate secret polynomial f(x) with f(0) = secret.
	fPoly, err := polynomial.Random(k-1, secret, grp.f)
	if err != nil {
		return nil, nil, fmt.Errorf("vss: generating secret polynomial: %w", err)
	}

	// Generate blinding polynomial r(x) with random r(0).
	r0, err := grp.f.Rand()
	if err != nil {
		return nil, nil, fmt.Errorf("vss: generating blinding constant: %w", err)
	}
	rPoly, err := polynomial.Random(k-1, r0, grp.f)
	if err != nil {
		return nil, nil, fmt.Errorf("vss: generating blinding polynomial: %w", err)
	}

	// Evaluate both polynomials at x = 1..n.
	fPoints := fPoly.EvaluateAt(n, grp.f)
	rPoints := rPoly.EvaluateAt(n, grp.f)

	shares := make([]PedersenShare, n)
	for i := range shares {
		shares[i] = PedersenShare{
			X: fPoints[i].X,
			Y: fPoints[i].Y,
			T: rPoints[i].Y,
		}
	}

	// Commitments: C_j = g^{a_j} * h^{b_j} mod p.
	fCoeffs := fPoly.Coefficients
	rCoeffs := rPoly.Coefficients
	values := make([]*big.Int, len(fCoeffs))
	for j := range fCoeffs {
		gaj := grp.exp(grp.g, fCoeffs[j].Value())
		hbj := grp.exp(grp.h, rCoeffs[j].Value())
		values[j] = grp.mul(gaj, hbj)
	}

	return shares, &Commitment{Values: values}, nil
}

// PedersenVerify checks that a share is consistent with the published
// commitments. Returns true if g^{y_i} * h^{t_i} = ∏ C_j^{x_i^j} (mod p).
//
// Unlike Feldman, the commitments are information-theoretically hiding:
// even with unbounded computational power, the commitments reveal nothing
// about the secret.
func PedersenVerify(share PedersenShare, c *Commitment, grp *Group) bool {
	if grp.h == nil {
		return false
	}

	// LHS: g^{y_i} * h^{t_i} mod p
	gs := grp.exp(grp.g, share.Y.Value())
	ht := grp.exp(grp.h, share.T.Value())
	lhs := grp.mul(gs, ht)

	// RHS: ∏ C_j^{x^j} mod p
	rhs := big.NewInt(1)
	x := share.X.Value()
	xPow := big.NewInt(1) // x^0 = 1

	for _, cj := range c.Values {
		term := grp.exp(cj, xPow)
		rhs = grp.mul(rhs, term)
		xPow = new(big.Int).Mul(xPow, x)
		xPow.Mod(xPow, grp.q)
	}

	return lhs.Cmp(rhs) == 0
}

// PedersenToSSS extracts basic SSS shares from Pedersen shares for
// reconstruction via sss.Combine.
func PedersenToSSS(shares []PedersenShare) []sss.Share {
	out := make([]sss.Share, len(shares))
	for i, s := range shares {
		out[i] = sss.Share{X: s.X, Y: s.Y}
	}
	return out
}
