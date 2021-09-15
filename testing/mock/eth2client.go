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

package mock

import (
	"context"
	"time"

	eth2client "github.com/attestantio/go-eth2-client"
	api "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/phase0"
)

// GenesisTimeProvider is a mock for eth2client.GenesisTimeProvider.
type GenesisTimeProvider struct {
	genesisTime time.Time
}

// NewGenesisTimeProvider returns a mock genesis time provider with the provided value.
func NewGenesisTimeProvider(genesisTime time.Time) eth2client.GenesisTimeProvider {
	return &GenesisTimeProvider{
		genesisTime: genesisTime,
	}
}

// GenesisTime is a mock.
func (m *GenesisTimeProvider) GenesisTime(ctx context.Context) (time.Time, error) {
	return m.genesisTime, nil
}

// SpecProvider is a mock for eth2client.SpecProvider.
type SpecProvider struct {
	spec map[string]interface{}
}

// NewSpecProvider returns a mock spec provider with the provided values.
func NewSpecProvider(slotDuration time.Duration,
	slotsPerEpoch uint64,
	epochsPerSyncCommitteePeriod uint64,
) eth2client.SpecProvider {
	return &SpecProvider{
		spec: map[string]interface{}{
			"SECONDS_PER_SLOT":                 slotDuration,
			"SLOTS_PER_EPOCH":                  slotsPerEpoch,
			"EPOCHS_PER_SYNC_COMMITTEE_PERIOD": epochsPerSyncCommitteePeriod,
		},
	}
}

// Spec is a mock.
func (m *SpecProvider) Spec(ctx context.Context) (map[string]interface{}, error) {
	return m.spec, nil
}

// ForkScheduleProvider is a mock for eth2client.ForkScheduleProvider.
type ForkScheduleProvider struct {
	schedule []*phase0.Fork
}

// NewForkScheduleProvider returns a mock spec provider with the provided values.
func NewForkScheduleProvider(schedule []*phase0.Fork) eth2client.ForkScheduleProvider {
	return &ForkScheduleProvider{
		schedule: schedule,
	}
}

// ForkSchedule is a mock.
func (m *ForkScheduleProvider) ForkSchedule(ctx context.Context) ([]*phase0.Fork, error) {
	return m.schedule, nil
}

// SlotsPerEpochProvider is a mock for eth2client.SlotsPerEpochProvider.
type SlotsPerEpochProvider struct {
	slotsPerEpoch uint64
}

// NewSlotsPerEpochProvider returns a mock slots per epoch provider with the provided value.
func NewSlotsPerEpochProvider(slotsPerEpoch uint64) eth2client.SlotsPerEpochProvider {
	return &SlotsPerEpochProvider{
		slotsPerEpoch: slotsPerEpoch,
	}
}

// SlotsPerEpoch is a mock.
func (m *SlotsPerEpochProvider) SlotsPerEpoch(ctx context.Context) (uint64, error) {
	return m.slotsPerEpoch, nil
}

// AttestationsSubmitter is a mock for eth2client.AttestationsSubmitter.
type AttestationsSubmitter struct{}

// NewAttestationSubmitter returns a mock attestations submitter with the provided value.
func NewAttestationSubmitter() eth2client.AttestationsSubmitter {
	return &AttestationsSubmitter{}
}

// SubmitAttestations is a mock.
func (m *AttestationsSubmitter) SubmitAttestations(ctx context.Context, attestations []*phase0.Attestation) error {
	return nil
}

// BeaconBlockSubmitter is a mock for eth2client.BeaconBlockSubmitter.
type BeaconBlockSubmitter struct{}

// NewBeaconBlockSubmitter returns a mock beacon block submitter with the provided value.
func NewBeaconBlockSubmitter() eth2client.BeaconBlockSubmitter {
	return &BeaconBlockSubmitter{}
}

// SubmitBeaconBlock is a mock.
func (m *BeaconBlockSubmitter) SubmitBeaconBlock(ctx context.Context, bloc *spec.VersionedSignedBeaconBlock) error {
	return nil
}

// AggregateAttestationsSubmitter is a mock for eth2client.AggregateAttestationsSubmitter.
type AggregateAttestationsSubmitter struct{}

// NewAggregateAttestationsSubmitter returns a mock aggregate attestation submitter with the provided value.
func NewAggregateAttestationsSubmitter() eth2client.AggregateAttestationsSubmitter {
	return &AggregateAttestationsSubmitter{}
}

// SubmitAggregateAttestations is a mock.
func (m *AggregateAttestationsSubmitter) SubmitAggregateAttestations(ctx context.Context, aggregates []*phase0.SignedAggregateAndProof) error {
	return nil
}

// BeaconCommitteeSubscriptionsSubmitter is a mock for eth2client.BeaconCommitteeSubscriptionsSubmitter.
type BeaconCommitteeSubscriptionsSubmitter struct{}

// NewBeaconCommitteeSubscriptionsSubmitter returns a mock beacon committee subscription submitter with the provided value.
func NewBeaconCommitteeSubscriptionsSubmitter() eth2client.BeaconCommitteeSubscriptionsSubmitter {
	return &BeaconCommitteeSubscriptionsSubmitter{}
}

// SubmitBeaconCommitteeSubscriptions is a mock.
func (m *BeaconCommitteeSubscriptionsSubmitter) SubmitBeaconCommitteeSubscriptions(ctx context.Context, subscriptions []*api.BeaconCommitteeSubscription) error {
	return nil
}
