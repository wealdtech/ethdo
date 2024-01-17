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

package inclusion

import (
	"context"
	"fmt"
	"net/http"

	eth2client "github.com/attestantio/go-eth2-client"
	"github.com/attestantio/go-eth2-client/api"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/altair"
	"github.com/pkg/errors"
	standardchaintime "github.com/wealdtech/ethdo/services/chaintime/standard"
	"github.com/wealdtech/ethdo/util"
)

func (c *command) process(ctx context.Context) error {
	// Obtain information we need to process.
	if err := c.setup(ctx); err != nil {
		return err
	}

	validator, err := util.ParseValidator(ctx, c.eth2Client.(eth2client.ValidatorsProvider), c.validator, "head")
	if err != nil {
		return err
	}

	c.epoch, err = util.ParseEpoch(ctx, c.chainTime, c.epochStr)
	if err != nil {
		return err
	}

	syncCommitteeResponse, err := c.eth2Client.(eth2client.SyncCommitteesProvider).SyncCommittee(ctx, &api.SyncCommitteeOpts{
		State: "head",
		Epoch: &c.epoch,
	})
	if err != nil {
		return errors.Wrap(err, "failed to obtain sync committee information")
	}
	syncCommittee := syncCommitteeResponse.Data

	if syncCommittee == nil {
		return errors.New("no sync committee returned")
	}

	for i := range syncCommittee.Validators {
		if syncCommittee.Validators[i] == validator.Index {
			c.inCommittee = true
			c.committeeIndex = uint64(i)
			break
		}
	}

	if c.inCommittee {
		firstSlot := c.chainTime.FirstSlotOfEpoch(c.epoch)
		lastSlot := c.chainTime.LastSlotOfEpoch(c.epoch)
		// This validator is in the sync committee.  Check blocks to see where it has been included.
		c.inclusions = make([]int, 0)
		if lastSlot > c.chainTime.CurrentSlot() {
			lastSlot = c.chainTime.CurrentSlot()
		}
		for slot := firstSlot; slot <= lastSlot; slot++ {
			blockResponse, err := c.eth2Client.(eth2client.SignedBeaconBlockProvider).SignedBeaconBlock(ctx, &api.SignedBeaconBlockOpts{
				Block: fmt.Sprintf("%d", slot),
			})
			if err != nil {
				var apiErr *api.Error
				if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
					c.inclusions = append(c.inclusions, 0)
					continue
				}
				return errors.Wrap(err, "failed to obtain beacon block")
			}

			block := blockResponse.Data
			if block == nil {
				c.inclusions = append(c.inclusions, 0)
				continue
			}
			var aggregate *altair.SyncAggregate
			switch block.Version {
			case spec.DataVersionAltair:
				aggregate = block.Altair.Message.Body.SyncAggregate
				if aggregate.SyncCommitteeBits.BitAt(c.committeeIndex) {
					c.inclusions = append(c.inclusions, 1)
				} else {
					c.inclusions = append(c.inclusions, 2)
				}
			case spec.DataVersionBellatrix:
				aggregate = block.Bellatrix.Message.Body.SyncAggregate
				if aggregate.SyncCommitteeBits.BitAt(c.committeeIndex) {
					c.inclusions = append(c.inclusions, 1)
				} else {
					c.inclusions = append(c.inclusions, 2)
				}
			case spec.DataVersionCapella:
				aggregate = block.Capella.Message.Body.SyncAggregate
				if aggregate.SyncCommitteeBits.BitAt(c.committeeIndex) {
					c.inclusions = append(c.inclusions, 1)
				} else {
					c.inclusions = append(c.inclusions, 2)
				}
			case spec.DataVersionDeneb:
				aggregate = block.Deneb.Message.Body.SyncAggregate
				if aggregate.SyncCommitteeBits.BitAt(c.committeeIndex) {
					c.inclusions = append(c.inclusions, 1)
				} else {
					c.inclusions = append(c.inclusions, 2)
				}
			default:
				return fmt.Errorf("unhandled block version %v", block.Version)
			}
		}
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
		return err
	}

	c.chainTime, err = standardchaintime.New(ctx,
		standardchaintime.WithSpecProvider(c.eth2Client.(eth2client.SpecProvider)),
		standardchaintime.WithGenesisProvider(c.eth2Client.(eth2client.GenesisProvider)),
	)
	if err != nil {
		return errors.Wrap(err, "failed to set up chaintime service")
	}

	return nil
}
