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

package inclusion

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	eth2client "github.com/attestantio/go-eth2-client"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/altair"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
	standardchaintime "github.com/wealdtech/ethdo/services/chaintime/standard"
	"github.com/wealdtech/ethdo/util"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

func (c *command) process(ctx context.Context) error {
	// Obtain information we need to process.
	if err := c.setup(ctx); err != nil {
		return err
	}

	firstSlot, lastSlot := c.calculateSlots(ctx)

	validatorIndex, err := c.validatorIndex(ctx)
	if err != nil {
		return err
	}

	syncCommittee, err := c.eth2Client.(eth2client.SyncCommitteesProvider).SyncCommitteeAtEpoch(ctx, "head", phase0.Epoch(c.epoch))
	if err != nil {
		return errors.Wrap(err, "failed to obtain sync committee information")
	}

	if syncCommittee == nil {
		return errors.New("no sync committee returned")
	}

	for i := range syncCommittee.Validators {
		if syncCommittee.Validators[i] == validatorIndex {
			c.inCommittee = true
			c.committeeIndex = uint64(i)
			break
		}
	}

	if c.inCommittee {
		// This validator is in the sync committee.  Check blocks to see where it has been included.
		c.inclusions = make([]int, 0)
		if lastSlot > c.chainTime.CurrentSlot() {
			lastSlot = c.chainTime.CurrentSlot()
		}
		for slot := firstSlot; slot < lastSlot; slot++ {
			block, err := c.eth2Client.(eth2client.SignedBeaconBlockProvider).SignedBeaconBlock(ctx, fmt.Sprintf("%d", slot))
			if err != nil {
				return err
			}
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
	c.eth2Client, err = util.ConnectToBeaconNode(ctx, c.connection, c.timeout, c.allowInsecureConnections)
	if err != nil {
		return err
	}

	c.chainTime, err = standardchaintime.New(ctx,
		standardchaintime.WithSpecProvider(c.eth2Client.(eth2client.SpecProvider)),
		standardchaintime.WithForkScheduleProvider(c.eth2Client.(eth2client.ForkScheduleProvider)),
		standardchaintime.WithGenesisTimeProvider(c.eth2Client.(eth2client.GenesisTimeProvider)),
	)
	if err != nil {
		return errors.Wrap(err, "failed to set up chaintime service")
	}

	return nil
}

// validatorIndex obtains the index of a validator.
func (c *command) validatorIndex(ctx context.Context) (phase0.ValidatorIndex, error) {
	switch {
	case c.account != "":
		ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
		defer cancel()
		_, account, err := util.WalletAndAccountFromPath(ctx, c.account)
		if err != nil {
			return 0, errors.Wrap(err, "failed to obtain account")
		}
		return accountToIndex(ctx, account, c.eth2Client)
	case c.pubKey != "":
		pubKeyBytes, err := hex.DecodeString(strings.TrimPrefix(c.pubKey, "0x"))
		if err != nil {
			return 0, errors.Wrap(err, fmt.Sprintf("failed to decode public key %s", c.pubKey))
		}
		account, err := util.NewScratchAccount(nil, pubKeyBytes)
		if err != nil {
			return 0, errors.Wrap(err, fmt.Sprintf("invalid public key %s", c.pubKey))
		}
		return accountToIndex(ctx, account, c.eth2Client)
	case c.index != "":
		val, err := strconv.ParseUint(c.index, 10, 64)
		if err != nil {
			return 0, err
		}
		return phase0.ValidatorIndex(val), nil
	default:
		return 0, errors.New("no validator")
	}
}

func accountToIndex(ctx context.Context, account e2wtypes.Account, eth2Client eth2client.Service) (phase0.ValidatorIndex, error) {
	pubKey, err := util.BestPublicKey(account)
	if err != nil {
		return 0, err
	}

	pubKeys := make([]phase0.BLSPubKey, 1)
	copy(pubKeys[0][:], pubKey.Marshal())
	validators, err := eth2Client.(eth2client.ValidatorsProvider).ValidatorsByPubKey(ctx, "head", pubKeys)
	if err != nil {
		return 0, err
	}

	for index := range validators {
		return index, nil
	}
	return 0, errors.New("validator not found")
}

func (c *command) calculateSlots(ctx context.Context) (phase0.Slot, phase0.Slot) {
	var firstSlot phase0.Slot
	var lastSlot phase0.Slot
	if c.epoch == -1 {
		c.epoch = int64(c.chainTime.CurrentEpoch()) - 1
	}
	firstSlot = c.chainTime.FirstSlotOfEpoch(phase0.Epoch(c.epoch))
	lastSlot = c.chainTime.FirstSlotOfEpoch(phase0.Epoch(c.epoch) + 1)

	return firstSlot, lastSlot
}
