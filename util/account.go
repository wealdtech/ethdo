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
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/wealdtech/go-ecodec"
	util "github.com/wealdtech/go-eth2-util"
	keystorev4 "github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"
	nd "github.com/wealdtech/go-eth2-wallet-nd/v2"
	scratch "github.com/wealdtech/go-eth2-wallet-store-scratch"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
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
	switch {
	case accountStr == "":
		return nil, errors.New("no account specified")
	case strings.HasPrefix(accountStr, "0x"):
		// A key.
		return parseAccountFromKey(ctx, accountStr, unlock)
	case strings.HasPrefix(accountStr, "{"):
		// This could be a keystore.
		return parseAccountFromKeystore(ctx, accountStr, supplementary, unlock)
	case strings.Contains(accountStr, "/"):
		// An account specifier.
		account, err := parseAccountFromSpecifier(ctx, accountStr, supplementary, unlock)
		if err != nil {
			// It is possible that this is actually a path to a keystore, so try that instead.
			if _, statErr := os.Stat(accountStr); statErr == nil {
				account, err = parseAccountFromKeystorePath(ctx, accountStr, supplementary, unlock)
			}
		}
		if err != nil {
			return nil, err
		}
		return account, nil
	case strings.Contains(accountStr, " "):
		// A mnemonic.
		return parseAccountFromMnemonic(ctx, accountStr, supplementary, unlock)
	default:
		// This could be the path to a keystore.
		if _, err := os.Stat(accountStr); err != nil {
			return nil, fmt.Errorf("unknown account specifier %s", accountStr)
		}
		account, err := parseAccountFromKeystorePath(ctx, accountStr, supplementary, unlock)
		if err != nil {
			return nil, err
		}
		return account, nil
	}
}

func parseAccountFromKey(ctx context.Context,
	accountStr string,
	unlock bool,
) (
	e2wtypes.Account,
	error,
) {
	var account e2wtypes.Account
	var err error

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
			return nil, errors.Wrap(err, "failed to create account from private key")
		}
		if unlock {
			_, err = UnlockAccount(ctx, account, nil)
			if err != nil {
				return nil, err
			}
		}
	default:
		return nil, fmt.Errorf("key of length %d neither public nor private key", len(data))
	}

	return account, nil
}

func parseAccountFromSpecifier(ctx context.Context,
	accountStr string,
	supplementary []string,
	unlock bool,
) (
	e2wtypes.Account,
	error,
) {
	var account e2wtypes.Account
	var err error

	_, account, err = WalletAndAccountFromPath(ctx, accountStr)
	if err != nil {
		return nil, errors.Wrap(err, "unable to obtain account")
	}
	if unlock {
		// Supplementary will be the unlock passphrase(s).
		_, err = UnlockAccount(ctx, account, supplementary)
		if err != nil {
			return nil, err
		}
	}

	return account, nil
}

func parseAccountFromMnemonic(ctx context.Context,
	accountStr string,
	supplementary []string,
	unlock bool,
) (
	e2wtypes.Account,
	error,
) {
	var account e2wtypes.Account
	var err error

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

	return account, nil
}

func parseAccountFromKeystore(ctx context.Context,
	accountStr string,
	supplementary []string,
	unlock bool,
) (
	e2wtypes.Account,
	error,
) {
	var account e2wtypes.Account
	var err error

	// Need to import the keystore in to a temporary wallet to fetch the private key.
	store := scratch.New()
	encryptor := keystorev4.New()

	// Need to add a couple of fields to the keystore to make it compliant.
	var keystore map[string]any
	if err := json.Unmarshal([]byte(accountStr), &keystore); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal keystore")
	}
	keystore["name"] = "Import"
	keystore["encryptor"] = "keystore"
	keystoreData, err := json.Marshal(keystore)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal keystore")
	}

	walletData := fmt.Sprintf(`{"wallet":{"name":"Import","type":"non-deterministic","uuid":"e1526407-1dc7-4f3f-9d05-ab696f40707c","version":1},"accounts":[%s]}`, keystoreData)
	encryptedData, err := ecodec.Encrypt([]byte(walletData), []byte(`password`))
	if err != nil {
		return nil, err
	}
	wallet, err := nd.Import(ctx, encryptedData, []byte(`password`), store, encryptor)
	if err != nil {
		return nil, errors.Wrap(err, "failed to import account")
	}

	account = <-wallet.Accounts(ctx)
	if unlock {
		if locker, isLocker := account.(e2wtypes.AccountLocker); isLocker {
			unlocked := false
			for _, passphrase := range supplementary {
				if err = locker.Unlock(ctx, []byte(passphrase)); err == nil {
					unlocked = true
					break
				}
			}
			if !unlocked {
				return nil, errors.New("failed to unlock account")
			}
		}
	}

	return account, nil
}

func parseAccountFromKeystorePath(ctx context.Context,
	accountStr string,
	supplementary []string,
	unlock bool,
) (
	e2wtypes.Account,
	error,
) {
	data, err := os.ReadFile(accountStr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read keystore file")
	}
	return parseAccountFromKeystore(ctx, string(data), supplementary, unlock)
}

func accountFromMnemonicAndPath(mnemonic string, path string) (e2wtypes.Account, error) {
	seed, err := SeedFromMnemonic(mnemonic)
	if err != nil {
		return nil, err
	}

	// Ensure the path is valid.
	match := hdPathRegex.MatchString(path)
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
