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

package validatorsummary

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
	if len(c.summary.NonParticipatingValidators) > 0 {
		builder.WriteString("  Non-participating validators:\n")
		for _, validator := range c.summary.NonParticipatingValidators {
			builder.WriteString(fmt.Sprintf("    %d (slot %d, committee %d)\n", validator.Validator, validator.Slot, validator.Committee))
		}
	}
	if len(c.summary.IncorrectHeadValidators) > 0 {
		builder.WriteString("  Incorrect head validators:\n")
		for _, validator := range c.summary.IncorrectHeadValidators {
			builder.WriteString(fmt.Sprintf("    %d (slot %d, committee %d)\n", validator.Validator, validator.AttestationData.Slot, validator.AttestationData.Index))
		}
	}
	if len(c.summary.UntimelyHeadValidators) > 0 {
		builder.WriteString("  Untimely head validators:\n")
		for _, validator := range c.summary.UntimelyHeadValidators {
			builder.WriteString(fmt.Sprintf("    %d (slot %d, committee %d, inclusion distance %d)\n", validator.Validator, validator.AttestationData.Slot, validator.AttestationData.Index, validator.InclusionDistance))
		}
	}
	if len(c.summary.UntimelySourceValidators) > 0 {
		builder.WriteString("  Untimely source validators:\n")
		for _, validator := range c.summary.UntimelySourceValidators {
			builder.WriteString(fmt.Sprintf("    %d (slot %d, committee %d, inclusion distance %d)\n", validator.Validator, validator.AttestationData.Slot, validator.AttestationData.Index, validator.InclusionDistance))
		}
	}
	if len(c.summary.IncorrectTargetValidators) > 0 {
		builder.WriteString("  Incorrect target validators:\n")
		for _, validator := range c.summary.IncorrectTargetValidators {
			builder.WriteString(fmt.Sprintf("    %d (slot %d, committee %d)\n", validator.Validator, validator.AttestationData.Slot, validator.AttestationData.Index))
		}
	}
	if len(c.summary.UntimelyTargetValidators) > 0 {
		builder.WriteString("  Untimely target validators:\n")
		for _, validator := range c.summary.UntimelyTargetValidators {
			builder.WriteString(fmt.Sprintf("    %d (slot %d, committee %d, inclusion distance %d)\n", validator.Validator, validator.AttestationData.Slot, validator.AttestationData.Index, validator.InclusionDistance))
		}
	}

	return builder.String(), nil
}
