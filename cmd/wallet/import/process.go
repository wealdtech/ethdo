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

package walletimport

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/wealdtech/go-ecodec"
	e2wallet "github.com/wealdtech/go-eth2-wallet"
)

func process(_ context.Context, data *dataIn) (*dataOut, error) {
	if data == nil {
		return nil, errors.New("no data")
	}
	if data.data == nil {
		return nil, errors.New("import data is required")
	}

	ext := &export{}
	if data.verify {
		data, err := ecodec.Decrypt(data.data, []byte(data.passphrase))
		if err != nil {
			return nil, errors.Wrap(err, "failed to decrypt export")
		}
		if err := json.Unmarshal(data, ext); err != nil {
			return nil, errors.Wrap(err, "failed to read export")
		}
	} else if _, err := e2wallet.ImportWallet(data.data, []byte(data.passphrase)); err != nil {
		return nil, errors.Wrap(err, "failed to import wallet")
	}

	results := &dataOut{
		verify: data.verify,
		export: ext,
	}

	return results, nil
}
