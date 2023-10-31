// Copyright © 2022 Weald Technology Trading.
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

package inclusion

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

func (c *command) output(_ context.Context) (string, error) {
	if c.quiet {
		return "", nil
	}

	builder := strings.Builder{}

	if c.verbose {
		builder.WriteString("Epoch: ")
		builder.WriteString(fmt.Sprintf("%d\n", c.epoch))
	}

	if !c.inCommittee {
		builder.WriteString("Validator not in sync committee")
	} else {
		if c.verbose {
			builder.WriteString("Validator sync committee index ")
			builder.WriteString(fmt.Sprintf("%d\n", c.committeeIndex))
		}

		noBlock := 0
		included := 0
		missed := 0
		for _, inclusion := range c.inclusions {
			switch inclusion {
			case 0:
				noBlock++
			case 1:
				included++
			case 2:
				missed++
			}
		}
		builder.WriteString("Expected: ")
		builder.WriteString(strconv.Itoa(len(c.inclusions)))
		builder.WriteString("\nIncluded: ")
		builder.WriteString(strconv.Itoa(included))
		builder.WriteString("\nMissed: ")
		builder.WriteString(strconv.Itoa(missed))
		builder.WriteString("\nNo block: ")
		builder.WriteString(strconv.Itoa(noBlock))

		builder.WriteString("\nPer-slot result: ")
		for i, inclusion := range c.inclusions {
			switch inclusion {
			case 0:
				builder.WriteString("-")
			case 1:
				builder.WriteString("✓")
			case 2:
				builder.WriteString("✕")
			}
			if i%8 == 7 && i != len(c.inclusions)-1 {
				builder.WriteString(" ")
			}
		}
	}

	return builder.String(), nil
}
