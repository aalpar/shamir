# Shamir

Multi-party threshold cryptography library in Go.

A zero-dependency library implementing Shamir's Secret Sharing, Verifiable
Secret Sharing (Feldman, Pedersen), and proactive share refresh.

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

## Domain Types

The `threshold/` package defines platform-independent types for threshold
secret lifecycle management: scheme selection, share configuration, participant
tracking, and phase transitions.

## Status

Early development. The crypto primitives are implemented; higher-level
coordination protocols are next.

## License

Apache License 2.0. See [LICENSE](LICENSE).
