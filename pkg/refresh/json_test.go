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
	"math/big"
	"testing"

	"github.com/aalpar/shamir/pkg/sss"
	"github.com/aalpar/shamir/pkg/vss"
)

// TestSubShareJSONRoundtrip verifies that a sub-share survives
// marshal → unmarshal and the reconstructed field matches.
func TestSubShareJSONRoundtrip(t *testing.T) {
	grp := testGroup(t)
	subs, err := ZeroSharing(5, 3, grp.Field())
	if err != nil {
		t.Fatalf("ZeroSharing() error: %v", err)
	}

	for i, orig := range subs {
		data, err := json.Marshal(orig)
		if err != nil {
			t.Fatalf("Marshal sub-share %d: %v", i, err)
		}

		var decoded SubShare
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Unmarshal sub-share %d: %v", i, err)
		}

		if !decoded.X.Equal(orig.X) {
			t.Errorf("sub-share %d: X = %s, want %s", i, decoded.X, orig.X)
		}
		if !decoded.Delta.Equal(orig.Delta) {
			t.Errorf("sub-share %d: Delta = %s, want %s", i, decoded.Delta, orig.Delta)
		}
		if decoded.X.Field().Prime().Cmp(grp.Field().Prime()) != 0 {
			t.Errorf("sub-share %d: prime mismatch", i)
		}
	}
}

// TestSubShareJSONFormat verifies hex encoding in the wire format.
func TestSubShareJSONFormat(t *testing.T) {
	grp := testGroup(t)
	f := grp.Field()
	sub := SubShare{
		X:     f.NewElement(big.NewInt(3)),
		Delta: f.NewElement(big.NewInt(7)),
	}

	data, err := json.Marshal(sub)
	if err != nil {
		t.Fatalf("Marshal() error: %v", err)
	}

	var raw map[string]string
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal raw: %v", err)
	}

	// q = 11 (the field prime for p=23 safe prime). 11 decimal = b hex.
	if raw["p"] != "b" {
		t.Errorf("p = %q, want \"b\"", raw["p"])
	}
	if raw["x"] != "3" {
		t.Errorf("x = %q, want \"3\"", raw["x"])
	}
	if raw["delta"] != "7" {
		t.Errorf("delta = %q, want \"7\"", raw["delta"])
	}
}

// TestSubShareJSONUnmarshalErrors covers malformed input rejection.
func TestSubShareJSONUnmarshalErrors(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{"not json", `not json`},
		{"empty object", `{}`},
		{"invalid p", `{"p":"zz","x":"3","delta":"7"}`},
		{"zero p", `{"p":"0","x":"3","delta":"7"}`},
		{"negative p", `{"p":"-1","x":"3","delta":"7"}`},
		{"invalid x", `{"p":"b","x":"zz","delta":"7"}`},
		{"invalid delta", `{"p":"b","x":"3","delta":"zz"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s SubShare
			if err := json.Unmarshal([]byte(tt.json), &s); err == nil {
				t.Error("expected error")
			}
		})
	}
}

// TestContributionJSONRoundtrip is the core property test:
// marshal → unmarshal → verify sub-shares against the deserialized commitment.
func TestContributionJSONRoundtrip(t *testing.T) {
	grp := testGroup(t)

	contrib, err := VerifiableZeroSharing(5, 3, grp)
	if err != nil {
		t.Fatalf("VerifiableZeroSharing() error: %v", err)
	}

	data, err := json.Marshal(contrib)
	if err != nil {
		t.Fatalf("Marshal() error: %v", err)
	}

	var decoded Contribution
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}

	// Sub-share count preserved.
	if len(decoded.SubShares) != len(contrib.SubShares) {
		t.Fatalf("sub-share count = %d, want %d", len(decoded.SubShares), len(contrib.SubShares))
	}

	// Sub-share values preserved.
	for i := range contrib.SubShares {
		if !decoded.SubShares[i].X.Equal(contrib.SubShares[i].X) {
			t.Errorf("sub-share %d: X mismatch", i)
		}
		if !decoded.SubShares[i].Delta.Equal(contrib.SubShares[i].Delta) {
			t.Errorf("sub-share %d: Delta mismatch", i)
		}
	}

	// Commitment values preserved.
	if len(decoded.Commitment.Values) != len(contrib.Commitment.Values) {
		t.Fatalf("commitment length = %d, want %d", len(decoded.Commitment.Values), len(contrib.Commitment.Values))
	}
	for j := range contrib.Commitment.Values {
		if decoded.Commitment.Values[j].Cmp(contrib.Commitment.Values[j]) != 0 {
			t.Errorf("commitment value[%d] mismatch", j)
		}
	}
}

// TestContributionJSONVerifyAfterRoundtrip verifies that deserialized
// sub-shares still pass Feldman verification against the deserialized commitment.
func TestContributionJSONVerifyAfterRoundtrip(t *testing.T) {
	grp := testGroup(t)

	contrib, err := VerifiableZeroSharing(5, 3, grp)
	if err != nil {
		t.Fatalf("VerifiableZeroSharing() error: %v", err)
	}

	data, err := json.Marshal(contrib)
	if err != nil {
		t.Fatalf("Marshal() error: %v", err)
	}

	var decoded Contribution
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}

	// Re-verify all sub-shares using the deserialized commitment.
	for i, sub := range decoded.SubShares {
		if !VerifySubShare(sub, decoded.Commitment, grp) {
			t.Errorf("deserialized sub-share %d failed verification", i)
		}
	}
}

// TestContributionJSONRefreshRoundtrip tests the full workflow:
// deal → refresh with serialized contribution → reconstruct secret.
func TestContributionJSONRefreshRoundtrip(t *testing.T) {
	grp := testGroup(t)
	secret := grp.Field().NewElement(big.NewInt(7))

	shares, _, err := vss.FeldmanDeal(secret, 5, 3, grp)
	if err != nil {
		t.Fatalf("FeldmanDeal() error: %v", err)
	}

	// Generate contribution, serialize, deserialize.
	contrib, err := VerifiableZeroSharing(5, 3, grp)
	if err != nil {
		t.Fatalf("VerifiableZeroSharing() error: %v", err)
	}

	data, err := json.Marshal(contrib)
	if err != nil {
		t.Fatalf("Marshal() error: %v", err)
	}

	var decoded Contribution
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}

	// Apply deserialized sub-shares.
	refreshed := make([]sss.Share, 5)
	for i := range shares {
		refreshed[i], err = Apply(shares[i], []SubShare{decoded.SubShares[i]})
		if err != nil {
			t.Fatalf("Apply(share %d) error: %v", i, err)
		}
	}

	// Reconstruct — secret must be unchanged.
	got, err := sss.Combine(refreshed[:3])
	if err != nil {
		t.Fatalf("Combine() error: %v", err)
	}
	if !got.Equal(secret) {
		t.Errorf("Combine() = %s, want %s", got, secret)
	}
}

// TestContributionJSONUnmarshalErrors covers malformed input rejection.
func TestContributionJSONUnmarshalErrors(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{"not json", `not json`},
		{"empty sub_shares", `{"sub_shares":[],"commitment":{"p":"17","g":"4","values":["1"]}}`},
		{"bad sub_share", `{"sub_shares":[{"p":"zz"}],"commitment":{"p":"17","g":"4","values":["1"]}}`},
		{"bad commitment", `{"sub_shares":[{"p":"b","x":"1","delta":"2"}],"commitment":{"p":"zz"}}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var c Contribution
			if err := json.Unmarshal([]byte(tt.json), &c); err == nil {
				t.Error("expected error")
			}
		})
	}
}
