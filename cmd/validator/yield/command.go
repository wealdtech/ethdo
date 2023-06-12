// Copyright © 2022, 2023 Weald Technology Trading.
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
	"time"

	eth2client "github.com/attestantio/go-eth2-client"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"github.com/spf13/viper"
)

type command struct {
	quiet   bool
	verbose bool
	debug   bool
	json    bool

	// Beacon node connection.
	timeout                  time.Duration
	connection               string
	allowInsecureConnections bool

	// Input.
	validators string
	epoch      string

	// Data access.
	eth2Client eth2client.Service

	// Output.
	results *output
}

type output struct {
	BaseReward                       decimal.Decimal `json:"base_reward"`
	ActiveValidators                 decimal.Decimal `json:"active_validators"`
	ActiveValidatorBalance           decimal.Decimal `json:"active_validator_balance"`
	ValidatorRewardsPerEpoch         decimal.Decimal `json:"validator_rewards_per_epoch"`
	ValidatorRewardsPerYear          decimal.Decimal `json:"validator_rewards_per_year"`
	ValidatorRewardsAllCorrect       decimal.Decimal `json:"validator_rewards_all_correct"`
	ExpectedValidatorRewardsPerEpoch decimal.Decimal `json:"expected_validator_rewards_per_epoch"`
	MaxIssuancePerEpoch              decimal.Decimal `json:"max_issuance_per_epoch"`
	MaxIssuancePerYear               decimal.Decimal `json:"max_issuance_per_year"`
	Yield                            decimal.Decimal `json:"yield"`
}

func newCommand(_ context.Context) (*command, error) {
	c := &command{
		quiet:   viper.GetBool("quiet"),
		verbose: viper.GetBool("verbose"),
		debug:   viper.GetBool("debug"),
		json:    viper.GetBool("json"),
		epoch:   viper.GetString("epoch"),
		results: &output{},
	}

	// Timeout.
	if viper.GetDuration("timeout") == 0 {
		return nil, errors.New("timeout is required")
	}
	c.timeout = viper.GetDuration("timeout")

	c.connection = viper.GetString("connection")
	c.allowInsecureConnections = viper.GetBool("allow-insecure-connections")

	c.validators = viper.GetString("validators")

	return c, nil
}
