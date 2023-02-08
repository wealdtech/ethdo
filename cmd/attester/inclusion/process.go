// Copyright © 2019 - 2022 Weald Technology Trading
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
	"bytes"
	"context"
	"fmt"

	eth2client "github.com/attestantio/go-eth2-client"
	api "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
	standardchaintime "github.com/wealdtech/ethdo/services/chaintime/standard"
	"github.com/wealdtech/ethdo/util"
)

func process(ctx context.Context, data *dataIn) (*dataOut, error) {
	if data == nil {
		return nil, errors.New("no data")
	}

	var err error

	data.chainTime, err = standardchaintime.New(ctx,
		standardchaintime.WithSpecProvider(data.eth2Client.(eth2client.SpecProvider)),
		standardchaintime.WithGenesisTimeProvider(data.eth2Client.(eth2client.GenesisTimeProvider)),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to set up chaintime service")
	}

	validatorIndex, err := util.ValidatorIndex(ctx, data.eth2Client, data.account, data.pubKey, data.index)
	if err != nil {
		return nil, errors.Wrap(err, "failed to obtain validator index")
	}

	validators, err := data.eth2Client.(eth2client.ValidatorsProvider).Validators(ctx, fmt.Sprintf("%d", uint64(data.epoch)*data.slotsPerEpoch), []phase0.ValidatorIndex{validatorIndex})
	if err != nil {
		return nil, errors.New("failed to obtain validator information")
	}
	if len(validators) == 0 {
		return nil, errors.New("validator is not known")
	}
	var validator *api.Validator
	for _, v := range validators {
		validator = v
	}

	results := &dataOut{
		debug:   data.debug,
		quiet:   data.quiet,
		verbose: data.verbose,
	}

	duty, err := duty(ctx, data.eth2Client, validator, data.epoch, data.slotsPerEpoch)
	if err != nil {
		return nil, errors.Wrap(err, "failed to obtain duty for validator")
	}
	if data.debug {
		fmt.Printf("Duty is %s\n", duty.String())
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
		blockSlot, err := signedBlock.Slot()
		if err != nil {
			return nil, errors.Wrap(err, "failed to obtain block slot")
		}
		if blockSlot != slot {
			continue
		}
		if data.debug {
			fmt.Printf("Fetched block for slot %d\n", slot)
		}
		attestations, err := signedBlock.Attestations()
		if err != nil {
			return nil, errors.Wrap(err, "failed to obtain block attestations")
		}
		for i, attestation := range attestations {
			if attestation.Data.Slot == duty.Slot &&
				attestation.Data.Index == duty.CommitteeIndex &&
				attestation.AggregationBits.BitAt(duty.ValidatorCommitteeIndex) {
				headCorrect := false
				targetCorrect := false
				if data.verbose {
					headCorrect, err = calcHeadCorrect(ctx, data, attestation)
					if err != nil {
						return nil, errors.Wrap(err, "failed to obtain head correct result")
					}
					targetCorrect, err = calcTargetCorrect(ctx, data, attestation)
					if err != nil {
						return nil, errors.Wrap(err, "failed to obtain target correct result")
					}
				}
				results.found = true
				results.attestation = attestation
				results.slot = slot
				results.attestationIndex = uint64(i)
				results.inclusionDelay = slot - duty.Slot
				results.sourceTimely = results.inclusionDelay <= 5 // sqrt(32)
				results.targetCorrect = targetCorrect
				results.targetTimely = targetCorrect && results.inclusionDelay <= 32
				results.headCorrect = headCorrect
				results.headTimely = headCorrect && results.inclusionDelay == 1
				if data.debug {
					fmt.Printf("Attestation is %s\n", attestation.String())
				}
				return results, nil
			}
		}
	}
	return results, nil
}

func calcHeadCorrect(ctx context.Context, data *dataIn, attestation *phase0.Attestation) (bool, error) {
	slot := attestation.Data.Slot
	for {
		header, err := data.eth2Client.(eth2client.BeaconBlockHeadersProvider).BeaconBlockHeader(ctx, fmt.Sprintf("%d", slot))
		if err != nil {
			return false, err
		}
		if header == nil {
			// No block.
			slot--
			continue
		}
		if !header.Canonical {
			// Not canonical.
			slot--
			continue
		}
		return bytes.Equal(header.Root[:], attestation.Data.BeaconBlockRoot[:]), nil
	}
}

func calcTargetCorrect(ctx context.Context, data *dataIn, attestation *phase0.Attestation) (bool, error) {
	// Start with first slot of the target epoch.
	slot := data.chainTime.FirstSlotOfEpoch(attestation.Data.Target.Epoch)
	for {
		header, err := data.eth2Client.(eth2client.BeaconBlockHeadersProvider).BeaconBlockHeader(ctx, fmt.Sprintf("%d", slot))
		if err != nil {
			return false, err
		}
		if header == nil {
			// No block.
			slot--
			continue
		}
		if !header.Canonical {
			// Not canonical.
			slot--
			continue
		}
		return bytes.Equal(header.Root[:], attestation.Data.Target.Root[:]), nil
	}
}

func duty(ctx context.Context, eth2Client eth2client.Service, validator *api.Validator, epoch phase0.Epoch, slotsPerEpoch uint64) (*api.AttesterDuty, error) {
	// Find the attesting slot for the given epoch.
	duties, err := eth2Client.(eth2client.AttesterDutiesProvider).AttesterDuties(ctx, epoch, []phase0.ValidatorIndex{validator.Index})
	if err != nil {
		return nil, errors.Wrap(err, "failed to obtain attester duties")
	}

	if len(duties) == 0 {
		return nil, errors.New("validator does not have duty for that epoch")
	}

	return duties[0], nil
}
