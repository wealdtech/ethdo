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

package accountcreate

import (
	"context"
	"regexp"

	"github.com/pkg/errors"
	"github.com/wealdtech/ethdo/util"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

func process(ctx context.Context, data *dataIn) (*dataOut, error) {
	if data == nil {
		return nil, errors.New("no data")
	}
	if data.passphrase != "" && !util.AcceptablePassphrase(data.passphrase) {
		return nil, errors.New("supplied passphrase is weak; use a stronger one or run with the --allow-weak-passphrases flag")
	}
	locker, isLocker := data.wallet.(e2wtypes.WalletLocker)
	if isLocker {
		if err := locker.Unlock(ctx, []byte(data.walletPassphrase)); err != nil {
			return nil, errors.Wrap(err, "failed to unlock wallet")
		}
		defer func() {
			if err := locker.Lock(ctx); err != nil {
				util.Log.Trace().Err(err).Msg("Failed to lock wallet")
			}
		}()
	}
	if data.participants == 0 {
		return nil, errors.New("participants is required")
	}

	// Create style of account based on input.
	switch {
	case data.participants > 1:
		return processDistributed(ctx, data)
	case data.path != "":
		return processPathed(ctx, data)
	default:
		return processStandard(ctx, data)
	}
}

func processStandard(ctx context.Context, data *dataIn) (*dataOut, error) {
	if data == nil {
		return nil, errors.New("no data")
	}
	if data.passphrase == "" {
		return nil, errors.New("passphrase is required")
	}

	results := &dataOut{}

	creator, isCreator := data.wallet.(e2wtypes.WalletAccountCreator)
	if !isCreator {
		return nil, errors.New("wallet does not support account creation")
	}
	ctx, cancel := context.WithTimeout(ctx, data.timeout)
	defer cancel()
	account, err := creator.CreateAccount(ctx, data.accountName, []byte(data.passphrase))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create account")
	}
	results.account = account
	return results, nil
}

func processPathed(ctx context.Context, data *dataIn) (*dataOut, error) {
	if data == nil {
		return nil, errors.New("no data")
	}
	if data.passphrase == "" {
		return nil, errors.New("passphrase is required")
	}
	match, err := regexp.MatchString("^m/[0-9]+/[0-9]+(/[0-9+])+", data.path)
	if err != nil {
		return nil, errors.Wrap(err, "unable to match path to regular expression")
	}
	if !match {
		return nil, errors.New("path does not match expected format m/…")
	}

	results := &dataOut{}

	creator, isCreator := data.wallet.(e2wtypes.WalletPathedAccountCreator)
	if !isCreator {
		return nil, errors.New("wallet does not support account creation with an explicit path")
	}

	ctx, cancel := context.WithTimeout(ctx, data.timeout)
	defer cancel()
	account, err := creator.CreatePathedAccount(ctx, data.path, data.accountName, []byte(data.passphrase))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create account")
	}
	results.account = account
	return results, nil
}

func processDistributed(ctx context.Context, data *dataIn) (*dataOut, error) {
	if data == nil {
		return nil, errors.New("no data")
	}
	if data.signingThreshold == 0 {
		return nil, errors.New("signing threshold required")
	}
	if data.signingThreshold <= data.participants/2 {
		return nil, errors.New("signing threshold must be more than half the number of participants")
	}
	if data.signingThreshold > data.participants {
		return nil, errors.New("signing threshold cannot be higher than the number of participants")
	}

	results := &dataOut{}

	creator, isCreator := data.wallet.(e2wtypes.WalletDistributedAccountCreator)
	if !isCreator {
		return nil, errors.New("wallet does not support distributed account creation")
	}

	ctx, cancel := context.WithTimeout(ctx, data.timeout)
	defer cancel()
	account, err := creator.CreateDistributedAccount(ctx,
		data.accountName,
		data.participants,
		data.signingThreshold,
		[]byte(data.passphrase))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create account")
	}
	results.account = account
	return results, nil
}
