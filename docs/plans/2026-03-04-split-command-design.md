# Split Command Design

## Interface

```
echo -n 'my secret' | shamir split -k 3 -n 5 [--scheme sss|feldman|pedersen] [--bits 2048]
```

### Flags

| Flag | Short | Required | Default | Description |
|------|-------|----------|---------|-------------|
| `--threshold` | `-k` | yes | — | Minimum shares to reconstruct |
| `--shares` | `-n` | yes | — | Total shares to generate |
| `--scheme` | — | no | `sss` | `sss`, `feldman`, or `pedersen` |
| `--bits` | — | no | `2048` | Safe prime bit size (VSS schemes only) |

### Input

Raw bytes from stdin, interpreted as a big-endian `big.Int`.

### Output

- **stdout**: one JSON share per line (n lines)
- **stderr**: commitment JSON (feldman/pedersen only)

## Flow

1. Parse flags via `flag.NewFlagSet("split", ...)`
2. Read stdin to EOF -> `[]byte` -> `big.Int` (big-endian)
3. Validate: secret must be non-empty, k >= 2, k <= n
4. Based on scheme:
   - **sss**: find next prime > secret, create `field.Field`, call `sss.Split`
   - **feldman**: generate safe prime of `--bits` size, find generator g, `vss.NewGroup(p, g, nil)`, call `vss.FeldmanDeal`
   - **pedersen**: generate safe prime, find two independent generators g and h, `vss.NewGroup(p, g, h)`, call `vss.PedersenDeal`
5. Marshal each share as JSON, write one per line to stdout
6. If VSS: marshal commitment JSON to stderr

## Design Decisions

### Prime selection for SSS

Use the next prime after the secret value. The field must be larger than the
secret for correctness. This keeps SSS fast (no safe prime generation) and
simple. The share JSON already embeds `p`, so the field is self-describing.

If users want to hide approximate secret size, they pad their input.

### Safe prime generation for VSS

Generate on the fly. The `--bits` flag controls size (default 2048). This is
slow (~seconds) but fully self-contained. No external setup or separate
`gengroup` command needed.

### Commitment output to stderr

Commitment goes to stderr so stdout contains only shares (one per line).
Users redirect stderr to capture: `shamir split 2>commitment.json`.
This keeps share output pipeable to line-based tools.

### Secret as raw bytes

Stdin raw bytes avoids encoding ambiguity. The bytes are interpreted as a
big-endian unsigned integer. This is the natural format for piping binary
secrets. Hex/base64 encoding can be done outside the tool.

## File Structure

Single new file: `cmd/shamir/split.go`

The existing `main.go` dispatches on subcommand; `split.go` exports a
`runSplit(args []string) error` function called from the switch case.

## Dependencies

Uses only existing packages:
- `pkg/field` — `field.New`, `field.Element`
- `pkg/sss` — `sss.Split`, `sss.Share`
- `pkg/vss` — `vss.NewGroup`, `vss.FeldmanDeal`, `vss.PedersenDeal`
- `crypto/rand` — safe prime generation
- `encoding/json` — share/commitment marshaling (via TextMarshaler)

## Open Questions

- Safe prime generation: use `crypto/rand.Prime` for the base prime, then
  check `2p+1` for primality? Or use a dedicated safe-prime finder? The
  stdlib approach is simple but may need multiple attempts.
- Generator finding: for a safe prime p = 2q+1, any element of order q in
  Z*_p works. Pick random candidates, verify g^q = 1 mod p and g^2 != 1 mod p.
