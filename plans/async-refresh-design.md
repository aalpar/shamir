# Async Refresh Design — Research State

Date: 2026-03-04
Status: In progress

## Problem

The current refresh model (`pkg/refresh/`) is synchronous: all holders
enter a "Refreshing" phase, all contributions arrive, all shares update.
The `crdt-async-refresh.md` research doc established that refresh is
algebraically a CmRDT (commutative operations over GF(p)). The question
is how to upgrade `pkg/refresh/` to support async contribution application
without breaking the existing API or introducing a coordination dependency.

## What Exists

`Contribution` currently bundles `[]SubShare` + `*vss.Commitment`. It has
no identity — you can't tell which participant generated it or for which
epoch. `Apply` is already order-independent (field addition commutes).
The math doesn't change; only metadata and tracking do.

## Required Changes (from crdt-async-refresh.md)

1. **Contribution identity**: Add `ContributorID` and `Epoch` to
   `Contribution` struct
2. **Applied-set tracking**: Each holder tracks `set[(ContributorID, Epoch)]`
   to prevent double-application
3. **Stable cut**: Coordinator computes min epoch fully applied across
   all holders (garbage collection point for old commitments)

## Open Questions

### Q1: Where does applied-set tracking live?

Option A: Inside `pkg/refresh/` — a `HolderState` struct that wraps a
share and tracks applied contributions. Pro: self-contained. Con: refresh
package now has state, not just functions.

Option B: In `threshold/` — the `Holder` domain type already has an `ID`
and `Ready` flag, could track applied contributions. Pro: keeps crypto
packages stateless. Con: threshold package needs to understand refresh
internals.

Option C: At the application layer (caller's responsibility). Pro: no
changes to library packages. Con: every caller reimplements tracking.

**Leaning toward A**: a `HolderState` in `pkg/refresh/` that wraps a
`sss.Share` and manages applied-set. The crypto math stays in pure
functions; HolderState is the coordination adapter.

### Q2: What's a ContributorID?

In the crypto layer, participants are identified by their field element
X (evaluation point). Is `ContributorID = X`? Or is it a separate
identifier (e.g., a string or uint)?

Using X directly: natural, no new types, but conflates "which evaluation
point is this share at" with "who generated this contribution."

Using a separate ID: cleaner separation, but adds a type that the field
package doesn't know about.

**Decision needed**: X is the simplest option. A participant's contribution
is identified by `(X, Epoch)` where X is the contributor's evaluation
point.

### Q3: Epoch semantics

Is an epoch a monotonic counter? A timestamp? Who increments it?

The sync model has implicit epochs (each refresh round = one epoch). In
async, epochs are explicit. The simplest model: epochs are uint64 counters.
The coordinator (or any participant) can "open" a new epoch. Contributions
reference the epoch they belong to. A contribution is valid for exactly
one epoch.

### Q4: Stable cut computation

The stable cut is `min_over_all_holders(max_epoch_fully_applied)`. This
requires knowing each holder's applied-set. Two options:

- Push: holders report their state to coordinator
- Pull: coordinator queries holders

This is a coordination concern, not a crypto concern. The library should
provide the primitives (query a holder's applied set, compute the cut)
but not dictate the communication pattern.

## Paper Analysis

### Levrat et al. 2022 — Key Takeaways

- Refresh does NOT require consensus on a unique set of shares
- Parties only need a common *partial* view of committed shares
- Eventual consistency suffices for correctness
- Achieves PSS for any t < n (not just t < n/3)
- **Implication**: no "Refreshing" phase needed in state machine

### Cachin et al. 2002 — Privacy-Correctness Trade-off

- In truly async networks, you can't guarantee both privacy and correctness
- If adversary can delay messages indefinitely: either secret is lost
  (privacy preserved) or learned (correctness preserved)
- Bounded-delay environments sidestep this
- **Implication**: library should document the assumption about bounded
  delays; the trade-off is a deployment concern, not a protocol concern

### Zhou et al. 2005 — APSS Protocol Structure

- Full async proactive refresh: inherently more robust against DoS
- Node failures ≡ async message delays
- **TODO**: extract the contribution/verification/application phases
  and compare with our existing Contribution → VerifySubShare → Apply flow

### Günther et al. 2022 — Practical Constructions

- O(n³) basic construction, O(cn²) with sortition
- **TODO**: compare communication patterns with our coordinator model
- Question: does the sortition approach make sense for small n?

### Yurek et al. 2023 — Async DPSS

- Shows sync DPSS degrades O(n³) → O(n⁴) during async periods
- Dynamic committee support (shareholders can join/leave)
- **TODO**: read the committee change protocol for Phase 5

### Hu et al. 2025 — Optimistic Async DPSS

- O(n²) optimistic, O(n³) worst case
- State of the art efficiency
- **TODO**: read and compare with our simpler model

## Implementation Sketch

```go
// In pkg/refresh/

// ContributionTag uniquely identifies a contribution.
type ContributionTag struct {
    ContributorX field.Element
    Epoch        uint64
}

// Contribution adds identity fields.
type Contribution struct {
    Tag        ContributionTag
    SubShares  []SubShare
    Commitment *vss.Commitment
}

// HolderState tracks a holder's share and applied contributions.
type HolderState struct {
    share   sss.Share
    applied map[ContributionTag]struct{}
}

// ApplyContribution applies a contribution if not already applied.
// Returns (applied bool, err error).
func (h *HolderState) ApplyContribution(sub SubShare, tag ContributionTag) (bool, error) {
    if _, seen := h.applied[tag]; seen {
        return false, nil // idempotent
    }
    refreshed, err := Apply(h.share, []SubShare{sub})
    if err != nil {
        return false, err
    }
    h.share = refreshed
    h.applied[tag] = struct{}{}
    return true, nil
}
```

## Next Steps

1. Read Zhou 2005 and Günther 2022 for protocol structure comparison
2. Decide Q1-Q4 (leaning toward answers above)
3. Prototype HolderState with tests
4. Verify backward compatibility: existing tests should pass unchanged
