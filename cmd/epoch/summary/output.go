// Copyright © 2022, 2025 Weald Technology Trading.
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
	"math/big"
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
	missedProposals := make([]string, 0, len(c.summary.Proposals))
	for _, proposal := range c.summary.Proposals {
		if !proposal.Block {
			missedProposals = append(missedProposals, fmt.Sprintf("\n    Slot %d (validator %d)", proposal.Slot, proposal.ValidatorIndex))
		} else {
			proposedBlocks++
		}
	}
	builder.WriteString(fmt.Sprintf("  Proposals: %d/%d (%0.2f%%)", proposedBlocks, len(missedProposals)+proposedBlocks, 100.0*float64(proposedBlocks)/float64(len(missedProposals)+proposedBlocks)))
	if c.verbose {
		for _, proposal := range c.summary.Proposals {
			if proposal.Block {
				continue
			}
			builder.WriteString("\n    Slot ")
			builder.WriteString(fmt.Sprintf("%d (%d/%d)", proposal.Slot, uint64(proposal.Slot)%uint64(len(c.summary.Proposals)), len(c.summary.Proposals)))
			builder.WriteString(" validator ")
			builder.WriteString(fmt.Sprintf("%d", proposal.ValidatorIndex))
			builder.WriteString(" not proposed or not included")
		}
	}

	gweiToEth := big.NewInt(1e9)
	mul := big.NewInt(10000)
	participatingBalancePct := new(big.Int).Div(new(big.Int).Mul(c.summary.ParticipatingBalance, mul), c.summary.ActiveBalance)
	builder.WriteString(fmt.Sprintf("\n  Attesting balance: %s/%s (%0.2f%%)",
		new(big.Int).Div(c.summary.ParticipatingBalance, gweiToEth).String(),
		new(big.Int).Div(c.summary.ActiveBalance, gweiToEth).String(),
		float64(participatingBalancePct.Uint64())/100.0,
	))

	sourceTimelyBalancePct := new(big.Int).Div(new(big.Int).Mul(c.summary.SourceTimelyBalance, mul), c.summary.ActiveBalance)
	builder.WriteString(fmt.Sprintf("\n    Source timely: %s/%s (%0.2f%%)",
		new(big.Int).Div(c.summary.SourceTimelyBalance, gweiToEth).String(),
		new(big.Int).Div(c.summary.ActiveBalance, gweiToEth).String(),
		float64(sourceTimelyBalancePct.Uint64())/100.0,
	))

	targetCorrectBalancePct := new(big.Int).Div(new(big.Int).Mul(c.summary.TargetCorrectBalance, mul), c.summary.ActiveBalance)
	builder.WriteString(fmt.Sprintf("\n    Target correct: %s/%s (%0.2f%%)",
		new(big.Int).Div(c.summary.TargetCorrectBalance, gweiToEth).String(),
		new(big.Int).Div(c.summary.ActiveBalance, gweiToEth).String(),
		float64(targetCorrectBalancePct.Uint64())/100.0,
	))

	targetTimelyBalancePct := new(big.Int).Div(new(big.Int).Mul(c.summary.TargetTimelyBalance, mul), c.summary.ActiveBalance)
	builder.WriteString(fmt.Sprintf("\n    Target timely: %s/%s (%0.2f%%)",
		new(big.Int).Div(c.summary.TargetTimelyBalance, gweiToEth).String(),
		new(big.Int).Div(c.summary.ActiveBalance, gweiToEth).String(),
		float64(targetTimelyBalancePct.Uint64())/100.0,
	))

	headCorrectBalancePct := new(big.Int).Div(new(big.Int).Mul(c.summary.HeadCorrectBalance, mul), c.summary.ActiveBalance)
	builder.WriteString(fmt.Sprintf("\n    Head correct: %s/%s (%0.2f%%)",
		new(big.Int).Div(c.summary.HeadCorrectBalance, gweiToEth).String(),
		new(big.Int).Div(c.summary.ActiveBalance, gweiToEth).String(),
		float64(headCorrectBalancePct.Uint64())/100.0,
	))

	headTimelyBalancePct := new(big.Int).Div(new(big.Int).Mul(c.summary.HeadTimelyBalance, mul), c.summary.ActiveBalance)
	builder.WriteString(fmt.Sprintf("\n    Head timely: %s/%s (%0.2f%%)",
		new(big.Int).Div(c.summary.HeadTimelyBalance, gweiToEth).String(),
		new(big.Int).Div(c.summary.ActiveBalance, gweiToEth).String(),
		float64(headTimelyBalancePct.Uint64())/100.0,
	))

	builder.WriteString(fmt.Sprintf("\n  Attesting validators: %d/%d (%0.2f%%)",
		c.summary.ParticipatingValidators,
		c.summary.ActiveValidators,
		100.0*float64(c.summary.ParticipatingValidators)/float64(c.summary.ActiveValidators),
	))
	builder.WriteString(fmt.Sprintf("\n    Source timely: %d/%d (%0.2f%%)",
		c.summary.SourceTimelyValidators,
		c.summary.ActiveValidators,
		100.0*float64(c.summary.SourceTimelyValidators)/float64(c.summary.ActiveValidators),
	))
	builder.WriteString(fmt.Sprintf("\n    Target correct: %d/%d (%0.2f%%)",
		c.summary.TargetCorrectValidators,
		c.summary.ActiveValidators,
		100.0*float64(c.summary.TargetCorrectValidators)/float64(c.summary.ActiveValidators),
	))
	builder.WriteString(fmt.Sprintf("\n    Target timely: %d/%d (%0.2f%%)",
		c.summary.TargetTimelyValidators,
		c.summary.ActiveValidators,
		100.0*float64(c.summary.TargetTimelyValidators)/float64(c.summary.ActiveValidators),
	))
	builder.WriteString(fmt.Sprintf("\n    Head correct: %d/%d (%0.2f%%)",
		c.summary.HeadCorrectValidators,
		c.summary.ActiveValidators,
		100.0*float64(c.summary.HeadCorrectValidators)/float64(c.summary.ActiveValidators),
	))
	builder.WriteString(fmt.Sprintf("\n    Head timely: %d/%d (%0.2f%%)",
		c.summary.HeadTimelyValidators,
		c.summary.ActiveValidators,
		100.0*float64(c.summary.HeadTimelyValidators)/float64(c.summary.ActiveValidators),
	))
	if c.verbose {
		// Sort list by validator index.
		for _, validator := range c.summary.NonParticipatingValidators {
			builder.WriteString("\n    Slot ")
			builder.WriteString(fmt.Sprintf("%d", validator.Slot))
			builder.WriteString(" committee ")
			builder.WriteString(fmt.Sprintf("%d", validator.Committee))
			builder.WriteString(" validator ")
			builder.WriteString(fmt.Sprintf("%d", validator.Validator))
			builder.WriteString(" failed to participate")
		}
	}

	if c.summary.Epoch >= c.chainTime.AltairInitialEpoch() {
		contributions := proposedBlocks * 512 // SYNC_COMMITTEE_SIZE
		totalMissed := 0
		for _, contribution := range c.summary.SyncCommittee {
			totalMissed += contribution.Missed
		}
		builder.WriteString(fmt.Sprintf("\n  Sync committees: %d/%d (%0.2f%%)", contributions-totalMissed, contributions, 100.0*float64(contributions-totalMissed)/float64(contributions)))
		if c.verbose {
			for _, syncCommittee := range c.summary.SyncCommittee {
				builder.WriteString("\n    Validator ")
				builder.WriteString(fmt.Sprintf("%d", syncCommittee.ValidatorIndex))
				builder.WriteString(" included ")
				builder.WriteString(fmt.Sprintf("%d/%d", proposedBlocks-syncCommittee.Missed, proposedBlocks))
				builder.WriteString(fmt.Sprintf(" (%0.2f%%)", 100.0*float64(proposedBlocks-syncCommittee.Missed)/float64(proposedBlocks)))
			}
		}
	}

	return builder.String(), nil
}
