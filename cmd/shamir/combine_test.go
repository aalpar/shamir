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
	"bytes"
	"strings"
	"testing"
)

func TestCombineRoundTripExactlyK(t *testing.T) {
	secret := []byte("hello world")
	var splitOut bytes.Buffer

	split := &SplitCommand{Threshold: 3, Shares: 5, Scheme: "sss"}
	if err := split.run(bytes.NewReader(secret), &splitOut, nil); err != nil {
		t.Fatalf("split: %v", err)
	}

	// Take exactly k=3 shares (first 3 lines)
	lines := strings.Split(strings.TrimSpace(splitOut.String()), "\n")
	kShares := strings.Join(lines[:3], "\n") + "\n"

	var combineOut bytes.Buffer
	combine := &CombineCommand{}
	if err := combine.run(strings.NewReader(kShares), &combineOut); err != nil {
		t.Fatalf("combine: %v", err)
	}

	if !bytes.Equal(combineOut.Bytes(), secret) {
		t.Errorf("got %q, want %q", combineOut.String(), secret)
	}
}

func TestCombineRoundTripAllN(t *testing.T) {
	secret := []byte("hello world")
	var splitOut bytes.Buffer

	split := &SplitCommand{Threshold: 3, Shares: 5, Scheme: "sss"}
	if err := split.run(bytes.NewReader(secret), &splitOut, nil); err != nil {
		t.Fatalf("split: %v", err)
	}

	var combineOut bytes.Buffer
	combine := &CombineCommand{}
	if err := combine.run(strings.NewReader(splitOut.String()), &combineOut); err != nil {
		t.Fatalf("combine: %v", err)
	}

	if !bytes.Equal(combineOut.Bytes(), secret) {
		t.Errorf("got %q, want %q", combineOut.String(), secret)
	}
}

func TestCombineEmptyStdin(t *testing.T) {
	var out bytes.Buffer
	combine := &CombineCommand{}
	err := combine.run(bytes.NewReader(nil), &out)
	if err == nil {
		t.Fatal("expected error for empty stdin")
	}
	if !strings.Contains(err.Error(), "no shares") {
		t.Errorf("error = %q, want substring %q", err, "no shares")
	}
}

func TestCombineMalformedInput(t *testing.T) {
	var out bytes.Buffer
	combine := &CombineCommand{}
	err := combine.run(strings.NewReader("not json\n"), &out)
	if err == nil {
		t.Fatal("expected error for malformed input")
	}
	if !strings.Contains(err.Error(), "line 1") {
		t.Errorf("error = %q, want substring %q", err, "line 1")
	}
}

func TestCombineFewerThanK(t *testing.T) {
	secret := []byte("hello world")
	var splitOut bytes.Buffer

	split := &SplitCommand{Threshold: 3, Shares: 5, Scheme: "sss"}
	if err := split.run(bytes.NewReader(secret), &splitOut, nil); err != nil {
		t.Fatalf("split: %v", err)
	}

	// Take only 2 shares (less than k=3) — should produce wrong result, not error
	lines := strings.Split(strings.TrimSpace(splitOut.String()), "\n")
	tooFew := strings.Join(lines[:2], "\n") + "\n"

	var combineOut bytes.Buffer
	combine := &CombineCommand{}
	if err := combine.run(strings.NewReader(tooFew), &combineOut); err != nil {
		t.Fatalf("combine should not error with < k shares: %v", err)
	}

	if bytes.Equal(combineOut.Bytes(), secret) {
		t.Error("fewer than k shares should NOT reconstruct the correct secret")
	}
}
