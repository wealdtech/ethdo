// Copyright © 2019, 2020 Weald Technology Trading
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

package blockinfo

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	eth2client "github.com/attestantio/go-eth2-client"
	api "github.com/attestantio/go-eth2-client/api/v1"
	spec "github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
)

var jsonOutput bool
var results *dataOut

func process(ctx context.Context, data *dataIn) (*dataOut, error) {
	if data == nil {
		return nil, errors.New("no data")
	}

	results = &dataOut{
		debug:      data.debug,
		verbose:    data.verbose,
		eth2Client: data.eth2Client,
	}

	config, err := results.eth2Client.(eth2client.SpecProvider).Spec(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to obtain configuration information")
	}
	genesis, err := results.eth2Client.(eth2client.GenesisProvider).Genesis(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to obtain genesis information")
	}
	results.genesisTime = genesis.GenesisTime
	results.slotDuration = config["SECONDS_PER_SLOT"].(time.Duration)
	results.slotsPerEpoch = config["SLOTS_PER_EPOCH"].(uint64)

	signedBlock, err := results.eth2Client.(eth2client.SignedBeaconBlockProvider).SignedBeaconBlock(ctx, data.blockID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to obtain beacon block")
	}

	if err := outputBlock(ctx, data.jsonOutput, signedBlock); err != nil {
		return nil, errors.Wrap(err, "failed to output block")
	}

	if data.stream {
		jsonOutput = data.jsonOutput
		err := data.eth2Client.(eth2client.EventsProvider).Events(ctx, []string{"head"}, headEventHandler)
		if err != nil {
			return nil, errors.Wrap(err, "failed to start block stream")
		}
		<-ctx.Done()
	}

	return &dataOut{}, nil
}

func headEventHandler(event *api.Event) {
	// Only interested in head events.
	if event.Topic != "head" {
		return
	}

	blockID := fmt.Sprintf("%#x", event.Data.(*api.HeadEvent).Block[:])
	signedBlock, err := results.eth2Client.(eth2client.SignedBeaconBlockProvider).SignedBeaconBlock(context.Background(), blockID)
	if err != nil {
		fmt.Printf("Failed to obtain block: %v\n", err)
	}
	if err := outputBlock(context.Background(), jsonOutput, signedBlock); err != nil {
		fmt.Printf("Failed to display block: %v\n", err)
	}
}

func outputBlock(ctx context.Context, jsonOutput bool, signedBlock *spec.SignedBeaconBlock) error {
	switch {
	case jsonOutput:
		data, err := json.Marshal(signedBlock)
		if err != nil {
			return errors.Wrap(err, "failed to generate JSON")
		}
		fmt.Printf("%s\n", string(data))
	default:
		data, err := outputBlockText(ctx, results, signedBlock)
		if err != nil {
			return errors.Wrap(err, "failed to generate text")
		}
		fmt.Printf("%s\n", data)
	}
	return nil
}
