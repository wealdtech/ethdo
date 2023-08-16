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

package proposerduties

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
	slot       string
	jsonOutput bool

	// Data access.
	eth2Client             eth2client.Service
	chainTime              chaintime.Service
	proposerDutiesProvider eth2client.ProposerDutiesProvider

	// Results.
	results *results
}

type results struct {
	Epoch  phase0.Epoch          `json:"epoch"`
	Duties []*apiv1.ProposerDuty `json:"duties"`
}

func newCommand(_ context.Context) (*command, error) {
	c := &command{
		quiet:                    viper.GetBool("quiet"),
		verbose:                  viper.GetBool("verbose"),
		debug:                    viper.GetBool("debug"),
		timeout:                  viper.GetDuration("timeout"),
		connection:               viper.GetString("connection"),
		allowInsecureConnections: viper.GetBool("allow-insecure-connections"),
		epoch:                    viper.GetString("epoch"),
		slot:                     viper.GetString("slot"),
		jsonOutput:               viper.GetBool("json"),
		results:                  &results{},
	}

	// Timeout.
	if c.timeout == 0 {
		return nil, errors.New("timeout is required")
	}

	return c, nil
}
