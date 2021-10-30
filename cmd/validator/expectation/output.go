// Copyright © 2021 Weald Technology Trading.
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
	"strings"

	"github.com/hako/durafmt"
)

func (c *command) output(ctx context.Context) (string, error) {
	if c.quiet {
		return "", nil
	}

	builder := strings.Builder{}

	builder.WriteString("Expected time between block proposals: ")
	builder.WriteString(durafmt.Parse(c.timeBetweenProposals).LimitFirstN(2).String())
	builder.WriteString("\n")

	builder.WriteString("Expected time between sync committees: ")
	builder.WriteString(durafmt.Parse(c.timeBetweenSyncCommittees).LimitFirstN(2).String())
	builder.WriteString("\n")

	return builder.String(), nil
}
