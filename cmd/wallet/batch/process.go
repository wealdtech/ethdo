// Copyright © 2023 Weald Technology Trading.
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

package walletbatch

import (
	"context"

	"github.com/pkg/errors"
	"github.com/wealdtech/ethdo/util"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

func (c *command) process(ctx context.Context) error {
	// Obtain the wallet.
	opCtx, cancel := context.WithTimeout(ctx, c.timeout)
	wallet, err := util.WalletFromInput(opCtx)
	cancel()
	if err != nil {
		return errors.Wrap(err, "failed to obtain wallet")
	}
	batchCreator, isBatchCreator := wallet.(e2wtypes.WalletBatchCreator)
	if !isBatchCreator {
		return errors.New("wallet does not support batching")
	}

	// Create the batch.
	if err := batchCreator.BatchWallet(ctx, util.GetPassphrases(), c.batchPassphrase); err != nil {
		return errors.Wrap(err, "failed to batch wallet")
	}

	return nil
}
