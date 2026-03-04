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
// Package threshold defines domain types for threshold-protected secrets.
package threshold

import "fmt"

// Scheme identifies the secret sharing scheme.
type Scheme string

const (
	SSS      Scheme = "sss"
	Feldman  Scheme = "feldman"
	Pedersen Scheme = "pedersen"
)

// ValidSchemes is the set of recognized schemes.
var ValidSchemes = []Scheme{SSS, Feldman, Pedersen}

// Valid reports whether s is a recognized scheme.
func (s Scheme) Valid() bool {
	for _, v := range ValidSchemes {
		if s == v {
			return true
		}
	}
	return false
}

// Phase represents the lifecycle state of a threshold secret.
type Phase string

const (
	Pending    Phase = "pending"
	Splitting  Phase = "splitting"
	Active     Phase = "active"
	Refreshing Phase = "refreshing"
	Degraded   Phase = "degraded"
	Failed     Phase = "failed"
)

// validTransitions defines the legal phase transitions.
//
// State machine:
//
//	Pending → Splitting → Active → Refreshing → Active
//	                        ↓          ↓
//	                     Degraded   Failed
var validTransitions = map[Phase][]Phase{
	Pending:    {Splitting},
	Splitting:  {Active},
	Active:     {Refreshing, Degraded},
	Refreshing: {Active, Failed},
	Degraded:   {},
	Failed:     {},
}

// CanTransition reports whether moving from phase p to next is a legal
// transition in the threshold secret lifecycle.
func (p Phase) CanTransition(next Phase) bool {
	for _, allowed := range validTransitions[p] {
		if next == allowed {
			return true
		}
	}
	return false
}

// Config defines the parameters for a threshold-protected secret.
type Config struct {
	Threshold int    // k — minimum shares to reconstruct
	Shares    int    // n — total shares to generate
	Scheme    Scheme // secret sharing scheme
}

// Validate checks that the config is well-formed:
//   - Threshold >= 2 (a 1-of-n scheme is just replication)
//   - Threshold <= Shares
//   - Scheme is recognized
func (c Config) Validate() error {
	if c.Threshold < 2 {
		return fmt.Errorf("threshold: k=%d must be >= 2", c.Threshold)
	}
	if c.Threshold > c.Shares {
		return fmt.Errorf("threshold: k=%d exceeds n=%d", c.Threshold, c.Shares)
	}
	if !c.Scheme.Valid() {
		return fmt.Errorf("threshold: unrecognized scheme %q", c.Scheme)
	}
	return nil
}

// Holder represents a participant that holds a share.
type Holder struct {
	ID    string // unique identifier for this holder
	Ready bool   // can participate in share operations
}

// HoldersUnique reports whether all holders have distinct IDs.
func HoldersUnique(holders []Holder) bool {
	seen := make(map[string]struct{}, len(holders))
	for _, h := range holders {
		if _, dup := seen[h.ID]; dup {
			return false
		}
		seen[h.ID] = struct{}{}
	}
	return true
}
