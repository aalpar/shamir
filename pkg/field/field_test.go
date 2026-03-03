package field

import (
	"math/big"
	"testing"
)

// Small prime for readable tests.
var p17 = big.NewInt(17)

// Large prime for realistic tests (256-bit).
var p256, _ = new(big.Int).SetString(
	"FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEFFFFFC2F", 16,
) // secp256k1 prime

func TestNew(t *testing.T) {
	t.Run("accepts prime", func(t *testing.T) {
		f := New(p17)
		if f.Prime().Cmp(p17) != 0 {
			t.Errorf("Prime() = %s, want %s", f.Prime(), p17)
		}
	})

	t.Run("panics on composite", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic on composite input")
			}
		}()
		New(big.NewInt(15))
	})

	t.Run("panics on 1", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic on 1")
			}
		}()
		New(big.NewInt(1))
	})
}

func TestNewElement(t *testing.T) {
	f := New(p17)

	tests := []struct {
		name string
		in   int64
		want int64
	}{
		{"zero", 0, 0},
		{"in range", 5, 5},
		{"at prime", 17, 0},
		{"above prime", 20, 3},
		{"negative", -1, 16},
		{"large negative", -18, 16},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := f.NewElement(big.NewInt(tt.in))
			if e.Value().Int64() != tt.want {
				t.Errorf("NewElement(%d) = %s, want %d", tt.in, e, tt.want)
			}
		})
	}
}

func TestAdd(t *testing.T) {
	f := New(p17)

	tests := []struct {
		a, b, want int64
	}{
		{3, 4, 7},
		{10, 10, 3},   // 20 mod 17 = 3
		{0, 5, 5},
		{16, 1, 0},    // wraps to 0
		{16, 16, 15},  // 32 mod 17 = 15
	}

	for _, tt := range tests {
		a := f.NewElement(big.NewInt(tt.a))
		b := f.NewElement(big.NewInt(tt.b))
		got := a.Add(b)
		if got.Value().Int64() != tt.want {
			t.Errorf("%d + %d = %s, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestSub(t *testing.T) {
	f := New(p17)

	tests := []struct {
		a, b, want int64
	}{
		{7, 3, 4},
		{3, 7, 13},  // -4 mod 17 = 13
		{0, 1, 16},
		{0, 0, 0},
	}

	for _, tt := range tests {
		a := f.NewElement(big.NewInt(tt.a))
		b := f.NewElement(big.NewInt(tt.b))
		got := a.Sub(b)
		if got.Value().Int64() != tt.want {
			t.Errorf("%d - %d = %s, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestNeg(t *testing.T) {
	f := New(p17)

	tests := []struct {
		a, want int64
	}{
		{0, 0},
		{1, 16},
		{16, 1},
		{8, 9},
	}

	for _, tt := range tests {
		a := f.NewElement(big.NewInt(tt.a))
		got := a.Neg()
		if got.Value().Int64() != tt.want {
			t.Errorf("-%d = %s, want %d", tt.a, got, tt.want)
		}
	}
}

func TestMul(t *testing.T) {
	f := New(p17)

	tests := []struct {
		a, b, want int64
	}{
		{3, 4, 12},
		{5, 5, 8},    // 25 mod 17 = 8
		{0, 10, 0},
		{1, 16, 16},
		{16, 16, 1},  // (-1)(-1) = 1
	}

	for _, tt := range tests {
		a := f.NewElement(big.NewInt(tt.a))
		b := f.NewElement(big.NewInt(tt.b))
		got := a.Mul(b)
		if got.Value().Int64() != tt.want {
			t.Errorf("%d * %d = %s, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestInv(t *testing.T) {
	f := New(p17)

	// For every nonzero element, a * a^{-1} = 1.
	for i := int64(1); i < 17; i++ {
		a := f.NewElement(big.NewInt(i))
		inv := a.Inv()
		product := a.Mul(inv)
		if !product.Equal(f.One()) {
			t.Errorf("%d * Inv(%d) = %s, want 1", i, i, product)
		}
	}
}

func TestInvPanicsOnZero(t *testing.T) {
	f := New(p17)
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on Inv(0)")
		}
	}()
	f.Zero().Inv()
}

func TestExp(t *testing.T) {
	f := New(p17)

	tests := []struct {
		base int64
		exp  int64
		want int64
	}{
		{2, 0, 1},
		{2, 1, 2},
		{2, 4, 16},
		{2, 5, 15},    // 32 mod 17 = 15
		{3, 16, 1},    // Fermat's little theorem: a^{p-1} = 1
		{5, 16, 1},
	}

	for _, tt := range tests {
		a := f.NewElement(big.NewInt(tt.base))
		got := a.Exp(big.NewInt(tt.exp))
		if got.Value().Int64() != tt.want {
			t.Errorf("%d^%d = %s, want %d", tt.base, tt.exp, got, tt.want)
		}
	}
}

func TestFermatLittleTheorem(t *testing.T) {
	// a^{p-1} ≡ 1 (mod p) for all a ≠ 0.
	f := New(p17)
	exp := big.NewInt(16) // p - 1

	for i := int64(1); i < 17; i++ {
		a := f.NewElement(big.NewInt(i))
		got := a.Exp(exp)
		if !got.Equal(f.One()) {
			t.Errorf("Fermat: %d^16 = %s, want 1", i, got)
		}
	}
}

func TestEqual(t *testing.T) {
	f := New(p17)
	a := f.NewElement(big.NewInt(5))
	b := f.NewElement(big.NewInt(5))
	c := f.NewElement(big.NewInt(6))

	if !a.Equal(b) {
		t.Error("equal elements not equal")
	}
	if a.Equal(c) {
		t.Error("different elements reported equal")
	}
}

func TestCrossFieldPanics(t *testing.T) {
	f1 := New(p17)
	f2 := New(big.NewInt(19))

	a := f1.NewElement(big.NewInt(3))
	b := f2.NewElement(big.NewInt(3))

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on cross-field operation")
		}
	}()
	a.Add(b)
}

func TestRand(t *testing.T) {
	f := New(p17)

	// Generate several random elements; all must be in [0, p).
	for range 100 {
		e, err := f.Rand()
		if err != nil {
			t.Fatalf("Rand() error: %v", err)
		}
		if e.Value().Sign() < 0 || e.Value().Cmp(p17) >= 0 {
			t.Errorf("Rand() = %s, out of range [0, %s)", e, p17)
		}
	}
}

func TestLargePrime(t *testing.T) {
	f := New(p256)

	// Verify basic properties hold with a 256-bit prime.
	a, err := f.Rand()
	if err != nil {
		t.Fatalf("Rand() error: %v", err)
	}
	b, err := f.Rand()
	if err != nil {
		t.Fatalf("Rand() error: %v", err)
	}

	// Commutativity: a + b = b + a
	if !a.Add(b).Equal(b.Add(a)) {
		t.Error("addition not commutative")
	}

	// Commutativity: a * b = b * a
	if !a.Mul(b).Equal(b.Mul(a)) {
		t.Error("multiplication not commutative")
	}

	// Inverse: a * a^{-1} = 1 (for nonzero a)
	if !a.IsZero() {
		inv := a.Inv()
		if !a.Mul(inv).Equal(f.One()) {
			t.Error("a * Inv(a) != 1")
		}
	}

	// Additive inverse: a + (-a) = 0
	if !a.Add(a.Neg()).Equal(f.Zero()) {
		t.Error("a + Neg(a) != 0")
	}
}
