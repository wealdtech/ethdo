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
	"strconv"

	eth2client "github.com/attestantio/go-eth2-client"
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
	state, err := c.beaconStateProvider.BeaconState(ctx, fmt.Sprintf("%d", fetchSlot))
	if err != nil {
		return errors.Wrap(err, "failed to obtain state")
	}
	if state == nil {
		return errors.New("state not returned by beacon node")
	}

	if c.debug {
		data, err := json.Marshal(state)
		if err == nil {
			fmt.Printf("%s\n", string(data))
		}
	}

	switch state.Version {
	case spec.DataVersionPhase0:
		c.slot = phase0.Slot(state.Phase0.Slot)
		c.incumbent = state.Phase0.ETH1Data
		c.eth1DataVotes = state.Phase0.ETH1DataVotes
	case spec.DataVersionAltair:
		c.slot = state.Altair.Slot
		c.incumbent = state.Altair.ETH1Data
		c.eth1DataVotes = state.Altair.ETH1DataVotes
	case spec.DataVersionBellatrix:
		c.slot = state.Bellatrix.Slot
		c.incumbent = state.Bellatrix.ETH1Data
		c.eth1DataVotes = state.Bellatrix.ETH1DataVotes
	case spec.DataVersionCapella:
		c.slot = state.Capella.Slot
		c.incumbent = state.Capella.ETH1Data
		c.eth1DataVotes = state.Capella.ETH1DataVotes
	default:
		return fmt.Errorf("unhandled beacon state version %v", state.Version)
	}

	c.period = uint64(c.epoch) / c.epochsPerEth1VotingPeriod

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
	c.eth2Client, err = util.ConnectToBeaconNode(ctx, c.connection, c.timeout, c.allowInsecureConnections)
	if err != nil {
		return errors.Wrap(err, "failed to connect to beacon node")
	}

	c.chainTime, err = standardchaintime.New(ctx,
		standardchaintime.WithSpecProvider(c.eth2Client.(eth2client.SpecProvider)),
		standardchaintime.WithForkScheduleProvider(c.eth2Client.(eth2client.ForkScheduleProvider)),
		standardchaintime.WithGenesisTimeProvider(c.eth2Client.(eth2client.GenesisTimeProvider)),
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

	spec, err := specProvider.Spec(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to obtain spec")
	}

	tmp, exists := spec["SLOTS_PER_EPOCH"]
	if !exists {
		return errors.New("spec did not contain SLOTS_PER_EPOCH")
	}
	var good bool
	c.slotsPerEpoch, good = tmp.(uint64)
	if !good {
		return errors.New("SLOTS_PER_EPOCH value invalid")
	}
	tmp, exists = spec["EPOCHS_PER_ETH1_VOTING_PERIOD"]
	if !exists {
		return errors.New("spec did not contain EPOCHS_PER_ETH1_VOTING_PERIOD")
	}
	c.epochsPerEth1VotingPeriod, good = tmp.(uint64)
	if !good {
		return errors.New("EPOCHS_PER_ETH1_VOTING_PERIOD value invalid")
	}

	return nil
}
