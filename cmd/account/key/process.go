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

package accountkey

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
	if len(data.passphrases) == 0 {
		return nil, errors.New("passphrase is required")
	}

	results := &dataOut{}

	privateKeyProvider, isPrivateKeyProvider := data.account.(e2wtypes.AccountPrivateKeyProvider)
	if !isPrivateKeyProvider {
		return nil, errors.New("account does not provide its private key")
	}

	if locker, isLocker := data.account.(e2wtypes.AccountLocker); isLocker {
		unlocked, err := locker.IsUnlocked(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "failed to find out if account is locked")
		}
		if !unlocked {
			for _, passphrase := range data.passphrases {
				err = locker.Unlock(ctx, []byte(passphrase))
				if err == nil {
					unlocked = true
					break
				}
			}
			if !unlocked {
				return nil, errors.New("failed to unlock account")
			}
			// Because we unlocked the accout we should re-lock it when we're done.
			defer func() {
				if err := locker.Lock(ctx); err != nil {
					util.Log.Trace().Err(err).Msg("Failed to lock account")
				}
			}()
		}
	}
	key, err := privateKeyProvider.PrivateKey(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to obtain private key")
	}
	results.key = key.Marshal()

	return results, nil
}
