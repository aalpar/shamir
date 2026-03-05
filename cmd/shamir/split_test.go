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
	"encoding/json"
	"math/big"
	"strings"
	"testing"

	"github.com/aalpar/shamir/pkg/sss"
	"github.com/jessevdk/go-flags"
)

func TestSplitRunEmptyStdin(t *testing.T) {
	cmd := &SplitCommand{Threshold: 3, Shares: 5, Scheme: "sss"}
	err := cmd.run(bytes.NewReader(nil), nil, nil)
	if err == nil {
		t.Fatal("expected error for empty stdin")
	}
	if !strings.Contains(err.Error(), "empty secret") {
		t.Errorf("error = %q, want substring %q", err, "empty secret")
	}
}

func TestSplitRunNotYetImplemented(t *testing.T) {
	cmd := &SplitCommand{Threshold: 3, Shares: 5, Scheme: "feldman"}
	var stdout bytes.Buffer
	err := cmd.run(bytes.NewReader([]byte("hello")), &stdout, nil)
	if err == nil {
		t.Fatal("expected 'not yet implemented' error")
	}
	if !strings.Contains(err.Error(), "not yet implemented") {
		t.Errorf("error = %q, want substring %q", err, "not yet implemented")
	}
}

func TestSplitFlagParsing(t *testing.T) {
	var cmd SplitCommand
	parser := flags.NewParser(&cmd, flags.Default & ^flags.PrintErrors)
	_, err := parser.ParseArgs([]string{"-k", "3", "-n", "5", "--scheme", "feldman"})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if cmd.Threshold != 3 {
		t.Errorf("threshold = %d, want 3", cmd.Threshold)
	}
	if cmd.Shares != 5 {
		t.Errorf("shares = %d, want 5", cmd.Shares)
	}
	if cmd.Scheme != "feldman" {
		t.Errorf("scheme = %q, want feldman", cmd.Scheme)
	}
	if cmd.Bits != 2048 {
		t.Errorf("bits = %d, want 2048", cmd.Bits)
	}
}

func TestSplitFlagDefaults(t *testing.T) {
	var cmd SplitCommand
	parser := flags.NewParser(&cmd, flags.Default & ^flags.PrintErrors)
	_, err := parser.ParseArgs([]string{"-k", "2", "-n", "3"})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if cmd.Scheme != "sss" {
		t.Errorf("scheme = %q, want default %q", cmd.Scheme, "sss")
	}
	if cmd.Bits != 2048 {
		t.Errorf("bits = %d, want default 2048", cmd.Bits)
	}
}

func TestSplitInvalidScheme(t *testing.T) {
	var cmd SplitCommand
	parser := flags.NewParser(&cmd, flags.Default & ^flags.PrintErrors)
	_, err := parser.ParseArgs([]string{"-k", "3", "-n", "5", "--scheme", "bogus"})
	if err == nil {
		t.Fatal("expected error for invalid scheme")
	}
}

func TestSplitMissingRequired(t *testing.T) {
	var cmd SplitCommand
	parser := flags.NewParser(&cmd, flags.Default & ^flags.PrintErrors)
	_, err := parser.ParseArgs([]string{"--scheme", "sss"})
	if err == nil {
		t.Fatal("expected error for missing required flags -k and -n")
	}
}

func TestSplitSSS(t *testing.T) {
	secret := []byte("hello world")
	var stdout, stderr bytes.Buffer

	cmd := &SplitCommand{Threshold: 3, Shares: 5, Scheme: "sss"}
	err := cmd.run(bytes.NewReader(secret), &stdout, &stderr)
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	if len(lines) != 5 {
		t.Fatalf("got %d lines, want 5", len(lines))
	}

	// Verify each line is a valid sss.Share
	shares := make([]sss.Share, 5)
	for i, line := range lines {
		if err := json.Unmarshal([]byte(line), &shares[i]); err != nil {
			t.Fatalf("share %d: unmarshal: %v", i, err)
		}
	}

	// Combine any 3 shares and verify we get the original secret back
	recovered, err := sss.Combine(shares[:3])
	if err != nil {
		t.Fatalf("Combine: %v", err)
	}

	original := new(big.Int).SetBytes(secret)
	if recovered.Value().Cmp(original) != 0 {
		t.Errorf("recovered %s, want %s", recovered.Value(), original)
	}

	// Stderr should be empty for SSS (no commitment)
	if stderr.Len() != 0 {
		t.Errorf("stderr should be empty for SSS, got: %s", stderr.String())
	}
}
