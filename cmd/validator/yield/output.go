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

package validatoryield

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/shopspring/decimal"
	"github.com/wealdtech/go-string2eth"
)

func (c *command) output(_ context.Context) (string, error) {
	if c.quiet {
		return "", nil
	}

	if c.json {
		data, err := json.Marshal(c.results)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}

	builder := strings.Builder{}

	if c.verbose {
		builder.WriteString("Per-validator rewards per epoch: ")
		builder.WriteString(string2eth.WeiToGWeiString(c.results.ValidatorRewardsPerEpoch.BigInt()))
		builder.WriteString("\n")

		builder.WriteString("Per-validator rewards per year: ")
		builder.WriteString(string2eth.WeiToString(c.results.ValidatorRewardsPerYear.BigInt(), true))
		builder.WriteString("\n")

		builder.WriteString("Expected per-validator rewards per epoch (with full participation): ")
		builder.WriteString(string2eth.WeiToGWeiString(c.results.ExpectedValidatorRewardsPerEpoch.BigInt()))
		builder.WriteString("\n")

		builder.WriteString("Maximum chain issuance per epoch: ")
		builder.WriteString(string2eth.WeiToString(c.results.MaxIssuancePerEpoch.BigInt(), true))
		builder.WriteString("\n")

		builder.WriteString("Maximum chain issuance per year: ")
		builder.WriteString(string2eth.WeiToString(c.results.MaxIssuancePerYear.BigInt(), true))
		builder.WriteString("\n")
	}

	builder.WriteString("Yield: ")
	builder.WriteString(c.results.Yield.Mul(decimal.New(100, 0)).StringFixed(2))
	builder.WriteString("%\n")

	return builder.String(), nil
}
