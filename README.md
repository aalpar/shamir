# Shamir

Threshold cryptography library and Kubernetes operator.

A zero-dependency Go library implementing Shamir's Secret Sharing, Verifiable
Secret Sharing (Feldman, Pedersen), and proactive share refresh — plus a
Kubernetes operator that uses these primitives to protect Secrets with
k-of-n threshold schemes.

## Crypto Library

The `pkg/` packages are importable by any Go project with no transitive
dependencies beyond stdlib:

```go
import (
    "github.com/aalpar/shamir/pkg/field"
    "github.com/aalpar/shamir/pkg/sss"
)

// Create a prime field
f := field.New(prime)

// Split a secret into 5 shares with threshold 3
shares, err := sss.Split(secret, 5, 3, f)

// Reconstruct from any 3 shares
recovered, err := sss.Combine(shares[:3])
```

## Status

Early development. The crypto primitives are being implemented first;
the Kubernetes operator will follow.

## License

Apache License 2.0. See [LICENSE](LICENSE).
