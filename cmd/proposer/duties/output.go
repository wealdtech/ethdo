// Copyright © 2022, 2023 Weald Technology Trading.
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

package proposerduties

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

func (c *command) output(ctx context.Context) (string, error) {
	if c.quiet {
		return "", nil
	}

	if c.jsonOutput {
		return c.outputJSON(ctx)
	}

	return c.outputTxt(ctx)
}

func (c *command) outputJSON(_ context.Context) (string, error) {
	data, err := json.Marshal(c.results)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (c *command) outputTxt(_ context.Context) (string, error) {
	builder := strings.Builder{}

	if len(c.results.Duties) == 1 {
		// Only have a single slot, just print the validator.
		duty := c.results.Duties[0]
		builder.WriteString("Validator ")
		builder.WriteString(fmt.Sprintf("%d", duty.ValidatorIndex))
		if c.verbose {
			builder.WriteString(" (pubkey ")
			builder.WriteString(fmt.Sprintf("%#x)", duty.PubKey))
		}
		builder.WriteString("\n")
	} else {
		// Have multiple slots, print per-slot information.
		builder.WriteString("Epoch ")
		builder.WriteString(fmt.Sprintf("%d:\n", c.results.Epoch))

		for _, duty := range c.results.Duties {
			builder.WriteString("  Slot ")
			builder.WriteString(fmt.Sprintf("%d: ", duty.Slot))
			builder.WriteString("validator ")
			builder.WriteString(fmt.Sprintf("%d", duty.ValidatorIndex))
			if c.verbose {
				builder.WriteString(" (pubkey ")
				builder.WriteString(fmt.Sprintf("%#x)", duty.PubKey))
			}
			builder.WriteString("\n")
		}
	}

	return strings.TrimSuffix(builder.String(), "\n"), nil
}
