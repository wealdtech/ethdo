// Copyright © 2019, 2020 Weald Technology Trading
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

package accountimport

import (
	"context"
	"encoding/hex"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/wealdtech/ethdo/core"
	"github.com/wealdtech/ethdo/util"
	e2wallet "github.com/wealdtech/go-eth2-wallet"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

type dataIn struct {
	timeout          time.Duration
	wallet           e2wtypes.Wallet
	key              []byte
	accountName      string
	passphrase       string
	walletPassphrase string
}

func input(ctx context.Context) (*dataIn, error) {
	var err error
	data := &dataIn{}

	if viper.GetDuration("timeout") == 0 {
		return nil, errors.New("timeout is required")
	}
	data.timeout = viper.GetDuration("timeout")

	// Account name.
	if viper.GetString("account") == "" {
		return nil, errors.New("account is required")
	}
	_, data.accountName, err = e2wallet.WalletAndAccountNames(viper.GetString("account"))
	if err != nil {
		return nil, errors.Wrap(err, "failed to obtain account name")
	}
	if data.accountName == "" {
		return nil, errors.New("account name is required")
	}

	// Wallet.
	ctx, cancel := context.WithTimeout(ctx, data.timeout)
	defer cancel()
	data.wallet, err = core.WalletFromInput(ctx)
	cancel()
	if err != nil {
		return nil, errors.Wrap(err, "failed to obtain wallet")
	}

	// Passphrase.
	data.passphrase, err = util.GetOptionalPassphrase()
	if err != nil {
		return nil, errors.Wrap(err, "failed to obtain passphrase")
	}

	// Wallet passphrase.
	data.walletPassphrase = util.GetWalletPassphrase()

	// Key.
	if viper.GetString("key") == "" {
		return nil, errors.New("key is required")
	}
	data.key, err = hex.DecodeString(strings.TrimPrefix(viper.GetString("key"), "0x"))
	if err != nil {
		return nil, errors.Wrap(err, "key is malformed")
	}

	return data, nil
}
