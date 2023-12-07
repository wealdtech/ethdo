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
	"fmt"
	"math/big"
	"strconv"

	eth2client "github.com/attestantio/go-eth2-client"
	"github.com/attestantio/go-eth2-client/api"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	standardchaintime "github.com/wealdtech/ethdo/services/chaintime/standard"
	"github.com/wealdtech/ethdo/util"
)

func (c *command) process(ctx context.Context) error {
	// Obtain information we need to process.
	if err := c.setup(ctx); err != nil {
		return err
	}

	if c.debug {
		fmt.Printf("Active validators: %v\n", c.results.ActiveValidators)
		fmt.Printf("Active validator balance: %v\n", c.results.ActiveValidatorBalance)
	}

	return c.calculateYield(ctx)
}

var (
	weiPerGwei    = decimal.New(1e9, 0)
	one           = decimal.New(1, 0)
	epochsPerYear = decimal.New(225*365, 0)
)

// calculateYield calculates yield from the number of active validators.
func (c *command) calculateYield(ctx context.Context) error {
	specResponse, err := c.eth2Client.(eth2client.SpecProvider).Spec(ctx, &api.SpecOpts{})
	if err != nil {
		return err
	}

	tmp, exists := specResponse.Data["BASE_REWARD_FACTOR"]
	if !exists {
		return errors.New("spec missing BASE_REWARD_FACTOR")
	}
	baseReward, isType := tmp.(uint64)
	if !isType {
		return errors.New("BASE_REWARD_FACTOR of incorrect type")
	}
	if c.debug {
		fmt.Printf("Base reward: %v\n", baseReward)
	}
	c.results.BaseReward = decimal.New(int64(baseReward), 0)

	numerator := decimal.New(32, 0).Mul(weiPerGwei).Mul(c.results.BaseReward)
	if c.debug {
		fmt.Printf("Numerator: %v\n", numerator)
	}
	activeValidatorsBalanceInGwei := c.results.ActiveValidatorBalance.Div(weiPerGwei)
	denominator := decimal.NewFromBigInt(new(big.Int).Sqrt(activeValidatorsBalanceInGwei.BigInt()), 0)
	if c.debug {
		fmt.Printf("Denominator: %v\n", denominator)
	}
	c.results.ValidatorRewardsPerEpoch = numerator.Div(denominator).RoundDown(0).Mul(weiPerGwei)
	if c.debug {
		fmt.Printf("Validator rewards per epoch: %v\n", c.results.ValidatorRewardsPerEpoch)
	}
	c.results.ValidatorRewardsPerYear = c.results.ValidatorRewardsPerEpoch.Mul(epochsPerYear)
	if c.debug {
		fmt.Printf("Validator rewards per year: %v\n", c.results.ValidatorRewardsPerYear)
	}
	// Expected validator rewards assume that there is no proposal and no sync committee participation,
	// but that head/source/target are correct and timely: this gives 54/64 of the reward.
	// These values are obtained from https://github.com/ethereum/consensus-specs/blob/dev/specs/altair/beacon-chain.md#incentivization-weights
	c.results.ExpectedValidatorRewardsPerEpoch = c.results.ValidatorRewardsPerEpoch.Mul(decimal.New(54, 0)).Div(decimal.New(64, 0)).Div(weiPerGwei).RoundDown(0).Mul(weiPerGwei)
	if c.debug {
		fmt.Printf("Expected validator rewards per epoch: %v\n", c.results.ExpectedValidatorRewardsPerEpoch)
	}

	c.results.MaxIssuancePerEpoch = c.results.ValidatorRewardsPerEpoch.Mul(c.results.ActiveValidators)
	if c.debug {
		fmt.Printf("Chain rewards per epoch: %v\n", c.results.MaxIssuancePerEpoch)
	}
	c.results.MaxIssuancePerYear = c.results.MaxIssuancePerEpoch.Mul(epochsPerYear)
	if c.debug {
		fmt.Printf("Chain rewards per year: %v\n", c.results.MaxIssuancePerYear)
	}

	c.results.Yield = c.results.ValidatorRewardsPerYear.Div(weiPerGwei).Div(weiPerGwei).Div(decimal.New(32, 0))
	if c.debug {
		fmt.Printf("Yield: %v\n", c.results.Yield)
	}

	return nil
}

func (c *command) setup(ctx context.Context) error {
	var err error

	// Connect to the client.
	c.eth2Client, err = util.ConnectToBeaconNode(ctx, &util.ConnectOpts{
		Address:       c.connection,
		Timeout:       c.timeout,
		AllowInsecure: c.allowInsecureConnections,
		LogFallback:   !c.quiet,
	})
	if err != nil {
		return errors.Wrap(err, "failed to connect to beacon node")
	}

	if c.validators == "" {
		chainTime, err := standardchaintime.New(ctx,
			standardchaintime.WithSpecProvider(c.eth2Client.(eth2client.SpecProvider)),
			standardchaintime.WithGenesisProvider(c.eth2Client.(eth2client.GenesisProvider)),
		)
		if err != nil {
			return errors.Wrap(err, "failed to set up chaintime service")
		}

		// Obtain the number of active validators.
		var isProvider bool
		validatorsProvider, isProvider := c.eth2Client.(eth2client.ValidatorsProvider)
		if !isProvider {
			return errors.New("connection does not provide validator information")
		}

		epoch, err := util.ParseEpoch(ctx, chainTime, c.epoch)
		if err != nil {
			return errors.Wrap(err, "failed to parse epoch")
		}

		response, err := validatorsProvider.Validators(ctx, &api.ValidatorsOpts{
			State: fmt.Sprintf("%d", chainTime.FirstSlotOfEpoch(epoch)),
		})
		if err != nil {
			return err
		}

		activeValidators := decimal.Zero
		activeValidatorBalance := decimal.Zero
		for _, validator := range response.Data {
			if validator.Validator.ActivationEpoch <= epoch &&
				validator.Validator.ExitEpoch > epoch {
				activeValidators = activeValidators.Add(one)
				activeValidatorBalance = activeValidatorBalance.Add(decimal.NewFromInt(int64(validator.Validator.EffectiveBalance)))
			}
		}
		c.results.ActiveValidators = activeValidators
		c.results.ActiveValidatorBalance = activeValidatorBalance.Mul(weiPerGwei)
	} else {
		activeValidators, err := strconv.ParseInt(c.validators, 0, 64)
		if err != nil {
			return errors.Wrap(err, "failed to parse number of validators")
		}
		if activeValidators <= 0 {
			return errors.New("number of validators must be greater than 0")
		}

		c.results.ActiveValidators = decimal.New(activeValidators, 0)
		c.results.ActiveValidatorBalance = decimal.New(32, 0).Mul(c.results.ActiveValidators).Mul(weiPerGwei).Mul(weiPerGwei)
		if c.debug {
			fmt.Println("Assuming 32Ξ per validator")
		}
	}

	return nil
}
