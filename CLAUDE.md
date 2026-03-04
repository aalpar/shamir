# Shamir

## Versioning

- Shamir is at v0.1.0 with zero consumers. Break freely — no stability guarantees until real users exist.

## Architecture

Two layers: a zero-dependency crypto library (`pkg/`) and platform-independent domain types (`threshold/`).

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

### Layer 2: Domain Types (`threshold/`)

Plain Go structs representing threshold secret lifecycle — no platform or deployment dependencies.

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

## Package Map

| Package | Key Types | Purpose |
|---------|-----------|---------|
| `pkg/field/` | Field, Element | Finite field arithmetic |
| `pkg/polynomial/` | Polynomial, Point | Polynomial operations |
| `pkg/sss/` | Share | Shamir's Secret Sharing |
| `pkg/vss/` | Commitment, VerifiableShare | Verifiable Secret Sharing |
| `pkg/refresh/` | RefreshShare | Proactive share refresh |
| `threshold/` | Config, Holder, Scheme, Phase | Threshold secret domain types |
| `cmd/shamir/` | main | CLI: split, combine, verify |

## Testing

- `go test ./...` — standard `testing` with table-driven tests, known vectors, property checks
- **After changes**: `make lint && make && make test`

## Commits

- Direct push to master is fine at this stage
