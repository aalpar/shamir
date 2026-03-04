// Copyright 2026 Aaron Alpar
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"fmt"
	"os"
)

// BuildVersion is set at build time via ldflags.
var BuildVersion = "dev"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "version":
		fmt.Println(BuildVersion)
	case "split", "combine", "verify":
		fmt.Fprintf(os.Stderr, "shamir %s: not yet implemented\n", os.Args[1])
		os.Exit(1)
	default:
		fmt.Fprintf(os.Stderr, "shamir: unknown command %q\n", os.Args[1])
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "Usage: shamir <command>")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Commands:")
	fmt.Fprintln(os.Stderr, "  split     Split a secret into shares")
	fmt.Fprintln(os.Stderr, "  combine   Reconstruct a secret from shares")
	fmt.Fprintln(os.Stderr, "  verify    Verify a share against commitments")
	fmt.Fprintln(os.Stderr, "  version   Print version")
}
