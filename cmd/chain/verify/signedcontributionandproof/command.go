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
	"time"

	eth2client "github.com/attestantio/go-eth2-client"
	api "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec/altair"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type command struct {
	quiet   bool
	verbose bool
	debug   bool

	// Beacon node connection.
	timeout                  time.Duration
	connection               string
	allowInsecureConnections bool

	// Input.
	data string
	item *altair.SignedContributionAndProof

	// Data access.
	eth2Client         eth2client.Service
	validatorsProvider eth2client.ValidatorsProvider

	// Data.
	spec          map[string]interface{}
	validator     *api.Validator
	syncCommittee *api.SyncCommittee

	// Output.
	itemStructureValid                       bool
	validatorKnown                           bool
	validatorInSyncCommittee                 bool
	validatorIsAggregator                    bool
	contributionSignatureValidFormat         bool
	contributionAndProofSignatureValidFormat bool
	contributionAndProofSignatureValid       bool
	additionalInfo                           string
}

func newCommand(_ context.Context) (*command, error) {
	c := &command{
		quiet:   viper.GetBool("quiet"),
		verbose: viper.GetBool("verbose"),
		debug:   viper.GetBool("debug"),
	}

	// Timeout.
	if viper.GetDuration("timeout") == 0 {
		return nil, errors.New("timeout is required")
	}
	c.timeout = viper.GetDuration("timeout")

	if viper.GetString("data") == "" {
		return nil, errors.New("data is required")
	}
	c.data = viper.GetString("data")

	c.connection = viper.GetString("connection")
	c.allowInsecureConnections = viper.GetBool("allow-insecure-connections")

	return c, nil
}
