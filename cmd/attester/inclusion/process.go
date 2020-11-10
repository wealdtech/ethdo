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

package attesterinclusion

import (
	"context"
	"fmt"

	eth2client "github.com/attestantio/go-eth2-client"
	api "github.com/attestantio/go-eth2-client/api/v1"
	spec "github.com/attestantio/go-eth2-client/spec/phase0"
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

	duty, err := duty(ctx, data.eth2Client, data.validator, data.epoch, data.slotsPerEpoch)
	if err != nil {
		return nil, errors.Wrap(err, "failed to obtain duty for validator")
	}

	startSlot := duty.Slot + 1
	endSlot := startSlot + 32
	for slot := startSlot; slot < endSlot; slot++ {
		signedBlock, err := data.eth2Client.(eth2client.SignedBeaconBlockProvider).SignedBeaconBlock(ctx, fmt.Sprintf("%d", slot))
		if err != nil {
			return nil, errors.Wrap(err, "failed to obtain block")
		}
		if signedBlock == nil {
			continue
		}
		if signedBlock.Message.Slot != slot {
			continue
		}
		if data.debug {
			fmt.Printf("Fetched block for slot %d\n", slot)
		}
		for i, attestation := range signedBlock.Message.Body.Attestations {
			if attestation.Data.Slot == duty.Slot &&
				attestation.Data.Index == duty.CommitteeIndex &&
				attestation.AggregationBits.BitAt(duty.ValidatorCommitteeIndex) {
				results.slot = slot
				results.attestationIndex = uint64(i)
				results.inclusionDelay = slot - duty.Slot
				results.found = true
				return results, nil
			}
		}
	}
	return nil, errors.New("not found")
}

func duty(ctx context.Context, eth2Client eth2client.Service, validator *api.Validator, epoch spec.Epoch, slotsPerEpoch uint64) (*api.AttesterDuty, error) {
	// Find the attesting slot for the given epoch.
	duties, err := eth2Client.(eth2client.AttesterDutiesProvider).AttesterDuties(ctx, epoch, []spec.ValidatorIndex{validator.Index})
	if err != nil {
		return nil, errors.Wrap(err, "failed to obtain attester duties")
	}

	if len(duties) == 0 {
		return nil, errors.New("validator does not have duty for that epoch")
	}

	return duties[0], nil
}
