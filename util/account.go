// Copyright © 2020, 2022 Weald Technology Trading
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

package util

import (
	"context"

	"github.com/pkg/errors"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

// UnlockAccount attempts to unlock an account.  It returns true if the account was already unlocked.
func UnlockAccount(ctx context.Context, account e2wtypes.Account, passphrases []string) (bool, error) {
	locker, isAccountLocker := account.(e2wtypes.AccountLocker)
	if !isAccountLocker {
		// This account doesn't support unlocking; return okay.
		return true, nil
	}

	alreadyUnlocked, err := locker.IsUnlocked(ctx)
	if err != nil {
		return false, errors.Wrap(err, "unable to ascertain if account is unlocked")
	}

	if alreadyUnlocked {
		return true, nil
	}

	// Not already unlocked; attempt to unlock it.
	for _, passphrase := range passphrases {
		err = locker.Unlock(ctx, []byte(passphrase))
		if err == nil {
			// Unlocked.
			return false, nil
		}
	}

	// Failed to unlock it.
	return false, errors.New("failed to unlock account")
}

// LockAccount attempts to lock an account.
func LockAccount(ctx context.Context, account e2wtypes.Account) error {
	locker, isAccountLocker := account.(e2wtypes.AccountLocker)
	if !isAccountLocker {
		// This account doesn't support locking; return okay.
		return nil
	}

	return locker.Lock(ctx)
}
