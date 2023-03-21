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

package chainverifysignedcontributionandproof

import (
	"context"
	"strings"
)

func (c *command) output(_ context.Context) (string, error) {
	if c.quiet {
		return "", nil
	}

	builder := strings.Builder{}

	builder.WriteString("Valid data structure: ")
	if c.itemStructureValid {
		builder.WriteString("✓\n")
	} else {
		builder.WriteString("✕")
		if c.additionalInfo != "" {
			builder.WriteString(" (")
			builder.WriteString(c.additionalInfo)
			builder.WriteString(")")
		}
		builder.WriteString("\n")
		return builder.String(), nil
	}

	builder.WriteString("Validator known: ")
	if c.validatorKnown {
		builder.WriteString("✓\n")
	} else {
		builder.WriteString("✕")
		if c.additionalInfo != "" {
			builder.WriteString(" (")
			builder.WriteString(c.additionalInfo)
			builder.WriteString(")")
		}
		builder.WriteString("\n")
		return builder.String(), nil
	}

	builder.WriteString("Validator in sync committee: ")
	if c.validatorInSyncCommittee {
		builder.WriteString("✓\n")
	} else {
		builder.WriteString("✕")
		if c.additionalInfo != "" {
			builder.WriteString(" (")
			builder.WriteString(c.additionalInfo)
			builder.WriteString(")")
		}
		builder.WriteString("\n")
		return builder.String(), nil
	}

	builder.WriteString("Validator is aggregator: ")
	if c.validatorIsAggregator {
		builder.WriteString("✓\n")
	} else {
		builder.WriteString("✕")
		if c.additionalInfo != "" {
			builder.WriteString(" (")
			builder.WriteString(c.additionalInfo)
			builder.WriteString(")")
		}
		builder.WriteString("\n")
		return builder.String(), nil
	}

	builder.WriteString("Contribution signature has valid format: ")
	if c.contributionSignatureValidFormat {
		builder.WriteString("✓\n")
	} else {
		builder.WriteString("✕")
		if c.additionalInfo != "" {
			builder.WriteString(" (")
			builder.WriteString(c.additionalInfo)
			builder.WriteString(")")
		}
		builder.WriteString("\n")
		return builder.String(), nil
	}

	builder.WriteString("Contribution and proof signature has valid format: ")
	if c.contributionAndProofSignatureValidFormat {
		builder.WriteString("✓\n")
	} else {
		builder.WriteString("✕")
		if c.additionalInfo != "" {
			builder.WriteString(" (")
			builder.WriteString(c.additionalInfo)
			builder.WriteString(")")
		}
		builder.WriteString("\n")
		return builder.String(), nil
	}

	builder.WriteString("Contribution and proof signature is valid: ")
	if c.contributionAndProofSignatureValid {
		builder.WriteString("✓\n")
	} else {
		builder.WriteString("✕")
		if c.additionalInfo != "" {
			builder.WriteString(" (")
			builder.WriteString(c.additionalInfo)
			builder.WriteString(")")
		}
		builder.WriteString("\n")
		return builder.String(), nil
	}

	return builder.String(), nil
}
