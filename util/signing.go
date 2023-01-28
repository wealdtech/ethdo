// Copyright © 2020 Weald Technology Trading
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

	spec "github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
	"github.com/prysmaticlabs/go-ssz"
	"github.com/spf13/viper"
	e2types "github.com/wealdtech/go-eth2-types/v2"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

// SignRoot signs the hash tree root of a data structure.
func SignRoot(account e2wtypes.Account, root spec.Root, domain spec.Domain) (e2types.Signature, error) {
	if _, isProtectingSigner := account.(e2wtypes.AccountProtectingSigner); isProtectingSigner {
		// Signer builds the signing data.
		return signGeneric(account, root, domain)
	}

	// Build the signing data manually.
	container := &spec.SigningData{
		ObjectRoot: root,
		Domain:     domain,
	}
	// outputIf(debug, fmt.Sprintf("Signing container:\n root: %#x\n domain: %#x", container.ObjectRoot, container.Domain))
	signingRoot, err := container.HashTreeRoot()
	if err != nil {
		return nil, err
	}
	// outputIf(debug, fmt.Sprintf("Signing root: %#x", signingRoot))
	return sign(account, signingRoot[:])
}

// VerifyRoot verifies the hash tree root of a data structure.
func VerifyRoot(account e2wtypes.Account, root spec.Root, domain spec.Domain, signature e2types.Signature) (bool, error) {
	// Build the signing data manually.
	container := &spec.SigningData{
		ObjectRoot: root,
		Domain:     domain,
	}
	// outputIf(debug, fmt.Sprintf("Signing container:\n root: %#x\n domain: %#x", container.ObjectRoot, container.Domain))
	signingRoot, err := ssz.HashTreeRoot(container)
	if err != nil {
		return false, err
	}
	// outputIf(debug, fmt.Sprintf("Signing root: %#x", signingRoot))
	pubKey, err := BestPublicKey(account)
	if err != nil {
		return false, errors.Wrap(err, "failed to obtain account public key")
	}
	return signature.Verify(signingRoot[:], pubKey), nil
}

// signGeneric signs generic data.
func signGeneric(account e2wtypes.Account, data spec.Root, domain spec.Domain) (e2types.Signature, error) {
	alreadyUnlocked, err := unlock(account)
	if err != nil {
		return nil, err
	}
	// outputIf(debug, fmt.Sprintf("Signing %x (%d)", data, len(data)))
	ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
	defer cancel()

	signer, isProtectingSigner := account.(e2wtypes.AccountProtectingSigner)
	if !isProtectingSigner {
		return nil, errors.New("account does not provide generic signing")
	}

	signature, err := signer.SignGeneric(ctx, data[:], domain[:])
	// errCheck(err, "failed to sign")
	if !alreadyUnlocked {
		if err := lock(account); err != nil {
			return nil, errors.Wrap(err, "failed to lock account")
		}
	}
	return signature, err
}

// sign signs arbitrary data, handling unlocking and locking as required.
func sign(account e2wtypes.Account, data []byte) (e2types.Signature, error) {
	alreadyUnlocked, err := unlock(account)
	if err != nil {
		return nil, err
	}
	// outputIf(debug, fmt.Sprintf("Signing %x (%d)", data, len(data)))
	ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
	defer cancel()

	signer, isSigner := account.(e2wtypes.AccountSigner)
	if !isSigner {
		return nil, errors.New("account does not provide signing")
	}

	signature, err := signer.Sign(ctx, data)
	// errCheck(err, "failed to sign")
	if !alreadyUnlocked {
		if err := lock(account); err != nil {
			return nil, errors.Wrap(err, "failed to lock account")
		}
	}
	return signature, err
}

// unlock attempts to unlock an account.  It returns true if the account was already unlocked.
func unlock(account e2wtypes.Account) (bool, error) {
	locker, isAccountLocker := account.(e2wtypes.AccountLocker)
	if !isAccountLocker {
		// outputIf(debug, "Account does not support unlocking")
		// This account doesn't support unlocking; return okay.
		return true, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
	alreadyUnlocked, err := locker.IsUnlocked(ctx)
	cancel()
	if err != nil {
		return false, errors.Wrap(err, "unable to ascertain if account is unlocked")
	}

	if alreadyUnlocked {
		return true, nil
	}

	// Not already unlocked; attempt to unlock it.
	for _, passphrase := range GetPassphrases() {
		ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
		err = locker.Unlock(ctx, []byte(passphrase))
		cancel()
		if err == nil {
			// Unlocked.
			return false, nil
		}
	}

	// Failed to unlock it.
	return false, errors.New("failed to unlock account")
}

// lock attempts to lock an account.
func lock(account e2wtypes.Account) error {
	locker, isAccountLocker := account.(e2wtypes.AccountLocker)
	if !isAccountLocker {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
	defer cancel()
	return locker.Lock(ctx)
}
