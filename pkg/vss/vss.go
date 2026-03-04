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
// Package vss implements Verifiable Secret Sharing schemes.
//
// VSS extends Shamir's Secret Sharing so that shareholders can verify
// the correctness of their shares without interaction, detecting a
// dishonest dealer.
//
// Both schemes operate over a multiplicative subgroup of Z_p* of prime
// order q, where p is a safe prime (p = 2q + 1). Shares are field
// elements in GF(q); commitments are group elements in Z_p*.
//
// Two schemes are provided:
//
// Feldman VSS — the dealer publishes commitments C_j = g^{a_j} mod p
// for each polynomial coefficient. Computationally hiding (discrete log).
// See: BIBLIOGRAPHY.md — Feldman 1987.
//
// Pedersen VSS — uses two generators g, h with unknown log_g(h).
// Commitments C_j = g^{a_j} * h^{b_j}. Information-theoretically hiding.
// See: BIBLIOGRAPHY.md — Pedersen 1991.
package vss

import (
	"math/big"

	"github.com/aalpar/shamir/pkg/field"
)

var bigOne = big.NewInt(1)

// Group holds parameters for a multiplicative subgroup of Z_p* of prime
// order q, where p = 2q + 1 is a safe prime.
//
// Shares live in GF(q) (polynomial arithmetic). Commitments live in the
// group (modular exponentiation mod p).
type Group struct {
	p *big.Int
	q *big.Int
	g *big.Int
	h *big.Int     // second generator for Pedersen; nil if not needed
	f *field.Field // GF(q) for polynomial/share arithmetic
}

// NewGroup creates a group from a safe prime p and generator g.
// The generator must have order q = (p-1)/2 in Z_p*.
// Pass h = nil if only Feldman VSS is needed; Pedersen requires both
// generators with unknown log_g(h).
func NewGroup(p, g, h *big.Int) (*Group, error) {
	if !p.ProbablyPrime(20) {
		return nil, &Error{Op: "NewGroup", Kind: ErrNotPrime}
	}

	q := new(big.Int).Sub(p, bigOne)
	q.Rsh(q, 1) // q = (p-1)/2

	if !q.ProbablyPrime(20) {
		return nil, &Error{Op: "NewGroup", Kind: ErrNotSafePrime, Detail: "q = " + q.String()}
	}

	if err := validateGenerator(g, q, p, "g"); err != nil {
		return nil, err
	}

	grp := &Group{
		p: new(big.Int).Set(p),
		q: new(big.Int).Set(q),
		g: new(big.Int).Set(g),
		f: field.New(q),
	}

	if h != nil {
		if err := validateGenerator(h, q, p, "h"); err != nil {
			return nil, err
		}
		if g.Cmp(h) == 0 {
			return nil, &Error{Op: "NewGroup", Kind: ErrGeneratorsEqual}
		}
		grp.h = new(big.Int).Set(h)
	}

	return grp, nil
}

func validateGenerator(gen, q, p *big.Int, name string) error {
	if gen.Cmp(bigOne) <= 0 || gen.Cmp(p) >= 0 {
		return &Error{Op: "NewGroup", Kind: ErrGeneratorRange, Detail: name}
	}
	// Must have order q: gen^q ≡ 1 (mod p).
	if new(big.Int).Exp(gen, q, p).Cmp(bigOne) != 0 {
		return &Error{Op: "NewGroup", Kind: ErrGeneratorOrder, Detail: name}
	}
	return nil
}

// Field returns GF(q) for creating secrets and working with shares.
func (grp *Group) Field() *field.Field { return grp.f }

// P returns the safe prime.
func (grp *Group) P() *big.Int { return new(big.Int).Set(grp.p) }

// Q returns the subgroup order.
func (grp *Group) Q() *big.Int { return new(big.Int).Set(grp.q) }

// exp computes base^exp mod p.
func (grp *Group) exp(base, e *big.Int) *big.Int {
	return new(big.Int).Exp(base, e, grp.p)
}

// mul computes a * b mod p.
func (grp *Group) mul(a, b *big.Int) *big.Int {
	r := new(big.Int).Mul(a, b)
	return r.Mod(r, grp.p)
}

// Commitment holds the dealer's published values that shareholders use
// to verify their shares. For Feldman: C_j = g^{a_j}. For Pedersen:
// C_j = g^{a_j} * h^{b_j}.
//
// grp is retained so the commitment can serialize its group parameters
// and stand alone without out-of-band context.
type Commitment struct {
	Values []*big.Int
	grp    *Group
}

// NewCommitment creates a Commitment with the given values and group.
// The group is retained for serialization.
func NewCommitment(values []*big.Int, grp *Group) *Commitment {
	return &Commitment{Values: values, grp: grp}
}

// Group returns the group associated with this commitment, or nil if none.
// TODO: does not belong here long term.  Needs to be decoupled from the commitment and shareed out of band.
func (c *Commitment) Group() *Group { return c.grp }

// PedersenShare is a share produced by Pedersen VSS (Pedersen 1991).
//
// Unlike Feldman VSS (which reuses sss.Share directly), Pedersen VSS
// requires a second polynomial r(x) — the "blinding polynomial" — to
// achieve information-theoretic hiding. The dealer generates two random
// degree-(k-1) polynomials:
//
//	f(x)  with f(0) = secret      (the secret polynomial, same as Feldman/SSS)
//	r(x)  with r(0) = random      (the blinding polynomial, Pedersen-specific)
//
// Each shareholder i receives three values:
//
//	X  = i           evaluation point (same across both polynomials)
//	Y  = f(i)        the actual secret share (same as sss.Share.Y)
//	T  = r(i)        the blinding component (unique to Pedersen)
//
// Verification uses both components: the verifier checks that
// g^{Y} * h^{T} equals the product of commitments C_j^{i^j}, where
// C_j = g^{a_j} * h^{b_j} and a_j, b_j are coefficients of f and r.
//
// For reconstruction, only X and Y are needed — T is used solely for
// verification and can be discarded once the share is verified.
//
// See: PedersenDeal (deal shares), PedersenVerify (verify a share).
// Reference: BIBLIOGRAPHY.md — Pedersen 1991.
type PedersenShare struct {
	X, Y field.Element // share: (i, f(i))
	T    field.Element // blinding: t_i = r(i)
}
