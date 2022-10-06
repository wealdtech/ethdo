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
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/tyler-smith/go-bip39"
	util "github.com/wealdtech/go-eth2-util"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
	"golang.org/x/text/unicode/norm"
)

// ParseAccount parses input to obtain an account.
func ParseAccount(ctx context.Context,
	accountStr string,
	supplementary []string,
	unlock bool,
) (
	e2wtypes.Account,
	error,
) {
	if accountStr == "" {
		return nil, errors.New("no account specified")
	}

	var account e2wtypes.Account
	var err error

	switch {
	case strings.HasPrefix(accountStr, "0x"):
		// A key.  Could be public or private.
		data, err := hex.DecodeString(strings.TrimPrefix(accountStr, "0x"))
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse account key")
		}
		switch len(data) {
		case 48:
			// Public key.
			account, err = newScratchAccountFromPubKey(data)
			if err != nil {
				return nil, errors.Wrap(err, "failed to create account from public key")
			}
			if unlock {
				return nil, errors.New("cannot unlock an account specified by its public key")
			}
		case 32:
			// Private key.
			account, err = newScratchAccountFromPrivKey(data)
			if err != nil {
				return nil, errors.Wrap(err, "failed to create account from public key")
			}
			if unlock {
				_, err = UnlockAccount(ctx, account, nil)
				if err != nil {
					return nil, errors.Wrap(err, "failed to unlock account")
				}
			}
		default:
			return nil, fmt.Errorf("key of length %d neither public nor private key", len(data))
		}
	case strings.Contains(accountStr, "/"):
		// An account.
		_, account, err = WalletAndAccountFromPath(ctx, accountStr)
		if err != nil {
			return nil, errors.Wrap(err, "unable to obtain account")
		}
		if unlock {
			// Supplementary will be the unlock passphrase(s).
			_, err = UnlockAccount(ctx, account, supplementary)
			if err != nil {
				return nil, errors.Wrap(err, "failed to unlock account")
			}
		}
	case strings.Contains(accountStr, " "):
		// A mnemonic.
		// Supplementary will be the path.
		if len(supplementary) == 0 {
			return nil, errors.New("missing derivation path")
		}
		account, err = accountFromMnemonicAndPath(accountStr, supplementary[0])
		if err != nil {
			return nil, err
		}
		if unlock {
			err = account.(e2wtypes.AccountLocker).Unlock(ctx, nil)
			if err != nil {
				return nil, errors.Wrap(err, "failed to unlock account")
			}
		}
	default:
		return nil, fmt.Errorf("unknown account specifier %s", accountStr)
	}

	return account, nil
}

// hdPathRegex is the regular expression that matches an HD path.
var hdPathRegex = regexp.MustCompile("^m/[0-9]+/[0-9]+(/[0-9+])+")

func accountFromMnemonicAndPath(mnemonic string, path string) (e2wtypes.Account, error) {
	// If there are more than 24 words we treat the additional characters as the passphrase.
	mnemonicParts := strings.Split(mnemonic, " ")
	mnemonicPassphrase := ""
	if len(mnemonicParts) > 24 {
		mnemonic = strings.Join(mnemonicParts[:24], " ")
		mnemonicPassphrase = strings.Join(mnemonicParts[24:], " ")
	}
	// Normalise the input.
	mnemonic = string(norm.NFKD.Bytes([]byte(mnemonic)))
	mnemonicPassphrase = string(norm.NFKD.Bytes([]byte(mnemonicPassphrase)))

	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, errors.New("mnemonic is invalid")
	}

	// Create seed from mnemonic and passphrase.
	seed := bip39.NewSeed(mnemonic, mnemonicPassphrase)

	// Ensure the path is valid.
	match := hdPathRegex.Match([]byte(path))
	if !match {
		return nil, errors.New("path does not match expected format m/…")
	}

	// Derive private key from seed and path.
	key, err := util.PrivateKeyFromSeedAndPath(seed, path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate key")
	}

	// Create a scratch account given the private key.
	account, err := newScratchAccountFromPrivKey(key.Marshal())
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate scratch account")
	}

	return account, nil
}

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

	// Also attempt to unlock without any passphrase.
	err = locker.Unlock(ctx, nil)
	if err == nil {
		// Unlocked.
		return false, nil
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
