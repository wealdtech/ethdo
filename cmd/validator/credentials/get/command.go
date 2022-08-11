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

package validatorcredentialsget

import (
	"context"
	"time"

	eth2client "github.com/attestantio/go-eth2-client"
	apiv1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type command struct {
	quiet   bool
	verbose bool
	debug   bool

	// Input.
	account string
	index   string
	pubKey  string

	// Beacon node connection.
	timeout                  time.Duration
	connection               string
	allowInsecureConnections bool

	// Data access.
	consensusClient    eth2client.Service
	validatorsProvider eth2client.ValidatorsProvider

	// Output.
	validator *apiv1.Validator
}

func newCommand(ctx context.Context) (*command, error) {
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

	c.connection = viper.GetString("connection")
	c.allowInsecureConnections = viper.GetBool("allow-insecure-connections")

	c.account = viper.GetString("account")
	c.index = viper.GetString("index")
	c.pubKey = viper.GetString("pubkey")
	nonNil := 0
	if c.account != "" {
		nonNil++
	}
	if c.index != "" {
		nonNil++
	}
	if c.pubKey != "" {
		nonNil++
	}
	if nonNil == 0 {
		return nil, errors.New("one of account, index or pubkey required")
	}
	if nonNil > 1 {
		return nil, errors.New("only one of account, index and pubkey allowed")
	}

	return c, nil
}
