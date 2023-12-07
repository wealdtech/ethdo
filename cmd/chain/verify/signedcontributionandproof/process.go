// Copyright © 2021 Weald Technology Trading.
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

package chainverifysignedcontributionandproof

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"

	eth2client "github.com/attestantio/go-eth2-client"
	"github.com/attestantio/go-eth2-client/api"
	"github.com/attestantio/go-eth2-client/spec/altair"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
	"github.com/wealdtech/ethdo/util"
	e2types "github.com/wealdtech/go-eth2-types/v2"
)

func (c *command) process(ctx context.Context) error {
	// Parse the data.
	if c.data == "" {
		return errors.New("no data supplied")
	}
	c.item = &altair.SignedContributionAndProof{}
	err := json.Unmarshal([]byte(c.data), c.item)
	if err != nil {
		c.additionalInfo = err.Error()
		//nolint:nilerr
		return nil
	}
	c.itemStructureValid = true

	// Obtain information we need to process.
	if err := c.setup(ctx); err != nil {
		return err
	}

	for _, validatorIndex := range c.syncCommittee.Validators {
		if validatorIndex == c.item.Message.AggregatorIndex {
			c.validatorInSyncCommittee = true
			break
		}
	}
	if !c.validatorInSyncCommittee {
		return nil
	}

	// Ensure the validator is an aggregator.
	isAggregator, err := c.isAggregator(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to ascertain if sync committee member is aggregator")
	}
	if !isAggregator {
		return nil
	}
	c.validatorIsAggregator = true

	// Confirm the contribution signature.
	if err := c.confirmContributionSignature(ctx); err != nil {
		return errors.Wrap(err, "failed to confirm the contribution signature")
	}

	// Confirm the contribution and proof signature.
	if err := c.confirmContributionAndProofSignature(ctx); err != nil {
		return errors.Wrap(err, "failed to confirm the contribution and proof signature")
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

	// Obtain the validator.
	var isProvider bool
	c.validatorsProvider, isProvider = c.eth2Client.(eth2client.ValidatorsProvider)
	if !isProvider {
		return errors.New("connection does not provide validator information")
	}

	stateID := fmt.Sprintf("%d", c.item.Message.Contribution.Slot)
	response, err := c.validatorsProvider.Validators(ctx, &api.ValidatorsOpts{
		State:   stateID,
		Indices: []phase0.ValidatorIndex{c.item.Message.AggregatorIndex},
	})
	if err != nil {
		return errors.Wrap(err, "failed to obtain validator information")
	}

	if len(response.Data) == 0 || response.Data[c.item.Message.AggregatorIndex] == nil {
		return nil
	}
	c.validatorKnown = true
	c.validator = response.Data[c.item.Message.AggregatorIndex]

	// Obtain the sync committee
	syncCommitteesProvider, isProvider := c.eth2Client.(eth2client.SyncCommitteesProvider)
	if !isProvider {
		return errors.New("connection does not provide sync committee information")
	}
	syncCommitteeResponse, err := syncCommitteesProvider.SyncCommittee(ctx, &api.SyncCommitteeOpts{
		State: stateID,
	})
	if err != nil {
		return errors.Wrap(err, "failed to obtain sync committee information")
	}
	c.syncCommittee = syncCommitteeResponse.Data

	return nil
}

// isAggregator returns true if the given.
func (c *command) isAggregator(ctx context.Context) (bool, error) {
	// Calculate the modulo.
	specProvider, isProvider := c.eth2Client.(eth2client.SpecProvider)
	if !isProvider {
		return false, errors.New("connection does not provide spec information")
	}
	specResponse, err := specProvider.Spec(ctx, &api.SpecOpts{})
	if err != nil {
		return false, errors.Wrap(err, "failed to obtain spec information")
	}
	c.spec = specResponse.Data

	tmp, exists := c.spec["SYNC_COMMITTEE_SIZE"]
	if !exists {
		return false, errors.New("spec does not contain SYNC_COMMITTEE_SIZE")
	}
	if _, isUint64 := tmp.(uint64); !isUint64 {
		return false, errors.New("spec returned non-integer value for SYNC_COMMITTEE_SIZE")
	}
	syncCommitteeSize := tmp.(uint64)
	if c.debug {
		fmt.Fprintf(os.Stderr, "sync committee size is %d\n", syncCommitteeSize)
	}

	tmp, exists = c.spec["SYNC_COMMITTEE_SUBNET_COUNT"]
	if !exists {
		return false, errors.New("spec does not contain SYNC_COMMITTEE_SUBNET_COUNT")
	}
	if _, isUint64 := tmp.(uint64); !isUint64 {
		return false, errors.New("spec returned non-integer value for SYNC_COMMITTEE_SUBNET_COUNT")
	}
	syncCommitteeSubnetCount := tmp.(uint64)
	if c.debug {
		fmt.Fprintf(os.Stderr, "sync committee subnet count is %d\n", syncCommitteeSubnetCount)
	}

	tmp, exists = c.spec["TARGET_AGGREGATORS_PER_SYNC_SUBCOMMITTEE"]
	if !exists {
		return false, errors.New("spec does not contain TARGET_AGGREGATORS_PER_SYNC_SUBCOMMITTEE")
	}
	if _, isUint64 := tmp.(uint64); !isUint64 {
		return false, errors.New("spec returned non-integer value for TARGET_AGGREGATORS_PER_SYNC_SUBCOMMITTEE")
	}
	targetAggregatorsPerSyncSubcommittee := tmp.(uint64)
	if c.debug {
		fmt.Fprintf(os.Stderr, "target aggregators per sync subcommittee is %d\n", targetAggregatorsPerSyncSubcommittee)
	}

	modulo := syncCommitteeSize / syncCommitteeSubnetCount / targetAggregatorsPerSyncSubcommittee
	if modulo < 1 {
		modulo = 1
	}
	if c.debug {
		fmt.Fprintf(os.Stderr, "modulo is %d\n", modulo)
	}

	// Hash the selection proof.
	sigHash := sha256.New()
	n, err := sigHash.Write(c.item.Message.SelectionProof[:])
	if err != nil {
		return false, errors.Wrap(err, "failed to hash the selection proof")
	}
	if n != len(c.item.Signature[:]) {
		return false, errors.New("failed to write all bytes of the selection proof to the hash")
	}
	hash := sigHash.Sum(nil)
	if c.debug {
		fmt.Fprintf(os.Stderr, "hash of selection proof is %#x\n", hash)
	}

	return binary.LittleEndian.Uint64(hash[:8])%modulo == 0, nil
}

func (c *command) confirmContributionSignature(ctx context.Context) error {
	sigBytes := make([]byte, 96)
	copy(sigBytes, c.item.Message.Contribution.Signature[:])
	_, err := e2types.BLSSignatureFromBytes(sigBytes)
	if err != nil {
		c.additionalInfo = err.Error()
		//nolint:nilerr
		return nil
	}
	c.contributionSignatureValidFormat = true

	subCommittee := c.syncCommittee.ValidatorAggregates[c.item.Message.Contribution.SubcommitteeIndex]
	includedIndices := make([]phase0.ValidatorIndex, 0, len(subCommittee))
	for i := uint64(0); i < c.item.Message.Contribution.AggregationBits.Len(); i++ {
		if c.item.Message.Contribution.AggregationBits.BitAt(i) {
			includedIndices = append(includedIndices, subCommittee[int(i)])
		}
	}
	if c.debug {
		fmt.Fprintf(os.Stderr, "Contribution validator indices: %v (%d)\n", includedIndices, len(includedIndices))
	}

	response, err := c.validatorsProvider.Validators(ctx, &api.ValidatorsOpts{
		State:   "head",
		Indices: includedIndices,
	})
	if err != nil {
		return errors.Wrap(err, "failed to obtain subcommittee validators")
	}
	if len(response.Data) == 0 {
		return errors.New("obtained empty subcommittee validator list")
	}

	var aggregatePubKey *e2types.BLSPublicKey
	for _, v := range response.Data {
		pubKeyBytes := make([]byte, 48)
		copy(pubKeyBytes, v.Validator.PublicKey[:])
		pubKey, err := e2types.BLSPublicKeyFromBytes(pubKeyBytes)
		if err != nil {
			return errors.Wrap(err, "failed to aggregate public key")
		}
		if aggregatePubKey == nil {
			aggregatePubKey = pubKey
		} else {
			aggregatePubKey.Aggregate(pubKey)
		}
	}
	if c.debug {
		fmt.Fprintf(os.Stderr, "Aggregate public key is %#x\n", aggregatePubKey.Marshal())
	}

	// Don't have the ability to carry out the batch verification at current.

	return nil
}

func (c *command) confirmContributionAndProofSignature(ctx context.Context) error {
	sigBytes := make([]byte, 96)
	copy(sigBytes, c.item.Signature[:])
	sig, err := e2types.BLSSignatureFromBytes(sigBytes)
	if err != nil {
		c.additionalInfo = err.Error()
		//nolint:nilerr
		return nil
	}
	c.contributionAndProofSignatureValidFormat = true

	pubKeyBytes := make([]byte, 48)
	copy(pubKeyBytes, c.validator.Validator.PublicKey[:])
	pubKey, err := e2types.BLSPublicKeyFromBytes(pubKeyBytes)
	if err != nil {
		return errors.Wrap(err, "failed to configure public key")
	}

	objectRoot, err := c.item.Message.HashTreeRoot()
	if err != nil {
		return errors.Wrap(err, "failed to obtain signging root")
	}

	tmp, exists := c.spec["DOMAIN_CONTRIBUTION_AND_PROOF"]
	if !exists {
		return errors.New("spec does not contain DOMAIN_CONTRIBUTION_AND_PROOF")
	}
	if _, isUint64 := tmp.(phase0.DomainType); !isUint64 {
		return errors.New("spec returned non-domain type value for DOMAIN_CONTRIBUTION_AND_PROOF")
	}
	contributionAndProofDomainType := tmp.(phase0.DomainType)
	if c.debug {
		fmt.Fprintf(os.Stderr, "contribution and proof domain type is %#x\n", contributionAndProofDomainType)
	}
	domain, err := c.eth2Client.(eth2client.DomainProvider).Domain(ctx, contributionAndProofDomainType, phase0.Epoch(c.item.Message.Contribution.Slot/32))
	if err != nil {
		return errors.Wrap(err, "failed to obtain domain")
	}

	container := &phase0.SigningData{
		ObjectRoot: objectRoot,
		Domain:     domain,
	}
	signingRoot, err := container.HashTreeRoot()
	if err != nil {
		return errors.Wrap(err, "failed to obtain signging root")
	}

	c.contributionAndProofSignatureValid = sig.Verify(signingRoot[:], pubKey)

	return nil
}
