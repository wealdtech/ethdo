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

package walletcreate

import (
	"context"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/wealdtech/ethdo/util"
	e2wallet "github.com/wealdtech/go-eth2-wallet"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

type dataIn struct {
	// System.
	timeout time.Duration
	quiet   bool
	verbose bool
	debug   bool
	// For all wallets.
	store      e2wtypes.Store
	walletType string
	walletName string
	// For HD wallets.
	passphrase string
	mnemonic   string
}

func input(_ context.Context) (*dataIn, error) {
	var err error
	data := &dataIn{}

	if viper.GetString("remote") != "" {
		return nil, errors.New("cannot create remote wallets")
	}

	if viper.GetDuration("timeout") == 0 {
		return nil, errors.New("timeout is required")
	}
	data.timeout = viper.GetDuration("timeout")
	data.quiet = viper.GetBool("quiet")
	data.verbose = viper.GetBool("verbose")
	data.debug = viper.GetBool("debug")

	store, isStore := viper.Get("store").(e2wtypes.Store)
	if !isStore {
		return nil, errors.New("store is required")
	}
	data.store = store

	// Wallet name.
	if viper.GetString("wallet") == "" {
		return nil, errors.New("wallet is required")
	}
	data.walletName, _, err = e2wallet.WalletAndAccountNames(viper.GetString("wallet"))
	if err != nil {
		return nil, errors.Wrap(err, "failed to obtain wallet name")
	}
	if data.walletName == "" {
		return nil, errors.New("wallet name is required")
	}

	// Type.
	data.walletType = strings.ToLower(viper.GetString("type"))
	if data.walletType == "" {
		return nil, errors.New("wallet type is required")
	}

	// Passphrase.
	data.passphrase = util.GetWalletPassphrase()

	// Mnemonic.
	data.mnemonic = viper.GetString("mnemonic")

	return data, nil
}
