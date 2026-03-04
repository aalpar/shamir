# Split Command Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Wire the `split` CLI subcommand to read a secret from stdin and output Shamir shares as JSON.

**Architecture:** Single new file `cmd/shamir/split.go` containing `runSplit(args []string) error`. The existing `main.go` dispatches to it. Three scheme paths (sss, feldman, pedersen) share flag parsing and output logic, diverging only in crypto setup. Group parameter generation (safe prime + generators) lives in a helper `cmd/shamir/group.go` since it's reusable by the future `verify` command.

**Tech Stack:** stdlib `flag`, `crypto/rand`, `math/big`, `io`, `encoding/json`; internal `pkg/field`, `pkg/sss`, `pkg/vss`.

---

### Task 1: Flag Parsing and Input Reading

Create the split subcommand skeleton with flag parsing and stdin reading.
No crypto yet — just the CLI plumbing.

**Files:**
- Create: `cmd/shamir/split.go`
- Modify: `cmd/shamir/main.go:30-35` (replace stub with `runSplit` call)
- Test: `cmd/shamir/split_test.go`

**Step 1: Write the failing test**

Create `cmd/shamir/split_test.go`:

```go
package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunSplitMissingFlags(t *testing.T) {
	err := runSplit([]string{}, bytes.NewReader(nil), nil, nil)
	if err == nil {
		t.Fatal("expected error for missing flags")
	}
}

func TestRunSplitInvalidScheme(t *testing.T) {
	err := runSplit([]string{"-k", "3", "-n", "5", "--scheme", "bogus"},
		bytes.NewReader([]byte("secret")), nil, nil)
	if err == nil {
		t.Fatal("expected error for invalid scheme")
	}
	if !strings.Contains(err.Error(), "scheme") {
		t.Errorf("error should mention scheme, got: %v", err)
	}
}

func TestRunSplitEmptyStdin(t *testing.T) {
	err := runSplit([]string{"-k", "3", "-n", "5"},
		bytes.NewReader(nil), nil, nil)
	if err == nil {
		t.Fatal("expected error for empty stdin")
	}
}
```

Note: `runSplit` signature is `func runSplit(args []string, stdin io.Reader, stdout, stderr io.Writer) error`. Accepting `io.Reader`/`io.Writer` makes it testable without real stdio.

**Step 2: Run test to verify it fails**

Run: `go test ./cmd/shamir/ -run TestRunSplit -v`
Expected: FAIL — `runSplit` undefined

**Step 3: Write the implementation**

Create `cmd/shamir/split.go`:

```go
package main

import (
	"flag"
	"fmt"
	"io"
	"math/big"
)

func runSplit(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("split", flag.ContinueOnError)
	fs.SetOutput(stderr)

	k := fs.Int("k", 0, "threshold (minimum shares to reconstruct)")
	n := fs.Int("n", 0, "total number of shares")
	scheme := fs.String("scheme", "sss", "sharing scheme: sss, feldman, pedersen")
	bits := fs.Int("bits", 2048, "safe prime bit size (VSS schemes only)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *k == 0 || *n == 0 {
		return fmt.Errorf("split: -k and -n flags are required")
	}

	switch *scheme {
	case "sss", "feldman", "pedersen":
	default:
		return fmt.Errorf("split: unknown scheme %q (want sss, feldman, or pedersen)", *scheme)
	}

	secretBytes, err := io.ReadAll(stdin)
	if err != nil {
		return fmt.Errorf("split: reading stdin: %w", err)
	}
	if len(secretBytes) == 0 {
		return fmt.Errorf("split: empty secret (no data on stdin)")
	}

	secret := new(big.Int).SetBytes(secretBytes)
	_ = secret
	_ = bits

	return fmt.Errorf("split: scheme %s not yet implemented", *scheme)
}
```

**Step 4: Wire main.go**

In `cmd/shamir/main.go`, replace the `"split", "combine", "verify"` case:

```go
	case "split":
		if err := runSplit(os.Args[2:], os.Stdin, os.Stdout, os.Stderr); err != nil {
			fmt.Fprintf(os.Stderr, "shamir split: %v\n", err)
			os.Exit(1)
		}
	case "combine", "verify":
		fmt.Fprintf(os.Stderr, "shamir %s: not yet implemented\n", os.Args[1])
		os.Exit(1)
```

**Step 5: Run tests and verify they pass**

Run: `go test ./cmd/shamir/ -run TestRunSplit -v`
Expected: PASS — all three tests pass (missing flags, invalid scheme, empty stdin all return errors)

**Step 6: Commit**

```bash
git add cmd/shamir/split.go cmd/shamir/split_test.go cmd/shamir/main.go
git commit -m "Add split subcommand skeleton with flag parsing and stdin reading"
```

---

### Task 2: SSS Split Path

Implement the plain Shamir scheme path: find next prime > secret, create field, split, output JSON.

**Files:**
- Modify: `cmd/shamir/split.go`
- Modify: `cmd/shamir/split_test.go`

**Step 1: Write the failing test**

Add to `cmd/shamir/split_test.go`:

```go
import (
	"encoding/json"
	"math/big"

	"github.com/aalpar/shamir/pkg/sss"
)

func TestRunSplitSSS(t *testing.T) {
	secret := []byte("hello world")
	var stdout, stderr bytes.Buffer

	err := runSplit([]string{"-k", "3", "-n", "5"},
		bytes.NewReader(secret), &stdout, &stderr)
	if err != nil {
		t.Fatalf("runSplit: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	if len(lines) != 5 {
		t.Fatalf("got %d lines, want 5", len(lines))
	}

	// Verify each line is a valid sss.Share
	shares := make([]sss.Share, 5)
	for i, line := range lines {
		if err := json.Unmarshal([]byte(line), &shares[i]); err != nil {
			t.Fatalf("share %d: unmarshal: %v", i, err)
		}
	}

	// Combine any 3 shares and verify we get the original secret back
	recovered, err := sss.Combine(shares[:3])
	if err != nil {
		t.Fatalf("Combine: %v", err)
	}

	original := new(big.Int).SetBytes(secret)
	if recovered.Value().Cmp(original) != 0 {
		t.Errorf("recovered %s, want %s", recovered.Value(), original)
	}

	// Stderr should be empty for SSS (no commitment)
	if stderr.Len() != 0 {
		t.Errorf("stderr should be empty for SSS, got: %s", stderr.String())
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./cmd/shamir/ -run TestRunSplitSSS -v`
Expected: FAIL — "scheme sss not yet implemented"

**Step 3: Implement SSS path**

Replace the placeholder at the bottom of `runSplit` in `cmd/shamir/split.go`. The function body after `secret := new(big.Int).SetBytes(secretBytes)` becomes:

```go
	switch *scheme {
	case "sss":
		return splitSSS(secret, *n, *k, stdout)
	case "feldman":
		return splitFeldman(secret, *n, *k, *bits, stdout, stderr)
	case "pedersen":
		return splitPedersen(secret, *n, *k, *bits, stdout, stderr)
	}
	return nil
```

Add `splitSSS` function:

```go
import (
	"crypto/rand"
	"encoding/json"

	"github.com/aalpar/shamir/pkg/field"
	"github.com/aalpar/shamir/pkg/sss"
)

func splitSSS(secret *big.Int, n, k int, stdout io.Writer) error {
	// Find next prime > secret for the field modulus.
	p := nextPrime(secret)
	f := field.New(p)
	elem := f.NewElement(secret)

	shares, err := sss.Split(elem, n, k, f)
	if err != nil {
		return fmt.Errorf("split: %w", err)
	}

	return writeShares(shares, stdout)
}

// nextPrime returns the smallest prime > n.
func nextPrime(n *big.Int) *big.Int {
	// Start at n+1, step by 2 (skip evens after handling n+1).
	p := new(big.Int).Add(n, big.NewInt(1))
	if p.Bit(0) == 0 {
		p.Add(p, big.NewInt(1))
	}
	for !p.ProbablyPrime(20) {
		p.Add(p, big.NewInt(2))
	}
	return p
}

func writeShares(shares []sss.Share, w io.Writer) error {
	enc := json.NewEncoder(w)
	for i := range shares {
		if err := enc.Encode(&shares[i]); err != nil {
			return fmt.Errorf("split: writing share %d: %w", i, err)
		}
	}
	return nil
}
```

Note: `json.Encoder.Encode` appends a newline, so each share is one JSON object per line — exactly the output format we want.

Add stubs for the two VSS paths so it compiles:

```go
func splitFeldman(secret *big.Int, n, k, bits int, stdout, stderr io.Writer) error {
	return fmt.Errorf("feldman: not yet implemented")
}

func splitPedersen(secret *big.Int, n, k, bits int, stdout, stderr io.Writer) error {
	return fmt.Errorf("pedersen: not yet implemented")
}
```

**Step 4: Run tests**

Run: `go test ./cmd/shamir/ -run TestRunSplit -v`
Expected: PASS — all 4 tests pass

**Step 5: Commit**

```bash
git add cmd/shamir/split.go cmd/shamir/split_test.go
git commit -m "Implement SSS split path with next-prime field selection"
```

---

### Task 3: Safe Prime and Generator Utilities

Build the group parameter generation needed by Feldman and Pedersen.

**Files:**
- Create: `cmd/shamir/group.go`
- Create: `cmd/shamir/group_test.go`

**Step 1: Write the failing test**

Create `cmd/shamir/group_test.go`:

```go
package main

import (
	"math/big"
	"testing"
)

func TestGenSafePrime(t *testing.T) {
	// Use small bit size for fast tests.
	p, err := genSafePrime(64)
	if err != nil {
		t.Fatalf("genSafePrime: %v", err)
	}

	if p.BitLen() < 64 {
		t.Errorf("prime bit length = %d, want >= 64", p.BitLen())
	}

	if !p.ProbablyPrime(20) {
		t.Error("p is not prime")
	}

	// q = (p-1)/2 should also be prime (safe prime property).
	q := new(big.Int).Sub(p, big.NewInt(1))
	q.Rsh(q, 1)
	if !q.ProbablyPrime(20) {
		t.Error("q = (p-1)/2 is not prime")
	}
}

func TestFindGenerator(t *testing.T) {
	p, err := genSafePrime(64)
	if err != nil {
		t.Fatalf("genSafePrime: %v", err)
	}

	q := new(big.Int).Sub(p, big.NewInt(1))
	q.Rsh(q, 1)

	g, err := findGenerator(p, q)
	if err != nil {
		t.Fatalf("findGenerator: %v", err)
	}

	// g^q mod p == 1 (order divides q)
	result := new(big.Int).Exp(g, q, p)
	if result.Cmp(big.NewInt(1)) != 0 {
		t.Errorf("g^q mod p = %s, want 1", result)
	}

	// g^2 mod p != 1 (order is not 2, so order is q)
	result.Exp(g, big.NewInt(2), p)
	if result.Cmp(big.NewInt(1)) == 0 {
		t.Error("g^2 mod p = 1, generator has order 2")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./cmd/shamir/ -run "TestGenSafePrime|TestFindGenerator" -v`
Expected: FAIL — `genSafePrime` undefined

**Step 3: Implement**

Create `cmd/shamir/group.go`:

```go
package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

var bigOne = big.NewInt(1)
var bigTwo = big.NewInt(2)

// genSafePrime generates a safe prime p = 2q+1 where q is also prime.
// bits specifies the minimum bit length of p.
func genSafePrime(bits int) (*big.Int, error) {
	for {
		// Generate a random prime q, then check if p = 2q+1 is prime.
		q, err := rand.Prime(rand.Reader, bits-1)
		if err != nil {
			return nil, fmt.Errorf("genSafePrime: %w", err)
		}

		p := new(big.Int).Mul(q, bigTwo)
		p.Add(p, bigOne)

		if p.ProbablyPrime(20) {
			return p, nil
		}
	}
}

// findGenerator finds a generator of the order-q subgroup of Z*_p
// where p = 2q+1 is a safe prime.
func findGenerator(p, q *big.Int) (*big.Int, error) {
	for i := 0; i < 1000; i++ {
		// Pick random h in [2, p-2].
		h, err := rand.Int(rand.Reader, new(big.Int).Sub(p, bigTwo))
		if err != nil {
			return nil, fmt.Errorf("findGenerator: %w", err)
		}
		h.Add(h, bigTwo) // shift from [0, p-4] to [2, p-2]

		// g = h^2 mod p guarantees g is in the order-q subgroup.
		g := new(big.Int).Exp(h, bigTwo, p)

		// Reject g = 1 (degenerate).
		if g.Cmp(bigOne) == 0 {
			continue
		}

		return g, nil
	}
	return nil, fmt.Errorf("findGenerator: failed after 1000 attempts")
}
```

`★ Insight ─────────────────────────────────────`
**Why `g = h^2 mod p` instead of checking `g^q == 1`:** For a safe prime p = 2q+1, Z\*\_p has order p-1 = 2q. Every element has order 1, 2, q, or 2q. Squaring any `h` gives `h^2` which has order dividing q (since `(h^2)^q = h^(2q) = h^(p-1) = 1`). As long as `h^2 != 1`, the order is exactly q. This is simpler and faster than trial-checking random candidates.
`─────────────────────────────────────────────────`

**Step 4: Run tests**

Run: `go test ./cmd/shamir/ -run "TestGenSafePrime|TestFindGenerator" -v`
Expected: PASS

**Step 5: Commit**

```bash
git add cmd/shamir/group.go cmd/shamir/group_test.go
git commit -m "Add safe prime generation and generator finding utilities"
```

---

### Task 4: Feldman Split Path

Wire `splitFeldman` using the group utilities.

**Files:**
- Modify: `cmd/shamir/split.go`
- Modify: `cmd/shamir/split_test.go`

**Step 1: Write the failing test**

Add to `cmd/shamir/split_test.go`:

```go
import "github.com/aalpar/shamir/pkg/vss"

func TestRunSplitFeldman(t *testing.T) {
	if testing.Short() {
		t.Skip("safe prime generation is slow")
	}

	secret := []byte("feldman secret")
	var stdout, stderr bytes.Buffer

	// Use small bits for test speed.
	err := runSplit([]string{"-k", "3", "-n", "5", "--scheme", "feldman", "--bits", "64"},
		bytes.NewReader(secret), &stdout, &stderr)
	if err != nil {
		t.Fatalf("runSplit: %v", err)
	}

	// Stdout: 5 share lines
	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	if len(lines) != 5 {
		t.Fatalf("got %d lines, want 5", len(lines))
	}

	shares := make([]sss.Share, 5)
	for i, line := range lines {
		if err := json.Unmarshal([]byte(line), &shares[i]); err != nil {
			t.Fatalf("share %d: unmarshal: %v", i, err)
		}
	}

	// Stderr: commitment JSON
	if stderr.Len() == 0 {
		t.Fatal("expected commitment on stderr")
	}
	var commitment vss.Commitment
	if err := json.Unmarshal(stderr.Bytes(), &commitment); err != nil {
		t.Fatalf("commitment unmarshal: %v", err)
	}

	// Verify each share against the commitment
	for i, share := range shares {
		if err := vss.FeldmanVerify(share, &commitment); err != nil {
			t.Errorf("share %d failed verification: %v", i, err)
		}
	}

	// Combine and verify secret
	recovered, err := sss.Combine(shares[:3])
	if err != nil {
		t.Fatalf("Combine: %v", err)
	}
	original := new(big.Int).SetBytes(secret)
	if recovered.Value().Cmp(original) != 0 {
		t.Errorf("recovered %s, want %s", recovered.Value(), original)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./cmd/shamir/ -run TestRunSplitFeldman -v`
Expected: FAIL — "feldman: not yet implemented"

**Step 3: Implement**

Replace `splitFeldman` stub in `cmd/shamir/split.go`:

```go
func splitFeldman(secret *big.Int, n, k, bits int, stdout, stderr io.Writer) error {
	p, err := genSafePrime(bits)
	if err != nil {
		return fmt.Errorf("split feldman: %w", err)
	}

	q := new(big.Int).Sub(p, bigOne)
	q.Rsh(q, 1)

	g, err := findGenerator(p, q)
	if err != nil {
		return fmt.Errorf("split feldman: %w", err)
	}

	grp, err := vss.NewGroup(p, g, nil)
	if err != nil {
		return fmt.Errorf("split feldman: %w", err)
	}

	elem := grp.Field().NewElement(secret)

	shares, commitment, err := vss.FeldmanDeal(elem, n, k, grp)
	if err != nil {
		return fmt.Errorf("split feldman: %w", err)
	}

	if err := writeShares(shares, stdout); err != nil {
		return err
	}

	return writeCommitment(commitment, stderr)
}

func writeCommitment(c *vss.Commitment, w io.Writer) error {
	data, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("split: marshaling commitment: %w", err)
	}
	_, err = w.Write(data)
	return err
}
```

Add import for `"github.com/aalpar/shamir/pkg/vss"` to `split.go`.

**Step 4: Run tests**

Run: `go test ./cmd/shamir/ -run TestRunSplit -v`
Expected: PASS — SSS test still passes, Feldman test passes

**Step 5: Commit**

```bash
git add cmd/shamir/split.go cmd/shamir/split_test.go
git commit -m "Implement Feldman split path with group generation"
```

---

### Task 5: Pedersen Split Path

Wire `splitPedersen` — needs two independent generators.

**Files:**
- Modify: `cmd/shamir/split.go`
- Modify: `cmd/shamir/split_test.go`

**Step 1: Write the failing test**

Add to `cmd/shamir/split_test.go`:

```go
func TestRunSplitPedersen(t *testing.T) {
	if testing.Short() {
		t.Skip("safe prime generation is slow")
	}

	secret := []byte("pedersen secret")
	var stdout, stderr bytes.Buffer

	err := runSplit([]string{"-k", "3", "-n", "5", "--scheme", "pedersen", "--bits", "64"},
		bytes.NewReader(secret), &stdout, &stderr)
	if err != nil {
		t.Fatalf("runSplit: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	if len(lines) != 5 {
		t.Fatalf("got %d lines, want 5", len(lines))
	}

	shares := make([]vss.PedersenShare, 5)
	for i, line := range lines {
		if err := json.Unmarshal([]byte(line), &shares[i]); err != nil {
			t.Fatalf("share %d: unmarshal: %v", i, err)
		}
	}

	if stderr.Len() == 0 {
		t.Fatal("expected commitment on stderr")
	}
	var commitment vss.Commitment
	if err := json.Unmarshal(stderr.Bytes(), &commitment); err != nil {
		t.Fatalf("commitment unmarshal: %v", err)
	}

	for i, share := range shares {
		if err := vss.PedersenVerify(share, &commitment); err != nil {
			t.Errorf("share %d failed verification: %v", i, err)
		}
	}
}
```

Note: Pedersen uses `vss.PedersenShare` (has X, Y, T fields), not `sss.Share`. The output lines will be `PedersenShare` JSON.

**Step 2: Run test to verify it fails**

Run: `go test ./cmd/shamir/ -run TestRunSplitPedersen -v`
Expected: FAIL — "pedersen: not yet implemented"

**Step 3: Implement**

Replace `splitPedersen` stub:

```go
func splitPedersen(secret *big.Int, n, k, bits int, stdout, stderr io.Writer) error {
	p, err := genSafePrime(bits)
	if err != nil {
		return fmt.Errorf("split pedersen: %w", err)
	}

	q := new(big.Int).Sub(p, bigOne)
	q.Rsh(q, 1)

	g, err := findGenerator(p, q)
	if err != nil {
		return fmt.Errorf("split pedersen: %w", err)
	}

	h, err := findGenerator(p, q)
	if err != nil {
		return fmt.Errorf("split pedersen: %w", err)
	}

	// Ensure g != h (extremely unlikely with random generation, but check).
	for g.Cmp(h) == 0 {
		h, err = findGenerator(p, q)
		if err != nil {
			return fmt.Errorf("split pedersen: %w", err)
		}
	}

	grp, err := vss.NewGroup(p, g, h)
	if err != nil {
		return fmt.Errorf("split pedersen: %w", err)
	}

	elem := grp.Field().NewElement(secret)

	shares, commitment, err := vss.PedersenDeal(elem, n, k, grp)
	if err != nil {
		return fmt.Errorf("split pedersen: %w", err)
	}

	if err := writePedersenShares(shares, stdout); err != nil {
		return err
	}

	return writeCommitment(commitment, stderr)
}

func writePedersenShares(shares []vss.PedersenShare, w io.Writer) error {
	enc := json.NewEncoder(w)
	for i := range shares {
		if err := enc.Encode(&shares[i]); err != nil {
			return fmt.Errorf("split: writing share %d: %w", i, err)
		}
	}
	return nil
}
```

**Step 4: Run tests**

Run: `go test ./cmd/shamir/ -run TestRunSplit -v`
Expected: PASS — all 5 tests pass

Run: `make lint && make test`
Expected: PASS — full suite green

**Step 5: Commit**

```bash
git add cmd/shamir/split.go cmd/shamir/split_test.go
git commit -m "Implement Pedersen split path with dual generators"
```

---

### Task 6: Validate Secret < Field Modulus for VSS

The VSS schemes use a safe prime of `--bits` size. If the secret is larger than q (the field order), `NewElement` silently reduces it mod q, destroying the original value. Add a check.

**Files:**
- Modify: `cmd/shamir/split.go`
- Modify: `cmd/shamir/split_test.go`

**Step 1: Write the failing test**

```go
func TestRunSplitSecretTooLargeForBits(t *testing.T) {
	// Secret that won't fit in a 64-bit safe prime field.
	bigSecret := make([]byte, 32) // 256 bits
	for i := range bigSecret {
		bigSecret[i] = 0xff
	}

	var stdout, stderr bytes.Buffer
	err := runSplit([]string{"-k", "3", "-n", "5", "--scheme", "feldman", "--bits", "64"},
		bytes.NewReader(bigSecret), &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error when secret exceeds field size")
	}
	if !strings.Contains(err.Error(), "too large") {
		t.Errorf("error should mention 'too large', got: %v", err)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./cmd/shamir/ -run TestRunSplitSecretTooLarge -v`
Expected: FAIL — no "too large" error (secret silently reduced)

**Step 3: Implement the check**

Add to both `splitFeldman` and `splitPedersen`, after creating the group and before calling Deal:

```go
	q := grp.Field().Prime()
	if secret.Cmp(q) >= 0 {
		return fmt.Errorf("split: secret (%d bits) too large for field (%d bits); increase --bits",
			secret.BitLen(), q.BitLen())
	}
```

Note: `grp.Field()` returns the `*field.Field` whose prime is q (the subgroup order), not p. This is correct — the secret polynomial is evaluated in GF(q).

**Step 4: Run tests**

Run: `go test ./cmd/shamir/ -run TestRunSplit -v`
Expected: PASS

**Step 5: Commit**

```bash
git add cmd/shamir/split.go cmd/shamir/split_test.go
git commit -m "Validate secret fits in VSS field, error with guidance on --bits"
```

---

### Task 7: Lint and Full Suite

Final pass — ensure everything is clean.

**Files:** No new files

**Step 1: Run full suite**

Run: `make lint && make test`
Expected: PASS

**Step 2: Run with race detector**

Run: `make test-race`
Expected: PASS (no shared mutable state in split)

**Step 3: Build binary and smoke test**

```bash
make build
echo -n 'hello world' | ./dist/shamir split -k 3 -n 5
echo -n 'hello world' | ./dist/shamir split -k 3 -n 5 --scheme feldman --bits 64
```

Expected: 5 JSON lines on stdout, commitment on stderr for Feldman

**Step 4: Commit if any fixups needed**

```bash
git add -A
git commit -m "Fix lint issues from split command implementation"
```

(Skip if no changes needed.)

---

### Task 8: Update TODO.md

Check off completed items from Phase 1.1.

**Files:**
- Modify: `TODO.md`

**Step 1: Mark completed items**

In TODO.md under "### 1.1 `split` Command", check off:
- [x] Parse flags: `--threshold/-k`, `--shares/-n`, `--scheme` (sss|feldman|pedersen)
- [x] Read secret from stdin (or `--secret` flag for testing)
- [x] Generate safe prime + generator for Feldman/Pedersen schemes
- [x] Call `sss.Split` or `vss.FeldmanDeal` / `vss.PedersenDeal`
- [x] Serialize and output shares (one per line? one per file?)
- [x] For VSS schemes: also output commitment
- [x] Tests: split -> combine round-trip via CLI
- [x] Tests: invalid flag combinations

Note: "Read secret from stdin" — we chose stdin-only (no `--secret` flag). Update the TODO text to reflect the actual decision.

**Step 2: Commit**

```bash
git add TODO.md
git commit -m "Update TODO: mark split command items complete"
```
