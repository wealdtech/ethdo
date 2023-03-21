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

package blockanalyze

import (
	"context"
	"time"

	eth2client "github.com/attestantio/go-eth2-client"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
	"github.com/prysmaticlabs/go-bitfield"
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
	blockID    string
	stream     bool
	jsonOutput bool

	// Data access.
	eth2Client           eth2client.Service
	chainTime            chaintime.Service
	blocksProvider       eth2client.SignedBeaconBlockProvider
	blockHeadersProvider eth2client.BeaconBlockHeadersProvider

	// Constants.
	timelySourceWeight uint64
	timelyTargetWeight uint64
	timelyHeadWeight   uint64
	syncRewardWeight   uint64
	proposerWeight     uint64
	weightDenominator  uint64

	// Processing.
	priorAttestations map[string]*attestationData
	// Head roots provides the root of the head slot at given slots.
	headRoots map[phase0.Slot]phase0.Root
	// Target roots provides the root of the target epoch at given slots.
	targetRoots map[phase0.Slot]phase0.Root

	// Block info.
	// Map is slot -> committee index -> validator committee index -> votes.
	votes map[phase0.Slot]map[phase0.CommitteeIndex]bitfield.Bitlist

	// Results.
	analysis *blockAnalysis
}

type blockAnalysis struct {
	Slot         phase0.Slot            `json:"slot"`
	Attestations []*attestationAnalysis `json:"attestations"`
	SyncCommitee *syncCommitteeAnalysis `json:"sync_committee"`
	Value        float64                `json:"value"`
}

type attestationAnalysis struct {
	Head          phase0.Root      `json:"head"`
	Target        phase0.Root      `json:"target"`
	Distance      int              `json:"distance"`
	Duplicate     *attestationData `json:"duplicate,omitempty"`
	NewVotes      int              `json:"new_votes"`
	Votes         int              `json:"votes"`
	PossibleVotes int              `json:"possible_votes"`
	HeadCorrect   bool             `json:"head_correct"`
	HeadTimely    bool             `json:"head_timely"`
	SourceTimely  bool             `json:"source_timely"`
	TargetCorrect bool             `json:"target_correct"`
	TargetTimely  bool             `json:"target_timely"`
	Score         float64          `json:"score"`
	Value         float64          `json:"value"`
}

type syncCommitteeAnalysis struct {
	Contributions         int     `json:"contributions"`
	PossibleContributions int     `json:"possible_contributions"`
	Score                 float64 `json:"score"`
	Value                 float64 `json:"value"`
}

type attestationData struct {
	Block phase0.Slot `json:"block"`
	Index int         `json:"index"`
}

func newCommand(_ context.Context) (*command, error) {
	c := &command{
		quiet:             viper.GetBool("quiet"),
		verbose:           viper.GetBool("verbose"),
		debug:             viper.GetBool("debug"),
		priorAttestations: make(map[string]*attestationData),
		headRoots:         make(map[phase0.Slot]phase0.Root),
		targetRoots:       make(map[phase0.Slot]phase0.Root),
		votes:             make(map[phase0.Slot]map[phase0.CommitteeIndex]bitfield.Bitlist),
	}

	// Timeout.
	if viper.GetDuration("timeout") == 0 {
		return nil, errors.New("timeout is required")
	}
	c.timeout = viper.GetDuration("timeout")

	c.connection = viper.GetString("connection")
	c.allowInsecureConnections = viper.GetBool("allow-insecure-connections")

	c.blockID = viper.GetString("blockid")
	c.stream = viper.GetBool("stream")
	c.jsonOutput = viper.GetBool("json")

	return c, nil
}
