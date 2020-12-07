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

package depositdata

import (
	"context"
	"fmt"

	spec "github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
	"github.com/wealdtech/ethdo/signing"
	"github.com/wealdtech/ethdo/util"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

func process(data *dataIn) ([]*dataOut, error) {
	if data == nil {
		return nil, errors.New("no data")
	}

	results := make([]*dataOut, 0)

	for _, validatorAccount := range data.validatorAccounts {
		validatorPubKey, err := util.BestPublicKey(validatorAccount)
		if err != nil {
			return nil, errors.Wrap(err, "validator account does not provide a public key")
		}

		var pubKey spec.BLSPubKey
		copy(pubKey[:], validatorPubKey.Marshal())
		depositMessage := &spec.DepositMessage{
			PublicKey:             pubKey,
			WithdrawalCredentials: data.withdrawalCredentials,
			Amount:                data.amount,
		}
		root, err := depositMessage.HashTreeRoot()
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate deposit message root")
		}
		var depositMessageRoot spec.Root
		copy(depositMessageRoot[:], root[:])

		sig, err := signing.SignRoot(context.Background(), validatorAccount, data.passphrases, depositMessageRoot, *data.domain)
		if err != nil {
			return nil, errors.Wrap(err, "failed to sign deposit message")
		}

		depositData := &spec.DepositData{
			PublicKey:             pubKey,
			WithdrawalCredentials: data.withdrawalCredentials,
			Amount:                data.amount,
			Signature:             sig,
		}

		root, err = depositData.HashTreeRoot()
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate deposit data root")
		}
		var depositDataRoot spec.Root
		copy(depositDataRoot[:], root[:])

		validatorWallet := validatorAccount.(e2wtypes.AccountWalletProvider).Wallet()
		results = append(results, &dataOut{
			format:                data.format,
			account:               fmt.Sprintf("%s/%s", validatorWallet.Name(), validatorAccount.Name()),
			validatorPubKey:       &pubKey,
			withdrawalCredentials: data.withdrawalCredentials,
			amount:                data.amount,
			signature:             &sig,
			forkVersion:           data.forkVersion,
			depositMessageRoot:    &depositMessageRoot,
			depositDataRoot:       &depositDataRoot,
		})
	}
	return results, nil
}
