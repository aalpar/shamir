# CRDT Structure in Proactive Share Refresh

Date: 2026-03-03

## Core Observation

Proactive share refresh (`pkg/refresh/`) is algebraically a CRDT operation.
The `Apply` function computes `new_y = old_y + Σ delta_i` over GF(p) — field
addition is commutative, associative, and has identity element zero. Contributions
from different participants can be applied in any order with the same result.

No published work explicitly frames proactive share refresh as a CRDT. The
concepts exist in separate literatures (distributed systems vs. threshold
cryptography). The algebraic structure is the same: commutative operations
over a group with unique operation identifiers.

## CRDT ↔ Shamir Refresh Mapping

| CRDT concept     | Shamir refresh analogue                                      |
|------------------|--------------------------------------------------------------|
| State            | A share (x, y) where y is the current polynomial evaluation  |
| Operation        | A zero-sharing delta δ_i from participant i                  |
| Merge            | y' = y + Σ δ_i (addition in GF(p))                          |
| Commutativity    | δ_a + δ_b = δ_b + δ_a (field addition commutes)             |
| Idempotence      | Detectable via Feldman commitment (not algebraic idempotence) |
| Convergence      | All holders applying same delta set → shares on same polynomial |

The share state doesn't form a natural semilattice (addition isn't idempotent).
But with Feldman commitments tagging each contribution, you get exactly-once
delivery semantics — each delta has a verifiable identity, and duplicates can
be detected and rejected. This is the same pattern as an OR-Set CRDT: each
operation has a unique tag, you track which tags you've applied, merge is
"apply all tags you haven't seen."

## Implications for Coordination

### Current Model (Synchronous Epochs)

The state machine assumes synchronized refresh rounds:

```
Active → Refreshing → Active
```

All holders enter "Refreshing" together, all contributions arrive, all shares
update, then back to "Active." This is the Herzberg et al. (1995) model.

### Proposed Model (Asynchronous, CRDT-Style)

```
Active (continuously accepting refresh contributions)
  ├── Contribution received → verify via Feldman, apply if new, track tag
  ├── All n contributions for epoch E received → epoch E complete
  ├── Old commitment updated → publish new commitment
  └── Degraded (< k holders reachable)
```

No "Refreshing" phase. No global synchronization point. Each shareholder
independently:

1. Generates a `VerifiableZeroSharing` contribution
2. Broadcasts it (or a coordinator distributes it)
3. Each holder applies contributions as they arrive (in any order)
4. Feldman verification prevents tampering
5. The coordinator tracks which contributions each holder has applied

### Checkpoint / Stable Cut

The coordinator computes "minimum epoch fully applied across all holders" — the
component-wise minimum across all holders' contribution sets. Below that cut,
old Feldman commitments can be garbage collected. This is a stable cut in the
causal graph: the latest point where all replicas have converged.

Detecting the stable cut requires knowing what every holder has applied. The
coordination layer reads each shareholder's state — this is cheap coordination
(read-only) that enables garbage collection.

### Privacy-Correctness Trade-off

Cachin et al. (2002) identified an inherent tension in asynchronous refresh:
you can guarantee privacy or correctness, but not both if the adversary can
delay messages indefinitely. In bounded-delay environments (not truly
adversarial async networks), this trade-off is manageable.

## Required Changes to `pkg/refresh/`

1. **Contribution identity** — `Contribution` struct needs contributor ID and
   epoch number so holders can track "which deltas have I applied?"
2. **Applied-set tracking** — each holder maintains a set of applied
   contribution tags (contributor, epoch) to prevent double-application
3. **No changes to `Apply` itself** — the math is already order-independent

## Key References

### Asynchronous Proactive Secret Sharing

- Cachin, Kursawe, Lysyanskaya, Strobl (2002). "Asynchronous Verifiable
  Secret Sharing and Proactive Cryptosystems." CCS '02.
  <https://eprint.iacr.org/2002/134.pdf>
  First practical async VSS + proactive refresh. Identifies the
  privacy-correctness trade-off in async settings.

- Zhou, Schneider, van Renesse (2005). "APSS: Proactive Secret Sharing in
  Asynchronous Systems." ACM TISSEC, vol 8, no 3.
  <https://dl.acm.org/doi/10.1145/1085126.1085127>
  Full async proactive refresh protocol. More robust against DoS than
  synchronous PSS.

- Levrat, Rambaud, Urban (2022). "Breaking the t<n/3 Consensus Bound:
  Asynchronous Dynamic Proactive Secret Sharing under Honest Majority."
  IACR ePrint 2022/619.
  <https://eprint.iacr.org/2022/619>
  Key result: refresh works without consensus on a unique set of shares.
  Parties only need a common (partial) view of committed shares — eventual
  consistency suffices for correctness. Achieves PSS for any t < n.

- Günther, Das, Kokoris-Kogias (2022). "Practical Asynchronous Proactive
  Secret Sharing and Key Refresh." IACR ePrint 2022/1586.
  <https://eprint.iacr.org/2022/1586>
  Two practical constructions: O(n³) and O(cn²) communication.

- Yurek, Luo, Gu, Kate (2023). "Robust Asynchronous DPSS and its
  Applications." USENIX Security '23.
  <https://www.usenix.org/system/files/usenixsecurity23-yurek.pdf>
  Robust async DPSS with dynamic committees. Shows synchronous DPSS suffers
  O(n³)→O(n⁴) degradation during async periods.

- Hu et al. (2025). "Optimistic Asynchronous Dynamic-committee Proactive
  Secret Sharing." IACR ePrint 2025/880.
  <https://eprint.iacr.org/2025/880>
  First async DPSS at O(n²) messages (optimistic), O(n³) worst case.

### CRDTs (Background)

- Shapiro, Preguiça, Baquero, Zawirski (2011). "A comprehensive study of
  Convergent and Commutative Replicated Data Types." INRIA TR 7506.
  The canonical CRDT formalization: join semilattice = commutative +
  associative + idempotent merge.

### Secure CRDTs (Adjacent Work)

- Tsaloli, Bai, Nergaard (2023). "General-Purpose Secure Conflict-free
  Replicated Data Types." IACR ePrint 2023/584.
  <https://eprint.iacr.org/2023/584.pdf>
  Goes the opposite direction — uses MPC to secure CRDTs, not CRDT structure
  within cryptographic protocols.
