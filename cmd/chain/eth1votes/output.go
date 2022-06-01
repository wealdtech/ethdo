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

package chaineth1votes

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/attestantio/go-eth2-client/spec/phase0"
)

type jsonOutput struct {
	Slot      phase0.Slot      `json:"slot"`
	Period    uint64           `json:"period"`
	Incumbent *phase0.ETH1Data `json:"incumbent"`
	Votes     []*vote          `json:"votes"`
}

func (c *command) output(ctx context.Context) (string, error) {
	if c.quiet {
		return "", nil
	}

	if c.json {
		return c.outputJSON(ctx)
	}
	return c.outputText(ctx)
}

func (c *command) outputJSON(ctx context.Context) (string, error) {
	votes := make([]*vote, 0, len(c.votes))
	totalVotes := 0
	for _, vote := range c.votes {
		votes = append(votes, vote)
		totalVotes += vote.Count
	}
	sort.Slice(votes, func(i int, j int) bool {
		if votes[i].Count != votes[j].Count {
			return votes[i].Count > votes[j].Count
		}
		return votes[i].Vote.DepositCount < votes[j].Vote.DepositCount
	})

	output := &jsonOutput{
		Slot:      c.slot,
		Period:    c.period,
		Incumbent: c.incumbent,
		Votes:     votes,
	}
	data, err := json.Marshal(output)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (c *command) outputText(ctx context.Context) (string, error) {
	builder := strings.Builder{}

	if c.verbose {
		builder.WriteString("Slot: ")
		builder.WriteString(fmt.Sprintf("%d\n", c.slot))
	}

	builder.WriteString("Voting period: ")
	builder.WriteString(fmt.Sprintf("%d\n", c.period))

	if c.verbose {
		builder.WriteString("Incumbent: ")
		builder.WriteString(fmt.Sprintf("block %#x, deposit count %d\n", c.incumbent.BlockHash, c.incumbent.DepositCount))
	}

	votes := make([]*vote, 0, len(c.votes))
	totalVotes := 0
	for _, vote := range c.votes {
		votes = append(votes, vote)
		totalVotes += vote.Count
	}
	sort.Slice(votes, func(i int, j int) bool {
		if votes[i].Count != votes[j].Count {
			return votes[i].Count > votes[j].Count
		}
		return votes[i].Vote.DepositCount < votes[j].Vote.DepositCount
	})

	slot := c.chainTime.CurrentSlot()
	if slot > c.slot {
		slot = c.slot
	}

	builder.WriteString("Slots through period: ")
	builder.WriteString(fmt.Sprintf("%d\n", slot-phase0.Slot(c.period*(c.slotsPerEpoch*c.epochsPerEth1VotingPeriod))))

	builder.WriteString("Votes this period: ")
	builder.WriteString(fmt.Sprintf("%d\n", totalVotes))

	if len(votes) > 0 {
		if c.verbose {
			for _, vote := range votes {
				builder.WriteString(fmt.Sprintf("  block %#x, deposit count %d: %d vote", vote.Vote.BlockHash, vote.Vote.DepositCount, vote.Count))
				if vote.Count != 1 {
					builder.WriteString("s\n")
				} else {
					builder.WriteString("\n")
				}
			}
		} else {
			builder.WriteString(fmt.Sprintf("Leading vote is for block %#x with %d votes\n", votes[0].Vote.BlockHash, votes[0].Count))
		}
	}

	return strings.TrimSuffix(builder.String(), "\n"), nil
}
