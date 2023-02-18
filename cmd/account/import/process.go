// Copyright © 2019 -2022 Weald Technology Trading
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
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/wealdtech/ethdo/util"
	"github.com/wealdtech/go-ecodec"
	keystorev4 "github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"
	nd "github.com/wealdtech/go-eth2-wallet-nd/v2"
	scratch "github.com/wealdtech/go-eth2-wallet-store-scratch"
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
		defer func() {
			if err := locker.Lock(ctx); err != nil {
				util.Log.Trace().Err(err).Msg("Failed to lock wallet")
			}
		}()
	}

	if len(data.key) > 0 {
		return processFromKey(ctx, data)
	}
	if len(data.keystore) > 0 {
		return processFromKeystore(ctx, data)
	}
	return nil, errors.New("unsupported import mechanism")
}

func processFromKey(ctx context.Context, data *dataIn) (*dataOut, error) {
	results := &dataOut{}

	importer, isImporter := data.wallet.(e2wtypes.WalletAccountImporter)
	if !isImporter {
		return nil, fmt.Errorf("%s wallets do not support importing accounts", data.wallet.Type())
	}
	account, err := importer.ImportAccount(ctx, data.accountName, data.key, []byte(data.passphrase))
	if err != nil {
		return nil, errors.Wrap(err, "failed to import wallet")
	}
	results.account = account

	return results, nil
}

func processFromKeystore(ctx context.Context, data *dataIn) (*dataOut, error) {
	// Need to import the keystore in to a temporary wallet to fetch the private key.
	store := scratch.New()
	encryptor := keystorev4.New()

	// Need to add a couple of fields to the keystore to make it compliant.
	var keystore map[string]any
	if err := json.Unmarshal(data.keystore, &keystore); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal keystore")
	}
	keystore["name"] = data.accountName
	keystore["encryptor"] = "keystore"
	keystoreData, err := json.Marshal(keystore)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal keystore")
	}

	walletData := fmt.Sprintf(`{"wallet":{"name":"Import","type":"non-deterministic","uuid":"e1526407-1dc7-4f3f-9d05-ab696f40707c","version":1},"accounts":[%s]}`, keystoreData)
	encryptedData, err := ecodec.Encrypt([]byte(walletData), data.keystorePassphrase)
	if err != nil {
		return nil, err
	}
	wallet, err := nd.Import(ctx, encryptedData, data.keystorePassphrase, store, encryptor)
	if err != nil {
		return nil, errors.Wrap(err, "failed to import account")
	}

	account := <-wallet.Accounts(ctx)
	privateKeyProvider, isPrivateKeyProvider := account.(e2wtypes.AccountPrivateKeyProvider)
	if !isPrivateKeyProvider {
		return nil, errors.New("account does not provide its private key")
	}
	if locker, isLocker := account.(e2wtypes.AccountLocker); isLocker {
		if err = locker.Unlock(ctx, data.keystorePassphrase); err != nil {
			return nil, errors.Wrap(err, "failed to unlock account")
		}
	}
	key, err := privateKeyProvider.PrivateKey(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to obtain private key")
	}
	data.key = key.Marshal()
	// We have the key from the keystore; import it.
	return processFromKey(ctx, data)
}
