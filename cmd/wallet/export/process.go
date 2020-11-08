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

package walletexport

import (
	"context"

	"github.com/pkg/errors"
	"github.com/wealdtech/ethdo/util"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

func process(ctx context.Context, data *dataIn) (*dataOut, error) {
	if data == nil {
		return nil, errors.New("no data")
	}
	if data.wallet == nil {
		return nil, errors.New("wallet is required")
	}
	if !util.AcceptablePassphrase(data.passphrase) {
		return nil, errors.New("supplied passphrase is weak; use a stronger one or run with the --allow-weak-passphrases flag")
	}

	exporter, isExporter := data.wallet.(e2wtypes.WalletExporter)
	if !isExporter {
		return nil, errors.New("wallet does not provide export")
	}

	export, err := exporter.Export(ctx, []byte(data.passphrase))
	if err != nil {
		return nil, errors.Wrap(err, "failed to export wallet")
	}

	results := &dataOut{
		export: export,
	}

	return results, nil
}
