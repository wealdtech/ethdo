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

package chaineth1votes

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
	json    bool

	// Beacon node connection.
	timeout                  time.Duration
	connection               string
	allowInsecureConnections bool

	// Input.
	xepoch  string
	xperiod string

	// Data access.
	eth2Client                eth2client.Service
	chainTime                 chaintime.Service
	beaconStateProvider       eth2client.BeaconStateProvider
	slotsPerEpoch             uint64
	epochsPerEth1VotingPeriod uint64

	// Output.
	slot          phase0.Slot
	epoch         phase0.Epoch
	period        uint64
	periodStart   time.Time
	periodEnd     time.Time
	incumbent     *phase0.ETH1Data
	eth1DataVotes []*phase0.ETH1Data
	votes         map[string]*vote
}

type vote struct {
	Vote  *phase0.ETH1Data `json:"vote"`
	Count int              `json:"count"`
}

func newCommand(_ context.Context) (*command, error) {
	c := &command{
		quiet:   viper.GetBool("quiet"),
		verbose: viper.GetBool("verbose"),
		debug:   viper.GetBool("debug"),
		json:    viper.GetBool("json"),
	}

	// Timeout.
	if viper.GetDuration("timeout") == 0 {
		return nil, errors.New("timeout is required")
	}
	c.timeout = viper.GetDuration("timeout")

	c.xepoch = viper.GetString("epoch")
	c.xperiod = viper.GetString("period")

	c.connection = viper.GetString("connection")
	c.allowInsecureConnections = viper.GetBool("allow-insecure-connections")

	return c, nil
}
