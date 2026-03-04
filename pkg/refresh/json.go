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
package refresh

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/aalpar/shamir/pkg/field"
	"github.com/aalpar/shamir/pkg/vss"
)

// MarshalText implements encoding.TextMarshaler using the JSON format.
func (s SubShare) MarshalText() ([]byte, error) {
	return s.MarshalJSON()
}

// UnmarshalText implements encoding.TextUnmarshaler using the JSON format.
func (s *SubShare) UnmarshalText(data []byte) error {
	return s.UnmarshalJSON(data)
}

// MarshalText implements encoding.TextMarshaler using the JSON format.
func (c Contribution) MarshalText() ([]byte, error) {
	return c.MarshalJSON()
}

// UnmarshalText implements encoding.TextUnmarshaler using the JSON format.
func (c *Contribution) UnmarshalText(data []byte) error {
	return c.UnmarshalJSON(data)
}

// subShareJSON is the wire format for a SubShare.
// All values are hex-encoded big integers.
type subShareJSON struct {
	P     string `json:"p"`
	X     string `json:"x"`
	Delta string `json:"delta"`
}

// MarshalJSON encodes a SubShare as self-describing JSON.
// The field prime p is included so each sub-share stands alone.
func (s SubShare) MarshalJSON() ([]byte, error) {
	return json.Marshal(subShareJSON{
		P:     s.X.Field().Prime().Text(16),
		X:     s.X.Value().Text(16),
		Delta: s.Delta.Value().Text(16),
	})
}

// UnmarshalJSON decodes a SubShare from its JSON representation.
// The field is reconstructed from the embedded prime p.
func (s *SubShare) UnmarshalJSON(data []byte) error {
	var raw subShareJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return &Error{Op: "UnmarshalJSON", Kind: ErrUnmarshal, Err: err}
	}

	p, ok := new(big.Int).SetString(raw.P, 16)
	if !ok || p.Sign() <= 0 {
		return &Error{Op: "UnmarshalJSON", Kind: ErrUnmarshal, Detail: "invalid prime " + raw.P}
	}

	x, ok := new(big.Int).SetString(raw.X, 16)
	if !ok {
		return &Error{Op: "UnmarshalJSON", Kind: ErrUnmarshal, Detail: "invalid x " + raw.X}
	}

	delta, ok := new(big.Int).SetString(raw.Delta, 16)
	if !ok {
		return &Error{Op: "UnmarshalJSON", Kind: ErrUnmarshal, Detail: "invalid delta " + raw.Delta}
	}

	f := field.New(p)
	s.X = f.NewElement(x)
	s.Delta = f.NewElement(delta)
	return nil
}

// contributionJSON is the wire format for a Contribution.
// SubShares are inline; Commitment delegates to vss.Commitment's own serialization.
type contributionJSON struct {
	SubShares  []json.RawMessage `json:"sub_shares"`
	Commitment json.RawMessage   `json:"commitment"`
}

// MarshalJSON encodes a Contribution as JSON.
// Each SubShare is serialized independently; the Commitment uses its own MarshalJSON.
func (c Contribution) MarshalJSON() ([]byte, error) {
	subs := make([]json.RawMessage, len(c.SubShares))
	for i, s := range c.SubShares {
		b, err := json.Marshal(s)
		if err != nil {
			return nil, &Error{Op: "MarshalJSON", Kind: ErrUnmarshal, Detail: fmt.Sprintf("sub-share %d", i), Err: err}
		}
		subs[i] = b
	}

	commitData, err := json.Marshal(c.Commitment)
	if err != nil {
		return nil, &Error{Op: "MarshalJSON", Err: err}
	}

	return json.Marshal(contributionJSON{
		SubShares:  subs,
		Commitment: commitData,
	})
}

// UnmarshalJSON decodes a Contribution from its JSON representation.
func (c *Contribution) UnmarshalJSON(data []byte) error {
	var raw contributionJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return &Error{Op: "UnmarshalJSON", Kind: ErrUnmarshal, Err: err}
	}

	if len(raw.SubShares) == 0 {
		return &Error{Op: "UnmarshalJSON", Kind: ErrUnmarshal, Detail: "no sub-shares"}
	}

	subs := make([]SubShare, len(raw.SubShares))
	for i, b := range raw.SubShares {
		if err := json.Unmarshal(b, &subs[i]); err != nil {
			return &Error{Op: "UnmarshalJSON", Kind: ErrUnmarshal, Detail: fmt.Sprintf("sub-share %d", i), Err: err}
		}
	}

	var commit vss.Commitment
	if err := json.Unmarshal(raw.Commitment, &commit); err != nil {
		return &Error{Op: "UnmarshalJSON", Kind: ErrUnmarshal, Detail: "commitment", Err: err}
	}

	c.SubShares = subs
	c.Commitment = &commit
	return nil
}
