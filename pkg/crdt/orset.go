// Package crdt implements conflict-free replicated data types over finite
// field elements.
package crdt

import (
	"maps"

	"github.com/aalpar/shamir/pkg/field"
)

// Tag uniquely identifies an Add operation by its originating replica and
// a monotonically increasing sequence number within that replica.
type Tag struct {
	Replica uint64
	Seq     uint64
}

// ORSet is an Observed-Remove Set over field.Element values.
//
// Each Add generates a globally unique Tag. Remove tombstones all active
// tags for a given element. Merge computes the semilattice join of two
// replicas: union of entries, union of tombstones, then effective set =
// entries \ tombstones.
//
// The "add-wins" semantics emerge from the data structure: a concurrent
// Add on another replica creates a *different* tag that the local Remove
// never saw, so it survives the merge.
type ORSet struct {
	field   *field.Field
	replica uint64
	seq     uint64
	entries map[Tag]field.Element // add set
	removed map[Tag]struct{}      // tombstone set
}

// New creates an empty ORSet bound to the given field and replica identity.
func New(f *field.Field, replica uint64) *ORSet {
	return &ORSet{
		field:   f,
		replica: replica,
		entries: make(map[Tag]field.Element),
		removed: make(map[Tag]struct{}),
	}
}

// sameField panics if e belongs to a different field than the set.
func (s *ORSet) sameField(e field.Element) {
	if e.Field() != s.field {
		panic("crdt: element from different field")
	}
}

// Add inserts an element into the set with a fresh unique tag.
func (s *ORSet) Add(e field.Element) {
	s.sameField(e)
	s.seq++
	tag := Tag{Replica: s.replica, Seq: s.seq}
	s.entries[tag] = e
}

// Remove tombstones all active (non-removed) tags that currently map to e.
// Removing an element not in the set is a no-op.
func (s *ORSet) Remove(e field.Element) {
	s.sameField(e)
	for tag, elem := range s.entries {
		if elem.Equal(e) {
			if _, ok := s.removed[tag]; !ok {
				s.removed[tag] = struct{}{}
			}
		}
	}
}

// Lookup reports whether e is in the effective set (has at least one
// non-tombstoned tag).
func (s *ORSet) Lookup(e field.Element) bool {
	s.sameField(e)
	for tag, elem := range s.entries {
		if elem.Equal(e) {
			if _, ok := s.removed[tag]; !ok {
				return true
			}
		}
	}
	return false
}

// Elements returns the deduplicated effective set (entries minus tombstones).
func (s *ORSet) Elements() []field.Element {
	seen := make(map[string]struct{})
	var result []field.Element
	for tag, elem := range s.entries {
		if _, ok := s.removed[tag]; ok {
			continue
		}
		key := elem.String()
		if _, dup := seen[key]; !dup {
			seen[key] = struct{}{}
			result = append(result, elem)
		}
	}
	return result
}

// Merge computes the semilattice join of two ORSets, returning a new ORSet.
// Both sets must be bound to the same field.
//
// TODO: Implement this. The merge of two OR-Sets is:
//   - entries = union of both entries maps
//   - removed = union of both removed sets
//   - effective set = entries \ removed
//
// Think about why set union already gives you commutativity, associativity,
// and idempotence for free.
func Merge(a, b *ORSet) *ORSet {
	if a.field != b.field {
		panic("crdt: merge of sets from different fields")
	}
	out := &ORSet{
		field:   a.field,
		entries: make(map[Tag]field.Element, len(a.entries)+len(b.entries)),
		removed: make(map[Tag]struct{}, len(a.removed)+len(b.removed)),
	}
	maps.Copy(out.entries, a.entries)
	maps.Copy(out.entries, b.entries)
	maps.Copy(out.removed, a.removed)
	maps.Copy(out.removed, b.removed)
	return out
}

// equal reports whether two ORSets have the same effective set. Used by
// tests to verify semilattice properties.
func equal(a, b *ORSet) bool {
	ae := a.Elements()
	be := b.Elements()
	if len(ae) != len(be) {
		return false
	}
	// Build lookup from b's elements.
	bset := make(map[string]struct{}, len(be))
	for _, e := range be {
		bset[e.String()] = struct{}{}
	}
	for _, e := range ae {
		if _, ok := bset[e.String()]; !ok {
			return false
		}
	}
	return true
}
