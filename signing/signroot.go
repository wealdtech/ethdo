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

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	e2types "github.com/wealdtech/go-eth2-types/v2"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

// SignRoot signs a root with a domain.
func SignRoot(account e2wtypes.Account, root []byte, domain []byte) ([]byte, error) {
	// Ensure input is as expected.
	if account == nil {
		return nil, errors.New("account not specified")
	}
	if len(root) != 32 {
		return nil, errors.New("root must be 32 bytes in length")
	}
	if len(domain) != 32 {
		return nil, errors.New("domain must be 32 bytes in length")
	}

	alreadyUnlocked, err := unlock(account)
	if err != nil {
		return nil, err
	}

	var signature e2types.Signature
	// outputIf(debug, fmt.Sprintf("Signing %x (%d)", data, len(data)))
	if protectingSigner, isProtectingSigner := account.(e2wtypes.AccountProtectingSigner); isProtectingSigner {
		// Signer takes root and domain.
		signature, err = signProtected(protectingSigner, root[:], domain)
	} else if signer, isSigner := account.(e2wtypes.AccountSigner); isSigner {
		signature, err = sign(signer, root[:], domain)
	} else {
		return nil, errors.New("account does not provide signing facility")
	}
	if err != nil {
		return nil, err
	}

	if !alreadyUnlocked {
		if err := lock(account); err != nil {
			return nil, errors.Wrap(err, "failed to lock account")
		}
	}

	return signature.Marshal(), nil
}

func sign(account e2wtypes.AccountSigner, root []byte, domain []byte) (e2types.Signature, error) {
	container := &Container{
		Root:   root,
		Domain: domain,
	}
	signingRoot, err := container.HashTreeRoot()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate hash tree root")
	}

	ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
	defer cancel()
	signature, err := account.Sign(ctx, signingRoot[:])
	if err != nil {
		return nil, errors.Wrap(err, "failed to sign")
	}

	return signature, err
}

func signProtected(account e2wtypes.AccountProtectingSigner, data []byte, domain []byte) (e2types.Signature, error) {
	ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
	defer cancel()
	signature, err := account.SignGeneric(ctx, data, domain)
	if err != nil {
		return nil, errors.Wrap(err, "failed to sign")
	}

	return signature, err
}
