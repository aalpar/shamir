// Package refresh implements proactive share refresh for secret sharing.
//
// Proactive refresh allows shareholders to generate new shares of the
// same secret, rendering previously compromised shares useless. Each
// participant generates a random "zero sharing" — shares of the zero
// polynomial — and adds sub-shares to their existing share.
//
// After refresh, all shares are algebraically different but still
// reconstruct the same secret. The security model shifts from "never
// compromise k shares" to "never compromise k shares within a single
// refresh period."
//
// See: BIBLIOGRAPHY.md — Herzberg et al. 1995 (proactive secret sharing).
package refresh
