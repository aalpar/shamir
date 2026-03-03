// Package vss implements Verifiable Secret Sharing schemes.
//
// VSS extends Shamir's Secret Sharing so that shareholders can verify
// the correctness of their shares without interaction, detecting a
// dishonest dealer.
//
// Two schemes are provided:
//
// Feldman VSS — the dealer publishes commitments C_j = g^{a_j} mod p
// for each polynomial coefficient. Computationally hiding (discrete log).
// See: BIBLIOGRAPHY.md — Feldman 1987.
//
// Pedersen VSS — uses two generators g, h with unknown log_g(h).
// Commitments C_j = g^{a_j} * h^{b_j}. Information-theoretically hiding.
// See: BIBLIOGRAPHY.md — Pedersen 1991.
package vss
