// Package polynomial implements polynomial arithmetic over a finite field.
//
// Polynomials are represented as a slice of field.Element coefficients in
// ascending degree order: coefficients[i] is the coefficient of x^i.
//
// The key operations for secret sharing are:
//   - Evaluate: compute f(x) using Horner's method
//   - Random: generate a random polynomial with a fixed constant term
//   - LagrangeZero: Lagrange interpolation evaluated at x=0 only
//
// See: BIBLIOGRAPHY.md — Shoup Ch. 12, 17.
package polynomial

import (
	"fmt"
	"math/big"

	"github.com/aalpar/shamir/pkg/field"
)

// Polynomial over a finite field with coefficients in ascending degree order.
type Polynomial struct {
	Coefficients []field.Element
}

// Point is an (x, y) pair representing a polynomial evaluation.
type Point struct {
	X, Y field.Element
}

// New creates a polynomial from the given coefficients (ascending degree).
func New(coefficients []field.Element) Polynomial {
	return Polynomial{Coefficients: coefficients}
}

// Random generates a random polynomial of the given degree with f(0) = secret.
// The remaining coefficients are uniformly random over the field.
func Random(degree int, secret field.Element, f *field.Field) (Polynomial, error) {
	if degree < 0 {
		return Polynomial{}, fmt.Errorf("polynomial: negative degree %d", degree)
	}
	coeffs := make([]field.Element, degree+1)
	coeffs[0] = secret
	for i := 1; i <= degree; i++ {
		c, err := f.Rand()
		if err != nil {
			return Polynomial{}, fmt.Errorf("polynomial: generating coefficient %d: %w", i, err)
		}
		coeffs[i] = c
	}
	return Polynomial{Coefficients: coeffs}, nil
}

// Evaluate computes f(x) using Horner's method.
//
// Horner's method evaluates a_0 + a_1*x + ... + a_n*x^n as
// a_0 + x*(a_1 + x*(a_2 + ... + x*a_n)). This is O(n) multiplications
// vs. O(n^2) for naive evaluation.
func (p Polynomial) Evaluate(x field.Element) field.Element {
	n := len(p.Coefficients)
	if n == 0 {
		return x.Field().Zero()
	}
	result := p.Coefficients[n-1]
	for i := n - 2; i >= 0; i-- {
		result = result.Mul(x).Add(p.Coefficients[i])
	}
	return result
}

// Degree returns the degree of the polynomial.
func (p Polynomial) Degree() int {
	return len(p.Coefficients) - 1
}

// LagrangeZero reconstructs f(0) from a set of points using Lagrange
// interpolation evaluated at x = 0 only.
//
// For points (x_1, y_1), ..., (x_k, y_k), the Lagrange basis polynomial
// L_j evaluated at 0 simplifies to:
//
//	L_j(0) = ∏_{m≠j} x_m / (x_m - x_j)
//
// and f(0) = Σ y_j * L_j(0).
//
// This is more efficient than full polynomial reconstruction when only
// the constant term is needed (the SSS case).
//
// See: BIBLIOGRAPHY.md — Shamir 1979, Shoup Ch. 17.
func LagrangeZero(points []Point) (field.Element, error) {
	k := len(points)
	if k == 0 {
		return field.Element{}, fmt.Errorf("polynomial: no points for interpolation")
	}

	f := points[0].X.Field()
	result := f.Zero()

	for j := 0; j < k; j++ {
		num := f.One()
		den := f.One()
		for m := 0; m < k; m++ {
			if m == j {
				continue
			}
			num = num.Mul(points[m].X)                    // x_m
			den = den.Mul(points[m].X.Sub(points[j].X))   // x_m - x_j
		}
		if den.IsZero() {
			return field.Element{}, fmt.Errorf("polynomial: duplicate x-coordinate at index %d", j)
		}
		lagrange := num.Mul(den.Inv())
		result = result.Add(points[j].Y.Mul(lagrange))
	}

	return result, nil
}

// EvaluateAt evaluates the polynomial at x = 1, 2, ..., n, returning
// the resulting points. Useful for generating shares.
func (p Polynomial) EvaluateAt(n int, f *field.Field) []Point {
	points := make([]Point, n)
	for i := 1; i <= n; i++ {
		x := f.NewElement(big.NewInt(int64(i)))
		points[i-1] = Point{X: x, Y: p.Evaluate(x)}
	}
	return points
}
