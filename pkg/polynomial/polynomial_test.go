// Copyright 2026 Aaron Alpar
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package polynomial

import (
	"math/big"
	"testing"

	"github.com/aalpar/shamir/pkg/field"
)

var p17 = big.NewInt(17)

func TestEvaluate(t *testing.T) {
	f := field.New(p17)

	// f(x) = 3 + 2x + x^2  (over GF(17))
	poly := New([]field.Element{
		f.NewElement(big.NewInt(3)),
		f.NewElement(big.NewInt(2)),
		f.NewElement(big.NewInt(1)),
	})

	tests := []struct {
		x    int64
		want int64
	}{
		{0, 3},  // f(0) = 3
		{1, 6},  // f(1) = 3 + 2 + 1 = 6
		{2, 11}, // f(2) = 3 + 4 + 4 = 11
		{3, 1},  // f(3) = 3 + 6 + 9 = 18 ≡ 1
		{4, 11}, // f(4) = 3 + 8 + 16 = 27 ≡ 10... let me compute: 3+8+16=27, 27 mod 17 = 10
	}
	// Fix: f(4) = 3 + 2*4 + 4^2 = 3 + 8 + 16 = 27 mod 17 = 10
	tests[4] = struct {
		x    int64
		want int64
	}{4, 10}

	for _, tt := range tests {
		x := f.NewElement(big.NewInt(tt.x))
		got := poly.Evaluate(x)
		if got.Value().Int64() != tt.want {
			t.Errorf("f(%d) = %s, want %d", tt.x, got, tt.want)
		}
	}
}

func TestEvaluateConstant(t *testing.T) {
	f := field.New(p17)

	// Constant polynomial f(x) = 7.
	poly := New([]field.Element{f.NewElement(big.NewInt(7))})

	for _, x := range []int64{0, 1, 5, 16} {
		got := poly.Evaluate(f.NewElement(big.NewInt(x)))
		if got.Value().Int64() != 7 {
			t.Errorf("constant poly at x=%d = %s, want 7", x, got)
		}
	}
}

func TestDegree(t *testing.T) {
	f := field.New(p17)

	tests := []struct {
		name   string
		coeffs []int64
		want   int
	}{
		{"constant", []int64{5}, 0},
		{"linear", []int64{1, 2}, 1},
		{"quadratic", []int64{1, 2, 3}, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			elems := make([]field.Element, len(tt.coeffs))
			for i, c := range tt.coeffs {
				elems[i] = f.NewElement(big.NewInt(c))
			}
			poly := New(elems)
			if poly.Degree() != tt.want {
				t.Errorf("Degree() = %d, want %d", poly.Degree(), tt.want)
			}
		})
	}
}

func TestRandom(t *testing.T) {
	f := field.New(p17)
	secret := f.NewElement(big.NewInt(7))

	poly, err := Random(2, secret, f)
	if err != nil {
		t.Fatalf("Random() error: %v", err)
	}

	// f(0) must equal the secret.
	got := poly.Evaluate(f.Zero())
	if !got.Equal(secret) {
		t.Errorf("f(0) = %s, want %s", got, secret)
	}

	// Degree must be 2.
	if poly.Degree() != 2 {
		t.Errorf("Degree() = %d, want 2", poly.Degree())
	}
}

func TestRandomNegativeDegree(t *testing.T) {
	f := field.New(p17)
	_, err := Random(-1, f.One(), f)
	if err == nil {
		t.Error("expected error for negative degree")
	}
}

func TestLagrangeZero(t *testing.T) {
	f := field.New(p17)

	// f(x) = 5 + 3x + x^2 over GF(17).
	// f(0) = 5. Verify by interpolating from 3 points.
	poly := New([]field.Element{
		f.NewElement(big.NewInt(5)),
		f.NewElement(big.NewInt(3)),
		f.NewElement(big.NewInt(1)),
	})

	// Evaluate at x = 1, 2, 3.
	points := poly.EvaluateAt(3, f)

	// Verify: f(1) = 5+3+1 = 9, f(2) = 5+6+4 = 15, f(3) = 5+9+9 = 23 ≡ 6
	expectations := []int64{9, 15, 6}
	for i, want := range expectations {
		if points[i].Y.Value().Int64() != want {
			t.Fatalf("f(%d) = %s, want %d", i+1, points[i].Y, want)
		}
	}

	// Reconstruct f(0) via Lagrange interpolation.
	got, err := LagrangeZero(points)
	if err != nil {
		t.Fatalf("LagrangeZero() error: %v", err)
	}
	if got.Value().Int64() != 5 {
		t.Errorf("LagrangeZero() = %s, want 5", got)
	}
}

func TestLagrangeZeroLinear(t *testing.T) {
	f := field.New(p17)

	// f(x) = 10 + 3x. f(0) = 10.
	poly := New([]field.Element{
		f.NewElement(big.NewInt(10)),
		f.NewElement(big.NewInt(3)),
	})

	// Need 2 points for degree-1 polynomial.
	points := poly.EvaluateAt(2, f)

	got, err := LagrangeZero(points)
	if err != nil {
		t.Fatalf("LagrangeZero() error: %v", err)
	}
	if got.Value().Int64() != 10 {
		t.Errorf("LagrangeZero() = %s, want 10", got)
	}
}

func TestLagrangeZeroDuplicateX(t *testing.T) {
	f := field.New(p17)

	points := []Point{
		{X: f.NewElement(big.NewInt(1)), Y: f.NewElement(big.NewInt(5))},
		{X: f.NewElement(big.NewInt(1)), Y: f.NewElement(big.NewInt(7))},
	}

	_, err := LagrangeZero(points)
	if err == nil {
		t.Error("expected error for duplicate x-coordinates")
	}
}

func TestLagrangeZeroEmpty(t *testing.T) {
	_, err := LagrangeZero(nil)
	if err == nil {
		t.Error("expected error for empty points")
	}
}

func TestEvaluateAt(t *testing.T) {
	f := field.New(p17)

	// f(x) = 2 + x
	poly := New([]field.Element{
		f.NewElement(big.NewInt(2)),
		f.NewElement(big.NewInt(1)),
	})

	points := poly.EvaluateAt(4, f)
	if len(points) != 4 {
		t.Fatalf("EvaluateAt(4) returned %d points, want 4", len(points))
	}

	// x values should be 1, 2, 3, 4.
	for i, pt := range points {
		wantX := int64(i + 1)
		wantY := (2 + wantX) % 17
		if pt.X.Value().Int64() != wantX {
			t.Errorf("point[%d].X = %s, want %d", i, pt.X, wantX)
		}
		if pt.Y.Value().Int64() != wantY {
			t.Errorf("point[%d].Y = %s, want %d", i, pt.Y, wantY)
		}
	}
}

func TestLagrangeRoundtripRandom(t *testing.T) {
	f := field.New(p17)

	// Generate a random degree-3 polynomial 10 times, interpolate f(0)
	// from 4 points each time.
	for range 10 {
		secret, err := f.Rand()
		if err != nil {
			t.Fatalf("Rand() error: %v", err)
		}

		poly, err := Random(3, secret, f)
		if err != nil {
			t.Fatalf("Random() error: %v", err)
		}

		points := poly.EvaluateAt(4, f)
		got, err := LagrangeZero(points)
		if err != nil {
			t.Fatalf("LagrangeZero() error: %v", err)
		}
		if !got.Equal(secret) {
			t.Errorf("roundtrip failed: got %s, want %s", got, secret)
		}
	}
}
