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

package walletsharedexport

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/wealdtech/ethdo/shamir"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

type sharedExport struct {
	Version      uint32 `json:"version"`
	Participants uint32 `json:"participants"`
	Threshold    uint32 `json:"threshold"`
	Data         string `json:"data"`
}

func process(ctx context.Context, data *dataIn) (*dataOut, error) {
	if data == nil {
		return nil, errors.New("no data")
	}
	if data.wallet == nil {
		return nil, errors.New("wallet is required")
	}

	passphrase := make([]byte, 64)
	n, err := rand.Read(passphrase)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate passphrase")
	}
	if n != 64 {
		return nil, errors.New("failed to obtain passphrase")
	}
	exporter, isExporter := data.wallet.(e2wtypes.WalletExporter)
	if !isExporter {
		return nil, errors.New("wallet does not provide export")
	}

	export, err := exporter.Export(ctx, passphrase)
	if err != nil {
		return nil, errors.Wrap(err, "failed to export wallet")
	}

	shares, err := shamir.Split(passphrase, int(data.participants), int(data.threshold))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create shamir shares")
	}

	sharedExport := &sharedExport{
		Version:      1,
		Participants: data.participants,
		Threshold:    data.threshold,
		Data:         fmt.Sprintf("%#x", export),
	}
	sharedFile, err := json.Marshal(sharedExport)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal shamir export")
	}

	if err := os.WriteFile(data.file, sharedFile, 0o600); err != nil {
		return nil, errors.Wrap(err, "failed to write export file")
	}

	results := &dataOut{
		shares: shares,
	}

	return results, nil
}
