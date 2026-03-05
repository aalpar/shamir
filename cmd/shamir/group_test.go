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
package main

import (
	"math/big"
	"testing"
)

func TestGenSafePrime(t *testing.T) {
	// Use small bit size for fast tests.
	p, err := genSafePrime(64)
	if err != nil {
		t.Fatalf("genSafePrime: %v", err)
	}

	if p.BitLen() < 64 {
		t.Errorf("prime bit length = %d, want >= 64", p.BitLen())
	}

	if !p.ProbablyPrime(20) {
		t.Error("p is not prime")
	}

	// q = (p-1)/2 should also be prime (safe prime property).
	q := new(big.Int).Sub(p, big.NewInt(1))
	q.Rsh(q, 1)
	if !q.ProbablyPrime(20) {
		t.Error("q = (p-1)/2 is not prime")
	}
}

func TestFindGenerator(t *testing.T) {
	p, err := genSafePrime(64)
	if err != nil {
		t.Fatalf("genSafePrime: %v", err)
	}

	q := new(big.Int).Sub(p, big.NewInt(1))
	q.Rsh(q, 1)

	g, err := findGenerator(p, q)
	if err != nil {
		t.Fatalf("findGenerator: %v", err)
	}

	// g^q mod p == 1 (order divides q)
	result := new(big.Int).Exp(g, q, p)
	if result.Cmp(big.NewInt(1)) != 0 {
		t.Errorf("g^q mod p = %s, want 1", result)
	}

	// g^2 mod p != 1 (order is not 2, so order must be q)
	result.Exp(g, big.NewInt(2), p)
	if result.Cmp(big.NewInt(1)) == 0 {
		t.Error("g^2 mod p = 1, generator has trivial order")
	}
}
