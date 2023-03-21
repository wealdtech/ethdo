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

package walletdelete

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

func process(_ context.Context, data *dataIn) (*dataOut, error) {
	if data == nil {
		return nil, errors.New("no data")
	}
	if data.wallet == nil {
		return nil, errors.New("wallet is required")
	}

	storeProvider, isProvider := data.wallet.(e2wtypes.StoreProvider)
	if !isProvider {
		return nil, errors.New("cannot obtain store for the wallet")
	}
	store := storeProvider.Store()

	if store.Name() != "filesystem" {
		return nil, fmt.Errorf("cannot delete %s wallet automatically, please remove manually", store.Name())
	}
	storeLocationProvider, isProvider := store.(e2wtypes.StoreLocationProvider)
	if !isProvider {
		return nil, errors.New("cannot obtain store location for the wallet")
	}
	walletLocation := filepath.Join(storeLocationProvider.Location(), data.wallet.ID().String())
	if err := os.RemoveAll(walletLocation); err != nil {
		return nil, errors.Wrap(err, "failed to delete wallet")
	}

	return &dataOut{}, nil
}
