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
package sss

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/aalpar/shamir/pkg/field"
)

// shareJSON is the wire format for a Share.
// All values are hex-encoded big integers.
type shareJSON struct {
	P string `json:"p"`
	X string `json:"x"`
	Y string `json:"y"`
}

// MarshalJSON encodes a Share as self-describing JSON.
// The field prime p is included so each share stands alone.
func (s Share) MarshalJSON() ([]byte, error) {
	return json.Marshal(shareJSON{
		P: s.X.Field().Prime().Text(16),
		X: s.X.Value().Text(16),
		Y: s.Y.Value().Text(16),
	})
}

// UnmarshalJSON decodes a Share from its JSON representation.
// The field is reconstructed from the embedded prime p.
func (s *Share) UnmarshalJSON(data []byte) error {
	var raw shareJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("sss: unmarshal share: %w", err)
	}

	p, ok := new(big.Int).SetString(raw.P, 16)
	if !ok || p.Sign() <= 0 {
		return fmt.Errorf("sss: invalid prime %q", raw.P)
	}

	x, ok := new(big.Int).SetString(raw.X, 16)
	if !ok {
		return fmt.Errorf("sss: invalid x %q", raw.X)
	}

	y, ok := new(big.Int).SetString(raw.Y, 16)
	if !ok {
		return fmt.Errorf("sss: invalid y %q", raw.Y)
	}

	f := field.New(p)
	s.X = f.NewElement(x)
	s.Y = f.NewElement(y)
	return nil
}
