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
	"crypto/rand"
	"fmt"
	"math/big"
)

var bigOne = big.NewInt(1)
var bigTwo = big.NewInt(2)

// genSafePrime generates a safe prime p = 2q+1 where q is also prime.
// bits specifies the minimum bit length of p.
func genSafePrime(bits int) (*big.Int, error) {
	for {
		q, err := rand.Prime(rand.Reader, bits-1)
		if err != nil {
			return nil, fmt.Errorf("genSafePrime: %w", err)
		}

		p := new(big.Int).Mul(q, bigTwo)
		p.Add(p, bigOne)

		if p.ProbablyPrime(20) {
			return p, nil
		}
	}
}

// findGenerator finds a generator of the order-q subgroup of Z*_p
// where p = 2q+1 is a safe prime. It uses the squaring trick: for a safe
// prime, h^2 mod p has order q whenever h^2 != 1.
func findGenerator(p, q *big.Int) (*big.Int, error) {
	for range 1000 {
		h, err := rand.Int(rand.Reader, new(big.Int).Sub(p, bigTwo))
		if err != nil {
			return nil, fmt.Errorf("findGenerator: %w", err)
		}
		h.Add(h, bigTwo) // shift from [0, p-4] to [2, p-2]

		// g = h^2 mod p guarantees g is in the order-q subgroup.
		g := new(big.Int).Exp(h, bigTwo, p)

		if g.Cmp(bigOne) == 0 {
			continue
		}

		return g, nil
	}
	return nil, fmt.Errorf("findGenerator: failed after 1000 attempts")
}
