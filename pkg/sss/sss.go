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
// Package sss implements Shamir's Secret Sharing scheme.
//
// A (k, n) threshold scheme splits a secret into n shares such that any
// k shares can reconstruct the secret, but k-1 shares reveal no information
// about the secret (information-theoretic security).
//
// The construction: choose a random polynomial f of degree k-1 where
// f(0) = secret. Each share is an evaluation (i, f(i)) at a distinct
// nonzero point. Reconstruction uses Lagrange interpolation at x = 0.
//
// See: BIBLIOGRAPHY.md — Shamir 1979, Shannon 1949 (perfect secrecy).
package sss

import (
	"fmt"

	"github.com/aalpar/shamir/pkg/field"
	"github.com/aalpar/shamir/pkg/polynomial"
)

// Share is a single share of a secret: an evaluation point (X, Y) where
// Y = f(X) for the secret-encoding polynomial.
type Share struct {
	X, Y field.Element
}

// Split divides a secret into n shares with threshold k.
// Any k shares can reconstruct the secret; k-1 shares reveal nothing.
//
// Requirements:
//   - 2 <= k <= n
//   - The secret must be an element of the given field
func Split(secret field.Element, n, k int, f *field.Field) ([]Share, error) {
	if k < 2 {
		return nil, fmt.Errorf("sss: threshold k=%d must be >= 2", k)
	}
	if k > n {
		return nil, fmt.Errorf("sss: threshold k=%d exceeds share count n=%d", k, n)
	}

	poly, err := polynomial.Random(k-1, secret, f)
	if err != nil {
		return nil, fmt.Errorf("sss: generating polynomial: %w", err)
	}

	points := poly.EvaluateAt(n, f)
	shares := make([]Share, n)
	for i, pt := range points {
		shares[i] = Share{X: pt.X, Y: pt.Y}
	}
	return shares, nil
}

// Combine reconstructs the secret from k or more shares using Lagrange
// interpolation at x = 0.
//
// The shares must be from the same (k, n) split. Providing fewer than k
// shares produces an incorrect result (not an error — this is by design,
// as the scheme is information-theoretically secure).
func Combine(shares []Share) (field.Element, error) {
	if len(shares) == 0 {
		return field.Element{}, fmt.Errorf("sss: no shares provided")
	}

	points := make([]polynomial.Point, len(shares))
	for i, s := range shares {
		points[i] = polynomial.Point{X: s.X, Y: s.Y}
	}

	secret, err := polynomial.LagrangeZero(points)
	if err != nil {
		return field.Element{}, fmt.Errorf("sss: reconstruction: %w", err)
	}
	return secret, nil
}
