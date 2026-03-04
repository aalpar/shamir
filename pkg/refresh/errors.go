package refresh

import "errors"

var (
	// ErrThreshold indicates k is out of the valid range [2, n].
	ErrThreshold = errors.New("threshold out of range")

	// ErrXMismatch indicates a sub-share's x-coordinate doesn't match the target share.
	ErrXMismatch = errors.New("sub-share x-coordinate mismatch")

	// ErrCommitmentLength indicates delta commitments have inconsistent lengths.
	ErrCommitmentLength = errors.New("commitment length mismatch")
)

// Error represents an error from the refresh package.
type Error struct {
	Op     string // function name: "ZeroSharing", "Apply", etc.
	Kind   error  // sentinel for errors.Is
	Detail string // human-readable context
	Err    error  // wrapped cause
}

func (e *Error) Error() string {
	s := "refresh: " + e.Op
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
