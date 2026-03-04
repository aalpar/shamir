package polynomial

import "errors"

var (
	// ErrNegativeDegree indicates a polynomial was requested with negative degree.
	ErrNegativeDegree = errors.New("negative degree")

	// ErrNoPoints indicates Lagrange interpolation was called with no points.
	ErrNoPoints = errors.New("no points for interpolation")

	// ErrDuplicateX indicates duplicate x-coordinates in interpolation input.
	ErrDuplicateX = errors.New("duplicate x-coordinate")
)

// Error represents an error from the polynomial package.
type Error struct {
	Op     string // function name: "Random", "LagrangeZero"
	Kind   error  // sentinel for errors.Is
	Detail string // human-readable context
	Err    error  // wrapped cause
}

func (e *Error) Error() string {
	s := "polynomial: " + e.Op
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
