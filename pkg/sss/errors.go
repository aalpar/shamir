package sss

import "errors"

var (
	// ErrThreshold indicates k is out of the valid range [2, n].
	ErrThreshold = errors.New("threshold out of range")

	// ErrNoShares indicates no shares were provided for reconstruction.
	ErrNoShares = errors.New("no shares provided")

	// ErrUnmarshal indicates invalid serialized share data.
	ErrUnmarshal = errors.New("invalid serialized data")
)

// Error represents an error from the sss package.
type Error struct {
	Op     string // function name: "Split", "Combine", "UnmarshalJSON"
	Kind   error  // sentinel for errors.Is
	Detail string // human-readable context
	Err    error  // wrapped cause
}

func (e *Error) Error() string {
	s := "sss: " + e.Op
	if e.Kind != nil {
		s += ": " + e.Kind.Error()
	}
	if e.Detail != "" {
		s += ": " + e.Detail
	}
	if e.Err != nil {
		s += ": " + e.Err.Error()
	}
	return s
}

func (e *Error) Unwrap() []error {
	var errs []error
	if e.Kind != nil {
		errs = append(errs, e.Kind)
	}
	if e.Err != nil {
		errs = append(errs, e.Err)
	}
	return errs
}
