// Copyright © 2021 Weald Technology Trading
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

package chaintime

import (
	"context"
	"strconv"
	"time"

	eth2client "github.com/attestantio/go-eth2-client"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
	standardchaintime "github.com/wealdtech/ethdo/services/chaintime/standard"
	"github.com/wealdtech/ethdo/util"
)

func process(ctx context.Context, data *dataIn) (*dataOut, error) {
	if data == nil {
		return nil, errors.New("no data")
	}

	eth2Client, err := util.ConnectToBeaconNode(ctx, &util.ConnectOpts{
		Address:       data.connection,
		Timeout:       data.timeout,
		AllowInsecure: data.allowInsecureConnections,
		LogFallback:   !data.quiet,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to Ethereum 2 beacon node")
	}

	chainTime, err := standardchaintime.New(ctx,
		standardchaintime.WithSpecProvider(eth2Client.(eth2client.SpecProvider)),
		standardchaintime.WithGenesisProvider(eth2Client.(eth2client.GenesisProvider)),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to set up chaintime service")
	}

	results := &dataOut{
		debug:   data.debug,
		quiet:   data.quiet,
		verbose: data.verbose,
	}

	// Calculate the slot given the input.
	switch {
	case data.slot != "":
		slot, err := strconv.ParseUint(data.slot, 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse slot")
		}
		results.slot = phase0.Slot(slot)
	case data.epoch != "":
		epoch, err := util.ParseEpoch(ctx, chainTime, data.epoch)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse epoch")
		}
		results.slot = chainTime.FirstSlotOfEpoch(epoch)
	case data.timestamp != "":
		timestamp, err := time.Parse("2006-01-02T15:04:05-0700", data.timestamp)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse timestamp")
		}
		results.slot = chainTime.TimestampToSlot(timestamp)
	}

	// Fill in the info given the slot.
	results.slotStart = chainTime.StartOfSlot(results.slot)
	results.slotEnd = chainTime.StartOfSlot(results.slot + 1)
	results.epoch = chainTime.SlotToEpoch(results.slot)
	results.epochStart = chainTime.StartOfEpoch(results.epoch)
	results.epochEnd = chainTime.StartOfEpoch(results.epoch + 1)
	if results.epoch >= chainTime.FirstEpochOfSyncPeriod(chainTime.AltairInitialSyncCommitteePeriod()) {
		results.hasSyncCommittees = true
		results.syncCommitteePeriod = chainTime.SlotToSyncCommitteePeriod(results.slot)
		results.syncCommitteePeriodEpochStart = chainTime.FirstEpochOfSyncPeriod(results.syncCommitteePeriod)
		results.syncCommitteePeriodEpochEnd = chainTime.FirstEpochOfSyncPeriod(results.syncCommitteePeriod + 1)
		results.syncCommitteePeriodStart = chainTime.StartOfEpoch(results.syncCommitteePeriodEpochStart)
		results.syncCommitteePeriodEnd = chainTime.StartOfEpoch(results.syncCommitteePeriodEpochEnd)
	}

	return results, nil
}
