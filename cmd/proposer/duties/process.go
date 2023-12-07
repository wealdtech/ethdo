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

package proposerduties

import (
	"context"

	eth2client "github.com/attestantio/go-eth2-client"
	"github.com/attestantio/go-eth2-client/api"
	apiv1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/pkg/errors"
	standardchaintime "github.com/wealdtech/ethdo/services/chaintime/standard"
	"github.com/wealdtech/ethdo/util"
)

func (c *command) process(ctx context.Context) error {
	// Obtain information we need to process.
	if err := c.setup(ctx); err != nil {
		return err
	}

	if c.slot != "" {
		return c.processSlot(ctx)
	}

	return c.processEpoch(ctx)
}

func (c *command) processSlot(ctx context.Context) error {
	var err error
	slot, err := util.ParseSlot(ctx, c.chainTime, c.slot)
	if err != nil {
		return errors.Wrap(err, "failed to parse slot")
	}

	c.results.Epoch = c.chainTime.SlotToEpoch(slot)

	response, err := c.proposerDutiesProvider.ProposerDuties(ctx, &api.ProposerDutiesOpts{
		Epoch: c.results.Epoch,
	})
	if err != nil {
		return errors.Wrap(err, "failed to obtain proposer duties")
	}

	c.results.Duties = make([]*apiv1.ProposerDuty, 0, 1)

	for _, duty := range response.Data {
		if duty.Slot == slot {
			c.results.Duties = append(c.results.Duties, duty)
			break
		}
	}

	return nil
}

func (c *command) processEpoch(ctx context.Context) error {
	var err error
	c.results.Epoch, err = util.ParseEpoch(ctx, c.chainTime, c.epoch)
	if err != nil {
		return errors.Wrap(err, "failed to parse epoch")
	}

	dutiesResponse, err := c.proposerDutiesProvider.ProposerDuties(ctx, &api.ProposerDutiesOpts{
		Epoch: c.results.Epoch,
	})
	if err != nil {
		return errors.Wrap(err, "failed to obtain proposer duties")
	}
	c.results.Duties = dutiesResponse.Data

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
	c.proposerDutiesProvider, isProvider = c.eth2Client.(eth2client.ProposerDutiesProvider)
	if !isProvider {
		return errors.New("connection does not provide proposer duties")
	}

	return nil
}
