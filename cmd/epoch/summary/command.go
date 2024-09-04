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

package epochsummary

import (
	"context"
	"time"

	eth2client "github.com/attestantio/go-eth2-client"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/wealdtech/ethdo/services/chaintime"
)

type command struct {
	quiet   bool
	verbose bool
	debug   bool

	// Beacon node connection.
	timeout                  time.Duration
	connection               string
	allowInsecureConnections bool

	// Operation.
	epoch         string
	validatorsStr []string
	validators    map[phase0.ValidatorIndex]struct{}
	stream        bool
	jsonOutput    bool

	// Data access.
	eth2Client                 eth2client.Service
	chainTime                  chaintime.Service
	proposerDutiesProvider     eth2client.ProposerDutiesProvider
	blocksProvider             eth2client.SignedBeaconBlockProvider
	syncCommitteesProvider     eth2client.SyncCommitteesProvider
	validatorsProvider         eth2client.ValidatorsProvider
	beaconCommitteesProvider   eth2client.BeaconCommitteesProvider
	beaconBlockHeadersProvider eth2client.BeaconBlockHeadersProvider

	// Caches.
	blocksCache map[string]*spec.VersionedSignedBeaconBlock

	// Results.
	summary *epochSummary
}

type epochSummary struct {
	Epoch                      phase0.Epoch          `json:"epoch"`
	FirstSlot                  phase0.Slot           `json:"first_slot"`
	LastSlot                   phase0.Slot           `json:"last_slot"`
	Blocks                     int                   `json:"blocks"`
	Proposals                  []*epochProposal      `json:"proposals"`
	SyncCommitteeValidators    int                   `json:"sync_committee_validators"`
	SyncCommittee              []*epochSyncCommittee `json:"sync_committees"`
	ActiveValidators           int                   `json:"active_validators"`
	ParticipatingValidators    int                   `json:"participating_validators"`
	HeadCorrectValidators      int                   `json:"head_correct_validators"`
	HeadTimelyValidators       int                   `json:"head_timely_validators"`
	SourceTimelyValidators     int                   `json:"source_timely_validators"`
	TargetCorrectValidators    int                   `json:"target_correct_validators"`
	TargetTimelyValidators     int                   `json:"target_timely_validators"`
	NonParticipatingValidators []*attestingValidator `json:"nonparticipating_validators"`
	NonHeadCorrectValidators   []*attestingValidator `json:"nonheadcorrect_validators"`
	NonHeadTimelyValidators    []*attestingValidator `json:"nonheadtimely_validators"`
	NonTargetCorrectValidators []*attestingValidator `json:"nontargetcorrect_validators"`
	NonSourceTimelyValidators  []*attestingValidator `json:"nonsourcetimely_validators"`
	Blobs                      int                   `json:"blobs"`
}

type epochProposal struct {
	ValidatorIndex phase0.ValidatorIndex `json:"validator_index"`
	Slot           phase0.Slot           `json:"slot"`
	Block          bool                  `json:"block"`
}

type epochSyncCommittee struct {
	ValidatorIndex phase0.ValidatorIndex `json:"validator_index"`
	Missed         int                   `json:"missed"`
	MissedSlots    []phase0.Slot         `json:"missed_slots"`
}

type attestingValidator struct {
	Validator     phase0.ValidatorIndex `json:"validator_index"`
	Slot          phase0.Slot           `json:"slot"`
	Committee     phase0.CommitteeIndex `json:"committee_index"`
	HeadVote      *phase0.Root          `json:"head_vote,omitempty"`
	Head          *phase0.Root          `json:"head,omitempty"`
	TargetVote    *phase0.Root          `json:"target_vote,omitempty"`
	Target        *phase0.Root          `json:"target,omitempty"`
	InclusionSlot phase0.Slot           `json:"inclusion_slot,omitempty"`
}

func newCommand(_ context.Context) (*command, error) {
	c := &command{
		quiet:         viper.GetBool("quiet"),
		verbose:       viper.GetBool("verbose"),
		debug:         viper.GetBool("debug"),
		validatorsStr: viper.GetStringSlice("validators"),
		summary: &epochSummary{
			Proposals: make([]*epochProposal, 0),
		},
		validators:  make(map[phase0.ValidatorIndex]struct{}),
		blocksCache: make(map[string]*spec.VersionedSignedBeaconBlock),
	}

	// Timeout.
	if viper.GetDuration("timeout") == 0 {
		return nil, errors.New("timeout is required")
	}
	c.timeout = viper.GetDuration("timeout")

	c.connection = viper.GetString("connection")
	c.allowInsecureConnections = viper.GetBool("allow-insecure-connections")

	c.epoch = viper.GetString("epoch")
	c.stream = viper.GetBool("stream")
	c.jsonOutput = viper.GetBool("json")

	return c, nil
}
