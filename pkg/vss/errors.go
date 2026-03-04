package vss

import "errors"

var (
	// ErrThreshold indicates k is out of the valid range [2, n].
	ErrThreshold = errors.New("threshold out of range")

	// ErrNotPrime indicates p is not prime.
	ErrNotPrime = errors.New("p is not prime")

	// ErrNotSafePrime indicates p is not a safe prime (q = (p-1)/2 is not prime).
	ErrNotSafePrime = errors.New("p is not a safe prime")

	// ErrGeneratorRange indicates a generator is outside the valid range (1, p).
	ErrGeneratorRange = errors.New("generator out of range")

	// ErrGeneratorOrder indicates a generator does not have order q in Z_p*.
	ErrGeneratorOrder = errors.New("generator does not have order q")

	// ErrGeneratorsEqual indicates g and h are the same value.
	ErrGeneratorsEqual = errors.New("generators must be distinct")

	// ErrNoSecondGenerator indicates Pedersen VSS was attempted without h.
	ErrNoSecondGenerator = errors.New("pedersen requires a second generator")

	// ErrUnmarshal indicates invalid serialized data.
	ErrUnmarshal = errors.New("invalid serialized data")

	// ErrNoGroup indicates a commitment has no associated group.
	ErrNoGroup = errors.New("commitment has no group")
)

// Error represents an error from the vss package.
type Error struct {
	Op     string // function name: "NewGroup", "FeldmanDeal", etc.
	Kind   error  // sentinel for errors.Is
	Detail string // human-readable context
	Err    error  // wrapped cause
}

func (e *Error) Error() string {
	s := "vss: " + e.Op
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
