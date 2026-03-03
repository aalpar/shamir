# Bibliography

Foundational references for the threshold cryptography, secret sharing
schemes, finite field arithmetic, and verifiable protocols used in Shamir.

## Secret Sharing

Shamir, A. (1979). "How to Share a Secret." *Communications of the ACM*,
vol 22, no 11, pp. 612-613.
<https://doi.org/10.1145/359168.359176>

> The foundational paper. Defines the (k, n) threshold scheme: a secret is
> split into n shares such that any k shares reconstruct it, but k-1 shares
> reveal nothing. The construction uses polynomial interpolation over a
> finite field — a degree-(k-1) polynomial where the secret is f(0) and
> shares are evaluations at distinct nonzero points. This paper is the
> direct basis for `pkg/sss/`.

Blakley, G.R. (1979). "Safeguarding Cryptographic Keys." *Proceedings of
the National Computer Conference*, AFIPS, vol 48, pp. 313-317.

> Independently proposes secret sharing using geometry: k shares define k
> hyperplanes in k-dimensional space whose intersection is the secret point.
> Mathematically equivalent in capability to Shamir's scheme but less
> efficient in practice (shares are larger). Worth understanding as the
> "road not taken" — it clarifies why polynomial interpolation won: shares
> are single field elements rather than geometric coordinates.

## Verifiable Secret Sharing

Feldman, P. (1987). "A Practical Scheme for Non-interactive Verifiable
Secret Sharing." *Proceedings of the 28th Annual Symposium on Foundations
of Computer Science (FOCS)*, pp. 427-438.
<https://doi.org/10.1109/SFCS.1987.4>

> Extends Shamir's scheme so shareholders can verify their shares without
> interaction. The dealer publishes commitments C_j = g^{a_j} mod p for
> each polynomial coefficient a_j. A shareholder with share (i, y_i) checks
> that g^{y_i} = product of C_j^{i^j}. Computationally hiding — the
> commitments reveal nothing about the secret *if* the discrete log problem
> is hard. This is the basis for `pkg/vss/` Feldman mode.

Pedersen, T.P. (1991). "Non-Interactive and Information-Theoretic Secure
Verifiable Secret Sharing." *Advances in Cryptology — CRYPTO '91*, Lecture
Notes in Computer Science, vol 576, pp. 129-140.
<https://doi.org/10.1007/3-540-46766-1_9>

> Strengthens Feldman's scheme to information-theoretic hiding: commitments
> reveal nothing about the secret even against unbounded adversaries. Uses
> two generators g, h where log_g(h) is unknown, with commitments
> C_j = g^{a_j} * h^{b_j}. The trade-off is a second random polynomial
> and larger commitment structures. This is the basis for `pkg/vss/`
> Pedersen mode.

## Proactive Security

Herzberg, A., Jarecki, S., Krawczyk, H., and Yung, M. (1995). "Proactive
Secret Sharing Or: How to Cope With Perpetual Leakage." *Advances in
Cryptology — CRYPTO '95*, Lecture Notes in Computer Science, vol 963,
pp. 339-352.
<https://doi.org/10.1007/3-540-44750-4_27>

> Addresses the fundamental weakness of static secret sharing: over time,
> an adversary may compromise k shares. Proactive refresh generates new
> shares of the same secret periodically — each participant adds shares of
> a random zero-polynomial to their existing share. Old shares become
> useless. The security model shifts from "compromise fewer than k shares
> ever" to "compromise fewer than k shares per refresh period." This is the
> basis for `pkg/refresh/`.

## Asynchronous Proactive Secret Sharing

Cachin, C., Kursawe, K., Lysyanskaya, A., and Strobl, R. (2002).
"Asynchronous Verifiable Secret Sharing and Proactive Cryptosystems."
*Proceedings of the 9th ACM Conference on Computer and Communications
Security (CCS '02)*, pp. 88-97.
<https://eprint.iacr.org/2002/134.pdf>

> First practical verifiable secret sharing protocol for asynchronous
> networks, with proactive refresh. Identifies the fundamental
> privacy–correctness trade-off in async settings: the adversary can either
> cause the secret to be lost (preserving privacy) or learn it (preserving
> correctness), and this trade-off "seems unavoidable in asynchronous
> networks." Relevant to the operator layer: Kubernetes offers bounded
> network delays (not truly adversarial async), so the trade-off is
> manageable, but the protocol structure informs how to handle partial
> refresh failures. See also: `docs/research/crdt-async-refresh.md`.

Zhou, L., Schneider, F.B., and van Renesse, R. (2005). "APSS: Proactive
Secret Sharing in Asynchronous Systems." *ACM Transactions on Information
and System Security (TISSEC)*, vol 8, no 3, pp. 259-286.
<https://dl.acm.org/doi/10.1145/1085126.1085127>

> Full asynchronous proactive refresh protocol. Key advantage over
> synchronous PSS: inherently more robust against denial-of-service attacks
> that slow processor execution or delay message delivery. Tolerates attacks
> that synchronous PSS protocols cannot. Directly relevant to the operator:
> K8s pod evictions and node failures during refresh are equivalent to async
> message delays.

Levrat, L., Rambaud, M., and Urban, G. (2022). "Breaking the t<n/3
Consensus Bound: Asynchronous Dynamic Proactive Secret Sharing under Honest
Majority." *IACR ePrint Archive*, Report 2022/619.
<https://eprint.iacr.org/2022/619>

> Key result for async refresh: proactive share refresh does NOT require
> consensus on a unique set of shares. Parties only need a common (partial)
> view of committed shares — an eventual-consistency paradigm suffices for
> correctness. Achieves PSS for any threshold t < n with guaranteed output
> delivery in a completely asynchronous network. This paper bridges
> threshold cryptography and distributed systems consistency models. See
> also: `docs/research/crdt-async-refresh.md` for the CRDT connection.

Günther, D., Das, S., and Kokoris-Kogias, L. (2022). "Practical
Asynchronous Proactive Secret Sharing and Key Refresh." *IACR ePrint
Archive*, Report 2022/1586.
<https://eprint.iacr.org/2022/1586>

> Two practical constructions for async proactive secret sharing: O(n³)
> communication complexity and a sortition-based variant at O(cn²). Targets
> decentralized systems (blockchains) but the protocol structure applies to
> any distributed threshold key management, including K8s operators.

Yurek, T., Luo, L., Gu, J., and Kate, A. (2023). "Robust Asynchronous
DPSS and its Applications." *Proceedings of the 32nd USENIX Security
Symposium*, pp. 1537-1554.
<https://www.usenix.org/system/files/usenixsecurity23-yurek.pdf>

> Robust async dynamic-committee proactive secret sharing (DPSS). Shows
> that synchronous DPSS protocols (like COBRA) suffer asymptotic
> performance hits (O(n³) → O(n⁴)) during periods of asynchrony. Async
> protocols limit damage from network partitions and slow nodes. Relevant
> to the operator: K8s cluster upgrades, node drains, and network policies
> can cause transient async behavior.

Hu, J., et al. (2025). "Optimistic Asynchronous Dynamic-committee
Proactive Secret Sharing." *IACR ePrint Archive*, Report 2025/880.
<https://eprint.iacr.org/2025/880>

> First async DPSS protocol with O(n²) message complexity in all scenarios.
> Achieves O(λn²) communication in the optimistic case (all nodes honest,
> synchronous network) and O(λn³) worst case. Represents the current
> state-of-the-art in async proactive secret sharing efficiency.

## Distributed Key Generation

Gennaro, R., Jarecki, S., Krawczyk, H., and Rabin, T. (1999). "Secure
Distributed Key Generation for Discrete-Log Based Cryptosystems."
*Advances in Cryptology — EUROCRYPT '99*, Lecture Notes in Computer
Science, vol 1592, pp. 295-310.
<https://doi.org/10.1007/3-540-48910-X_21>

> Removes the trusted dealer assumption. Instead of one party generating
> and distributing shares, all participants jointly generate a shared secret
> where no single party ever knows the complete secret. Uses Pedersen VSS
> as a building block — each participant acts as a dealer for a random
> value, and the final shared secret is the sum. Relevant for the
> Kubernetes operator's key ceremony workflow, where trusting a single
> controller with the plaintext secret defeats the purpose of threshold
> protection.

## Number Theory Foundations

Hardy, G.H. and Wright, E.M. (2008). *An Introduction to the Theory of
Numbers*. 6th ed. Oxford University Press. ISBN 978-0-19-921986-5.

> Classical introduction to number theory. Chapters on congruences and
> modular arithmetic (Ch. 5-6), quadratic residues (Ch. 6), and the
> distribution of primes (Ch. 22) provide the mathematical foundations
> for finite field arithmetic. The Extended Euclidean Algorithm for
> modular inverse — used in `pkg/field/` — traces to the treatment of
> continued fractions and the Euclidean algorithm in Chapter 10.

Shoup, V. (2009). *A Computational Introduction to Number Theory and
Algebra*. 2nd ed. Cambridge University Press.
<https://shoup.net/ntb/>

> Free textbook bridging number theory and practical computation. Chapter 2
> (congruences) and Chapter 12 (polynomial arithmetic over finite fields)
> directly inform the implementations in `pkg/field/` and `pkg/polynomial/`.
> The treatment of Lagrange interpolation in the context of error-correcting
> codes (Chapter 17) provides useful intuition for why polynomial-based
> secret sharing is robust.

## Information-Theoretic Security

Shannon, C.E. (1949). "Communication Theory of Secrecy Systems." *Bell
System Technical Journal*, vol 28, no 4, pp. 656-715.
<https://doi.org/10.1002/j.1538-7305.1949.tb00928.x>

> Defines perfect secrecy: an encryption scheme is perfectly secret if
> the ciphertext reveals no information about the plaintext, regardless
> of computational power. Shamir's scheme achieves this — any k-1 shares
> are statistically independent of the secret. This is not just a
> historical curiosity; it's the specific security guarantee that
> distinguishes SSS from computational schemes. Pedersen VSS preserves
> this property for commitments; Feldman VSS does not.

## Quorum Systems and Access Structures

Peleg, D. and Wool, A. (1995). "Crumbling Walls Do Not Make Good Building
Material." Unpublished manuscript.

> Constructs quorum systems with optimal trade-offs between quorum size,
> availability, and load. The "crumbling wall" construction arranges
> elements in rows of varying widths; a quorum is one full row plus one
> representative from every row below. The CWlog system achieves O(lg n)
> quorum size with Condorcet availability (failure probability → 0 for
> any element failure probability p < 1/2). Directly relevant to the
> operator layer: the question "which subsets of shareholders can
> reconstruct the secret?" is an access structure, and crumbling wall
> quorums define topology-aware reconstruction policies richer than
> simple (k, n) thresholds. The availability analysis maps to
> "probability no valid quorum of shareholders survives," and the load
> analysis governs how often each shareholder is accessed during
> reconstruction.

Naor, M. and Wool, A. (1998). "Access Control and Signatures via Quorum
Secret Sharing." *IEEE Transactions on Parallel and Distributed Systems*,
vol 9, no 9, pp. 909-922.
<https://doi.org/10.1109/71.722221>

> The direct bridge between quorum systems and secret sharing. Defines
> quorum secret sharing: an access structure where the authorized subsets
> are exactly the quorums of a quorum system. This generalizes Shamir's
> (k, n) threshold to arbitrary quorum topologies — a crumbling wall
> quorum system, for instance, encodes "any full team of nodes plus
> one representative from each other team" as an access policy. If the
> operator evolves beyond simple thresholds, this is the theoretical
> framework.

### Note on General Access Structures

Shamir's polynomial interpolation gives optimal share size (one field
element) for (k, n) threshold access structures. For general access
structures — including crumbling wall quorums — constructions exist via
monotone span programs and monotone formulas, but share sizes blow up
exponentially. The best known upper bound is O(2^{0.892n}) (Liu and
Vaikuntanathan, STOC 2018, subsequently improved). The gap between upper
and lower bounds has been open for ~40 years. This means crumbling wall
access structures cannot reuse the polynomial machinery from `pkg/sss/`;
they would require a fundamentally different algebraic layer. This is
research-grade territory and out of scope for this project.

See: Beimel, A. (2011). "Secret-Sharing Schemes: A Survey." *Coding and
Cryptology*, Lecture Notes in Computer Science, vol 6639, pp. 11-46.
<https://link.springer.com/chapter/10.1007/978-3-642-20901-7_2>

## Threshold Signatures (Future Extensions)

Gennaro, R. and Goldfeder, S. (2020). "One Round Threshold ECDSA with
Identifiable Abort." *Cryptology ePrint Archive*, Report 2020/540.
<https://eprint.iacr.org/2020/540>

> Threshold ECDSA allowing k-of-n parties to jointly produce a valid
> ECDSA signature without any party knowing the full signing key.
> "Identifiable abort" means if a party misbehaves, others can identify
> who. Relevant for the signing use case: distributed certificate signing
> or admission webhook signing in Kubernetes without a single key holder.

Boneh, D., Lynn, B., and Shacham, H. (2001). "Short Signatures from
the Weil Pairing." *Advances in Cryptology — ASIACRYPT 2001*, Lecture
Notes in Computer Science, vol 2248, pp. 514-532.
<https://doi.org/10.1007/3-540-45682-1_30>

> Introduces BLS signatures based on bilinear pairings on elliptic
> curves. BLS signatures are naturally threshold-friendly — unlike
> ECDSA, threshold BLS requires no interactive protocol beyond share
> distribution, because signature shares combine linearly. A natural
> candidate for threshold signing in the Kubernetes operator.

## Conflict-Free Replicated Data Types (CRDTs)

Shapiro, M., Preguiça, N., Baquero, C., and Zawirski, M. (2011). "A
comprehensive study of Convergent and Commutative Replicated Data Types."
*INRIA Technical Report*, RR-7506.

> The canonical CRDT formalization. State-based CRDTs (CvRDTs) require a
> join semilattice: merge must be commutative, associative, and idempotent.
> Operation-based CRDTs (CmRDTs) require commutative operations. Proactive
> share refresh is structurally a CmRDT — zero-sharing deltas are
> commutative operations over GF(p). See `docs/research/crdt-async-refresh.md`
> for the detailed mapping between CRDT concepts and share refresh.

Tsaloli, G., Bai, S., and Nergaard, H. (2023). "General-Purpose Secure
Conflict-free Replicated Data Types." *IACR ePrint Archive*, Report
2023/584.
<https://eprint.iacr.org/2023/584.pdf>

> Goes the opposite direction from our use case: uses MPC to *secure*
> CRDTs, not CRDT structure *within* cryptographic protocols. Each CRDT
> replica is realized by a group of MPC parties. Included for completeness
> as the closest published intersection of CRDTs and threshold cryptography.

Helland, P. (2007). "Life beyond Distributed Transactions: an Apostate's
Opinion." *Proceedings of the 3rd Biennial Conference on Innovative Data
Systems Research (CIDR)*.

> Argues for "apologize rather than ask permission" in distributed systems:
> use compensation and detection rather than coordination and prevention.
> This framing — post-condition detection vs. pre-condition enforcement —
> directly applies to async proactive refresh: instead of enforcing
> synchronous refresh rounds, detect and compensate for partial application.

Kleppmann, M., Wiggins, A., van Hardenberg, P., and McGranaghan, M. (2019).
"Local-First Software: You Own Your Data." *Proceedings of the 2019 ACM
SIGPLAN International Symposium on New Ideas, New Paradigms, and Reflections
on Programming and Software (Onward!)*, pp. 154-178.

> The local-first software manifesto. Argues that data should live on the
> user's device with peer-to-peer sync, not on centralized servers. CRDTs
> are the enabling technology. Relevant context for understanding why CRDT
> concepts apply to distributed key management: both problems involve
> replicas making independent updates that must converge without central
> coordination.
