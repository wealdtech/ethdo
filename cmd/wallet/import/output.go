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
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
)

type dataOut struct {
	verify  bool
	quiet   bool
	verbose bool
	export  *export
}

type accountInfo struct {
	Name string `json:"name"`
}
type walletInfo struct {
	ID   uuid.UUID `json:"uuid"`
	Name string    `json:"name"`
	Type string    `json:"type"`
}
type export struct {
	Wallet   *walletInfo    `json:"wallet"`
	Accounts []*accountInfo `json:"accounts"`
}

func output(_ context.Context, data *dataOut) (string, error) {
	if data == nil {
		return "", errors.New("no data")
	}

	res := ""
	if data.verify {
		if !data.quiet {
			res = fmt.Sprintf("Wallet name: %s\nWallet type: %s\nWallet UUID: %s\nWallet accounts: %d", data.export.Wallet.Name, data.export.Wallet.Type, data.export.Wallet.ID, len(data.export.Accounts))
			if data.verbose {
				for _, account := range data.export.Accounts {
					res = fmt.Sprintf("%s\n  %s", res, account.Name)
				}
			}
		}
	}

	return res, nil
}
