// Copyright © 2021 Weald Technology Trading
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

package walletsharedimport

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/wealdtech/ethdo/shamir"
	e2wallet "github.com/wealdtech/go-eth2-wallet"
)

type sharedExport struct {
	Version      uint32 `json:"version"`
	Participants uint32 `json:"participants"`
	Threshold    uint32 `json:"threshold"`
	Data         string `json:"data"`
}

func process(_ context.Context, data *dataIn) (*dataOut, error) {
	if data == nil {
		return nil, errors.New("no data")
	}
	if len(data.file) == 0 {
		return nil, errors.New("import file is required")
	}

	sharedExport := &sharedExport{}
	err := json.Unmarshal(data.file, sharedExport)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal export")
	}

	if len(data.shares) != int(sharedExport.Threshold) {
		return nil, fmt.Errorf("import requires %d shares, %d were provided", sharedExport.Threshold, len(data.shares))
	}

	shares := make([][]byte, len(data.shares))
	for i := range data.shares {
		shares[i], err = hex.DecodeString(data.shares[i])
		if err != nil {
			return nil, errors.Wrap(err, "invalid share")
		}
	}
	passphrase, err := shamir.Combine(shares)
	if err != nil {
		return nil, errors.Wrap(err, "failed to recreate passphrase from shares")
	}
	wallet, err := hex.DecodeString(strings.TrimPrefix(sharedExport.Data, "0x"))
	if err != nil {
		return nil, errors.Wrap(err, "failed to obtain data from export")
	}
	if _, err := e2wallet.ImportWallet(wallet, passphrase); err != nil {
		return nil, errors.Wrap(err, "failed to import wallet")
	}

	return nil, nil
}
