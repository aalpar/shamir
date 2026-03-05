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
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"os"

	"github.com/aalpar/shamir/pkg/field"
	"github.com/aalpar/shamir/pkg/sss"
)

// SplitCommand holds the flags for the split subcommand.
type SplitCommand struct {
	Threshold int    `short:"k" long:"threshold" required:"true" description:"Minimum shares to reconstruct"`
	Shares    int    `short:"n" long:"shares" required:"true" description:"Total number of shares"`
	Scheme    string `long:"scheme" default:"sss" choice:"sss" choice:"feldman" choice:"pedersen" description:"Sharing scheme"`
	Bits      int    `long:"bits" default:"2048" description:"Safe prime bit size (VSS schemes only)"`
}

// Execute implements the go-flags Commander interface.
func (c *SplitCommand) Execute(args []string) error {
	return c.run(os.Stdin, os.Stdout, os.Stderr)
}

func (c *SplitCommand) run(stdin io.Reader, stdout, stderr io.Writer) error {
	secretBytes, err := io.ReadAll(stdin)
	if err != nil {
		return fmt.Errorf("split: reading stdin: %w", err)
	}
	if len(secretBytes) == 0 {
		return fmt.Errorf("split: empty secret (no data on stdin)")
	}

	secret := new(big.Int).SetBytes(secretBytes)

	switch c.Scheme {
	case "sss":
		return c.splitSSS(secret, stdout)
	case "feldman":
		return fmt.Errorf("split: feldman not yet implemented")
	case "pedersen":
		return fmt.Errorf("split: pedersen not yet implemented")
	default:
		return fmt.Errorf("split: unknown scheme %q", c.Scheme)
	}
}

func (c *SplitCommand) splitSSS(secret *big.Int, stdout io.Writer) error {
	p := nextPrime(secret)
	f := field.New(p)
	elem := f.NewElement(secret)

	shares, err := sss.Split(elem, c.Shares, c.Threshold, f)
	if err != nil {
		return fmt.Errorf("split: %w", err)
	}

	return writeShares(shares, stdout)
}

// nextPrime returns the smallest prime > n.
func nextPrime(n *big.Int) *big.Int {
	p := new(big.Int).Add(n, big.NewInt(1))
	if p.Bit(0) == 0 {
		p.Add(p, big.NewInt(1))
	}
	for !p.ProbablyPrime(20) {
		p.Add(p, big.NewInt(2))
	}
	return p
}

func writeShares(shares []sss.Share, w io.Writer) error {
	enc := json.NewEncoder(w)
	for i := range shares {
		if err := enc.Encode(&shares[i]); err != nil {
			return fmt.Errorf("split: writing share %d: %w", i, err)
		}
	}
	return nil
}
