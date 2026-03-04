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

// Scheme identifies the secret sharing scheme.
type Scheme string

const (
	SSS      Scheme = "sss"
	Feldman  Scheme = "feldman"
	Pedersen Scheme = "pedersen"
)

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

// Config defines the parameters for a threshold-protected secret.
type Config struct {
	Threshold int    // k — minimum shares to reconstruct
	Shares    int    // n — total shares to generate
	Scheme    Scheme // secret sharing scheme
}

// Holder represents a participant that holds a share.
type Holder struct {
	ID    string // unique identifier for this holder
	Ready bool   // can participate in share operations
}
