// Copyright © 2023 Weald Technology Trading.
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

package validatorwithdrawl

import (
	"context"
	"fmt"
	"os"
	"time"

	consensusclient "github.com/attestantio/go-eth2-client"
	"github.com/attestantio/go-eth2-client/api"
	apiv1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
	standardchaintime "github.com/wealdtech/ethdo/services/chaintime/standard"
	"github.com/wealdtech/ethdo/util"
)

const (
	ethWithdrawalPrefix = 0x01
)

func (c *command) process(ctx context.Context) error {
	if err := c.setup(ctx); err != nil {
		return err
	}

	validator, err := util.ParseValidator(ctx, c.consensusClient.(consensusclient.ValidatorsProvider), c.validator, "head")
	if err != nil {
		return errors.Wrap(err, "failed to parse validator")
	}

	if validator.Validator.WithdrawalCredentials[0] != ethWithdrawalPrefix {
		return errors.New("validator does not have suitable withdrawal credentials")
	}
	if validator.Balance == 0 {
		return errors.New("validator has nothing to withdraw")
	}

	blockResponse, err := c.consensusClient.(consensusclient.SignedBeaconBlockProvider).SignedBeaconBlock(ctx, &api.SignedBeaconBlockOpts{
		Block: "head",
	})
	if err != nil {
		return errors.Wrap(err, "failed to obtain block")
	}
	block := blockResponse.Data
	slot, err := block.Slot()
	if err != nil {
		return errors.Wrap(err, "failed to obtain block slot")
	}
	if c.debug {
		fmt.Fprintf(os.Stderr, "Slot is %d\n", slot)
	}

	response, err := c.consensusClient.(consensusclient.ValidatorsProvider).Validators(ctx, &api.ValidatorsOpts{
		State: fmt.Sprintf("%d", slot),
	})
	if err != nil {
		return errors.Wrap(err, "failed to obtain validators")
	}
	validators := make([]*apiv1.Validator, len(response.Data))
	for _, validator := range response.Data {
		validators[validator.Index] = validator
	}

	withdrawals, err := block.Withdrawals()
	if err != nil {
		return errors.Wrap(err, "failed to obtain withdrawals from block")
	}
	if len(withdrawals) == 0 {
		return errors.New("block without withdrawals; cannot obtain next withdrawal validator index")
	}
	nextWithdrawalValidatorIndex := phase0.ValidatorIndex((int(withdrawals[len(withdrawals)-1].ValidatorIndex) + 1) % len(validators))

	if c.debug {
		fmt.Fprintf(os.Stderr, "Next withdrawal validator index is %d\n", nextWithdrawalValidatorIndex)
	}

	index := int(nextWithdrawalValidatorIndex)
	for {
		if index == len(validators) {
			index = 0
		}
		if index == int(validator.Index) {
			break
		}
		if validators[index].Validator.WithdrawalCredentials[0] == ethWithdrawalPrefix &&
			validators[index].Validator.EffectiveBalance == c.maxEffectiveBalance {
			c.res.WithdrawalsToGo++
		}
		index++
	}

	c.res.BlocksToGo = c.res.WithdrawalsToGo / c.maxWithdrawalsPerPayload
	if c.res.WithdrawalsToGo%c.maxWithdrawalsPerPayload != 0 {
		c.res.BlocksToGo++
	}
	c.res.Block = uint64(slot) + c.res.BlocksToGo
	c.res.Expected = c.chainTime.StartOfSlot(phase0.Slot(c.res.Block))
	c.res.Wait = time.Until(c.res.Expected)

	return nil
}

func (c *command) setup(ctx context.Context) error {
	// Connect to the consensus node.
	var err error
	c.consensusClient, err = util.ConnectToBeaconNode(ctx, &util.ConnectOpts{
		Address:       c.connection,
		Timeout:       c.timeout,
		AllowInsecure: c.allowInsecureConnections,
		LogFallback:   !c.quiet,
	})
	if err != nil {
		return err
	}

	// Set up chaintime.
	c.chainTime, err = standardchaintime.New(ctx,
		standardchaintime.WithGenesisProvider(c.consensusClient.(consensusclient.GenesisProvider)),
		standardchaintime.WithSpecProvider(c.consensusClient.(consensusclient.SpecProvider)),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create chaintime service")
	}

	specResponse, err := c.consensusClient.(consensusclient.SpecProvider).Spec(ctx, &api.SpecOpts{})
	if err != nil {
		return errors.Wrap(err, "failed to obtain spec")
	}

	if val, exists := specResponse.Data["MAX_WITHDRAWALS_PER_PAYLOAD"]; !exists {
		c.maxWithdrawalsPerPayload = 16
	} else {
		c.maxWithdrawalsPerPayload = val.(uint64)
	}

	if val, exists := specResponse.Data["MAX_EFFECTIVE_BALANCE"]; !exists {
		c.maxEffectiveBalance = 32000000000
	} else {
		c.maxEffectiveBalance = phase0.Gwei(val.(uint64))
	}

	return nil
}
