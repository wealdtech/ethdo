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

package blockanalyze

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	eth2client "github.com/attestantio/go-eth2-client"
	"github.com/attestantio/go-eth2-client/api"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
	"github.com/prysmaticlabs/go-bitfield"
	standardchaintime "github.com/wealdtech/ethdo/services/chaintime/standard"
	"github.com/wealdtech/ethdo/util"
)

func (c *command) process(ctx context.Context) error {
	// Obtain information we need to process.
	if err := c.setup(ctx); err != nil {
		return err
	}

	blockResponse, err := c.blocksProvider.SignedBeaconBlock(ctx, &api.SignedBeaconBlockOpts{
		Block: c.blockID,
	})
	if err != nil {
		var apiError *api.Error
		if errors.As(err, &apiError) && apiError.StatusCode == http.StatusNotFound {
			return errors.New("empty beacon block")
		}
		return errors.Wrap(err, "failed to obtain beacon block")
	}
	block := blockResponse.Data

	slot, err := block.Slot()
	if err != nil {
		return err
	}
	attestations, err := block.Attestations()
	if err != nil {
		return err
	}

	c.analysis = &blockAnalysis{
		Slot: slot,
	}

	// Calculate how many parents we need to fetch.
	minSlot := slot
	for _, attestation := range attestations {
		attestationData, err := attestation.Data()
		if err != nil {
			return errors.Wrap(err, "failed to obtain attestation data")
		}
		if attestationData.Slot < minSlot {
			minSlot = attestationData.Slot
		}
	}
	if c.debug {
		fmt.Printf("Need to fetch blocks to slot %d\n", minSlot)
	}

	if err := c.fetchParents(ctx, block, minSlot); err != nil {
		return err
	}

	return c.analyze(ctx, block)
}

func (c *command) analyze(ctx context.Context, block *spec.VersionedSignedBeaconBlock) error {
	if err := c.analyzeAttestations(ctx, block); err != nil {
		return err
	}

	return c.analyzeSyncCommittees(ctx, block)
}

func (c *command) analyzeAttestations(ctx context.Context, block *spec.VersionedSignedBeaconBlock) error {
	attestations, err := block.Attestations()
	if err != nil {
		return err
	}
	slot, err := block.Slot()
	if err != nil {
		return err
	}

	c.analysis.Attestations = make([]*attestationAnalysis, len(attestations))

	blockVotes := make(map[phase0.Slot]map[phase0.CommitteeIndex]bitfield.Bitlist)
	for i, attestation := range attestations {
		if c.debug {
			fmt.Printf("Processing attestation %d\n", i)
		}

		attestationData, err := attestation.Data()
		if err != nil {
			return errors.Wrap(err, "failed to obtain attestation data")
		}

		analysis := &attestationAnalysis{
			Head:     attestationData.BeaconBlockRoot,
			Target:   attestationData.Target.Root,
			Distance: int(slot - attestationData.Slot),
		}

		root, err := attestation.HashTreeRoot()
		if err != nil {
			return err
		}
		if info, exists := c.priorAttestations[fmt.Sprintf("%#x", root)]; exists {
			analysis.Duplicate = info
		} else {
			aggregationBits, err := attestation.AggregationBits()
			if err != nil {
				return err
			}
			_, exists := blockVotes[attestationData.Slot]
			if !exists {
				blockVotes[attestationData.Slot] = make(map[phase0.CommitteeIndex]bitfield.Bitlist)
			}
			_, exists = blockVotes[attestationData.Slot][attestationData.Index]
			if !exists {
				blockVotes[attestationData.Slot][attestationData.Index] = bitfield.NewBitlist(aggregationBits.Len())
			}

			// Count new votes.
			analysis.PossibleVotes = int(aggregationBits.Len())
			for j := range aggregationBits.Len() {
				if aggregationBits.BitAt(j) {
					analysis.Votes++
					if blockVotes[attestationData.Slot][attestationData.Index].BitAt(j) {
						// Already attested to in this block; skip.
						continue
					}
					if c.votes[attestationData.Slot][attestationData.Index].BitAt(j) {
						// Already attested to in a previous block; skip.
						continue
					}
					analysis.NewVotes++
					blockVotes[attestationData.Slot][attestationData.Index].SetBitAt(j, true)
				}
			}
			// Calculate head correct.
			analysis.HeadCorrect, err = c.calcHeadCorrect(ctx, attestation)
			if err != nil {
				return err
			}

			// Calculate head timely.
			analysis.HeadTimely = analysis.HeadCorrect && attestationData.Slot == slot-1

			// Calculate source timely.
			analysis.SourceTimely = attestationData.Slot >= slot-5

			// Calculate target correct.
			analysis.TargetCorrect, err = c.calcTargetCorrect(ctx, attestation)
			if err != nil {
				return err
			}

			// Calculate target timely.
			if block.Version < spec.DataVersionDeneb {
				analysis.TargetTimely = attestationData.Slot >= slot-32
			} else {
				analysis.TargetTimely = true
			}
		}

		// Calculate score and value.
		if analysis.TargetCorrect && analysis.TargetTimely {
			analysis.Score += float64(c.timelyTargetWeight) / float64(c.weightDenominator)
		}
		if analysis.SourceTimely {
			analysis.Score += float64(c.timelySourceWeight) / float64(c.weightDenominator)
		}
		if analysis.HeadCorrect && analysis.HeadTimely {
			analysis.Score += float64(c.timelyHeadWeight) / float64(c.weightDenominator)
		}
		analysis.Value = analysis.Score * float64(analysis.NewVotes)
		c.analysis.Value += analysis.Value

		c.analysis.Attestations[i] = analysis
	}

	return nil
}

func (c *command) fetchParents(ctx context.Context, block *spec.VersionedSignedBeaconBlock, minSlot phase0.Slot) error {
	parentRoot, err := block.ParentRoot()
	if err != nil {
		return err
	}
	root, err := block.Root()
	if err != nil {
		panic(err)
	}
	slot, err := block.Slot()
	if err != nil {
		panic(err)
	}
	if c.debug {
		fmt.Printf("Parent root of %#x@%d is %#x\n", root, slot, parentRoot)
	}

	// Obtain the parent block.
	parentBlockResponse, err := c.blocksProvider.SignedBeaconBlock(ctx, &api.SignedBeaconBlockOpts{
		Block: fmt.Sprintf("%#x", parentRoot),
	})
	if err != nil {
		var apiError *api.Error
		if errors.As(err, &apiError) && apiError.StatusCode == http.StatusNotFound {
			return errors.New("empty beacon block")
		}
		return err
	}
	parentBlock := parentBlockResponse.Data
	if parentBlock == nil {
		return fmt.Errorf("unable to obtain parent block %s", parentBlock)
	}

	parentSlot, err := parentBlock.Slot()
	if err != nil {
		return err
	}
	if parentSlot < minSlot {
		return nil
	}

	if err := c.processParentBlock(ctx, parentBlock); err != nil {
		return err
	}

	return c.fetchParents(ctx, parentBlock, minSlot)
}

func (c *command) processParentBlock(_ context.Context, block *spec.VersionedSignedBeaconBlock) error {
	attestations, err := block.Attestations()
	if err != nil {
		return err
	}
	slot, err := block.Slot()
	if err != nil {
		return err
	}
	if c.debug {
		fmt.Printf("Processing block %d\n", slot)
	}

	for i, attestation := range attestations {
		root, err := attestation.HashTreeRoot()
		if err != nil {
			return err
		}
		c.priorAttestations[fmt.Sprintf("%#x", root)] = &attestationDataInfo{
			Block: slot,
			Index: i,
		}

		attestationData, err := attestation.Data()
		if err != nil {
			return errors.Wrap(err, "failed to obtain attestation data")
		}
		aggregationBits, err := attestation.AggregationBits()
		if err != nil {
			return errors.Wrap(err, "failed to obtain attestation aggregation bits")
		}

		_, exists := c.votes[attestationData.Slot]
		if !exists {
			c.votes[attestationData.Slot] = make(map[phase0.CommitteeIndex]bitfield.Bitlist)
		}
		_, exists = c.votes[attestationData.Slot][attestationData.Index]
		if !exists {
			c.votes[attestationData.Slot][attestationData.Index] = bitfield.NewBitlist(aggregationBits.Len())
		}
		for j := range aggregationBits.Len() {
			if aggregationBits.BitAt(j) {
				c.votes[attestationData.Slot][attestationData.Index].SetBitAt(j, true)
			}
		}
	}

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

	// Obtain the number of active validators.
	var isProvider bool
	c.blocksProvider, isProvider = c.eth2Client.(eth2client.SignedBeaconBlockProvider)
	if !isProvider {
		return errors.New("connection does not provide signed beacon block information")
	}
	c.blockHeadersProvider, isProvider = c.eth2Client.(eth2client.BeaconBlockHeadersProvider)
	if !isProvider {
		return errors.New("connection does not provide beacon block header information")
	}

	specProvider, isProvider := c.eth2Client.(eth2client.SpecProvider)
	if !isProvider {
		return errors.New("connection does not provide spec information")
	}

	specResponse, err := specProvider.Spec(ctx, &api.SpecOpts{})
	if err != nil {
		return errors.Wrap(err, "failed to obtain spec")
	}

	tmp, exists := specResponse.Data["TIMELY_SOURCE_WEIGHT"]
	if !exists {
		// Set a default value based on the Altair spec.
		tmp = uint64(14)
	}
	var ok bool
	c.timelySourceWeight, ok = tmp.(uint64)
	if !ok {
		return errors.New("TIMELY_SOURCE_WEIGHT of unexpected type")
	}

	tmp, exists = specResponse.Data["TIMELY_TARGET_WEIGHT"]
	if !exists {
		// Set a default value based on the Altair spec.
		tmp = uint64(26)
	}
	c.timelyTargetWeight, ok = tmp.(uint64)
	if !ok {
		return errors.New("TIMELY_TARGET_WEIGHT of unexpected type")
	}

	tmp, exists = specResponse.Data["TIMELY_HEAD_WEIGHT"]
	if !exists {
		// Set a default value based on the Altair spec.
		tmp = uint64(14)
	}
	c.timelyHeadWeight, ok = tmp.(uint64)
	if !ok {
		return errors.New("TIMELY_HEAD_WEIGHT of unexpected type")
	}

	tmp, exists = specResponse.Data["SYNC_REWARD_WEIGHT"]
	if !exists {
		// Set a default value based on the Altair spec.
		tmp = uint64(2)
	}
	c.syncRewardWeight, ok = tmp.(uint64)
	if !ok {
		return errors.New("SYNC_REWARD_WEIGHT of unexpected type")
	}

	tmp, exists = specResponse.Data["PROPOSER_WEIGHT"]
	if !exists {
		// Set a default value based on the Altair spec.
		tmp = uint64(8)
	}
	c.proposerWeight, ok = tmp.(uint64)
	if !ok {
		return errors.New("PROPOSER_WEIGHT of unexpected type")
	}

	tmp, exists = specResponse.Data["WEIGHT_DENOMINATOR"]
	if !exists {
		// Set a default value based on the Altair spec.
		tmp = uint64(64)
	}
	c.weightDenominator, ok = tmp.(uint64)
	if !ok {
		return errors.New("WEIGHT_DENOMINATOR of unexpected type")
	}
	return nil
}

func (c *command) calcHeadCorrect(ctx context.Context, attestation *spec.VersionedAttestation) (bool, error) {
	attestationData, err := attestation.Data()
	if err != nil {
		return false, errors.Wrap(err, "failed to obtain attestation data")
	}

	slot := attestationData.Slot
	root, exists := c.headRoots[slot]
	if !exists {
		for {
			response, err := c.blockHeadersProvider.BeaconBlockHeader(ctx, &api.BeaconBlockHeaderOpts{
				Block: fmt.Sprintf("%d", slot),
			})
			if err != nil {
				var apiError *api.Error
				if errors.As(err, &apiError) && apiError.StatusCode == http.StatusNotFound {
					if c.debug {
						fmt.Printf("No block available for slot %d, assuming not in canonical chain", slot)
					}
					return false, nil
				}
				return false, err
			}
			if response.Data == nil {
				// No block.
				slot--
				continue
			}
			if !response.Data.Canonical {
				// Not canonical.
				slot--
				continue
			}
			c.headRoots[slot] = response.Data.Root
			root = response.Data.Root
			break
		}
	}

	return bytes.Equal(root[:], attestationData.BeaconBlockRoot[:]), nil
}

func (c *command) calcTargetCorrect(ctx context.Context, attestation *spec.VersionedAttestation) (bool, error) {
	attestationData, err := attestation.Data()
	if err != nil {
		return false, errors.Wrap(err, "failed to obtain attestation data")
	}

	root, exists := c.targetRoots[attestationData.Slot]
	if !exists {
		// Start with first slot of the target epoch.
		slot := c.chainTime.FirstSlotOfEpoch(attestationData.Target.Epoch)
		for {
			response, err := c.blockHeadersProvider.BeaconBlockHeader(ctx, &api.BeaconBlockHeaderOpts{
				Block: fmt.Sprintf("%d", slot),
			})
			if err != nil {
				var apiError *api.Error
				if errors.As(err, &apiError) && apiError.StatusCode == http.StatusNotFound {
					if c.debug {
						fmt.Printf("No block available for slot %d, assuming not in canonical chain", slot)
					}
					return false, nil
				}
			}
			if response.Data == nil {
				// No block.
				slot--
				continue
			}
			if !response.Data.Canonical {
				// Not canonical.
				slot--
				continue
			}
			c.targetRoots[attestationData.Slot] = response.Data.Root
			root = response.Data.Root
			break
		}
	}
	return bytes.Equal(root[:], attestationData.Target.Root[:]), nil
}

func (c *command) analyzeSyncCommittees(_ context.Context, block *spec.VersionedSignedBeaconBlock) error {
	c.analysis.SyncCommitee = &syncCommitteeAnalysis{}
	switch block.Version {
	case spec.DataVersionPhase0:
		return nil
	case spec.DataVersionAltair:
		c.analysis.SyncCommitee.Contributions = int(block.Altair.Message.Body.SyncAggregate.SyncCommitteeBits.Count())
		c.analysis.SyncCommitee.PossibleContributions = int(block.Altair.Message.Body.SyncAggregate.SyncCommitteeBits.Len())
		c.analysis.SyncCommitee.Score = float64(c.syncRewardWeight) / float64(c.weightDenominator)
		c.analysis.SyncCommitee.Value = c.analysis.SyncCommitee.Score * float64(c.analysis.SyncCommitee.Contributions)
		c.analysis.Value += c.analysis.SyncCommitee.Value
		return nil
	case spec.DataVersionBellatrix:
		c.analysis.SyncCommitee.Contributions = int(block.Bellatrix.Message.Body.SyncAggregate.SyncCommitteeBits.Count())
		c.analysis.SyncCommitee.PossibleContributions = int(block.Bellatrix.Message.Body.SyncAggregate.SyncCommitteeBits.Len())
		c.analysis.SyncCommitee.Score = float64(c.syncRewardWeight) / float64(c.weightDenominator)
		c.analysis.SyncCommitee.Value = c.analysis.SyncCommitee.Score * float64(c.analysis.SyncCommitee.Contributions)
		c.analysis.Value += c.analysis.SyncCommitee.Value
		return nil
	case spec.DataVersionCapella:
		c.analysis.SyncCommitee.Contributions = int(block.Capella.Message.Body.SyncAggregate.SyncCommitteeBits.Count())
		c.analysis.SyncCommitee.PossibleContributions = int(block.Capella.Message.Body.SyncAggregate.SyncCommitteeBits.Len())
		c.analysis.SyncCommitee.Score = float64(c.syncRewardWeight) / float64(c.weightDenominator)
		c.analysis.SyncCommitee.Value = c.analysis.SyncCommitee.Score * float64(c.analysis.SyncCommitee.Contributions)
		c.analysis.Value += c.analysis.SyncCommitee.Value
		return nil
	case spec.DataVersionDeneb:
		c.analysis.SyncCommitee.Contributions = int(block.Deneb.Message.Body.SyncAggregate.SyncCommitteeBits.Count())
		c.analysis.SyncCommitee.PossibleContributions = int(block.Deneb.Message.Body.SyncAggregate.SyncCommitteeBits.Len())
		c.analysis.SyncCommitee.Score = float64(c.syncRewardWeight) / float64(c.weightDenominator)
		c.analysis.SyncCommitee.Value = c.analysis.SyncCommitee.Score * float64(c.analysis.SyncCommitee.Contributions)
		c.analysis.Value += c.analysis.SyncCommitee.Value
		return nil
	case spec.DataVersionElectra:
		c.analysis.SyncCommitee.Contributions = int(block.Electra.Message.Body.SyncAggregate.SyncCommitteeBits.Count())
		c.analysis.SyncCommitee.PossibleContributions = int(block.Electra.Message.Body.SyncAggregate.SyncCommitteeBits.Len())
		c.analysis.SyncCommitee.Score = float64(c.syncRewardWeight) / float64(c.weightDenominator)
		c.analysis.SyncCommitee.Value = c.analysis.SyncCommitee.Score * float64(c.analysis.SyncCommitee.Contributions)
		c.analysis.Value += c.analysis.SyncCommitee.Value
		return nil
	case spec.DataVersionFulu:
		c.analysis.SyncCommitee.Contributions = int(block.Fulu.Message.Body.SyncAggregate.SyncCommitteeBits.Count())
		c.analysis.SyncCommitee.PossibleContributions = int(block.Fulu.Message.Body.SyncAggregate.SyncCommitteeBits.Len())
		c.analysis.SyncCommitee.Score = float64(c.syncRewardWeight) / float64(c.weightDenominator)
		c.analysis.SyncCommitee.Value = c.analysis.SyncCommitee.Score * float64(c.analysis.SyncCommitee.Contributions)
		c.analysis.Value += c.analysis.SyncCommitee.Value
		return nil
	default:
		return fmt.Errorf("unsupported block version %d", block.Version)
	}
}
