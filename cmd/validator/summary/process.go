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

package validatorsummary

import (
	"context"
	"fmt"
	"net/http"
	"sort"

	eth2client "github.com/attestantio/go-eth2-client"
	"github.com/attestantio/go-eth2-client/api"
	apiv1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
	standardchaintime "github.com/wealdtech/ethdo/services/chaintime/standard"
	"github.com/wealdtech/ethdo/util"
)

func (c *command) process(ctx context.Context) error {
	if len(c.validators) == 0 {
		return errors.New("no validators supplied")
	}

	// Obtain information we need to process.
	err := c.setup(ctx)
	if err != nil {
		return err
	}

	c.summary.Epoch, err = util.ParseEpoch(ctx, c.chainTime, c.epoch)
	if err != nil {
		return errors.Wrap(err, "failed to parse epoch")
	}
	c.summary.FirstSlot = c.chainTime.FirstSlotOfEpoch(c.summary.Epoch)
	c.summary.LastSlot = c.chainTime.FirstSlotOfEpoch(c.summary.Epoch+1) - 1
	c.summary.Slots = make([]*slot, 1+int(c.summary.LastSlot)-int(c.summary.FirstSlot))
	for i := range c.summary.Slots {
		c.summary.Slots[i] = &slot{
			Slot: c.summary.FirstSlot + phase0.Slot(i),
		}
	}

	c.summary.Validators, err = util.ParseValidators(ctx, c.validatorsProvider, c.validators, fmt.Sprintf("%d", c.summary.FirstSlot))
	if err != nil {
		return errors.Wrap(err, "failed to parse validators")
	}
	// Reorder validators by index.
	sort.Slice(c.summary.Validators, func(i int, j int) bool {
		return c.summary.Validators[i].Index < c.summary.Validators[j].Index
	})

	// Create a map for validator indices for easy lookup.
	c.validatorsByIndex = make(map[phase0.ValidatorIndex]*apiv1.Validator)
	for _, validator := range c.summary.Validators {
		c.validatorsByIndex[validator.Index] = validator
	}

	if err := c.processProposerDuties(ctx); err != nil {
		return err
	}

	if err := c.processAttesterDuties(ctx); err != nil {
		return err
	}

	// if err := c.processSyncCommitteeDuties(ctx); err != nil {
	// 	return err
	// }

	return nil
}

func (c *command) processProposerDuties(ctx context.Context) error {
	response, err := c.proposerDutiesProvider.ProposerDuties(ctx, &api.ProposerDutiesOpts{
		Epoch: c.summary.Epoch,
	})
	if err != nil {
		return errors.Wrap(err, "failed to obtain proposer duties")
	}
	for _, duty := range response.Data {
		if _, exists := c.validatorsByIndex[duty.ValidatorIndex]; !exists {
			continue
		}
		blockResponse, err := c.blocksProvider.SignedBeaconBlock(ctx, &api.SignedBeaconBlockOpts{
			Block: fmt.Sprintf("%d", duty.Slot),
		})
		if err != nil {
			var apiErr *api.Error
			if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
				return nil
			}

			return errors.Wrap(err, fmt.Sprintf("failed to obtain block for slot %d", duty.Slot))
		}
		block := blockResponse.Data
		present := block != nil
		c.summary.Proposals = append(c.summary.Proposals, &epochProposal{
			Slot:     duty.Slot,
			Proposer: duty.ValidatorIndex,
			Block:    present,
		})
	}

	return nil
}

func (c *command) activeValidators() (map[phase0.ValidatorIndex]*apiv1.Validator, []phase0.ValidatorIndex) {
	activeValidators := make(map[phase0.ValidatorIndex]*apiv1.Validator)
	activeValidatorIndices := make([]phase0.ValidatorIndex, 0, len(c.validatorsByIndex))
	for _, validator := range c.summary.Validators {
		if validator.Validator.ActivationEpoch <= c.summary.Epoch && validator.Validator.ExitEpoch > c.summary.Epoch {
			activeValidators[validator.Index] = validator
			activeValidatorIndices = append(activeValidatorIndices, validator.Index)
		}
	}

	return activeValidators, activeValidatorIndices
}

func (c *command) processAttesterDuties(ctx context.Context) error {
	activeValidators, activeValidatorIndices := c.activeValidators()

	// Obtain number of validators that voted for blocks in the epoch.
	// These votes can be included anywhere from the second slot of
	// the epoch to the first slot of the next-but-one epoch.
	firstSlot := c.chainTime.FirstSlotOfEpoch(c.summary.Epoch) + 1
	lastSlot := c.chainTime.FirstSlotOfEpoch(c.summary.Epoch + 2)
	if lastSlot > c.chainTime.CurrentSlot() {
		lastSlot = c.chainTime.CurrentSlot()
	}

	// Obtain the duties for the validators to know where they should be attesting.
	dutiesResponse, err := c.attesterDutiesProvider.AttesterDuties(ctx, &api.AttesterDutiesOpts{
		Epoch:   c.summary.Epoch,
		Indices: activeValidatorIndices,
	})
	if err != nil {
		return errors.Wrap(err, "failed to obtain attester duties")
	}
	duties := dutiesResponse.Data
	for slot := c.chainTime.FirstSlotOfEpoch(c.summary.Epoch); slot < c.chainTime.FirstSlotOfEpoch(c.summary.Epoch+1); slot++ {
		index := int(slot - c.chainTime.FirstSlotOfEpoch(c.summary.Epoch))
		c.summary.Slots[index].Attestations = &slotAttestations{}
	}

	// Need a cache of beacon block headers to reduce lookup times.
	headersCache := util.NewBeaconBlockHeaderCache(c.beaconBlockHeadersProvider)

	// Need a map of duties to easily find the attestations we care about.
	dutiesBySlot := make(map[phase0.Slot]map[phase0.CommitteeIndex][]*apiv1.AttesterDuty)
	dutiesByValidatorIndex := make(map[phase0.ValidatorIndex]*apiv1.AttesterDuty)
	for _, duty := range duties {
		index := int(duty.Slot - c.chainTime.FirstSlotOfEpoch(c.summary.Epoch))
		dutiesByValidatorIndex[duty.ValidatorIndex] = duty
		c.summary.Slots[index].Attestations.Expected++
		if _, exists := dutiesBySlot[duty.Slot]; !exists {
			dutiesBySlot[duty.Slot] = make(map[phase0.CommitteeIndex][]*apiv1.AttesterDuty)
		}
		if _, exists := dutiesBySlot[duty.Slot][duty.CommitteeIndex]; !exists {
			dutiesBySlot[duty.Slot][duty.CommitteeIndex] = make([]*apiv1.AttesterDuty, 0)
		}
		dutiesBySlot[duty.Slot][duty.CommitteeIndex] = append(dutiesBySlot[duty.Slot][duty.CommitteeIndex], duty)
	}

	c.summary.IncorrectHeadValidators = make([]*validatorFault, 0)
	c.summary.UntimelyHeadValidators = make([]*validatorFault, 0)
	c.summary.UntimelySourceValidators = make([]*validatorFault, 0)
	c.summary.IncorrectTargetValidators = make([]*validatorFault, 0)
	c.summary.UntimelyTargetValidators = make([]*validatorFault, 0)

	// Hunt through the blocks looking for attestations from the validators.
	votes := make(map[phase0.ValidatorIndex]struct{})
	for slot := firstSlot; slot <= lastSlot; slot++ {
		if err := c.processAttesterDutiesSlot(ctx, slot, dutiesBySlot, votes, headersCache, activeValidatorIndices); err != nil {
			return err
		}
	}

	// Use dutiesMap and votes to work out which validators didn't participate.
	c.summary.NonParticipatingValidators = make([]*nonParticipatingValidator, 0)
	for _, index := range activeValidatorIndices {
		if _, exists := votes[index]; !exists {
			// Didn't vote.
			duty := dutiesByValidatorIndex[index]
			c.summary.NonParticipatingValidators = append(c.summary.NonParticipatingValidators, &nonParticipatingValidator{
				Validator: index,
				Slot:      duty.Slot,
				Committee: duty.CommitteeIndex,
			})
		}
	}

	// Sort the non-participating validators list.
	sort.Slice(c.summary.NonParticipatingValidators, func(i int, j int) bool {
		if c.summary.NonParticipatingValidators[i].Slot != c.summary.NonParticipatingValidators[j].Slot {
			return c.summary.NonParticipatingValidators[i].Slot < c.summary.NonParticipatingValidators[j].Slot
		}
		if c.summary.NonParticipatingValidators[i].Committee != c.summary.NonParticipatingValidators[j].Committee {
			return c.summary.NonParticipatingValidators[i].Committee < c.summary.NonParticipatingValidators[j].Committee
		}
		return c.summary.NonParticipatingValidators[i].Validator < c.summary.NonParticipatingValidators[j].Validator
	})

	c.summary.ActiveValidators = len(activeValidators)
	c.summary.ParticipatingValidators = len(votes)

	return nil
}

func (c *command) processAttesterDutiesSlot(ctx context.Context,
	slot phase0.Slot,
	dutiesBySlot map[phase0.Slot]map[phase0.CommitteeIndex][]*apiv1.AttesterDuty,
	votes map[phase0.ValidatorIndex]struct{},
	headersCache *util.BeaconBlockHeaderCache,
	activeValidatorIndices []phase0.ValidatorIndex,
) error {
	blockResponse, err := c.blocksProvider.SignedBeaconBlock(ctx, &api.SignedBeaconBlockOpts{
		Block: fmt.Sprintf("%d", slot),
	})
	if err != nil {
		var apiErr *api.Error
		if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
			return nil
		}

		return errors.Wrap(err, "failed to obtain beacon block")
	}
	block := blockResponse.Data
	attestations, err := block.Attestations()
	if err != nil {
		return err
	}
	for _, attestation := range attestations {
		if _, exists := dutiesBySlot[attestation.Data.Slot]; !exists {
			// We do not have any attestations for this slot.
			continue
		}
		if _, exists := dutiesBySlot[attestation.Data.Slot][attestation.Data.Index]; !exists {
			// We do not have any attestations for this committee.
			continue
		}
		for _, duty := range dutiesBySlot[attestation.Data.Slot][attestation.Data.Index] {
			if attestation.AggregationBits.BitAt(duty.ValidatorCommitteeIndex) {
				// Found it.
				if _, exists := votes[duty.ValidatorIndex]; exists {
					// Duplicate; ignore.
					continue
				}
				votes[duty.ValidatorIndex] = struct{}{}

				// Update the metrics for the attestation.
				index := int(attestation.Data.Slot - c.chainTime.FirstSlotOfEpoch(c.summary.Epoch))
				c.summary.Slots[index].Attestations.Included++
				inclusionDelay := slot - duty.Slot

				fault := &validatorFault{
					Validator:         duty.ValidatorIndex,
					AttestationData:   attestation.Data,
					InclusionDistance: int(inclusionDelay),
				}

				headCorrect, err := util.AttestationHeadCorrect(ctx, headersCache, attestation)
				if err != nil {
					return errors.Wrap(err, "failed to calculate if attestation had correct head vote")
				}
				if headCorrect {
					c.summary.Slots[index].Attestations.CorrectHead++
					if inclusionDelay == 1 {
						c.summary.Slots[index].Attestations.TimelyHead++
					} else {
						c.summary.UntimelyHeadValidators = append(c.summary.UntimelyHeadValidators, fault)
					}
				} else {
					c.summary.IncorrectHeadValidators = append(c.summary.IncorrectHeadValidators, fault)
					if inclusionDelay > 1 {
						c.summary.UntimelyHeadValidators = append(c.summary.UntimelyHeadValidators, fault)
					}
				}

				if inclusionDelay <= 5 {
					c.summary.Slots[index].Attestations.TimelySource++
				} else {
					c.summary.UntimelySourceValidators = append(c.summary.UntimelySourceValidators, fault)
				}

				targetCorrect, err := util.AttestationTargetCorrect(ctx, headersCache, c.chainTime, attestation)
				if err != nil {
					return errors.Wrap(err, "failed to calculate if attestation had correct target vote")
				}
				if targetCorrect {
					c.summary.Slots[index].Attestations.CorrectTarget++
					if inclusionDelay <= 32 {
						c.summary.Slots[index].Attestations.TimelyTarget++
					} else {
						c.summary.UntimelyTargetValidators = append(c.summary.UntimelyTargetValidators, fault)
					}
				} else {
					c.summary.IncorrectTargetValidators = append(c.summary.IncorrectTargetValidators, fault)
					if inclusionDelay > 32 {
						c.summary.UntimelyTargetValidators = append(c.summary.UntimelyTargetValidators, fault)
					}
				}
			}
		}

		if len(votes) == len(activeValidatorIndices) {
			// Found them all.
			break
		}
	}

	return nil
}

// func (c *command) processSyncCommitteeDuties(ctx context.Context) error {
// 	if c.summary.Epoch < c.chainTime.AltairInitialEpoch() {
// 		// The epoch is pre-Altair.  No info but no error.
// 		return nil
// 	}
//
// 	committee, err := c.syncCommitteesProvider.SyncCommittee(ctx, fmt.Sprintf("%d", c.summary.FirstSlot))
// 	if err != nil {
// 		return errors.Wrap(err, "failed to obtain sync committee")
// 	}
// 	if len(committee.Validators) == 0 {
// 		return errors.Wrap(err, "empty sync committee")
// 	}
//
// 	missed := make(map[phase0.ValidatorIndex]int)
// 	for _, index := range committee.Validators {
// 		missed[index] = 0
// 	}
//
// 	for slot := c.summary.FirstSlot; slot <= c.summary.LastSlot; slot++ {
// 		block, err := c.blocksProvider.SignedBeaconBlock(ctx, fmt.Sprintf("%d", slot))
// 		if err != nil {
// 			return errors.Wrap(err, fmt.Sprintf("failed to obtain block for slot %d", slot))
// 		}
// 		if block == nil {
// 			// If the block is missed we don't count the sync aggregate miss.
// 			continue
// 		}
// 		var aggregate *altair.SyncAggregate
// 		switch block.Version {
// 		case spec.DataVersionPhase0:
// 			// No sync committees in this fork.
// 			return nil
// 		case spec.DataVersionAltair:
// 			aggregate = block.Altair.Message.Body.SyncAggregate
// 		case spec.DataVersionBellatrix:
// 			aggregate = block.Bellatrix.Message.Body.SyncAggregate
// 		default:
// 			return fmt.Errorf("unhandled block version %v", block.Version)
// 		}
// 		for i := uint64(0); i < aggregate.SyncCommitteeBits.Len(); i++ {
// 			if !aggregate.SyncCommitteeBits.BitAt(i) {
// 				missed[committee.Validators[int(i)]]++
// 			}
// 		}
// 	}
//
// 	c.summary.SyncCommittee = make([]*epochSyncCommittee, 0, len(missed))
// 	for index, count := range missed {
// 		if count > 0 {
// 			c.summary.SyncCommittee = append(c.summary.SyncCommittee, &epochSyncCommittee{
// 				Index:  index,
// 				Missed: count,
// 			})
// 		}
// 	}
//
// 	sort.Slice(c.summary.SyncCommittee, func(i int, j int) bool {
// 		missedDiff := c.summary.SyncCommittee[i].Missed - c.summary.SyncCommittee[j].Missed
// 		if missedDiff != 0 {
// 			// Actually want to order by missed descending, so invert the expected condition.
// 			return missedDiff > 0
// 		}
// 		// Then order by validator index.
// 		return c.summary.SyncCommittee[i].Index < c.summary.SyncCommittee[j].Index
// 	})
//
// 	return nil
// }

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
	c.attesterDutiesProvider, isProvider = c.eth2Client.(eth2client.AttesterDutiesProvider)
	if !isProvider {
		return errors.New("connection does not provide attester duties")
	}
	c.blocksProvider, isProvider = c.eth2Client.(eth2client.SignedBeaconBlockProvider)
	if !isProvider {
		return errors.New("connection does not provide signed beacon blocks")
	}
	c.syncCommitteesProvider, isProvider = c.eth2Client.(eth2client.SyncCommitteesProvider)
	if !isProvider {
		return errors.New("connection does not provide sync committee duties")
	}
	c.validatorsProvider, isProvider = c.eth2Client.(eth2client.ValidatorsProvider)
	if !isProvider {
		return errors.New("connection does not provide validators")
	}
	c.beaconCommitteesProvider, isProvider = c.eth2Client.(eth2client.BeaconCommitteesProvider)
	if !isProvider {
		return errors.New("connection does not provide beacon committees")
	}
	c.beaconBlockHeadersProvider, isProvider = c.eth2Client.(eth2client.BeaconBlockHeadersProvider)
	if !isProvider {
		return errors.New("connection does not provide beacon block headers")
	}

	return nil
}
