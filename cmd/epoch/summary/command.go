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
	epoch      string
	stream     bool
	jsonOutput bool

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
	Epoch                      phase0.Epoch                 `json:"epoch"`
	FirstSlot                  phase0.Slot                  `json:"first_slot"`
	LastSlot                   phase0.Slot                  `json:"last_slot"`
	Proposals                  []*epochProposal             `json:"proposals"`
	SyncCommittee              []*epochSyncCommittee        `json:"sync_committees"`
	ActiveValidators           int                          `json:"active_validators"`
	ParticipatingValidators    int                          `json:"participating_validators"`
	HeadCorrectValidators      int                          `json:"head_correct_validators"`
	HeadTimelyValidators       int                          `json:"head_timely_validators"`
	SourceTimelyValidators     int                          `json:"source_timely_validators"`
	TargetCorrectValidators    int                          `json:"target_correct_validators"`
	TargetTimelyValidators     int                          `json:"target_timely_validators"`
	NonParticipatingValidators []*nonParticipatingValidator `json:"nonparticipating_validators"`
	Blobs                      int                          `json:"blobs"`
}

type epochProposal struct {
	Slot     phase0.Slot           `json:"slot"`
	Proposer phase0.ValidatorIndex `json:"proposer"`
	Block    bool                  `json:"block"`
}

type epochSyncCommittee struct {
	Index  phase0.ValidatorIndex `json:"index"`
	Missed int                   `json:"missed"`
}

type nonParticipatingValidator struct {
	Validator phase0.ValidatorIndex `json:"validator_index"`
	Slot      phase0.Slot           `json:"slot"`
	Committee phase0.CommitteeIndex `json:"committee_index"`
}

func newCommand(_ context.Context) (*command, error) {
	c := &command{
		quiet:       viper.GetBool("quiet"),
		verbose:     viper.GetBool("verbose"),
		debug:       viper.GetBool("debug"),
		summary:     &epochSummary{},
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
