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

	"github.com/pkg/errors"
	"github.com/wealdtech/ethdo/util"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

func process(ctx context.Context, data *dataIn) (*dataOut, error) {
	if data == nil {
		return nil, errors.New("no data")
	}
	if data.passphrase == "" {
		return nil, errors.New("passphrase is required")
	}
	if !util.AcceptablePassphrase(data.passphrase) {
		return nil, errors.New("supplied passphrase is weak; use a stronger one or run with the --allow-weak-passphrases flag")
	}
	locker, isLocker := data.wallet.(e2wtypes.WalletLocker)
	if isLocker {
		if err := locker.Unlock(ctx, []byte(data.walletPassphrase)); err != nil {
			return nil, errors.Wrap(err, "failed to unlock wallet")
		}
		defer locker.Lock(ctx)
	}

	results := &dataOut{}

	account, err := data.wallet.(e2wtypes.WalletAccountImporter).ImportAccount(ctx, data.accountName, data.key, []byte(data.passphrase))
	if err != nil {
		return nil, errors.Wrap(err, "failed to import account")
	}
	results.account = account

	return results, nil
}
