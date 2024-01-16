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

package chaineth1votes

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	eth2client "github.com/attestantio/go-eth2-client"
	"github.com/attestantio/go-eth2-client/api"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
	standardchaintime "github.com/wealdtech/ethdo/services/chaintime/standard"
	"github.com/wealdtech/ethdo/util"
)

func (c *command) process(ctx context.Context) error {
	// Obtain information we need to process.
	if err := c.setup(ctx); err != nil {
		return err
	}

	var err error
	if c.xperiod != "" {
		period, err := strconv.ParseUint(c.xperiod, 10, 64)
		if err != nil {
			return err
		}
		c.epoch = phase0.Epoch(c.epochsPerEth1VotingPeriod*(period+1)) - 1
	} else {
		c.epoch, err = util.ParseEpoch(ctx, c.chainTime, c.xepoch)
		if err != nil {
			return err
		}
	}

	// Do not fetch from the future.
	if c.epoch > c.chainTime.CurrentEpoch() {
		c.epoch = c.chainTime.CurrentEpoch()
	}

	// Need to fetch the state from the last slot of the epoch.
	fetchSlot := c.chainTime.FirstSlotOfEpoch(c.epoch+1) - 1
	// Do not fetch from the future.
	if fetchSlot > c.chainTime.CurrentSlot() {
		fetchSlot = c.chainTime.CurrentSlot()
	}
	stateResponse, err := c.beaconStateProvider.BeaconState(ctx, &api.BeaconStateOpts{
		State: fmt.Sprintf("%d", fetchSlot),
	})
	if err != nil {
		return errors.Wrap(err, "failed to obtain state")
	}
	state := stateResponse.Data
	if state == nil {
		return errors.New("state not returned by beacon node")
	}

	if c.debug {
		data, err := json.Marshal(state)
		if err == nil {
			fmt.Printf("%s\n", string(data))
		}
	}

	c.slot, err = state.Slot()
	if err != nil {
		return errors.Wrap(err, "failed to obtain slot")
	}
	switch state.Version {
	case spec.DataVersionPhase0:
		c.incumbent = state.Phase0.ETH1Data
		c.eth1DataVotes = state.Phase0.ETH1DataVotes
	case spec.DataVersionAltair:
		c.incumbent = state.Altair.ETH1Data
		c.eth1DataVotes = state.Altair.ETH1DataVotes
	case spec.DataVersionBellatrix:
		c.incumbent = state.Bellatrix.ETH1Data
		c.eth1DataVotes = state.Bellatrix.ETH1DataVotes
	case spec.DataVersionCapella:
		c.incumbent = state.Capella.ETH1Data
		c.eth1DataVotes = state.Capella.ETH1DataVotes
	case spec.DataVersionDeneb:
		c.incumbent = state.Deneb.ETH1Data
		c.eth1DataVotes = state.Deneb.ETH1DataVotes
	default:
		return fmt.Errorf("unhandled beacon state version %v", state.Version)
	}

	c.period = uint64(c.epoch) / c.epochsPerEth1VotingPeriod
	c.periodStart = c.chainTime.StartOfEpoch(phase0.Epoch(c.period * c.epochsPerEth1VotingPeriod))
	c.periodEnd = c.chainTime.StartOfEpoch(phase0.Epoch((c.period + 1) * c.epochsPerEth1VotingPeriod))

	c.votes = make(map[string]*vote)
	for _, eth1Vote := range c.eth1DataVotes {
		key := fmt.Sprintf("%#x:%d", eth1Vote.BlockHash, eth1Vote.DepositCount)
		if _, exists := c.votes[key]; !exists {
			c.votes[key] = &vote{
				Vote: eth1Vote,
			}
		}
		c.votes[key].Count++
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

	c.chainTime, err = standardchaintime.New(ctx,
		standardchaintime.WithSpecProvider(c.eth2Client.(eth2client.SpecProvider)),
		standardchaintime.WithGenesisProvider(c.eth2Client.(eth2client.GenesisProvider)),
	)
	if err != nil {
		return errors.Wrap(err, "failed to set up chaintime service")
	}

	var isProvider bool
	c.beaconStateProvider, isProvider = c.eth2Client.(eth2client.BeaconStateProvider)
	if !isProvider {
		return errors.New("connection does not provide beacon state")
	}
	specProvider, isProvider := c.eth2Client.(eth2client.SpecProvider)
	if !isProvider {
		return errors.New("connection does not provide spec information")
	}

	specResponse, err := specProvider.Spec(ctx, &api.SpecOpts{})
	if err != nil {
		return errors.Wrap(err, "failed to obtain spec")
	}

	tmp, exists := specResponse.Data["SLOTS_PER_EPOCH"]
	if !exists {
		return errors.New("spec did not contain SLOTS_PER_EPOCH")
	}
	var good bool
	c.slotsPerEpoch, good = tmp.(uint64)
	if !good {
		return errors.New("SLOTS_PER_EPOCH value invalid")
	}
	tmp, exists = specResponse.Data["EPOCHS_PER_ETH1_VOTING_PERIOD"]
	if !exists {
		return errors.New("spec did not contain EPOCHS_PER_ETH1_VOTING_PERIOD")
	}
	c.epochsPerEth1VotingPeriod, good = tmp.(uint64)
	if !good {
		return errors.New("EPOCHS_PER_ETH1_VOTING_PERIOD value invalid")
	}

	return nil
}
