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

package attesterduties

import (
	"context"

	eth2client "github.com/attestantio/go-eth2-client"
	"github.com/attestantio/go-eth2-client/api"
	apiv1 "github.com/attestantio/go-eth2-client/api/v1"
	spec "github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
	"github.com/wealdtech/ethdo/util"
)

func process(ctx context.Context, data *dataIn) (*dataOut, error) {
	if data == nil {
		return nil, errors.New("no data")
	}

	validator, err := util.ParseValidator(ctx, data.eth2Client.(eth2client.ValidatorsProvider), data.validator, "head")
	if err != nil {
		return nil, err
	}

	results := &dataOut{
		debug:   data.debug,
		quiet:   data.quiet,
		verbose: data.verbose,
	}

	duty, err := duty(ctx, data.eth2Client, validator, data.epoch)
	if err != nil {
		return nil, errors.Wrap(err, "failed to obtain duty for validator")
	}

	results.duty = duty

	return results, nil
}

func duty(ctx context.Context, eth2Client eth2client.Service, validator *apiv1.Validator, epoch spec.Epoch) (*apiv1.AttesterDuty, error) {
	// Find the attesting slot for the given epoch.
	dutiesResponse, err := eth2Client.(eth2client.AttesterDutiesProvider).AttesterDuties(ctx, &api.AttesterDutiesOpts{
		Epoch:   epoch,
		Indices: []spec.ValidatorIndex{validator.Index},
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to obtain attester duties")
	}
	duties := dutiesResponse.Data

	if len(duties) == 0 {
		return nil, errors.New("validator does not have duty for that epoch")
	}

	return duties[0], nil
}
