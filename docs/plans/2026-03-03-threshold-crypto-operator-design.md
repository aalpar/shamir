# Threshold Cryptography Operator — Design

**Status: DEFERRED** — K8s operator is out of scope for v0.x. Focus is on
the crypto library (`pkg/`) and CRDT-based async refresh. This plan is
archived for reference when the operator layer is revisited.

Date: 2026-03-03

## Overview

`github.com/aalpar/shamir` — a threshold cryptography library. The `pkg/`
layer is the current focus. The Kubernetes operator described below is a
future extension that would use the library to protect Secrets with k-of-n
threshold schemes.

## Architecture

### Layer 1: Crypto Library (`pkg/`)

Zero external dependencies. Backed by `math/big` from stdlib.

| Package        | Purpose                                              |
|----------------|------------------------------------------------------|
| `pkg/field/`       | GF(p) arithmetic: Add, Sub, Mul, Inv, Exp, Neg, Rand |
| `pkg/polynomial/`  | Evaluation (Horner), Lagrange interpolation, random generation |
| `pkg/sss/`         | Shamir's Secret Sharing: Split, Combine              |
| `pkg/vss/`         | Verifiable Secret Sharing: Feldman, Pedersen          |
| `pkg/refresh/`     | Proactive share refresh                               |

Dependency graph (strictly bottom-up):

    refresh → vss → sss → polynomial → field

Key design choices:
- `Element` carries a reference to its `Field` for runtime invariant checking
- `Inv` uses Extended Euclidean algorithm (not Fermat's little theorem)
- Lagrange interpolation optimized for f(0) evaluation only
- Configurable prime with well-known primes as exported constants
- `math/big` is not constant-time; noted as future optimization target

### Layer 2: Kubernetes Operator (`internal/` + `api/`)

API group: `threshold.shamir.io/v1alpha1`

Two CRDs:

**ThresholdSecret** — orchestrator CR defining what to protect:
- `spec.threshold` (k), `spec.shares` (n), `spec.scheme` (sss|feldman|pedersen)
- `spec.secretRef` — reference to the Kubernetes Secret to protect
- `spec.refresh.interval` — proactive refresh cadence
- `status.phase` — Pending | Splitting | Active | Refreshing | Degraded | Failed

**ShareHolder** — represents a share-holding entity:
- `spec.type` — Node | Pod | External
- `spec.storage` — where this holder's shares live
- Many-to-many relationship with ThresholdSecrets

State machine:

    Pending → Splitting → Active → Refreshing → Active
                            ↓          ↓
                         Degraded   Failed

Degraded = fewer than n but >= k holders reachable (recoverable, below redundancy).

### Build Order

1. Project scaffolding — module, Makefile, CI, lint, goreleaser
2. Crypto primitives (`pkg/`) — field, polynomial, sss, vss, refresh
3. CLI (`cmd/shamir/`) — split, combine, verify commands
4. Kubernetes operator (`api/`, `internal/controller/`, `cmd/controller/`)

## Infrastructure

Matches Janus patterns:
- Kubebuilder scaffold with placeholder CRDs
- Makefile: dev/build/release/kubernetes target groups
- golangci-lint v2, goreleaser, VERSION file
- GitHub Actions: lint.yml, test.yml, release.yml
- Apache 2.0 license
- Standard `testing` for `pkg/`, Ginkgo + envtest for operator tests
- Distroless container image

## Testing

- `pkg/` — table-driven tests with known vectors, edge cases, property checks
- `internal/controller/` — Ginkgo v2 + envtest (when operator is built)
- E2E — Kind cluster (when operator is built)

## Notes

- `math/big` is not constant-time; side-channel resistance deferred to optimization phase
- Share storage mechanism (where shares physically live) deferred to operator phase
- BIBLIOGRAPHY.md tracks foundational references; code comments reference back to it
