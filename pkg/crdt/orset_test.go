package crdt

import (
	"math/big"
	"testing"

	"github.com/aalpar/shamir/pkg/field"
)

var p17 = big.NewInt(17)

func TestEmptySet(t *testing.T) {
	f := field.New(p17)
	s := New(f, 1)

	if s.Lookup(f.NewElement(big.NewInt(5))) {
		t.Error("empty set should not contain any element")
	}
	if len(s.Elements()) != 0 {
		t.Errorf("Elements() = %d elements, want 0", len(s.Elements()))
	}
}

func TestAddLookup(t *testing.T) {
	f := field.New(p17)
	s := New(f, 1)

	five := f.NewElement(big.NewInt(5))
	seven := f.NewElement(big.NewInt(7))

	s.Add(five)
	if !s.Lookup(five) {
		t.Error("added element not found")
	}
	if s.Lookup(seven) {
		t.Error("non-added element found")
	}
}

func TestAddRemove(t *testing.T) {
	f := field.New(p17)
	s := New(f, 1)

	five := f.NewElement(big.NewInt(5))
	s.Add(five)
	s.Remove(five)

	if s.Lookup(five) {
		t.Error("removed element should not be found")
	}
}

func TestAddRemoveAdd(t *testing.T) {
	f := field.New(p17)
	s := New(f, 1)

	five := f.NewElement(big.NewInt(5))
	s.Add(five)
	s.Remove(five)
	s.Add(five)

	if !s.Lookup(five) {
		t.Error("re-added element should be found (new tag)")
	}
}

func TestConcurrentAddRemove(t *testing.T) {
	f := field.New(p17)

	// Replica 1 adds element 5.
	r1 := New(f, 1)
	five := f.NewElement(big.NewInt(5))
	r1.Add(five)

	// Replica 2 independently adds then removes element 5.
	r2 := New(f, 2)
	r2.Add(five)
	r2.Remove(five)

	merged := Merge(r1, r2)
	if !merged.Lookup(five) {
		t.Error("concurrent add should win over remove (add-wins semantics)")
	}
}

func TestMergeCommutativity(t *testing.T) {
	f := field.New(p17)
	a := New(f, 1)
	b := New(f, 2)

	a.Add(f.NewElement(big.NewInt(3)))
	a.Add(f.NewElement(big.NewInt(7)))
	b.Add(f.NewElement(big.NewInt(5)))
	b.Add(f.NewElement(big.NewInt(7)))

	ab := Merge(a, b)
	ba := Merge(b, a)
	if !equal(ab, ba) {
		t.Error("Merge(a,b) ≠ Merge(b,a) — commutativity violated")
	}
}

func TestMergeAssociativity(t *testing.T) {
	f := field.New(p17)
	a := New(f, 1)
	b := New(f, 2)
	c := New(f, 3)

	a.Add(f.NewElement(big.NewInt(2)))
	b.Add(f.NewElement(big.NewInt(5)))
	b.Remove(f.NewElement(big.NewInt(5)))
	c.Add(f.NewElement(big.NewInt(9)))

	ab_c := Merge(Merge(a, b), c)
	a_bc := Merge(a, Merge(b, c))
	if !equal(ab_c, a_bc) {
		t.Error("Merge(Merge(a,b),c) ≠ Merge(a,Merge(b,c)) — associativity violated")
	}
}

func TestMergeIdempotence(t *testing.T) {
	f := field.New(p17)
	a := New(f, 1)

	a.Add(f.NewElement(big.NewInt(4)))
	a.Add(f.NewElement(big.NewInt(11)))
	a.Remove(f.NewElement(big.NewInt(4)))

	aa := Merge(a, a)
	if !equal(a, aa) {
		t.Error("Merge(a,a) ≠ a — idempotence violated")
	}
}

func TestCrossFieldPanic(t *testing.T) {
	f1 := field.New(p17)
	f2 := field.New(big.NewInt(19))

	s := New(f1, 1)
	e := f2.NewElement(big.NewInt(3))

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on cross-field Add")
		}
	}()
	s.Add(e)
}

func TestMultipleElements(t *testing.T) {
	f := field.New(p17)
	s := New(f, 1)

	vals := []int64{1, 3, 5, 7, 11}
	for _, v := range vals {
		s.Add(f.NewElement(big.NewInt(v)))
	}

	elems := s.Elements()
	if len(elems) != len(vals) {
		t.Errorf("Elements() has %d elements, want %d", len(elems), len(vals))
	}

	for _, v := range vals {
		e := f.NewElement(big.NewInt(v))
		if !s.Lookup(e) {
			t.Errorf("element %d not found", v)
		}
	}
}

func TestRemoveNonExistent(t *testing.T) {
	f := field.New(p17)
	s := New(f, 1)

	three := f.NewElement(big.NewInt(3))
	five := f.NewElement(big.NewInt(5))

	s.Add(three)
	s.Remove(five) // no-op

	if !s.Lookup(three) {
		t.Error("removing non-existent element should not affect other elements")
	}
	if s.Lookup(five) {
		t.Error("removed non-existent element should not appear")
	}
}
