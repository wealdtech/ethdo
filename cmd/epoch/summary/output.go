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

package epochsummary

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
	data, err := json.Marshal(c.summary)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (c *command) outputTxt(_ context.Context) (string, error) {
	builder := strings.Builder{}

	builder.WriteString("Epoch ")
	builder.WriteString(fmt.Sprintf("%d:\n", c.summary.Epoch))

	proposedBlocks := 0
	if c.verbose {
		for _, proposal := range c.summary.Proposals {
			builder.WriteString("  Slot ")
			builder.WriteString(fmt.Sprintf("%d (%d/%d):\n", proposal.Slot, uint64(proposal.Slot)%uint64(len(c.summary.Proposals)), len(c.summary.Proposals)))
			builder.WriteString("    Proposer: ")
			builder.WriteString(fmt.Sprintf("%d\n", proposal.Proposer))
			builder.WriteString("    Proposed: ")
			if proposal.Block {
				proposedBlocks++
				builder.WriteString("✓\n")
			} else {
				builder.WriteString("✕\n")
			}
		}
	} else {
		missedProposals := make([]string, 0, len(c.summary.Proposals))
		for _, proposal := range c.summary.Proposals {
			if !proposal.Block {
				missedProposals = append(missedProposals, fmt.Sprintf("    Slot %d (validator %d)\n", proposal.Slot, proposal.Proposer))
			} else {
				proposedBlocks++
			}
		}
		if len(missedProposals) > 0 {
			builder.WriteString("  Missed proposals:\n")
			for _, missedProposal := range missedProposals {
				builder.WriteString(missedProposal)
			}
		}
	}

	if c.verbose {
		for _, syncCommittee := range c.summary.SyncCommittee {
			builder.WriteString("  Sync committee validator ")
			builder.WriteString(fmt.Sprintf("%d:\n", syncCommittee.Index))
			builder.WriteString("    Chances: ")
			builder.WriteString(fmt.Sprintf("%d\n", proposedBlocks))
			builder.WriteString("    Included: ")
			builder.WriteString(fmt.Sprintf("%d\n", proposedBlocks-syncCommittee.Missed))
			builder.WriteString("    Inclusion %: ")
			builder.WriteString(fmt.Sprintf("%0.2f\n", 100.0*float64(proposedBlocks-syncCommittee.Missed)/float64(proposedBlocks)))
		}
	} else {
		missedSyncCommittees := make([]string, 0, len(c.summary.SyncCommittee))
		for _, syncCommittee := range c.summary.SyncCommittee {
			missedPct := 100.0 * float64(syncCommittee.Missed) / float64(proposedBlocks)
			missedSyncCommittees = append(missedSyncCommittees, fmt.Sprintf("    %d (%0.2f%%) by validator %d\n", syncCommittee.Missed, missedPct, syncCommittee.Index))
		}
		if len(missedSyncCommittees) > 0 {
			builder.WriteString("  Missed sync committees (excluding missed blocks):\n")
			for _, missedSyncCommittee := range missedSyncCommittees {
				builder.WriteString(missedSyncCommittee)
			}
		}
	}

	return builder.String(), nil
}
