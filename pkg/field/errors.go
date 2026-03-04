package field

import "errors"

// ErrRandom indicates that cryptographic random generation failed.
var ErrRandom = errors.New("random generation failed")

// Error represents an error from the field package.
type Error struct {
	Op     string // function name: "Rand"
	Kind   error  // sentinel for errors.Is
	Detail string // human-readable context
	Err    error  // wrapped cause
}

func (e *Error) Error() string {
	s := "field: " + e.Op
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
