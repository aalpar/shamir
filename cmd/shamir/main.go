// Copyright 2026 Aaron Alpar
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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

	"github.com/jessevdk/go-flags"
)

// BuildVersion is set at build time via ldflags.
var BuildVersion = "dev"

// GlobalOptions holds top-level flags that apply before any subcommand.
type GlobalOptions struct {
	Version func() `short:"V" long:"version" description:"Print version and exit"`
}

func main() {
	var opts GlobalOptions
	opts.Version = func() {
		fmt.Println(BuildVersion)
		os.Exit(0)
	}

	parser := flags.NewParser(&opts, flags.Default)
	parser.Name = "shamir"

	if _, err := parser.AddCommand("split",
		"Split a secret into shares",
		"Split a secret into shares using Shamir's Secret Sharing",
		&SplitCommand{}); err != nil {
		fmt.Fprintf(os.Stderr, "shamir: %v\n", err)
		os.Exit(1)
	}

	_, err := parser.Parse()
	if err != nil {
		flagsErr, ok := err.(*flags.Error)
		if ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		}
		os.Exit(1)
	}
}
