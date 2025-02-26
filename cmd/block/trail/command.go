// Copyright © 2025 Weald Technology Trading.
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

package blocktrail

import (
	"context"
	"time"

	eth2client "github.com/attestantio/go-eth2-client"
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
	blockID    string
	jsonOutput bool
	target     string
	maxBlocks  int

	// Data access.
	consensusClient      eth2client.Service
	chainTime            chaintime.Service
	blocksProvider       eth2client.SignedBeaconBlockProvider
	blockHeadersProvider eth2client.BeaconBlockHeadersProvider

	// Processing.
	justifiedCheckpoint *phase0.Checkpoint
	finalizedCheckpoint *phase0.Checkpoint

	// Results.
	steps []*step
	found bool
}

type step struct {
	Slot       phase0.Slot `json:"slot"`
	Root       phase0.Root `json:"root"`
	ParentRoot phase0.Root `json:"parent_root"`
	State      string      `json:"state,omitempty"`
	// Not a slot, but we're using it to steal the JSON processing.
	ExecutionBlock phase0.Slot   `json:"execution_block"`
	ExecutionHash  phase0.Hash32 `json:"execution_hash"`
}

func newCommand(_ context.Context) (*command, error) {
	c := &command{
		timeout:                  viper.GetDuration("timeout"),
		quiet:                    viper.GetBool("quiet"),
		verbose:                  viper.GetBool("verbose"),
		debug:                    viper.GetBool("debug"),
		jsonOutput:               viper.GetBool("json"),
		connection:               viper.GetString("connection"),
		allowInsecureConnections: viper.GetBool("allow-insecure-connections"),
		blockID:                  viper.GetString("blockid"),
		target:                   viper.GetString("target"),
		maxBlocks:                viper.GetInt("max-blocks"),
		steps:                    make([]*step, 0),
	}

	// Timeout.
	if c.timeout == 0 {
		return nil, errors.New("timeout is required")
	}

	return c, nil
}
