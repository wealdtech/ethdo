// Copyright © 2020 Weald Technology Trading
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

package validatorduties

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type dataIn struct {
	// System.
	timeout time.Duration
	quiet   bool
	verbose bool
	debug   bool
	// Ethereum 2 connection.
	eth2Client    string
	allowInsecure bool
	// Operation.
	account string
	pubKey  string
	index   string
}

func input(_ context.Context) (*dataIn, error) {
	data := &dataIn{}

	if viper.GetDuration("timeout") == 0 {
		return nil, errors.New("timeout is required")
	}
	data.timeout = viper.GetDuration("timeout")
	data.quiet = viper.GetBool("quiet")
	data.verbose = viper.GetBool("verbose")
	data.debug = viper.GetBool("debug")

	// Ethereum 2 connection.
	data.eth2Client = viper.GetString("connection")
	data.allowInsecure = viper.GetBool("allow-insecure-connections")

	// Account.
	data.account = viper.GetString("account")

	// PubKey.
	data.pubKey = viper.GetString("pubkey")

	// ID.
	data.index = viper.GetString("index")

	if data.account == "" && data.pubKey == "" && data.index == "" {
		return nil, errors.New("account, pubkey or index required")
	}

	return data, nil
}
