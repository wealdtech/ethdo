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
		c.summary.Proposals = append(c.summary.Proposals, &epochProposal{
			Slot:     duty.Slot,
			Proposer: duty.ValidatorIndex,
			Block:    present,
		})
	}

	return nil
}

func (c *command) activeValidators(ctx context.Context) (map[phase0.ValidatorIndex]*apiv1.Validator, error) {
	response, err := c.validatorsProvider.Validators(ctx, &api.ValidatorsOpts{
		State: fmt.Sprintf("%d", c.chainTime.FirstSlotOfEpoch(c.summary.Epoch)),
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to obtain validators for epoch")
	}
	activeValidators := make(map[phase0.ValidatorIndex]*apiv1.Validator)
	for _, validator := range response.Data {
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

	var votes map[phase0.ValidatorIndex]struct{}
	var participations map[phase0.ValidatorIndex]*nonParticipatingValidator
	c.summary.ParticipatingValidators, c.summary.HeadCorrectValidators, c.summary.HeadTimelyValidators, c.summary.SourceTimelyValidators, c.summary.TargetCorrectValidators, c.summary.TargetTimelyValidators, votes, participations, err = c.processSlots(ctx, firstSlot, lastSlot)
	if err != nil {
		return err
	}

	c.summary.NonParticipatingValidators = make([]*nonParticipatingValidator, 0, len(activeValidators)-len(votes))
	for activeValidatorIndex := range activeValidators {
		if _, exists := votes[activeValidatorIndex]; !exists {
			if _, exists := participations[activeValidatorIndex]; exists {
				c.summary.NonParticipatingValidators = append(c.summary.NonParticipatingValidators, participations[activeValidatorIndex])
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

func (c *command) processSlots(ctx context.Context,
	firstSlot phase0.Slot,
	lastSlot phase0.Slot,
) (
	int,
	int,
	int,
	int,
	int,
	int,
	map[phase0.ValidatorIndex]struct{},
	map[phase0.ValidatorIndex]*nonParticipatingValidator,
	error,
) {
	votes := make(map[phase0.ValidatorIndex]struct{})
	headCorrects := make(map[phase0.ValidatorIndex]struct{})
	headTimelys := make(map[phase0.ValidatorIndex]struct{})
	sourceTimelys := make(map[phase0.ValidatorIndex]struct{})
	targetCorrects := make(map[phase0.ValidatorIndex]struct{})
	targetTimelys := make(map[phase0.ValidatorIndex]struct{})
	allCommittees := make(map[phase0.Slot]map[phase0.CommitteeIndex][]phase0.ValidatorIndex)
	participations := make(map[phase0.ValidatorIndex]*nonParticipatingValidator)

	// Need a cache of beacon block headers to reduce lookup times.
	headersCache := util.NewBeaconBlockHeaderCache(c.beaconBlockHeadersProvider)

	for slot := firstSlot; slot <= lastSlot; slot++ {
		block, err := c.fetchBlock(ctx, fmt.Sprintf("%d", slot))
		if err != nil {
			return 0, 0, 0, 0, 0, 0, nil, nil, errors.Wrap(err, fmt.Sprintf("failed to obtain block for slot %d", slot))
		}
		if block == nil {
			// No block at this slot; that's fine.
			continue
		}
		slot, err := block.Slot()
		if err != nil {
			return 0, 0, 0, 0, 0, 0, nil, nil, err
		}
		attestations, err := block.Attestations()
		if err != nil {
			return 0, 0, 0, 0, 0, 0, nil, nil, err
		}
		for _, attestation := range attestations {
			if attestation.Data.Slot < c.chainTime.FirstSlotOfEpoch(c.summary.Epoch) || attestation.Data.Slot >= c.chainTime.FirstSlotOfEpoch(c.summary.Epoch+1) {
				// Outside of this epoch's range.
				continue
			}
			slotCommittees, exists := allCommittees[attestation.Data.Slot]
			if !exists {
				response, err := c.beaconCommitteesProvider.BeaconCommittees(ctx, &api.BeaconCommitteesOpts{
					State: fmt.Sprintf("%d", attestation.Data.Slot),
				})
				if err != nil {
					return 0, 0, 0, 0, 0, 0, nil, nil, errors.Wrap(err, fmt.Sprintf("failed to obtain committees for slot %d", attestation.Data.Slot))
				}
				for _, beaconCommittee := range response.Data {
					if _, exists := allCommittees[beaconCommittee.Slot]; !exists {
						allCommittees[beaconCommittee.Slot] = make(map[phase0.CommitteeIndex][]phase0.ValidatorIndex)
					}
					allCommittees[beaconCommittee.Slot][beaconCommittee.Index] = beaconCommittee.Validators
					for _, index := range beaconCommittee.Validators {
						participations[index] = &nonParticipatingValidator{
							Validator: index,
							Slot:      beaconCommittee.Slot,
							Committee: beaconCommittee.Index,
						}
					}
				}
				slotCommittees = allCommittees[attestation.Data.Slot]
			}
			committee := slotCommittees[attestation.Data.Index]

			inclusionDistance := slot - attestation.Data.Slot
			headCorrect, err := util.AttestationHeadCorrect(ctx, headersCache, attestation)
			if err != nil {
				return 0, 0, 0, 0, 0, 0, nil, nil, err
			}
			targetCorrect, err := util.AttestationTargetCorrect(ctx, headersCache, c.chainTime, attestation)
			if err != nil {
				return 0, 0, 0, 0, 0, 0, nil, nil, err
			}

			for i := uint64(0); i < attestation.AggregationBits.Len(); i++ {
				if attestation.AggregationBits.BitAt(i) {
					votes[committee[int(i)]] = struct{}{}
					if _, exists := headCorrects[committee[int(i)]]; !exists && headCorrect {
						headCorrects[committee[int(i)]] = struct{}{}
					}
					if _, exists := headTimelys[committee[int(i)]]; !exists && headCorrect && inclusionDistance == 1 {
						headTimelys[committee[int(i)]] = struct{}{}
					}
					if _, exists := sourceTimelys[committee[int(i)]]; !exists && inclusionDistance <= 5 {
						sourceTimelys[committee[int(i)]] = struct{}{}
					}
					if _, exists := targetCorrects[committee[int(i)]]; !exists && targetCorrect {
						targetCorrects[committee[int(i)]] = struct{}{}
					}
					if _, exists := targetTimelys[committee[int(i)]]; !exists && targetCorrect && inclusionDistance <= 32 {
						targetTimelys[committee[int(i)]] = struct{}{}
					}
				}
			}
		}
	}
	return len(votes),
		len(headCorrects),
		len(headTimelys),
		len(sourceTimelys),
		len(targetCorrects),
		len(targetTimelys),
		votes,
		participations,
		nil
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

	missed := make(map[phase0.ValidatorIndex]int)
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
		for i := uint64(0); i < aggregate.SyncCommitteeBits.Len(); i++ {
			if !aggregate.SyncCommitteeBits.BitAt(i) {
				missed[committee.Validators[int(i)]]++
			}
		}
	}

	c.summary.SyncCommittee = make([]*epochSyncCommittee, 0, len(missed))
	for index, count := range missed {
		if count > 0 {
			c.summary.SyncCommittee = append(c.summary.SyncCommittee, &epochSyncCommittee{
				Index:  index,
				Missed: count,
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
		return c.summary.SyncCommittee[i].Index < c.summary.SyncCommittee[j].Index
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
	for slot := c.summary.FirstSlot; slot <= c.summary.LastSlot; slot++ {
		block, err := c.fetchBlock(ctx, fmt.Sprintf("%d", slot))
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to obtain block for slot %d", slot))
		}
		if block == nil {
			continue
		}
		switch block.Version {
		case spec.DataVersionPhase0, spec.DataVersionAltair, spec.DataVersionBellatrix, spec.DataVersionCapella:
			// No blobs in these forks.
		case spec.DataVersionDeneb:
			c.summary.Blobs += len(block.Deneb.Message.Body.BlobKZGCommitments)
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
