# Groundwork

Things to read and digest before making design decisions. Each entry
exists because it unblocks a specific architectural choice in this project.
Not a general reading list — every item has a "why this matters for Shamir"
annotation.

Sections are ordered by dependency — later sections depend on decisions
made in earlier ones.

Status key: `[ ]` unread, `[~]` skimmed, `[x]` digested.

---

## 1. Threat Model

Need to formalize what "security" means for a K8s threshold secret
operator before making protocol choices. Everything else depends on this:
the threat model determines whether DKG is required, whether async refresh
is safe, and what "compromised" means.

- [ ] **Herzberg et al. (1995)** — already read for `pkg/refresh/`, but
  re-read the threat model sections specifically
  <https://doi.org/10.1007/3-540-44750-4_27>
  *Why:* The original proactive security model assumes a mobile
  adversary that can corrupt up to t servers per epoch. How does this
  map to K8s? Is a compromised node a "corrupted server"? What about
  a compromised controller?

### Decision this unblocks

What are the actual security guarantees the operator provides? "An
attacker who compromises fewer than k ShareHolders within a single
refresh interval learns nothing about the secret" — is that the claim?
What about an attacker who compromises the controller?

---

## 2. Asynchronous Proactive Refresh

The operator's refresh model (synchronous epochs vs. async contributions)
is the biggest open design question. These papers determine whether we need
a "Refreshing" phase in the state machine at all.

Depends on: **Threat Model** — the privacy-correctness trade-off in async
settings (Cachin 2002) only matters once we know our threat model. A K8s
environment with bounded delays may sidestep the impossibility result, but
we need to verify that against our specific threat assumptions.

- [ ] **Cachin et al. (2002)** — "Asynchronous Verifiable Secret Sharing
  and Proactive Cryptosystems"
  <https://eprint.iacr.org/2002/134.pdf>
  *Why:* First async VSS + proactive refresh. Identifies the
  privacy-correctness trade-off — can we guarantee both in a K8s
  environment with bounded delays? Need to understand whether our
  threat model sidesteps the impossibility result.

- [ ] **Zhou, Schneider, van Renesse (2005)** — "APSS: Proactive Secret
  Sharing in Asynchronous Systems"
  <https://dl.acm.org/doi/10.1145/1085126.1085127>
  *Why:* Full async PSS protocol. Maps directly to the K8s failure model
  (pod evictions, node drains = async message delays). Need to extract
  the protocol structure and see what maps to controller-runtime
  reconciliation loops.

- [~] **Levrat, Rambaud, Urban (2022)** — "Breaking the t<n/3 Consensus
  Bound: APSS under Honest Majority"
  <https://eprint.iacr.org/2022/619>
  *Why:* The key result. Proves refresh works without consensus on a
  unique set of shares — eventual consistency suffices. This is the
  paper that justifies dropping the synchronous "Refreshing" phase.
  Need to understand the exact conditions under which this holds and
  whether K8s satisfies them.

- [ ] **Günther, Das, Kokoris-Kogias (2022)** — "Practical Asynchronous
  PSS and Key Refresh"
  <https://eprint.iacr.org/2022/1586>
  *Why:* Practical constructions at O(n³) and O(cn²). Need to evaluate
  whether the communication patterns fit a K8s operator model (controller
  as relay vs. peer-to-peer between pods).

- [ ] **Yurek et al. (2023)** — "Robust Asynchronous DPSS"
  <https://www.usenix.org/system/files/usenixsecurity23-yurek.pdf>
  *Why:* Shows synchronous DPSS degrades O(n³)→O(n⁴) during async
  periods. Directly argues against our current synchronous model.
  Also handles dynamic committees — relevant if ShareHolders can join
  and leave.

- [ ] **Hu et al. (2025)** — "Optimistic Asynchronous DPSS"
  <https://eprint.iacr.org/2025/880>
  *Why:* State-of-the-art efficiency (O(n²) optimistic). If we're
  going async, this is the protocol to benchmark against.

### Decision this unblocks

Should the operator use synchronous refresh epochs (current design:
`Active → Refreshing → Active`) or async contribution-based refresh
(CRDT-style: holders apply verified deltas as they arrive, controller
tracks progress)?

See: `docs/research/crdt-async-refresh.md` for the preliminary analysis.

---

## 3. CRDTs and Distributed Consistency

The connection between proactive refresh and CRDTs. Understanding this
informs how the controller tracks refresh progress and when old
commitments can be garbage-collected.

Depends on: **Async Proactive Refresh** — this section only matters if
we go async. If we stay synchronous, it's interesting but not load-bearing.

- [~] **Shapiro et al. (2011)** — "A comprehensive study of Convergent
  and Commutative Replicated Data Types" (INRIA TR 7506)
  *Why:* The CRDT formalization. Need to verify whether share refresh
  satisfies CvRDT or CmRDT axioms precisely, and where the analogy
  breaks (idempotence via Feldman tags vs. algebraic idempotence).

- [~] **Helland (2007)** — "Life beyond Distributed Transactions: an
  Apostate's Opinion" (CIDR)
  *Why:* The "apologize rather than ask permission" framing. Directly
  applicable: instead of preventing partial refresh (synchronous
  rounds), detect and compensate. Need to understand the compensation
  model for failed partial refreshes.

### Decision this unblocks

How does the controller detect "all holders have applied all contributions
for epoch E" (the stable cut / checkpoint) without requiring synchronous
coordination? What metadata does each ShareHolder need to track?

---

## 4. Distributed Key Generation (DKG)

The operator needs a key ceremony where no single party sees the plaintext
secret. Already in BIBLIOGRAPHY.md but needs deeper study.

Depends on: **Threat Model** — whether the controller is trusted determines
whether DKG is required or just nice-to-have.

- [ ] **Gennaro et al. (1999)** — "Secure Distributed Key Generation for
  Discrete-Log Based Cryptosystems"
  <https://doi.org/10.1007/3-540-48910-X_21>
  *Why:* The controller generating and splitting the secret defeats
  threshold protection. DKG removes the trusted dealer. Need to
  understand whether the protocol works with our Feldman/Pedersen VSS
  primitives as building blocks, or requires new algebra.

### Decision this unblocks

Does the operator's initial secret splitting require a trusted dealer
(controller sees plaintext), or can we do DKG where the controller
orchestrates without seeing the secret?

---

## 5. Dynamic Committees

If ShareHolders can join and leave (pod scaling, node replacement), the
refresh protocol needs to handle committee changes.

Depends on: **Async Proactive Refresh** — committee changes during refresh
interact with the refresh protocol. Can't design this in isolation.

- [ ] **Yurek et al. (2023)** — same paper as above, but focus on the
  DPSS (dynamic-committee) aspects
  *Why:* The K8s operator will inevitably face ShareHolder churn. Need
  to understand whether share redistribution during committee changes
  can be folded into the refresh protocol or requires a separate
  mechanism.

- [~] **Hu et al. (2025)** — same paper as above, dynamic-committee
  focus
  *Why:* Optimistic-case efficiency for committee changes.

### Decision this unblocks

Can ShareHolder addition/removal be modeled as a special case of refresh
(reshare to new polynomial with different evaluation points), or does it
require a separate protocol?
