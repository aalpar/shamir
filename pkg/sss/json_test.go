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
package sss

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/aalpar/shamir/pkg/field"
)

// TestShareJSONRoundtrip is the core design property test:
// marshal → unmarshal produces a share that still reconstructs
// the original secret when combined with other shares.
//
// TODO(aalpar): implement this test. The structure:
//   1. Split a known secret into shares over a known field
//   2. Marshal each share to JSON
//   3. Unmarshal each share back
//   4. Combine the unmarshaled shares
//   5. Verify the reconstructed secret matches the original
//
// Consider: what happens if you marshal shares from a small field
// (p=17) vs a large field (secp256k1 prime)? Both should round-trip.
// Consider: does the unmarshaled share carry the same field as the
// original? (It won't be pointer-equal, but it must be value-equal.)
func TestShareJSONRoundtrip(t *testing.T) {
	t.Skip("TODO: implement round-trip test")
}

func TestShareJSONFormat(t *testing.T) {
	f := field.New(p17)
	s := Share{
		X: f.NewElement(big.NewInt(3)),
		Y: f.NewElement(big.NewInt(12)),
	}

	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("Marshal() error: %v", err)
	}

	// Verify the JSON contains hex-encoded values.
	var raw map[string]string
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal raw: %v", err)
	}

	if raw["p"] != "11" { // 17 decimal = 11 hex
		t.Errorf("p = %q, want \"11\"", raw["p"])
	}
	if raw["x"] != "3" {
		t.Errorf("x = %q, want \"3\"", raw["x"])
	}
	if raw["y"] != "c" { // 12 decimal = c hex
		t.Errorf("y = %q, want \"c\"", raw["y"])
	}
}

func TestShareJSONUnmarshalErrors(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{"empty object", `{}`},
		{"missing p", `{"x":"3","y":"c"}`},
		{"invalid p", `{"p":"zz","x":"3","y":"c"}`},
		{"zero p", `{"p":"0","x":"3","y":"c"}`},
		{"negative p", `{"p":"-1","x":"3","y":"c"}`},
		{"invalid x", `{"p":"11","x":"zz","y":"c"}`},
		{"invalid y", `{"p":"11","x":"3","y":"zz"}`},
		{"not json", `not json`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s Share
			if err := json.Unmarshal([]byte(tt.json), &s); err == nil {
				t.Error("expected error")
			}
		})
	}
}

func TestShareJSONLargePrime(t *testing.T) {
	f := field.New(p256)
	secret := f.NewElement(big.NewInt(42))

	shares, err := Split(secret, 5, 3, f)
	if err != nil {
		t.Fatalf("Split() error: %v", err)
	}

	// Marshal and unmarshal the first share.
	data, err := json.Marshal(shares[0])
	if err != nil {
		t.Fatalf("Marshal() error: %v", err)
	}

	var decoded Share
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}

	// The decoded share's field prime must match.
	if decoded.X.Field().Prime().Cmp(p256) != 0 {
		t.Errorf("decoded prime = %s, want %s", decoded.X.Field().Prime(), p256)
	}

	// The decoded values must match the originals.
	if !decoded.X.Value().IsInt64() || decoded.X.Value().Cmp(shares[0].X.Value()) != 0 {
		t.Errorf("decoded X = %s, want %s", decoded.X, shares[0].X)
	}
	if decoded.Y.Value().Cmp(shares[0].Y.Value()) != 0 {
		t.Errorf("decoded Y = %s, want %s", decoded.Y, shares[0].Y)
	}
}
