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

package slottime

import (
	"context"
	"strconv"
	"time"

	eth2client "github.com/attestantio/go-eth2-client"
	"github.com/attestantio/go-eth2-client/api"
	"github.com/pkg/errors"
)

func process(ctx context.Context, data *dataIn) (*dataOut, error) {
	if data == nil {
		return nil, errors.New("no data")
	}

	results := &dataOut{
		debug:   data.debug,
		quiet:   data.quiet,
		verbose: data.verbose,
	}

	slot, err := strconv.ParseInt(data.slot, 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, "invalid slot specified")
	}
	if slot < 0 {
		return nil, errors.New("slot must be a positive integer")
	}

	genesisResponse, err := data.eth2Client.(eth2client.GenesisProvider).Genesis(ctx, &api.GenesisOpts{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to obtain genesis information")
	}
	genesis := genesisResponse.Data

	specResponse, err := data.eth2Client.(eth2client.SpecProvider).Spec(ctx, &api.SpecOpts{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to obtain chain specifications")
	}

	slotDuration := specResponse.Data["SECONDS_PER_SLOT"].(time.Duration)

	results.startTime = genesis.GenesisTime.Add((time.Duration(slot*int64(slotDuration.Seconds())) * time.Second))
	results.endTime = results.startTime.Add(slotDuration)

	return results, nil
}
