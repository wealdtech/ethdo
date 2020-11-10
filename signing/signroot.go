// Copyright © 2019 Weald Technology Trading
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

package signing

import (
	"context"

	spec "github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
	e2types "github.com/wealdtech/go-eth2-types/v2"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

// SignRoot signs a root with a domain.
func SignRoot(ctx context.Context, account e2wtypes.Account, passphrases []string, root spec.Root, domain spec.Domain) (spec.BLSSignature, error) {
	// Ensure input is as expected.
	if account == nil {
		return spec.BLSSignature{}, errors.New("account not specified")
	}

	alreadyUnlocked, err := Unlock(ctx, account, passphrases)
	if err != nil {
		return spec.BLSSignature{}, err
	}

	var signature e2types.Signature
	// outputIf(debug, fmt.Sprintf("Signing %x (%d)", data, len(data)))
	if protectingSigner, isProtectingSigner := account.(e2wtypes.AccountProtectingSigner); isProtectingSigner {
		// Signer takes root and domain.
		signature, err = signProtected(ctx, protectingSigner, root, domain)
	} else if signer, isSigner := account.(e2wtypes.AccountSigner); isSigner {
		signature, err = sign(ctx, signer, root, domain)
	} else {
		return spec.BLSSignature{}, errors.New("account does not provide signing facility")
	}
	if err != nil {
		return spec.BLSSignature{}, err
	}

	if !alreadyUnlocked {
		if err := Lock(ctx, account); err != nil {
			return spec.BLSSignature{}, errors.Wrap(err, "failed to lock account")
		}
	}

	var sig spec.BLSSignature
	copy(sig[:], signature.Marshal())
	return sig, nil
}

func sign(ctx context.Context, account e2wtypes.AccountSigner, root spec.Root, domain spec.Domain) (e2types.Signature, error) {
	container := &Container{
		Root:   root[:],
		Domain: domain[:],
	}
	signingRoot, err := container.HashTreeRoot()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate hash tree root")
	}

	signature, err := account.Sign(ctx, signingRoot[:])
	if err != nil {
		return nil, errors.Wrap(err, "failed to sign")
	}

	return signature, err
}

func signProtected(ctx context.Context, account e2wtypes.AccountProtectingSigner, root spec.Root, domain spec.Domain) (e2types.Signature, error) {
	signature, err := account.SignGeneric(ctx, root[:], domain[:])
	if err != nil {
		return nil, errors.Wrap(err, "failed to sign")
	}

	return signature, err
}
