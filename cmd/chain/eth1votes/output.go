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
	Period      uint64           `json:"period"`
	PeriodStart int64            `json:"period_start"`
	PeriodEnd   int64            `json:"period_end"`
	Epoch       phase0.Epoch     `json:"epoch"`
	Slot        phase0.Slot      `json:"slot"`
	Incumbent   *phase0.ETH1Data `json:"incumbent"`
	Votes       []*vote          `json:"votes"`
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

func (c *command) outputJSON(_ context.Context) (string, error) {
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
		Period:      c.period,
		PeriodStart: c.periodStart.Unix(),
		PeriodEnd:   c.periodEnd.Unix(),
		Epoch:       c.epoch,
		Slot:        c.slot,
		Incumbent:   c.incumbent,
		Votes:       votes,
	}
	data, err := json.Marshal(output)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (c *command) outputText(_ context.Context) (string, error) {
	builder := strings.Builder{}

	builder.WriteString("Voting period: ")
	builder.WriteString(fmt.Sprintf("%d\n", c.period))

	if c.verbose {
		builder.WriteString("Period start: ")
		builder.WriteString(fmt.Sprintf("%s\n", c.periodStart))
		builder.WriteString("Period end: ")
		builder.WriteString(fmt.Sprintf("%s\n", c.periodEnd))

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

	slotsThroughPeriod := slot + 1 - phase0.Slot(c.period*(c.slotsPerEpoch*c.epochsPerEth1VotingPeriod))
	builder.WriteString("Slots through period: ")
	builder.WriteString(fmt.Sprintf("%d (%d)\n", slotsThroughPeriod, c.slot))

	builder.WriteString("Votes this period: ")
	builder.WriteString(fmt.Sprintf("%d\n", totalVotes))

	if len(votes) > 0 {
		if c.verbose {
			for _, vote := range votes {
				builder.WriteString(fmt.Sprintf("  block %#x, deposit count %d: %d vote", vote.Vote.BlockHash, vote.Vote.DepositCount, vote.Count))
				if vote.Count != 1 {
					builder.WriteString("s")
				}
				builder.WriteString(fmt.Sprintf(" (%0.2f%%)\n", 100.0*float64(vote.Count)/float64(slotsThroughPeriod)))
			}
		} else {
			builder.WriteString(fmt.Sprintf("Leading vote is for block %#x with %d votes (%0.2f%%)\n", votes[0].Vote.BlockHash, votes[0].Count, 100.0*float64(votes[0].Count)/float64(slotsThroughPeriod)))
		}
	}

	return strings.TrimSuffix(builder.String(), "\n"), nil
}
