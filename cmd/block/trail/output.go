// Copyright © 2025 Weald Technology Trading.
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

package blocktrail

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

type simpleOut struct {
	Start *step `json:"start"`
	End   *step `json:"end"`
	Steps int   `json:"distance"`
}

func (c *command) outputJSON(_ context.Context) (string, error) {
	var err error
	var data []byte
	if c.verbose {
		data, err = json.Marshal(c.steps)
	} else {
		basic := &simpleOut{
			Start: c.steps[0],
			End:   c.steps[len(c.steps)-1],
			Steps: len(c.steps) - 1,
		}
		data, err = json.Marshal(basic)
	}

	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (c *command) outputTxt(_ context.Context) (string, error) {
	if !c.found {
		return "Target not found", nil
	}

	builder := strings.Builder{}
	builder.WriteString("Target '")
	builder.WriteString(c.target)
	builder.WriteString("' found at a distance of ")
	builder.WriteString(fmt.Sprintf("%d", len(c.steps)-1))
	builder.WriteString(" block(s)")

	return builder.String(), nil
}
