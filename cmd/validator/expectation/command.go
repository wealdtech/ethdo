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

package validatorexpectation

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	eth2client "github.com/attestantio/go-eth2-client"
	"github.com/hako/durafmt"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
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
	validators int64

	// Data access.
	eth2Client         eth2client.Service
	validatorsProvider eth2client.ValidatorsProvider

	// Results.
	res *results
}

type results struct {
	activeValidators          uint64
	timeBetweenProposals      time.Duration
	timeBetweenSyncCommittees time.Duration
}

type resultsJSON struct {
	ActiveValidators          string `json:"active_validators"`
	TimeBetweenProposals      string `json:"time_between_proposals"`
	SecsBetweenProposals      string `json:"secs_between_proposals"`
	TimeBetweenSyncCommittees string `json:"time_between_sync_committees"`
	SecsBetweenSyncCommittees string `json:"secs_between_sync_committees"`
}

func (r *results) MarshalJSON() ([]byte, error) {
	data := &resultsJSON{
		ActiveValidators:          strconv.FormatUint(r.activeValidators, 10),
		TimeBetweenProposals:      durafmt.Parse(r.timeBetweenProposals).LimitFirstN(2).String(),
		SecsBetweenProposals:      strconv.FormatInt(int64(r.timeBetweenProposals.Seconds()), 10),
		TimeBetweenSyncCommittees: durafmt.Parse(r.timeBetweenSyncCommittees).LimitFirstN(2).String(),
		SecsBetweenSyncCommittees: strconv.FormatInt(int64(r.timeBetweenSyncCommittees.Seconds()), 10),
	}
	return json.Marshal(data)
}

func newCommand(_ context.Context) (*command, error) {
	c := &command{
		quiet:   viper.GetBool("quiet"),
		verbose: viper.GetBool("verbose"),
		debug:   viper.GetBool("debug"),
		json:    viper.GetBool("json"),
		res:     &results{},
	}

	// Timeout.
	if viper.GetDuration("timeout") == 0 {
		return nil, errors.New("timeout is required")
	}
	c.timeout = viper.GetDuration("timeout")

	c.connection = viper.GetString("connection")
	c.allowInsecureConnections = viper.GetBool("allow-insecure-connections")

	c.validators = viper.GetInt64("validators")
	if c.validators < 1 {
		return nil, errors.New("validators must be at least 1")
	}

	return c, nil
}
