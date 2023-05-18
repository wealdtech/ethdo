// Copyright © 2021, 2023 Weald Technology Trading.
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

package validatorexpectation

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hako/durafmt"
)

func (c *command) output(ctx context.Context) (string, error) {
	if c.quiet {
		return "", nil
	}

	if c.json {
		return c.outputJSON(ctx)
	}
	return c.outputTxt(ctx)
}

func (c *command) outputJSON(_ context.Context) (string, error) {
	data, err := json.Marshal(c.res)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s\n", string(data)), nil
}

func (c *command) outputTxt(_ context.Context) (string, error) {
	builder := strings.Builder{}

	builder.WriteString("Expected time between block proposals: ")
	builder.WriteString(durafmt.Parse(c.res.timeBetweenProposals).LimitFirstN(2).String())
	builder.WriteString("\n")

	builder.WriteString("Expected time between sync committees: ")
	builder.WriteString(durafmt.Parse(c.res.timeBetweenSyncCommittees).LimitFirstN(2).String())
	builder.WriteString("\n")

	return builder.String(), nil
}
