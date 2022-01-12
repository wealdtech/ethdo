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

package inclusion

import (
	"context"
	"time"

	eth2client "github.com/attestantio/go-eth2-client"
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

	// Input.
	account string
	pubKey  string
	index   string
	epoch   int64

	// Data access.
	eth2Client eth2client.Service
	chainTime  chaintime.Service

	// Output.
	inCommittee    bool
	committeeIndex uint64
	inclusions     []int
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

	if viper.GetString("connection") == "" {
		return nil, errors.New("connection is required")
	}
	c.connection = viper.GetString("connection")
	c.allowInsecureConnections = viper.GetBool("allow-insecure-connections")

	// Validator.
	c.account = viper.GetString("account")
	c.pubKey = viper.GetString("pubkey")
	c.index = viper.GetString("index")
	if c.account == "" && c.pubKey == "" && c.index == "" {
		return nil, errors.New("account, pubkey or index required")
	}

	// Epoch.
	c.epoch = viper.GetInt64("epoch")

	return c, nil
}
