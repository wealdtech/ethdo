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

package blockanalyze

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
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

type attestationAnalysisJSON struct {
	Head          string           `json:"head"`
	Target        string           `json:"target"`
	Distance      int              `json:"distance"`
	Duplicate     *attestationData `json:"duplicate,omitempty"`
	NewVotes      int              `json:"new_votes"`
	Votes         int              `json:"votes"`
	PossibleVotes int              `json:"possible_votes"`
	HeadCorrect   bool             `json:"head_correct"`
	HeadTimely    bool             `json:"head_timely"`
	SourceTimely  bool             `json:"source_timely"`
	TargetCorrect bool             `json:"target_correct"`
	TargetTimely  bool             `json:"target_timely"`
	Score         float64          `json:"score"`
	Value         float64          `json:"value"`
}

func (a *attestationAnalysis) MarshalJSON() ([]byte, error) {
	return json.Marshal(attestationAnalysisJSON{
		Head:          fmt.Sprintf("%#x", a.Head),
		Target:        fmt.Sprintf("%#x", a.Target),
		Distance:      a.Distance,
		Duplicate:     a.Duplicate,
		NewVotes:      a.NewVotes,
		Votes:         a.Votes,
		PossibleVotes: a.PossibleVotes,
		HeadCorrect:   a.HeadCorrect,
		HeadTimely:    a.HeadTimely,
		SourceTimely:  a.SourceTimely,
		TargetCorrect: a.TargetCorrect,
		TargetTimely:  a.TargetTimely,
		Score:         a.Score,
		Value:         a.Value,
	})
}

func (c *command) outputJSON(_ context.Context) (string, error) {
	data, err := json.Marshal(c.analysis)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (c *command) outputTxt(_ context.Context) (string, error) {
	builder := strings.Builder{}

	for i, attestation := range c.analysis.Attestations {
		if c.verbose {
			builder.WriteString("Attestation ")
			builder.WriteString(strconv.Itoa(i))
			builder.WriteString(": ")
			builder.WriteString("distance ")
			builder.WriteString(strconv.Itoa(attestation.Distance))
			builder.WriteString(", ")

			if attestation.Duplicate != nil {
				builder.WriteString("duplicate of attestation ")
				builder.WriteString(strconv.Itoa(attestation.Duplicate.Index))
				builder.WriteString(" in block ")
				builder.WriteString(fmt.Sprintf("%d", attestation.Duplicate.Block))
				builder.WriteString("\n")
				continue
			}

			builder.WriteString(strconv.Itoa(attestation.NewVotes))
			builder.WriteString("/")
			builder.WriteString(strconv.Itoa(attestation.Votes))
			builder.WriteString("/")
			builder.WriteString(strconv.Itoa(attestation.PossibleVotes))
			builder.WriteString(" new/total/possible votes")
			if attestation.NewVotes == 0 {
				builder.WriteString("\n")
				continue
			}
			builder.WriteString(", ")
			switch {
			case !attestation.HeadCorrect:
				builder.WriteString("head vote incorrect, ")
			case !attestation.HeadTimely:
				builder.WriteString("head vote correct but late, ")
			}

			if !attestation.SourceTimely {
				builder.WriteString("source vote late, ")
			}

			switch {
			case !attestation.TargetCorrect:
				builder.WriteString("target vote incorrect, ")
			case !attestation.TargetTimely:
				builder.WriteString("target vote correct but late, ")
			}

			builder.WriteString("score ")
			builder.WriteString(fmt.Sprintf("%0.3f", attestation.Score))
			builder.WriteString(", value ")
			builder.WriteString(fmt.Sprintf("%0.3f", attestation.Value))
			builder.WriteString("\n")
		}
	}

	if c.analysis.SyncCommitee.Contributions > 0 {
		if c.verbose {
			builder.WriteString("Sync committee contributions: ")
			builder.WriteString(strconv.Itoa(c.analysis.SyncCommitee.Contributions))
			builder.WriteString(" contributions, score ")
			builder.WriteString(fmt.Sprintf("%0.3f", c.analysis.SyncCommitee.Score))
			builder.WriteString(", value ")
			builder.WriteString(fmt.Sprintf("%0.3f", c.analysis.SyncCommitee.Value))
			builder.WriteString("\n")
		}
	}

	builder.WriteString("Value for block ")
	builder.WriteString(fmt.Sprintf("%d", c.analysis.Slot))
	builder.WriteString(": ")
	builder.WriteString(fmt.Sprintf("%0.3f", c.analysis.Value))
	builder.WriteString("\n")

	return builder.String(), nil
}
