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
// Package field implements finite field arithmetic over GF(p).
//
// All operations are performed modulo a prime p. The Field type holds
// the prime and provides arithmetic operations on Element values.
// Elements carry a reference to their field for runtime invariant checking.
//
// The current implementation uses math/big for arbitrary precision.
// This is not constant-time; side-channel resistance is a future concern.
//
// See: BIBLIOGRAPHY.md — Hardy & Wright Ch. 5-6, Shoup Ch. 2.
package field

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// Field represents GF(p) for a given prime p.
type Field struct {
	p *big.Int
}

// Element is a value in a finite field.
type Element struct {
	v     *big.Int
	field *Field
}

// New creates a finite field GF(p). Panics if p is not prime.
func New(p *big.Int) *Field {
	if !p.ProbablyPrime(20) {
		panic(fmt.Sprintf("field: %s is not prime", p))
	}
	return &Field{p: new(big.Int).Set(p)}
}

// Prime returns the field's prime modulus.
func (f *Field) Prime() *big.Int {
	return new(big.Int).Set(f.p)
}

// NewElement creates an element in this field, reducing v mod p.
func (f *Field) NewElement(v *big.Int) Element {
	r := new(big.Int).Mod(v, f.p)
	return Element{v: r, field: f}
}

// Zero returns the additive identity.
func (f *Field) Zero() Element {
	return Element{v: big.NewInt(0), field: f}
}

// One returns the multiplicative identity.
func (f *Field) One() Element {
	return Element{v: big.NewInt(1), field: f}
}

// Rand returns a uniformly random element in [0, p).
func (f *Field) Rand() (Element, error) {
	v, err := rand.Int(rand.Reader, f.p)
	if err != nil {
		return Element{}, &Error{Op: "Rand", Kind: ErrRandom, Err: err}
	}
	return Element{v: v, field: f}, nil
}

// Value returns the element's value as a new big.Int.
func (e Element) Value() *big.Int {
	return new(big.Int).Set(e.v)
}

// Field returns the element's field.
func (e Element) Field() *Field {
	return e.field
}

// IsZero reports whether the element is the additive identity.
func (e Element) IsZero() bool {
	return e.v.Sign() == 0
}

func (e Element) sameField(other Element) {
	if e.field.p.Cmp(other.field.p) != 0 {
		panic("field: operations on elements from different fields")
	}
}

// Add returns e + other (mod p).
func (e Element) Add(other Element) Element {
	e.sameField(other)
	r := new(big.Int).Add(e.v, other.v)
	r.Mod(r, e.field.p)
	return Element{v: r, field: e.field}
}

// Sub returns e - other (mod p).
func (e Element) Sub(other Element) Element {
	e.sameField(other)
	r := new(big.Int).Sub(e.v, other.v)
	r.Mod(r, e.field.p)
	return Element{v: r, field: e.field}
}

// Neg returns -e (mod p).
func (e Element) Neg() Element {
	r := new(big.Int).Neg(e.v)
	r.Mod(r, e.field.p)
	return Element{v: r, field: e.field}
}

// Mul returns e * other (mod p).
func (e Element) Mul(other Element) Element {
	e.sameField(other)
	r := new(big.Int).Mul(e.v, other.v)
	r.Mod(r, e.field.p)
	return Element{v: r, field: e.field}
}

// Inv returns the multiplicative inverse of e using the Extended Euclidean
// Algorithm. Panics if e is zero.
//
// See: BIBLIOGRAPHY.md — Hardy & Wright Ch. 10 (Euclidean algorithm),
// Shoup Ch. 2 (modular inverse).
func (e Element) Inv() Element {
	if e.IsZero() {
		panic("field: inverse of zero")
	}
	r := new(big.Int).ModInverse(e.v, e.field.p)
	return Element{v: r, field: e.field}
}

// Exp returns e^exp (mod p).
func (e Element) Exp(exp *big.Int) Element {
	r := new(big.Int).Exp(e.v, exp, e.field.p)
	return Element{v: r, field: e.field}
}

// Equal reports whether e and other have the same value in the same field.
func (e Element) Equal(other Element) bool {
	return e.field == other.field && e.v.Cmp(other.v) == 0
}

func (e Element) String() string {
	return e.v.String()
}
