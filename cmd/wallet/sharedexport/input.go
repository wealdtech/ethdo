// Copyright © 2021 Weald Technology Trading
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

package walletsharedexport

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/wealdtech/ethdo/util"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

type dataIn struct {
	// System.
	timeout      time.Duration
	verbose      bool
	debug        bool
	wallet       e2wtypes.Wallet
	file         string
	participants uint32
	threshold    uint32
}

func input(ctx context.Context) (*dataIn, error) {
	var err error
	data := &dataIn{}

	if viper.GetString("remote") != "" {
		return nil, errors.New("wallet export not available for remote wallets")
	}

	if viper.GetDuration("timeout") == 0 {
		return nil, errors.New("timeout is required")
	}
	data.timeout = viper.GetDuration("timeout")
	// Quiet is not allowed.
	if viper.GetBool("quiet") {
		return nil, errors.New("quiet not allowed")
	}
	data.verbose = viper.GetBool("verbose")
	data.debug = viper.GetBool("debug")

	// Wallet.
	wallet, err := util.WalletFromInput(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to access wallet")
	}
	data.wallet = wallet

	// File.
	data.file = viper.GetString("file")
	if data.file == "" {
		return nil, errors.New("file is required")
	}

	// Participants
	data.participants = viper.GetUint32("participants")
	if data.participants == 0 {
		return nil, errors.New("participants is required")
	}
	data.threshold = viper.GetUint32("threshold")
	if data.threshold == 0 {
		return nil, errors.New("threshold is required")
	}
	if data.threshold > data.participants {
		return nil, errors.New("threshold cannot be more than participants")
	}

	return data, nil
}
