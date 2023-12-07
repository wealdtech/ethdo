// Copyright © 2021 - 2023 Weald Technology Trading.
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

package standard

import (
	"context"
	"time"

	eth2client "github.com/attestantio/go-eth2-client"
	"github.com/attestantio/go-eth2-client/api"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	zerologger "github.com/rs/zerolog/log"
)

// Service provides chain time services.
type Service struct {
	genesisTime                  time.Time
	slotDuration                 time.Duration
	slotsPerEpoch                uint64
	epochsPerSyncCommitteePeriod uint64
	altairForkEpoch              phase0.Epoch
	bellatrixForkEpoch           phase0.Epoch
	capellaForkEpoch             phase0.Epoch
	denebForkEpoch               phase0.Epoch
}

// module-wide log.
var log zerolog.Logger

// New creates a new controller.
func New(ctx context.Context, params ...Parameter) (*Service, error) {
	parameters, err := parseAndCheckParameters(params...)
	if err != nil {
		return nil, errors.Wrap(err, "problem with parameters")
	}

	// Set logging.
	log = zerologger.With().Str("service", "chaintime").Str("impl", "standard").Logger().Level(parameters.logLevel)

	genesisResponse, err := parameters.genesisProvider.Genesis(ctx, &api.GenesisOpts{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to obtain genesis time")
	}
	log.Trace().Time("genesis_time", genesisResponse.Data.GenesisTime).Msg("Obtained genesis time")

	specResponse, err := parameters.specProvider.Spec(ctx, &api.SpecOpts{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to obtain spec")
	}

	tmp, exists := specResponse.Data["SECONDS_PER_SLOT"]
	if !exists {
		return nil, errors.New("SECONDS_PER_SLOT not found in spec")
	}
	slotDuration, ok := tmp.(time.Duration)
	if !ok {
		return nil, errors.New("SECONDS_PER_SLOT of unexpected type")
	}

	tmp, exists = specResponse.Data["SLOTS_PER_EPOCH"]
	if !exists {
		return nil, errors.New("SLOTS_PER_EPOCH not found in spec")
	}
	slotsPerEpoch, ok := tmp.(uint64)
	if !ok {
		return nil, errors.New("SLOTS_PER_EPOCH of unexpected type")
	}

	var epochsPerSyncCommitteePeriod uint64
	if tmp, exists := specResponse.Data["EPOCHS_PER_SYNC_COMMITTEE_PERIOD"]; exists {
		tmp2, ok := tmp.(uint64)
		if !ok {
			return nil, errors.New("EPOCHS_PER_SYNC_COMMITTEE_PERIOD of unexpected type")
		}
		epochsPerSyncCommitteePeriod = tmp2
	}

	altairForkEpoch, err := fetchAltairForkEpoch(ctx, parameters.specProvider)
	if err != nil {
		// Set to far future epoch.
		altairForkEpoch = 0xffffffffffffffff
	}
	log.Trace().Uint64("epoch", uint64(altairForkEpoch)).Msg("Obtained Altair fork epoch")

	bellatrixForkEpoch, err := fetchBellatrixForkEpoch(ctx, parameters.specProvider)
	if err != nil {
		// Set to far future epoch.
		bellatrixForkEpoch = 0xffffffffffffffff
	}
	log.Trace().Uint64("epoch", uint64(bellatrixForkEpoch)).Msg("Obtained Bellatrix fork epoch")

	capellaForkEpoch, err := fetchCapellaForkEpoch(ctx, parameters.specProvider)
	if err != nil {
		// Set to far future epoch.
		capellaForkEpoch = 0xffffffffffffffff
	}
	log.Trace().Uint64("epoch", uint64(capellaForkEpoch)).Msg("Obtained Capella fork epoch")

	denebForkEpoch, err := fetchDenebForkEpoch(ctx, parameters.specProvider)
	if err != nil {
		// Set to far future epoch.
		denebForkEpoch = 0xffffffffffffffff
	}
	log.Trace().Uint64("epoch", uint64(denebForkEpoch)).Msg("Obtained Deneb fork epoch")

	s := &Service{
		genesisTime:                  genesisResponse.Data.GenesisTime,
		slotDuration:                 slotDuration,
		slotsPerEpoch:                slotsPerEpoch,
		epochsPerSyncCommitteePeriod: epochsPerSyncCommitteePeriod,
		altairForkEpoch:              altairForkEpoch,
		bellatrixForkEpoch:           bellatrixForkEpoch,
		capellaForkEpoch:             capellaForkEpoch,
		denebForkEpoch:               denebForkEpoch,
	}

	return s, nil
}

// GenesisTime provides the time of the chain's genesis.
func (s *Service) GenesisTime() time.Time {
	return s.genesisTime
}

// SlotsPerEpoch provides the number of slots in the chain's epoch.
func (s *Service) SlotsPerEpoch() uint64 {
	return s.slotsPerEpoch
}

// SlotDuration provides the duration of the chain's slot.
func (s *Service) SlotDuration() time.Duration {
	return s.slotDuration
}

// StartOfSlot provides the time at which a given slot starts.
func (s *Service) StartOfSlot(slot phase0.Slot) time.Time {
	return s.genesisTime.Add(time.Duration(slot) * s.slotDuration)
}

// StartOfEpoch provides the time at which a given epoch starts.
func (s *Service) StartOfEpoch(epoch phase0.Epoch) time.Time {
	return s.genesisTime.Add(time.Duration(uint64(epoch)*s.slotsPerEpoch) * s.slotDuration)
}

// CurrentSlot provides the current slot.
func (s *Service) CurrentSlot() phase0.Slot {
	if s.genesisTime.After(time.Now()) {
		return 0
	}
	return phase0.Slot(uint64(time.Since(s.genesisTime).Seconds()) / uint64(s.slotDuration.Seconds()))
}

// CurrentEpoch provides the current epoch.
func (s *Service) CurrentEpoch() phase0.Epoch {
	return phase0.Epoch(uint64(s.CurrentSlot()) / s.slotsPerEpoch)
}

// CurrentSyncCommitteePeriod provides the current sync committee period.
func (s *Service) CurrentSyncCommitteePeriod() uint64 {
	return uint64(s.CurrentEpoch()) / s.epochsPerSyncCommitteePeriod
}

// SlotToEpoch provides the epoch of a given slot.
func (s *Service) SlotToEpoch(slot phase0.Slot) phase0.Epoch {
	return phase0.Epoch(uint64(slot) / s.slotsPerEpoch)
}

// SlotToSyncCommitteePeriod provides the sync committee period of the given slot.
func (s *Service) SlotToSyncCommitteePeriod(slot phase0.Slot) uint64 {
	return uint64(s.SlotToEpoch(slot)) / s.epochsPerSyncCommitteePeriod
}

// FirstSlotOfEpoch provides the first slot of the given epoch.
func (s *Service) FirstSlotOfEpoch(epoch phase0.Epoch) phase0.Slot {
	return phase0.Slot(uint64(epoch) * s.slotsPerEpoch)
}

// LastSlotOfEpoch provides the last slot of the given epoch.
func (s *Service) LastSlotOfEpoch(epoch phase0.Epoch) phase0.Slot {
	return phase0.Slot(uint64(epoch)*s.slotsPerEpoch + s.slotsPerEpoch - 1)
}

// TimestampToSlot provides the slot of the given timestamp.
func (s *Service) TimestampToSlot(timestamp time.Time) phase0.Slot {
	if timestamp.Before(s.genesisTime) {
		return 0
	}
	secondsSinceGenesis := uint64(timestamp.Sub(s.genesisTime).Seconds())
	return phase0.Slot(secondsSinceGenesis / uint64(s.slotDuration.Seconds()))
}

// TimestampToEpoch provides the epoch of the given timestamp.
func (s *Service) TimestampToEpoch(timestamp time.Time) phase0.Epoch {
	if timestamp.Before(s.genesisTime) {
		return 0
	}
	secondsSinceGenesis := uint64(timestamp.Sub(s.genesisTime).Seconds())
	return phase0.Epoch(secondsSinceGenesis / uint64(s.slotDuration.Seconds()) / s.slotsPerEpoch)
}

// FirstEpochOfSyncPeriod provides the first epoch of the given sync period.
// Note that epochs before the sync committee period will provide the Altair hard fork epoch.
func (s *Service) FirstEpochOfSyncPeriod(period uint64) phase0.Epoch {
	epoch := phase0.Epoch(period * s.epochsPerSyncCommitteePeriod)
	if epoch < s.altairForkEpoch {
		epoch = s.altairForkEpoch
	}
	return epoch
}

// AltairInitialEpoch provides the epoch at which the Altair hard fork takes place.
func (s *Service) AltairInitialEpoch() phase0.Epoch {
	return s.altairForkEpoch
}

// AltairInitialSyncCommitteePeriod provides the sync committee period in which the Altair hard fork takes place.
func (s *Service) AltairInitialSyncCommitteePeriod() uint64 {
	return uint64(s.altairForkEpoch) / s.epochsPerSyncCommitteePeriod
}

func fetchAltairForkEpoch(ctx context.Context,
	specProvider eth2client.SpecProvider,
) (
	phase0.Epoch,
	error,
) {
	// Fetch the fork version.
	specResponse, err := specProvider.Spec(ctx, &api.SpecOpts{})
	if err != nil {
		return 0, errors.Wrap(err, "failed to obtain spec")
	}
	tmp, exists := specResponse.Data["ALTAIR_FORK_EPOCH"]
	if !exists {
		return 0, errors.New("altair fork version not known by chain")
	}
	epoch, isEpoch := tmp.(uint64)
	if !isEpoch {
		//nolint:revive
		return 0, errors.New("ALTAIR_FORK_EPOCH is not a uint64!")
	}

	return phase0.Epoch(epoch), nil
}

// BellatrixInitialEpoch provides the epoch at which the Bellatrix hard fork takes place.
func (s *Service) BellatrixInitialEpoch() phase0.Epoch {
	return s.bellatrixForkEpoch
}

func fetchBellatrixForkEpoch(ctx context.Context,
	specProvider eth2client.SpecProvider,
) (
	phase0.Epoch,
	error,
) {
	// Fetch the fork version.
	specResponse, err := specProvider.Spec(ctx, &api.SpecOpts{})
	if err != nil {
		return 0, errors.Wrap(err, "failed to obtain spec")
	}
	tmp, exists := specResponse.Data["BELLATRIX_FORK_EPOCH"]
	if !exists {
		return 0, errors.New("bellatrix fork version not known by chain")
	}
	epoch, isEpoch := tmp.(uint64)
	if !isEpoch {
		//nolint:revive
		return 0, errors.New("BELLATRIX_FORK_EPOCH is not a uint64!")
	}

	return phase0.Epoch(epoch), nil
}

// CapellaInitialEpoch provides the epoch at which the Capella hard fork takes place.
func (s *Service) CapellaInitialEpoch() phase0.Epoch {
	return s.capellaForkEpoch
}

func fetchCapellaForkEpoch(ctx context.Context,
	specProvider eth2client.SpecProvider,
) (
	phase0.Epoch,
	error,
) {
	// Fetch the fork version.
	specResponse, err := specProvider.Spec(ctx, &api.SpecOpts{})
	if err != nil {
		return 0, errors.Wrap(err, "failed to obtain spec")
	}
	tmp, exists := specResponse.Data["CAPELLA_FORK_EPOCH"]
	if !exists {
		return 0, errors.New("capella fork version not known by chain")
	}
	epoch, isEpoch := tmp.(uint64)
	if !isEpoch {
		//nolint:revive
		return 0, errors.New("CAPELLA_FORK_EPOCH is not a uint64!")
	}

	return phase0.Epoch(epoch), nil
}

// DenebInitialEpoch provides the epoch at which the Deneb hard fork takes place.
func (s *Service) DenebInitialEpoch() phase0.Epoch {
	return s.denebForkEpoch
}

func fetchDenebForkEpoch(ctx context.Context,
	specProvider eth2client.SpecProvider,
) (
	phase0.Epoch,
	error,
) {
	// Fetch the fork version.
	specResponse, err := specProvider.Spec(ctx, &api.SpecOpts{})
	if err != nil {
		return 0, errors.Wrap(err, "failed to obtain spec")
	}
	tmp, exists := specResponse.Data["DENEB_FORK_EPOCH"]
	if !exists {
		return 0, errors.New("deneb fork version not known by chain")
	}
	epoch, isEpoch := tmp.(uint64)
	if !isEpoch {
		//nolint:revive
		return 0, errors.New("DENEB_FORK_EPOCH is not a uint64!")
	}

	return phase0.Epoch(epoch), nil
}
