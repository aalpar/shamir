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
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/aalpar/shamir/pkg/field"
)

// commitmentJSON is the wire format for a Commitment.
// All values are hex-encoded big integers. h is omitted for Feldman commitments.
type commitmentJSON struct {
	P      string   `json:"p"`
	G      string   `json:"g"`
	H      string   `json:"h,omitempty"`
	Values []string `json:"values"`
}

// MarshalJSON encodes a Commitment as self-describing JSON, embedding
// the group parameters so the receiver can verify without out-of-band context.
func (c *Commitment) MarshalJSON() ([]byte, error) {
	if c.grp == nil {
		return nil, fmt.Errorf("vss: commitment has no group (was it created by FeldmanDeal or PedersenDeal?)")
	}
	values := make([]string, len(c.Values))
	for i, v := range c.Values {
		values[i] = v.Text(16)
	}
	wire := commitmentJSON{
		P:      c.grp.p.Text(16),
		G:      c.grp.g.Text(16),
		Values: values,
	}
	if c.grp.h != nil {
		wire.H = c.grp.h.Text(16)
	}
	return json.Marshal(wire)
}

// UnmarshalJSON decodes a Commitment from its JSON representation,
// reconstructing the group from the embedded parameters.
func (c *Commitment) UnmarshalJSON(data []byte) error {
	var wire commitmentJSON
	if err := json.Unmarshal(data, &wire); err != nil {
		return fmt.Errorf("vss: unmarshal commitment: %w", err)
	}

	p, ok := new(big.Int).SetString(wire.P, 16)
	if !ok || p.Sign() <= 0 {
		return fmt.Errorf("vss: invalid group prime p %q", wire.P)
	}

	g, ok := new(big.Int).SetString(wire.G, 16)
	if !ok || g.Sign() <= 0 {
		return fmt.Errorf("vss: invalid generator g %q", wire.G)
	}

	var h *big.Int
	if wire.H != "" {
		h, ok = new(big.Int).SetString(wire.H, 16)
		if !ok || h.Sign() <= 0 {
			return fmt.Errorf("vss: invalid generator h %q", wire.H)
		}
	}

	grp, err := NewGroup(p, g, h)
	if err != nil {
		return fmt.Errorf("vss: reconstructing group: %w", err)
	}

	values := make([]*big.Int, len(wire.Values))
	for i, s := range wire.Values {
		v, ok := new(big.Int).SetString(s, 16)
		if !ok {
			return fmt.Errorf("vss: invalid commitment value[%d] %q", i, s)
		}
		if v.Sign() < 0 || v.Cmp(p) >= 0 {
			return fmt.Errorf("vss: commitment value[%d] out of range [0, p)", i)
		}
		values[i] = v
	}

	c.grp = grp
	c.Values = values
	return nil
}

// pedersenShareJSON is the wire format for a PedersenShare.
// q is the field prime (GF(q) where shares live). All values are hex-encoded.
type pedersenShareJSON struct {
	Q string `json:"q"`
	X string `json:"x"`
	Y string `json:"y"`
	T string `json:"t"`
}

// MarshalJSON encodes a PedersenShare as self-describing JSON.
// The field prime q is included so each share stands alone.
func (s *PedersenShare) MarshalJSON() ([]byte, error) {
	return json.Marshal(pedersenShareJSON{
		Q: s.X.Field().Prime().Text(16),
		X: s.X.Value().Text(16),
		Y: s.Y.Value().Text(16),
		T: s.T.Value().Text(16),
	})
}

// UnmarshalJSON decodes a PedersenShare from its JSON representation.
// The field is reconstructed from the embedded prime q.
func (s *PedersenShare) UnmarshalJSON(data []byte) error {
	var wire pedersenShareJSON
	if err := json.Unmarshal(data, &wire); err != nil {
		return fmt.Errorf("vss: unmarshal pedersen share: %w", err)
	}

	q, ok := new(big.Int).SetString(wire.Q, 16)
	if !ok || q.Sign() <= 0 {
		return fmt.Errorf("vss: invalid field prime q %q", wire.Q)
	}

	x, ok := new(big.Int).SetString(wire.X, 16)
	if !ok {
		return fmt.Errorf("vss: invalid x %q", wire.X)
	}

	y, ok := new(big.Int).SetString(wire.Y, 16)
	if !ok {
		return fmt.Errorf("vss: invalid y %q", wire.Y)
	}

	t, ok := new(big.Int).SetString(wire.T, 16)
	if !ok {
		return fmt.Errorf("vss: invalid t %q", wire.T)
	}

	f := field.New(q)
	s.X = f.NewElement(x)
	s.Y = f.NewElement(y)
	s.T = f.NewElement(t)
	return nil
}
