# TODO

Items ordered by dependency. Later phases depend on earlier ones.
Research state for complex portions saved in `plans/`.

---

## Phase 0: Hardening & Serialization

Foundation work that unblocks the CLI and exposes gaps in existing code.

### 0.1 Serialization Format

JSON serialization with hex-encoded big.Int values.

- [x] Design serialization for `sss.Share` (X, Y as big.Int) — `pkg/sss/json.go`
- [x] Design serialization for `vss.Commitment` ([]big.Int) — `pkg/vss/json.go`
- [x] Design serialization for `vss.PedersenShare` — `pkg/vss/json.go`
- [ ] Design serialization for `refresh.Contribution` ([]SubShare + Commitment)
- [x] Decide format: JSON with hex strings for first pass
- [ ] Implement `encoding.TextMarshaler`/`TextUnmarshaler` on key types
- [x] Tests: round-trip serialization for `sss.Share` — `pkg/sss/json_test.go`
- [x] Tests: malformed input rejection for `sss.Share` — `pkg/sss/json_test.go`
- [ ] Tests: round-trip and malformed input for `vss.Commitment`, `vss.PedersenShare`

**Decision**: field elements are `math/big.Int` — use hex encoding to
avoid base-10 ambiguity at large sizes. Commitment values are also
big.Int. JSON with hex strings is the simplest first pass.

**Decision**: `Commitment` serialization embeds group parameters (`p`, `q`,
`g`, `h`) for standalone verifiability. Revisit when a protocol layer exists
— in a real deployment, group parameters would be shared out-of-band and
embedding them in every commitment is verbose.
- [ ] Revisit Commitment serialization: embed group params vs. out-of-band
      when a protocol layer is added (phase 4+ or any network protocol)

### 0.2 threshold/ Domain Type Tests

Tests in `threshold/threshold_test.go`.

- [x] Test Config validation: k >= 2, k <= n, valid Scheme
- [x] Test Phase transitions: define which transitions are legal
- [x] Test Holder: ID uniqueness, Ready semantics

### 0.3 Error Types

Per-package `Error` struct + sentinel vars in `errors.go` files.
Callers use `errors.Is` for sentinels, `errors.As` for typed detail.
`Unwrap() []error` returns both Kind (sentinel) and Err (wrapped cause).

- [x] Define error types in each package — `pkg/*/errors.go`
- [x] Replace `fmt.Errorf` with typed errors at package boundaries
- [x] Tests: verify error types with `errors.Is` / `errors.As` — `pkg/*/errors_test.go`

---

## Phase 1: CLI Implementation (`cmd/shamir/`)

Wire up the three stubbed subcommands using existing crypto packages.
Depends on Phase 0.1 (serialization).

### 1.1 `split` Command

- [ ] Parse flags: `--threshold/-k`, `--shares/-n`, `--scheme` (sss|feldman|pedersen)
- [ ] Read secret from stdin (or `--secret` flag for testing)
- [ ] Generate safe prime + generator for Feldman/Pedersen schemes
- [ ] Call `sss.Split` or `vss.FeldmanDeal` / `vss.PedersenDeal`
- [ ] Serialize and output shares (one per line? one per file?)
- [ ] For VSS schemes: also output commitment
- [ ] Tests: split → combine round-trip via CLI
- [ ] Tests: invalid flag combinations

### 1.2 `combine` Command

- [ ] Parse flags: share inputs (files or stdin)
- [ ] Deserialize shares
- [ ] Call `sss.Combine`
- [ ] Output reconstructed secret
- [ ] Tests: combine with exactly k shares, with n shares, with < k shares (error)

### 1.3 `verify` Command

- [ ] Parse flags: share input, commitment input
- [ ] Deserialize share + commitment + group parameters
- [ ] Call `vss.FeldmanVerify` or `vss.PedersenVerify`
- [ ] Output pass/fail
- [ ] Tests: valid share, tampered share

### 1.4 CLI Integration Tests

- [ ] End-to-end: `split | combine` pipeline
- [ ] End-to-end: `split --scheme feldman`, then `verify` each share
- [ ] Error messages are human-readable and actionable

---

## Phase 2: Async Refresh Upgrade (`pkg/refresh/`)

**Research state**: `plans/async-refresh-design.md`
**Spec**: `docs/research/crdt-async-refresh.md`

### 2.1 Research (Papers to Read)

Papers are in BIBLIOGRAPHY.md. Focus areas annotated in GROUNDWORK.md §2.

- [ ] Cachin et al. (2002) — privacy-correctness trade-off in async
      Focus: bounded-delay assumption, protocol phases
- [ ] Zhou et al. (2005) — APSS protocol structure
      Focus: contribution/verification/application phases, compare with
      our existing Contribution → VerifySubShare → Apply flow
- [ ] Günther et al. (2022) — practical constructions
      Focus: O(n³) vs O(cn²), communication patterns
- [ ] Hu et al. (2025) — optimistic async DPSS
      Focus: O(n²) optimistic protocol, comparison with our model

Levrat et al. (2022) already skimmed — key result (no consensus needed)
captured in crdt-async-refresh.md.

### 2.2 Implementation

`Apply` math is already order-independent — no changes needed there.

- [ ] Add `ContributionTag` struct: `(ContributorX field.Element, Epoch uint64)`
- [ ] Add `Tag` field to `Contribution` struct
- [ ] Add `HolderState` struct wrapping `sss.Share` + applied-set
- [ ] Implement `HolderState.ApplyContribution` with idempotent semantics
- [ ] Implement `StableCut` function (min epoch fully applied across holders)
- [ ] Preserve backward compatibility: existing `Apply` function stays,
      `HolderState` is the new stateful layer on top

### 2.3 Testing

- [ ] Test: applying same contribution twice is idempotent (no change)
- [ ] Test: contributions applied in different orders produce same share
- [ ] Test: stable cut advances as all holders apply contributions
- [ ] Test: stable cut doesn't advance if one holder is behind
- [ ] Test: full async scenario — contributions arrive out of order,
      different holders apply different subsets, final reconstruct works
- [ ] Test: epoch rollover (uint64 overflow? probably just document)
- [ ] Regression: all existing refresh tests still pass unchanged

---

## Phase 3: Threat Model Formalization

Gates DKG decision and async refresh security claims.
Mostly reading and writing, not code.

- [ ] Re-read Herzberg et al. (1995) threat model sections
      Focus: mobile adversary model, epoch semantics, what "corrupted" means
- [ ] Write threat model document (`docs/threat-model.md`)
  - [ ] Define "compromised shareholder" precisely
  - [ ] Define coordinator trust level (trusted? semi-honest? untrusted?)
  - [ ] State the security claim: "attacker who compromises < k shareholders
        within a single refresh interval learns nothing about the secret"
  - [ ] Enumerate what the library does NOT protect against
        (side channels, malicious coordinator, etc.)
  - [ ] Document the bounded-delay assumption for async refresh
- [ ] Review threshold/ state machine against threat model
      (does Degraded → Active make sense? Under what conditions?)
- [ ] Note: Degraded and Failed are currently terminal states with no
      outgoing edges. Worth revisiting here — you might want Degraded → Active
      if a degraded holder comes back online and re-shares. Right now they're
      intentionally dead ends; the test in threshold/ locks that in so you'll
      notice if you change it.

---

## Phase 4: Distributed Key Generation (`pkg/dkg/`)

**Research state**: `plans/dkg-compatibility.md`

Depends on Phase 3 (threat model determines if DKG is required).

### 4.1 Research

- [ ] Read Gennaro et al. (1999) in full
      Focus: complaint protocol, disqualification logic
- [ ] Verify existing Pedersen VSS primitives are sufficient building blocks
- [ ] Determine if complaint protocol can be stateless (pure function)
      or requires round management

### 4.2 Implementation (if required)

- [ ] Create `pkg/dkg/` package (depends on `pkg/vss/`)
- [ ] Implement `Round1`: each participant generates and distributes Pedersen VSS
- [ ] Implement `Round2`: complaint verification, disqualification
- [ ] Implement `Finalize`: aggregate shares and commitments
- [ ] Extract combined public key (for threshold signature use cases)

### 4.3 Testing

- [ ] Test: all-honest case — n participants, all behave correctly
- [ ] Test: one malicious dealer — complaints filed, dealer disqualified,
      remaining participants reconstruct correct shared secret
- [ ] Test: output shares are valid Shamir shares (combine any k → secret)
- [ ] Test: output commitment is valid Feldman commitment (all shares verify)
- [ ] Test: DKG secret matches Σ s_i (where s_i are individual contributions)

---

## Phase 5: Dynamic Committees

Shareholder addition/removal. Depends on async refresh design (Phase 2).

### 5.1 Research

- [ ] Read Yurek et al. (2023) — dynamic-committee (DPSS) sections
      Focus: resharing protocol, committee transition
- [ ] Read Hu et al. (2025) — dynamic-committee focus
      Focus: optimistic-case efficiency for committee changes
- [ ] Decide: is committee change a special case of refresh (reshare to
      new polynomial with different evaluation points) or a separate protocol?

### 5.2 Implementation

Blocked on 5.1 decisions.

- [ ] Implement share redistribution (old committee → new committee)
- [ ] Handle threshold changes (k changes along with n)
- [ ] Commitment update for new committee
- [ ] Integration with async refresh (committee changes during a refresh epoch)

### 5.3 Testing

- [ ] Test: add one shareholder — old shares + new share reconstruct
- [ ] Test: remove one shareholder — remaining shares still reconstruct
- [ ] Test: change threshold (k=3,n=5 → k=4,n=7)
- [ ] Test: committee change during active refresh epoch

---

## 6 Other

- [ ] &Error constructors
- [ ] Put error sentinal names (Op:) in constants

## Future (Not Planned)

These appear in BIBLIOGRAPHY.md but are explicitly out of scope.

- **Threshold signatures** (BLS, threshold ECDSA) — requires elliptic
  curve arithmetic, significant new dependency
- **General access structures** (quorum-based, crumbling wall) — share
  sizes blow up exponentially, research-grade territory
- **Constant-time arithmetic** — `math/big` is not constant-time;
  documented as future optimization, not blocking
