// Copyright © 2019 - 2022 Weald Technology Trading.
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
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/wealdtech/ethdo/util"
	e2wallet "github.com/wealdtech/go-eth2-wallet"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

type dataIn struct {
	timeout            time.Duration
	wallet             e2wtypes.Wallet
	key                []byte
	accountName        string
	passphrase         string
	walletPassphrase   string
	keystore           []byte
	keystorePassphrase []byte
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
	data.wallet, err = util.WalletFromInput(ctx)
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

	if viper.GetString("key") == "" && viper.GetString("keystore") == "" {
		return nil, errors.New("key or keystore is required")
	}
	if viper.GetString("key") != "" && viper.GetString("keystore") != "" {
		return nil, errors.New("only one of key and keystore is required")
	}

	if viper.GetString("key") != "" {
		data.key, err = hex.DecodeString(strings.TrimPrefix(viper.GetString("key"), "0x"))
		if err != nil {
			return nil, errors.Wrap(err, "key is malformed")
		}
	}

	if viper.GetString("keystore") != "" {
		data.keystorePassphrase = []byte(viper.GetString("keystore-passphrase"))
		if len(data.keystorePassphrase) == 0 {
			return nil, errors.New("must supply keystore passphrase with keystore-passphrase when supplying keystore")
		}
		data.keystore, err = obtainKeystore(viper.GetString("keystore"))
		if err != nil {
			return nil, errors.Wrap(err, "invalid keystore")
		}
	}

	return data, nil
}

// obtainKeystore obtains keystore from an input, could be JSON itself or a path to JSON.
func obtainKeystore(input string) ([]byte, error) {
	var err error
	var data []byte
	// Input could be JSON or a path to JSON
	if strings.HasPrefix(input, "{") {
		// Looks like JSON
		data = []byte(input)
	} else {
		// Assume it's a path to JSON
		data, err = os.ReadFile(input)
		if err != nil {
			return nil, errors.Wrap(err, "failed to find deposit data file")
		}
	}
	return data, nil
	//	exitData := &util.ValidatorExitData{}
	//	err = json.Unmarshal(data, exitData)
	//	if err != nil {
	//		return nil, errors.Wrap(err, "data is not valid JSON")
	//	}

	//	return exitData, nil
}
