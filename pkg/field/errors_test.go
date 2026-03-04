package field

import (
	"errors"
	"testing"
)

func TestErrorIs(t *testing.T) {
	err := &Error{Op: "Rand", Kind: ErrRandom, Err: errors.New("entropy source failed")}

	if !errors.Is(err, ErrRandom) {
		t.Error("errors.Is(err, ErrRandom) = false, want true")
	}
}

func TestErrorAs(t *testing.T) {
	err := &Error{Op: "Rand", Kind: ErrRandom, Err: errors.New("entropy source failed")}

	var fe *Error
	if !errors.As(err, &fe) {
		t.Fatal("errors.As failed")
	}
	if fe.Op != "Rand" {
		t.Errorf("Op = %q, want %q", fe.Op, "Rand")
	}
}

func TestErrorMessage(t *testing.T) {
	tests := []struct {
		name string
		err  *Error
		want string
	}{
		{
			name: "kind only",
			err:  &Error{Op: "Rand", Kind: ErrRandom},
			want: "field: Rand: random generation failed",
		},
		{
			name: "kind and cause",
			err:  &Error{Op: "Rand", Kind: ErrRandom, Err: errors.New("blocked")},
			want: "field: Rand: random generation failed: blocked",
		},
		{
			name: "op only",
			err:  &Error{Op: "Rand"},
			want: "field: Rand",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestErrorUnwrapChain(t *testing.T) {
	inner := errors.New("hardware failure")
	err := &Error{Op: "Rand", Kind: ErrRandom, Err: inner}

	if !errors.Is(err, inner) {
		t.Error("errors.Is(err, inner) = false, want true")
	}
	if !errors.Is(err, ErrRandom) {
		t.Error("errors.Is(err, ErrRandom) = false, want true")
	}
}
