# Shamir

## Versioning

- Shamir is at v0.1.0 with zero consumers. Break freely — no stability guarantees until real users exist.
- API group: `threshold.shamir.io/v1alpha1`

## Architecture

Two layers: a zero-dependency crypto library (`pkg/`) and a Kubernetes operator (`internal/` + `api/`).

### Layer 1: Crypto Library (`pkg/`)

Zero external dependencies. Backed by `math/big` from stdlib.

| Package | Purpose |
|---------|---------|
| `pkg/field/` | GF(p) arithmetic: Add, Sub, Mul, Inv, Exp, Neg, Rand |
| `pkg/polynomial/` | Evaluation (Horner), Lagrange interpolation, random generation |
| `pkg/sss/` | Shamir's Secret Sharing: Split, Combine |
| `pkg/vss/` | Verifiable Secret Sharing: Feldman, Pedersen |
| `pkg/refresh/` | Proactive share refresh |

Dependency graph: `refresh → vss → sss → polynomial → field`

### Layer 2: Kubernetes Operator (`internal/` + `api/`)

Two CRDs: `ThresholdSecret` (orchestrator) and `ShareHolder` (share-holding entity).

State machine:

```
Pending → Splitting → Active → Refreshing → Active
                        ↓          ↓
                     Degraded   Failed
```

### Key Design Decisions

- `Element` carries a reference to its `Field` for runtime invariant checking
- `Inv` uses Extended Euclidean algorithm
- Lagrange interpolation optimized for f(0) evaluation only
- `math/big` is not constant-time; noted as future optimization target
- `pkg/` importable by external Go projects without pulling K8s dependencies

## Package Map

| Package | Key Types | Purpose |
|---------|-----------|---------|
| `pkg/field/` | Field, Element | Finite field arithmetic |
| `pkg/polynomial/` | Polynomial, Point | Polynomial operations |
| `pkg/sss/` | Share | Shamir's Secret Sharing |
| `pkg/vss/` | Commitment, VerifiableShare | Verifiable Secret Sharing |
| `pkg/refresh/` | RefreshShare | Proactive share refresh |
| `api/v1alpha1/` | ThresholdSecret, ShareHolder | CRD definitions |
| `internal/controller/` | ThresholdSecretReconciler | State machine, orchestration |
| `cmd/controller/` | main | Controller manager entry point |
| `cmd/shamir/` | main | CLI: split, combine, verify |

## Testing

- `pkg/` — standard `testing` with table-driven tests, known vectors, property checks
- `internal/controller/` — Ginkgo v2 + envtest (when operator is built)
- **After changes**: `make lint && make && make test`

## Commits

- Direct push to master is fine at this stage
