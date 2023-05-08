// Copyright © 2020, 2023 Weald Technology Trading
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

package accountderive

import (
	"context"

	"github.com/pkg/errors"
	"github.com/wealdtech/ethdo/util"
	e2types "github.com/wealdtech/go-eth2-types/v2"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

func process(ctx context.Context, data *dataIn) (*dataOut, error) {
	if data == nil {
		return nil, errors.New("no data")
	}

	account, err := util.ParseAccount(ctx, data.mnemonic, []string{data.path}, true)
	if err != nil {
		return nil, errors.Wrap(err, "failed to derive account")
	}

	key, err := account.(e2wtypes.AccountPrivateKeyProvider).PrivateKey(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to obtain account private key")
	}

	results := &dataOut{
		showPrivateKey:            data.showPrivateKey,
		showWithdrawalCredentials: data.showWithdrawalCredentials,
		generateKeystore:          data.generateKeystore,
		key:                       key.(*e2types.BLSPrivateKey),
		path:                      data.path,
	}

	return results, nil
}
