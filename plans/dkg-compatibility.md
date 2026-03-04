# DKG Compatibility — Research State

Date: 2026-03-04
Status: Not started (blocked on threat model)

## Problem

The current Split workflow requires a trusted dealer who sees the plaintext
secret. For many use cases, this defeats the purpose of threshold
protection — if you trust one party with the secret, why split it?

Distributed Key Generation (DKG) removes the trusted dealer: all n
participants jointly generate a shared secret where no single party ever
knows the complete value.

## Key Reference

Gennaro, Jarecki, Krawczyk, Rabin (1999). "Secure Distributed Key
Generation for Discrete-Log Based Cryptosystems."

## Compatibility Question

Does Gennaro's DKG work with our existing Feldman/Pedersen VSS primitives?

### Preliminary Analysis (from BIBLIOGRAPHY annotation)

The protocol structure is:
1. Each participant acts as a dealer for a random secret s_i
2. Each participant runs Pedersen VSS to distribute shares of s_i
3. The final shared secret is S = Σ s_i
4. Each participant's final share is the sum of their shares from all n VSS instances

This uses Pedersen VSS as a building block. We already have `PedersenDeal`
and `PedersenVerify` in `pkg/vss/`. The question is whether the existing
API surface is sufficient or needs extensions.

### What We'd Need

1. **Pedersen VSS** — already exists (`pkg/vss/pedersen.go`)
2. **Share aggregation** — summing shares from multiple VSS instances;
   this is field addition, already exists
3. **Commitment aggregation** — combining commitments from multiple
   dealers; structurally identical to `UpdateCommitment` in `pkg/refresh/`
4. **Complaint protocol** — if a dealer's shares don't verify, the other
   participants file complaints and the dealer must reveal; this is NEW
5. **Disqualification** — parties who fail the complaint round are
   excluded from the final secret computation; this is NEW

### Gap Analysis

| Requirement | Status | Notes |
|-------------|--------|-------|
| Pedersen VSS deal | ✓ exists | `vss.PedersenDeal` |
| Pedersen VSS verify | ✓ exists | `vss.PedersenVerify` |
| Share summation | ✓ trivial | `field.Element.Add` |
| Commitment aggregation | ✓ exists | `refresh.UpdateCommitment` (or similar) |
| Complaint protocol | ✗ new | Needs round-trip communication model |
| Disqualification logic | ✗ new | Subset selection after complaints |
| Extract share from Pedersen | ✓ exists | `vss.PedersenToSSS` |

## Blocked On

The threat model (GROUNDWORK.md §1) must answer:
- Is the coordinator trusted? If yes, DKG is nice-to-have. If no, DKG
  is required for the system to make sense.
- Can the coordinator see the plaintext secret during initial splitting?

## TODO When Unblocked

1. Read Gennaro 1999 in full — focus on complaint protocol
2. Determine whether complaint protocol can be a pure function
   (verify complaint, return disqualify/accept) or requires stateful
   round management
3. Decide package placement: `pkg/dkg/` as new package depending on
   `pkg/vss/`? Fits the existing DAG.
4. Prototype with existing test helpers
