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
	spec "github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
	"github.com/wealdtech/ethdo/util"
)

func process(ctx context.Context, data *dataIn) (*dataOut, error) {
	if data == nil {
		return nil, errors.New("no data")
	}

	eth2Client, err := util.ConnectToBeaconNode(ctx, data.connection, data.timeout, data.allowInsecureConnections)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to Ethereum 2 beacon node")
	}

	config, err := eth2Client.(eth2client.SpecProvider).Spec(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to obtain beacon chain configuration")
	}

	slotsPerEpoch := config["SLOTS_PER_EPOCH"].(uint64)
	slotDuration := config["SECONDS_PER_SLOT"].(time.Duration)
	genesis, err := eth2Client.(eth2client.GenesisProvider).Genesis(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to obtain genesis data")
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
		results.slot = spec.Slot(slot)
	case data.epoch != "":
		epoch, err := strconv.ParseUint(data.epoch, 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse epoch")
		}
		results.slot = spec.Slot(epoch * slotsPerEpoch)
	case data.timestamp != "":
		timestamp, err := time.Parse("2006-01-02T15:04:05-0700", data.timestamp)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse timestamp")
		}
		secs := timestamp.Sub(genesis.GenesisTime)
		if secs < 0 {
			return nil, errors.New("timestamp prior to genesis")
		}
		results.slot = spec.Slot(secs / slotDuration)
	}

	// Fill in the info given the slot.
	results.slotStart = genesis.GenesisTime.Add(time.Duration(results.slot) * slotDuration)
	results.slotEnd = genesis.GenesisTime.Add(time.Duration(results.slot+1) * slotDuration)
	results.epoch = spec.Epoch(uint64(results.slot) / slotsPerEpoch)
	results.epochStart = genesis.GenesisTime.Add(time.Duration(uint64(results.epoch)*slotsPerEpoch) * slotDuration)
	results.epochEnd = genesis.GenesisTime.Add(time.Duration(uint64(results.epoch+1)*slotsPerEpoch) * slotDuration)

	return results, nil
}
