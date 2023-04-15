// Copyright © 2023 Weald Technology Trading.
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

package validatorwithdrawl

import (
	"context"
	"encoding/json"
	"time"

	consensusclient "github.com/attestantio/go-eth2-client"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/wealdtech/ethdo/services/chaintime"
)

type command struct {
	quiet   bool
	verbose bool
	debug   bool
	offline bool
	json    bool

	// Input.
	validator string

	// Beacon node connection.
	timeout                  time.Duration
	connection               string
	allowInsecureConnections bool

	// Processing.
	consensusClient          consensusclient.Service
	chainTime                chaintime.Service
	maxWithdrawalsPerPayload uint64
	maxEffectiveBalance      phase0.Gwei

	// Output.
	res *res
}

func newCommand(_ context.Context) (*command, error) {
	c := &command{
		quiet:                    viper.GetBool("quiet"),
		verbose:                  viper.GetBool("verbose"),
		debug:                    viper.GetBool("debug"),
		offline:                  viper.GetBool("offline"),
		json:                     viper.GetBool("json"),
		timeout:                  viper.GetDuration("timeout"),
		connection:               viper.GetString("connection"),
		allowInsecureConnections: viper.GetBool("allow-insecure-connections"),
		validator:                viper.GetString("validator"),
		res:                      &res{},
	}

	// Timeout is required.
	if c.timeout == 0 {
		return nil, errors.New("timeout is required")
	}

	if c.validator == "" {
		return nil, errors.New("validator is required")
	}

	return c, nil
}

type res struct {
	WithdrawalsToGo uint64
	BlocksToGo      uint64
	Block           uint64
	Wait            time.Duration
	Expected        time.Time
}

type resJSON struct {
	WithdrawalsToGo   uint64 `json:"withdrawals_to_go"`
	BlocksToGo        uint64 `json:"blocks_to_go"`
	Block             uint64 `json:"block"`
	Wait              string `json:"wait"`
	WaitSecs          uint64 `json:"wait_secs"`
	Expected          string `json:"expected"`
	ExpectedTimestamp int64  `json:"expected_timestamp"`
}

func (r *res) MarshalJSON() ([]byte, error) {
	data := resJSON{
		WithdrawalsToGo:   r.WithdrawalsToGo,
		BlocksToGo:        r.BlocksToGo,
		Block:             r.Block,
		Wait:              r.Wait.Round(time.Second).String(),
		WaitSecs:          uint64(r.Wait.Round(time.Second).Seconds()),
		Expected:          r.Expected.Format("2006-01-02T15:04:05"),
		ExpectedTimestamp: r.Expected.Unix(),
	}
	return json.Marshal(data)
}
