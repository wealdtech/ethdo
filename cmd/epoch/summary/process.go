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

package epochsummary

import (
	"context"
	"fmt"
	"net/http"
	"sort"

	eth2client "github.com/attestantio/go-eth2-client"
	"github.com/attestantio/go-eth2-client/api"
	apiv1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
	standardchaintime "github.com/wealdtech/ethdo/services/chaintime/standard"
	"github.com/wealdtech/ethdo/util"
)

func (c *command) process(ctx context.Context) error {
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

	validators, err := util.ParseValidators(ctx, c.validatorsProvider, c.validatorsStr, "head")
	if err != nil {
		return errors.Wrap(err, "failed to parse validators")
	}
	for _, validator := range validators {
		c.validators[validator.Index] = struct{}{}
	}

	if err := c.processProposerDuties(ctx); err != nil {
		return err
	}
	if err := c.processAttesterDuties(ctx); err != nil {
		return err
	}
	if err := c.processSyncCommitteeDuties(ctx); err != nil {
		return err
	}

	return c.processBlobs(ctx)
}

func (c *command) processProposerDuties(ctx context.Context) error {
	response, err := c.proposerDutiesProvider.ProposerDuties(ctx, &api.ProposerDutiesOpts{
		Epoch: c.summary.Epoch,
	})
	if err != nil {
		return errors.Wrap(err, "failed to obtain proposer duties")
	}

	for _, duty := range response.Data {
		block, err := c.fetchBlock(ctx, fmt.Sprintf("%d", duty.Slot))
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to obtain block for slot %d", duty.Slot))
		}
		present := block != nil
		if present {
			c.summary.Blocks++
		}

		_, exists := c.validators[duty.ValidatorIndex]
		if len(c.validators) > 0 && !exists {
			// Not one of ours.
			continue
		}

		c.summary.Proposals = append(c.summary.Proposals, &epochProposal{
			Slot:           duty.Slot,
			ValidatorIndex: duty.ValidatorIndex,
			Block:          present,
		})
	}

	return nil
}

func (c *command) activeValidators(ctx context.Context) (map[phase0.ValidatorIndex]*apiv1.Validator, error) {
	validatorIndices := make([]phase0.ValidatorIndex, 0, len(c.validators))
	for validator := range c.validators {
		validatorIndices = append(validatorIndices, validator)
	}

	response, err := c.validatorsProvider.Validators(ctx, &api.ValidatorsOpts{
		State:   "head",
		Indices: validatorIndices,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to obtain validators for epoch")
	}
	activeValidators := make(map[phase0.ValidatorIndex]*apiv1.Validator)
	for _, validator := range response.Data {
		_, exists := c.validators[validator.Index]
		if len(c.validators) > 0 && !exists {
			continue
		}

		if validator.Validator.ActivationEpoch <= c.summary.Epoch && validator.Validator.ExitEpoch > c.summary.Epoch {
			activeValidators[validator.Index] = validator
		}
	}

	return activeValidators, nil
}

func (c *command) processAttesterDuties(ctx context.Context) error {
	activeValidators, err := c.activeValidators(ctx)
	if err != nil {
		return err
	}
	c.summary.ActiveValidators = len(activeValidators)

	// Obtain number of validators that voted for blocks in the epoch.
	// These votes can be included anywhere from the second slot of
	// the epoch to the first slot of the next-but-one epoch.
	firstSlot := c.chainTime.FirstSlotOfEpoch(c.summary.Epoch) + 1
	lastSlot := c.chainTime.FirstSlotOfEpoch(c.summary.Epoch + 2)
	if lastSlot > c.chainTime.CurrentSlot() {
		lastSlot = c.chainTime.CurrentSlot()
	}

	participatingValidators, headCorrectValidators, headTimelyValidators, sourceTimelyValidators, targetCorrectValidators, targetTimelyValidators, participations, err := c.processSlots(ctx, firstSlot, lastSlot)
	if err != nil {
		return err
	}

	c.summary.ParticipatingValidators = len(participatingValidators)
	c.summary.HeadCorrectValidators = len(headCorrectValidators)
	c.summary.HeadTimelyValidators = len(headTimelyValidators)
	c.summary.SourceTimelyValidators = len(sourceTimelyValidators)
	c.summary.TargetCorrectValidators = len(targetCorrectValidators)
	c.summary.TargetTimelyValidators = len(targetTimelyValidators)

	c.summary.NonParticipatingValidators = make([]*attestingValidator, 0, len(activeValidators)-len(participatingValidators))
	for activeValidatorIndex := range activeValidators {
		if _, exists := participatingValidators[activeValidatorIndex]; !exists {
			if _, exists := participations[activeValidatorIndex]; exists {
				c.summary.NonParticipatingValidators = append(c.summary.NonParticipatingValidators, participations[activeValidatorIndex])
			}
		}
		if _, exists := headCorrectValidators[activeValidatorIndex]; !exists {
			if _, exists := participations[activeValidatorIndex]; exists {
				c.summary.NonHeadCorrectValidators = append(c.summary.NonHeadCorrectValidators, participations[activeValidatorIndex])
			}
		}
		if _, exists := headTimelyValidators[activeValidatorIndex]; !exists {
			if _, exists := participations[activeValidatorIndex]; exists {
				c.summary.NonHeadTimelyValidators = append(c.summary.NonHeadTimelyValidators, participations[activeValidatorIndex])
			}
		}
		if _, exists := targetCorrectValidators[activeValidatorIndex]; !exists {
			if _, exists := participations[activeValidatorIndex]; exists {
				c.summary.NonTargetCorrectValidators = append(c.summary.NonTargetCorrectValidators, participations[activeValidatorIndex])
			}
		}
		if _, exists := sourceTimelyValidators[activeValidatorIndex]; !exists {
			if _, exists := participations[activeValidatorIndex]; exists {
				c.summary.NonSourceTimelyValidators = append(c.summary.NonSourceTimelyValidators, participations[activeValidatorIndex])
			}
		}
	}
	sort.Slice(c.summary.NonParticipatingValidators, func(i int, j int) bool {
		if c.summary.NonParticipatingValidators[i].Slot != c.summary.NonParticipatingValidators[j].Slot {
			return c.summary.NonParticipatingValidators[i].Slot < c.summary.NonParticipatingValidators[j].Slot
		}
		if c.summary.NonParticipatingValidators[i].Committee != c.summary.NonParticipatingValidators[j].Committee {
			return c.summary.NonParticipatingValidators[i].Committee < c.summary.NonParticipatingValidators[j].Committee
		}
		return c.summary.NonParticipatingValidators[i].Validator < c.summary.NonParticipatingValidators[j].Validator
	})

	return nil
}

//nolint:gocyclo
func (c *command) processSlots(ctx context.Context,
	firstSlot phase0.Slot,
	lastSlot phase0.Slot,
) (
	map[phase0.ValidatorIndex]struct{},
	map[phase0.ValidatorIndex]struct{},
	map[phase0.ValidatorIndex]struct{},
	map[phase0.ValidatorIndex]struct{},
	map[phase0.ValidatorIndex]struct{},
	map[phase0.ValidatorIndex]struct{},
	map[phase0.ValidatorIndex]*attestingValidator,
	error,
) {
	votes := make(map[phase0.ValidatorIndex]struct{})
	headCorrects := make(map[phase0.ValidatorIndex]struct{})
	headTimelys := make(map[phase0.ValidatorIndex]struct{})
	sourceTimelys := make(map[phase0.ValidatorIndex]struct{})
	targetCorrects := make(map[phase0.ValidatorIndex]struct{})
	targetTimelys := make(map[phase0.ValidatorIndex]struct{})
	allCommittees := make(map[phase0.Slot]map[phase0.CommitteeIndex][]phase0.ValidatorIndex)
	participations := make(map[phase0.ValidatorIndex]*attestingValidator)

	// Need a cache of beacon block headers to reduce lookup times.
	headersCache := util.NewBeaconBlockHeaderCache(c.beaconBlockHeadersProvider)

	for slot := firstSlot; slot <= lastSlot; slot++ {
		block, err := c.fetchBlock(ctx, fmt.Sprintf("%d", slot))
		if err != nil {
			return nil, nil, nil, nil, nil, nil, nil, errors.Wrap(err, fmt.Sprintf("failed to obtain block for slot %d", slot))
		}
		if block == nil {
			// No block at this slot; that's fine.
			continue
		}
		slot, err := block.Slot()
		if err != nil {
			return nil, nil, nil, nil, nil, nil, nil, err
		}
		attestations, err := block.Attestations()
		if err != nil {
			return nil, nil, nil, nil, nil, nil, nil, err
		}
		for _, attestation := range attestations {
			attestationData, err := attestation.Data()
			if err != nil {
				return nil, nil, nil, nil, nil, nil, nil, errors.Wrap(err, "failed to obtain attestation data")
			}
			if attestationData.Slot < c.chainTime.FirstSlotOfEpoch(c.summary.Epoch) || attestationData.Slot >= c.chainTime.FirstSlotOfEpoch(c.summary.Epoch+1) {
				// Outside of this epoch's range.
				continue
			}
			slotCommittees, exists := allCommittees[attestationData.Slot]
			if !exists {
				response, err := c.beaconCommitteesProvider.BeaconCommittees(ctx, &api.BeaconCommitteesOpts{
					State: fmt.Sprintf("%d", attestationData.Slot),
				})
				if err != nil {
					return nil, nil, nil, nil, nil, nil, nil, errors.Wrap(err, fmt.Sprintf("failed to obtain committees for slot %d", attestationData.Slot))
				}
				for _, beaconCommittee := range response.Data {
					if _, exists := allCommittees[beaconCommittee.Slot]; !exists {
						allCommittees[beaconCommittee.Slot] = make(map[phase0.CommitteeIndex][]phase0.ValidatorIndex)
					}

					allCommittees[beaconCommittee.Slot][beaconCommittee.Index] = beaconCommittee.Validators

					for _, index := range beaconCommittee.Validators {
						if len(c.validators) > 0 {
							if _, exists := c.validators[index]; !exists {
								// Not one of our validators.
								continue
							}
						}

						if _, exists := participations[index]; !exists {
							participations[index] = &attestingValidator{
								Validator: index,
								Slot:      beaconCommittee.Slot,
								Committee: beaconCommittee.Index,
							}
						}
					}
				}
				slotCommittees = allCommittees[attestationData.Slot]
			}
			if attestation.Version >= spec.DataVersionElectra {
				participations, votes, headCorrects, headTimelys, sourceTimelys, targetCorrects, targetTimelys, err = c.extractElectraAttestationData(
					ctx, attestation, attestationData, slotCommittees, slot, headersCache, participations, votes, headCorrects, headTimelys, sourceTimelys, targetCorrects, targetTimelys)
				if err != nil {
					return nil, nil, nil, nil, nil, nil, nil, err
				}
			} else {
				participations, votes, headCorrects, headTimelys, sourceTimelys, targetCorrects, targetTimelys, err = c.extractPhase0AttestationData(
					ctx, attestation, attestationData, slotCommittees, slot, headersCache, participations, votes, headCorrects, headTimelys, sourceTimelys, targetCorrects, targetTimelys)
				if err != nil {
					return nil, nil, nil, nil, nil, nil, nil, err
				}
			}
		}
	}

	return votes,
		headCorrects,
		headTimelys,
		sourceTimelys,
		targetCorrects,
		targetTimelys,
		participations,
		nil
}

func (c *command) extractPhase0AttestationData(ctx context.Context,
	attestation *spec.VersionedAttestation,
	attestationData *phase0.AttestationData,
	slotCommittees map[phase0.CommitteeIndex][]phase0.ValidatorIndex,
	slot phase0.Slot,
	headersCache *util.BeaconBlockHeaderCache,
	participations map[phase0.ValidatorIndex]*attestingValidator,
	votes map[phase0.ValidatorIndex]struct{},
	headCorrects map[phase0.ValidatorIndex]struct{},
	headTimelys map[phase0.ValidatorIndex]struct{},
	sourceTimelys map[phase0.ValidatorIndex]struct{},
	targetCorrects map[phase0.ValidatorIndex]struct{},
	targetTimelys map[phase0.ValidatorIndex]struct{},
) (
	map[phase0.ValidatorIndex]*attestingValidator,
	map[phase0.ValidatorIndex]struct{},
	map[phase0.ValidatorIndex]struct{},
	map[phase0.ValidatorIndex]struct{},
	map[phase0.ValidatorIndex]struct{},
	map[phase0.ValidatorIndex]struct{},
	map[phase0.ValidatorIndex]struct{},
	error,
) {
	committee := slotCommittees[attestationData.Index]

	inclusionDistance := slot - attestationData.Slot

	head, err := util.AttestationHead(ctx, headersCache, attestation)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, err
	}
	headCorrect, err := util.AttestationHeadCorrect(ctx, headersCache, attestation)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, err
	}
	target, err := util.AttestationTarget(ctx, headersCache, c.chainTime, attestation)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, err
	}
	targetCorrect, err := util.AttestationTargetCorrect(ctx, headersCache, c.chainTime, attestation)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, err
	}

	aggregationBits, err := attestation.AggregationBits()
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, errors.Wrap(err, "failed to obtain aggregation bits")
	}
	for i := range aggregationBits.Len() {
		if aggregationBits.BitAt(i) {
			validatorIndex := committee[int(i)]
			if len(c.validators) > 0 {
				if _, exists := c.validators[validatorIndex]; !exists {
					// Not one of our validators.
					continue
				}
			}

			// Only set the information from the first attestation we find for this validator.
			if participations[validatorIndex].InclusionSlot == 0 {
				participations[validatorIndex].HeadVote = &attestationData.BeaconBlockRoot
				participations[validatorIndex].Head = &head
				participations[validatorIndex].TargetVote = &attestationData.Target.Root
				participations[validatorIndex].Target = &target
				participations[validatorIndex].InclusionSlot = slot
			}

			votes[validatorIndex] = struct{}{}
			if _, exists := headCorrects[validatorIndex]; !exists && headCorrect {
				headCorrects[validatorIndex] = struct{}{}
			}
			if _, exists := headTimelys[validatorIndex]; !exists && headCorrect && inclusionDistance == 1 {
				headTimelys[validatorIndex] = struct{}{}
			}
			if _, exists := sourceTimelys[validatorIndex]; !exists && inclusionDistance <= 5 {
				sourceTimelys[validatorIndex] = struct{}{}
			}
			if _, exists := targetCorrects[validatorIndex]; !exists && targetCorrect {
				targetCorrects[validatorIndex] = struct{}{}
			}
			if _, exists := targetTimelys[validatorIndex]; !exists && targetCorrect && inclusionDistance <= 32 {
				targetTimelys[validatorIndex] = struct{}{}
			}
		}
	}
	return participations, votes, headCorrects, headTimelys, sourceTimelys, targetCorrects, targetTimelys, err
}

func (c *command) extractElectraAttestationData(ctx context.Context,
	attestation *spec.VersionedAttestation,
	attestationData *phase0.AttestationData,
	slotCommittees map[phase0.CommitteeIndex][]phase0.ValidatorIndex,
	slot phase0.Slot,
	headersCache *util.BeaconBlockHeaderCache,
	participations map[phase0.ValidatorIndex]*attestingValidator,
	votes map[phase0.ValidatorIndex]struct{},
	headCorrects map[phase0.ValidatorIndex]struct{},
	headTimelys map[phase0.ValidatorIndex]struct{},
	sourceTimelys map[phase0.ValidatorIndex]struct{},
	targetCorrects map[phase0.ValidatorIndex]struct{},
	targetTimelys map[phase0.ValidatorIndex]struct{},
) (
	map[phase0.ValidatorIndex]*attestingValidator,
	map[phase0.ValidatorIndex]struct{},
	map[phase0.ValidatorIndex]struct{},
	map[phase0.ValidatorIndex]struct{},
	map[phase0.ValidatorIndex]struct{},
	map[phase0.ValidatorIndex]struct{},
	map[phase0.ValidatorIndex]struct{},
	error,
) {
	committeeBits, err := attestation.CommitteeBits()
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, errors.Wrap(err, "failed to obtain committee bits")
	}
	for _, committeeIndex := range committeeBits.BitIndices() {
		committee := slotCommittees[phase0.CommitteeIndex(committeeIndex)]

		inclusionDistance := slot - attestationData.Slot

		head, err := util.AttestationHead(ctx, headersCache, attestation)
		if err != nil {
			return nil, nil, nil, nil, nil, nil, nil, err
		}
		headCorrect, err := util.AttestationHeadCorrect(ctx, headersCache, attestation)
		if err != nil {
			return nil, nil, nil, nil, nil, nil, nil, err
		}
		target, err := util.AttestationTarget(ctx, headersCache, c.chainTime, attestation)
		if err != nil {
			return nil, nil, nil, nil, nil, nil, nil, err
		}
		targetCorrect, err := util.AttestationTargetCorrect(ctx, headersCache, c.chainTime, attestation)
		if err != nil {
			return nil, nil, nil, nil, nil, nil, nil, err
		}

		aggregationBits, err := attestation.AggregationBits()
		if err != nil {
			return nil, nil, nil, nil, nil, nil, nil, errors.Wrap(err, "failed to obtain aggregation bits")
		}
		// Calculate the offset for the committee so we can extract the validator from the aggregate_bits.
		committeeOffset := calcCommitteeOffset(phase0.CommitteeIndex(committeeIndex), slotCommittees)

		// Range over the committee rather than the aggregate_bits as it's the smaller set.
		for i := range committee {
			aggregateIndex := committeeOffset + uint64(i)
			if aggregationBits.BitAt(aggregateIndex) {
				validatorIndex := committee[i]
				if len(c.validators) > 0 {
					if _, exists := c.validators[validatorIndex]; !exists {
						// Not one of our validators.
						continue
					}
				}

				// Only set the information from the first attestation we find for this validator.
				if participations[validatorIndex].InclusionSlot == 0 {
					participations[validatorIndex].HeadVote = &attestationData.BeaconBlockRoot
					participations[validatorIndex].Head = &head
					participations[validatorIndex].TargetVote = &attestationData.Target.Root
					participations[validatorIndex].Target = &target
					participations[validatorIndex].InclusionSlot = slot
				}

				votes[validatorIndex] = struct{}{}
				if _, exists := headCorrects[validatorIndex]; !exists && headCorrect {
					headCorrects[validatorIndex] = struct{}{}
				}
				if _, exists := headTimelys[validatorIndex]; !exists && headCorrect && inclusionDistance == 1 {
					headTimelys[validatorIndex] = struct{}{}
				}
				if _, exists := sourceTimelys[validatorIndex]; !exists && inclusionDistance <= 5 {
					sourceTimelys[validatorIndex] = struct{}{}
				}
				if _, exists := targetCorrects[validatorIndex]; !exists && targetCorrect {
					targetCorrects[validatorIndex] = struct{}{}
				}
				if _, exists := targetTimelys[validatorIndex]; !exists && targetCorrect && inclusionDistance <= 32 {
					targetTimelys[validatorIndex] = struct{}{}
				}
			}
		}
	}
	return participations, votes, headCorrects, headTimelys, sourceTimelys, targetCorrects, targetTimelys, err
}

func calcCommitteeOffset(committeeIndex phase0.CommitteeIndex, slotCommittees map[phase0.CommitteeIndex][]phase0.ValidatorIndex) uint64 {
	var total uint64
	for i := range committeeIndex {
		total += uint64(len(slotCommittees[i]))
	}
	return total
}

func (c *command) processSyncCommitteeDuties(ctx context.Context) error {
	if c.summary.Epoch < c.chainTime.AltairInitialEpoch() {
		// The epoch is pre-Altair.  No info but no error.
		return nil
	}

	committeeResponse, err := c.syncCommitteesProvider.SyncCommittee(ctx, &api.SyncCommitteeOpts{
		State: fmt.Sprintf("%d", c.summary.FirstSlot),
	})
	if err != nil {
		return errors.Wrap(err, "failed to obtain sync committee")
	}
	committee := committeeResponse.Data
	if len(committee.Validators) == 0 {
		return errors.Wrap(err, "empty sync committee")
	}

	for _, validatorIndex := range committee.Validators {
		if len(c.validators) == 0 {
			c.summary.SyncCommitteeValidators++
		} else {
			if _, exists := c.validators[validatorIndex]; exists {
				c.summary.SyncCommitteeValidators++
			}
		}
	}

	missed := make(map[phase0.ValidatorIndex]int)
	missedSlots := make(map[phase0.ValidatorIndex][]phase0.Slot)
	for _, index := range committee.Validators {
		missed[index] = 0
	}

	for slot := c.summary.FirstSlot; slot <= c.summary.LastSlot; slot++ {
		block, err := c.fetchBlock(ctx, fmt.Sprintf("%d", slot))
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to obtain block for slot %d", slot))
		}
		if block == nil {
			// If the block is missed we don't count the sync aggregate miss.
			continue
		}
		if block.Version == spec.DataVersionPhase0 {
			// No sync committees in this fork.
			return nil
		}

		aggregate, err := block.SyncAggregate()
		if err != nil {
			return errors.Wrapf(err, "failed to obtain sync aggregate for slot %d", slot)
		}
		for i := range aggregate.SyncCommitteeBits.Len() {
			validatorIndex := committee.Validators[int(i)]
			if _, exists := c.validators[validatorIndex]; !exists {
				// Not one of ours.
				continue
			}
			if !aggregate.SyncCommitteeBits.BitAt(i) {
				missed[validatorIndex]++
				missedSlots[validatorIndex] = append(missedSlots[validatorIndex], slot)
			}
		}
	}

	c.summary.SyncCommittee = make([]*epochSyncCommittee, 0, len(missed))
	for index, count := range missed {
		if count > 0 {
			c.summary.SyncCommittee = append(c.summary.SyncCommittee, &epochSyncCommittee{
				ValidatorIndex: index,
				Missed:         count,
				MissedSlots:    missedSlots[index],
			})
		}
	}

	sort.Slice(c.summary.SyncCommittee, func(i int, j int) bool {
		missedDiff := c.summary.SyncCommittee[i].Missed - c.summary.SyncCommittee[j].Missed
		if missedDiff != 0 {
			// Actually want to order by missed descending, so invert the expected condition.
			return missedDiff > 0
		}
		// Then order by validator index.
		return c.summary.SyncCommittee[i].ValidatorIndex < c.summary.SyncCommittee[j].ValidatorIndex
	})

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

func (c *command) processBlobs(ctx context.Context) error {
	for _, proposal := range c.summary.Proposals {
		block, err := c.fetchBlock(ctx, fmt.Sprintf("%d", proposal.Slot))
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to obtain block for slot %d", proposal.Slot))
		}
		if block == nil {
			continue
		}
		switch block.Version {
		case spec.DataVersionPhase0, spec.DataVersionAltair, spec.DataVersionBellatrix, spec.DataVersionCapella:
			// No blobs in these forks.
		case spec.DataVersionDeneb:
			c.summary.Blobs += len(block.Deneb.Message.Body.BlobKZGCommitments)
		case spec.DataVersionElectra:
			c.summary.Blobs += len(block.Electra.Message.Body.BlobKZGCommitments)
		default:
			return fmt.Errorf("unhandled block version %v", block.Version)
		}
	}

	return nil
}

func (c *command) fetchBlock(ctx context.Context,
	blockID string,
) (
	*spec.VersionedSignedBeaconBlock,
	error,
) {
	block, exists := c.blocksCache[blockID]
	if !exists {
		var err error
		blockResponse, err := c.blocksProvider.SignedBeaconBlock(ctx, &api.SignedBeaconBlockOpts{
			Block: blockID,
		})
		if err != nil {
			var apiErr *api.Error
			if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
				// No block for this slot, that's okay.
				return nil, nil
			}

			return nil, errors.Wrap(err, "failed to fetch block")
		}
		block = blockResponse.Data
		c.blocksCache[blockID] = block
	}
	return block, nil
}
