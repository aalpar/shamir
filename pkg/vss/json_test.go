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
package vss

import (
	"encoding/json"
	"math/big"
	"testing"
)

// TestCommitmentJSONRoundtrip is the core property test:
// marshal → unmarshal produces a commitment that still verifies shares.
func TestCommitmentJSONRoundtrip(t *testing.T) {
	grp := testGroup(t)
	secret := grp.Field().NewElement(big.NewInt(7))

	shares, commit, err := FeldmanDeal(secret, 5, 3, grp)
	if err != nil {
		t.Fatalf("FeldmanDeal() error: %v", err)
	}

	data, err := json.Marshal(commit)
	if err != nil {
		t.Fatalf("Marshal() error: %v", err)
	}

	var decoded Commitment
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}

	// Values preserved.
	if len(decoded.Values) != len(commit.Values) {
		t.Fatalf("len(Values) = %d, want %d", len(decoded.Values), len(commit.Values))
	}
	for j := range commit.Values {
		if decoded.Values[j].Cmp(commit.Values[j]) != 0 {
			t.Errorf("value[%d] = %s, want %s", j, decoded.Values[j], commit.Values[j])
		}
	}

	// Deserialized commitment verifies original shares.
	for i, s := range shares {
		if !FeldmanVerify(s, &decoded, grp) {
			t.Errorf("share %d failed verification against deserialized commitment", i)
		}
	}
}

// TestCommitmentJSONFormat verifies hex encoding in the wire format.
func TestCommitmentJSONFormat(t *testing.T) {
	grp := testGroup(t)
	secret := grp.Field().NewElement(big.NewInt(5))

	_, commit, err := FeldmanDeal(secret, 3, 2, grp)
	if err != nil {
		t.Fatalf("FeldmanDeal() error: %v", err)
	}

	data, err := json.Marshal(commit)
	if err != nil {
		t.Fatalf("Marshal() error: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal raw: %v", err)
	}

	// Must have p, g, values keys.
	for _, key := range []string{"p", "g", "values"} {
		if _, ok := raw[key]; !ok {
			t.Errorf("missing key %q", key)
		}
	}

	// p = 23 decimal = 17 hex.
	var p string
	json.Unmarshal(raw["p"], &p)
	if p != "17" {
		t.Errorf("p = %q, want \"17\"", p)
	}
}

// TestPedersenShareJSONRoundtrip verifies marshal → unmarshal preserves
// all three field elements (X, Y, T) and reconstructs the correct field.
func TestPedersenShareJSONRoundtrip(t *testing.T) {
	grp := testGroup(t)
	secret := grp.Field().NewElement(big.NewInt(7))

	shares, _, err := PedersenDeal(secret, 5, 3, grp)
	if err != nil {
		t.Fatalf("PedersenDeal() error: %v", err)
	}

	for i, orig := range shares {
		data, err := json.Marshal(&orig)
		if err != nil {
			t.Fatalf("Marshal share %d: %v", i, err)
		}

		var decoded PedersenShare
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Unmarshal share %d: %v", i, err)
		}

		if decoded.X.Value().Cmp(orig.X.Value()) != 0 {
			t.Errorf("share %d: X = %s, want %s", i, decoded.X, orig.X)
		}
		if decoded.Y.Value().Cmp(orig.Y.Value()) != 0 {
			t.Errorf("share %d: Y = %s, want %s", i, decoded.Y, orig.Y)
		}
		if decoded.T.Value().Cmp(orig.T.Value()) != 0 {
			t.Errorf("share %d: T = %s, want %s", i, decoded.T, orig.T)
		}
		if decoded.X.Field().Prime().Cmp(grp.Field().Prime()) != 0 {
			t.Errorf("share %d: prime mismatch", i)
		}
	}
}

// TestPedersenShareJSONFormat verifies hex encoding in the wire format.
func TestPedersenShareJSONFormat(t *testing.T) {
	grp := testGroup(t)
	f := grp.Field()
	s := PedersenShare{
		X: f.NewElement(big.NewInt(3)),
		Y: f.NewElement(big.NewInt(7)),
		T: f.NewElement(big.NewInt(9)),
	}

	data, err := json.Marshal(&s)
	if err != nil {
		t.Fatalf("Marshal() error: %v", err)
	}

	var raw map[string]string
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal raw: %v", err)
	}

	// q = 11 (field prime). 11 decimal = b hex.
	if raw["q"] != "b" {
		t.Errorf("q = %q, want \"b\"", raw["q"])
	}
	if raw["x"] != "3" {
		t.Errorf("x = %q, want \"3\"", raw["x"])
	}
	if raw["y"] != "7" {
		t.Errorf("y = %q, want \"7\"", raw["y"])
	}
	if raw["t"] != "9" {
		t.Errorf("t = %q, want \"9\"", raw["t"])
	}
}

// TestCommitmentJSONPedersenWithH verifies that Pedersen commitments
// (with h generator) serialize and deserialize correctly.
func TestCommitmentJSONPedersenWithH(t *testing.T) {
	grp := testGroup(t)
	secret := grp.Field().NewElement(big.NewInt(5))

	_, commit, err := PedersenDeal(secret, 5, 3, grp)
	if err != nil {
		t.Fatalf("PedersenDeal() error: %v", err)
	}

	data, err := json.Marshal(commit)
	if err != nil {
		t.Fatalf("Marshal() error: %v", err)
	}

	// Wire format must include h.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal raw: %v", err)
	}
	if _, ok := raw["h"]; !ok {
		t.Error("Pedersen commitment missing \"h\" in JSON")
	}

	var decoded Commitment
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}

	if len(decoded.Values) != len(commit.Values) {
		t.Fatalf("len(Values) = %d, want %d", len(decoded.Values), len(commit.Values))
	}
	for j := range commit.Values {
		if decoded.Values[j].Cmp(commit.Values[j]) != 0 {
			t.Errorf("value[%d] mismatch", j)
		}
	}
}
