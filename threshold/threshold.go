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
