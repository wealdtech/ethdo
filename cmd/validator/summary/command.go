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
	"time"

	eth2client "github.com/attestantio/go-eth2-client"
	apiv1 "github.com/attestantio/go-eth2-client/api/v1"
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
	validators []string
	jsonOutput bool

	// Data access.
	eth2Client                 eth2client.Service
	chainTime                  chaintime.Service
	proposerDutiesProvider     eth2client.ProposerDutiesProvider
	attesterDutiesProvider     eth2client.AttesterDutiesProvider
	blocksProvider             eth2client.SignedBeaconBlockProvider
	syncCommitteesProvider     eth2client.SyncCommitteesProvider
	validatorsProvider         eth2client.ValidatorsProvider
	beaconCommitteesProvider   eth2client.BeaconCommitteesProvider
	beaconBlockHeadersProvider eth2client.BeaconBlockHeadersProvider

	// Processing.
	validatorsByIndex map[phase0.ValidatorIndex]*apiv1.Validator

	// Results.
	summary *validatorSummary
}

type validatorSummary struct {
	Epoch                      phase0.Epoch                 `json:"epoch"`
	Validators                 []*apiv1.Validator           `json:"validators"`
	FirstSlot                  phase0.Slot                  `json:"first_slot"`
	LastSlot                   phase0.Slot                  `json:"last_slot"`
	ActiveValidators           int                          `json:"active_validators"`
	ParticipatingValidators    int                          `json:"participating_validators"`
	NonParticipatingValidators []*nonParticipatingValidator `json:"non_participating_validators"`
	IncorrectHeadValidators    []*validatorFault            `json:"incorrect_head_validators"`
	UntimelyHeadValidators     []*validatorFault            `json:"untimely_head_validators"`
	UntimelySourceValidators   []*validatorFault            `json:"untimely_source_validators"`
	IncorrectTargetValidators  []*validatorFault            `json:"incorrect_target_validators"`
	UntimelyTargetValidators   []*validatorFault            `json:"untimely_target_validators"`
	Slots                      []*slot                      `json:"slots"`
	Proposals                  []*epochProposal             `json:"-"`
	SyncCommittee              []*epochSyncCommittee        `json:"-"`
}

type slot struct {
	Slot         phase0.Slot       `json:"slot"`
	Attestations *slotAttestations `json:"attestations"`
}

type slotAttestations struct {
	Expected      int `json:"expected"`
	Included      int `json:"included"`
	CorrectHead   int `json:"correct_head"`
	TimelyHead    int `json:"timely_head"`
	CorrectTarget int `json:"correct_target"`
	TimelyTarget  int `json:"timely_target"`
	TimelySource  int `json:"timely_source"`
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

type validatorFault struct {
	Validator         phase0.ValidatorIndex   `json:"validator_index"`
	AttestationData   *phase0.AttestationData `json:"attestation_data,omitempty"`
	InclusionDistance int                     `json:"inclusion_delay"`
}

type nonParticipatingValidator struct {
	Validator phase0.ValidatorIndex `json:"validator_index"`
	Slot      phase0.Slot           `json:"slot"`
	Committee phase0.CommitteeIndex `json:"committee_index"`
}

func newCommand(_ context.Context) (*command, error) {
	c := &command{
		quiet:             viper.GetBool("quiet"),
		verbose:           viper.GetBool("verbose"),
		debug:             viper.GetBool("debug"),
		validatorsByIndex: make(map[phase0.ValidatorIndex]*apiv1.Validator),
		summary:           &validatorSummary{},
	}

	// Timeout.
	if viper.GetDuration("timeout") == 0 {
		return nil, errors.New("timeout is required")
	}
	c.timeout = viper.GetDuration("timeout")

	c.connection = viper.GetString("connection")
	c.allowInsecureConnections = viper.GetBool("allow-insecure-connections")

	c.epoch = viper.GetString("epoch")
	c.validators = viper.GetStringSlice("validators")
	c.jsonOutput = viper.GetBool("json")

	return c, nil
}
