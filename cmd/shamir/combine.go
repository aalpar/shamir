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
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/aalpar/shamir/pkg/sss"
)

// CombineCommand holds the flags for the combine subcommand.
type CombineCommand struct{}

// Execute implements the go-flags Commander interface.
func (c *CombineCommand) Execute(args []string) error {
	return c.run(os.Stdin, os.Stdout)
}

func (c *CombineCommand) run(stdin io.Reader, stdout io.Writer) error {
	shares, err := readShares(stdin)
	if err != nil {
		return err
	}
	if len(shares) == 0 {
		return fmt.Errorf("combine: no shares provided")
	}

	secret, err := sss.Combine(shares)
	if err != nil {
		return fmt.Errorf("combine: %w", err)
	}

	_, err = stdout.Write(secret.Value().Bytes())
	return err
}

func readShares(r io.Reader) ([]sss.Share, error) {
	var shares []sss.Share
	scanner := bufio.NewScanner(r)
	line := 0
	for scanner.Scan() {
		line++
		text := scanner.Text()
		if text == "" {
			continue
		}
		var share sss.Share
		if err := json.Unmarshal([]byte(text), &share); err != nil {
			return nil, fmt.Errorf("combine: line %d: %w", line, err)
		}
		shares = append(shares, share)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("combine: reading stdin: %w", err)
	}
	return shares, nil
}
