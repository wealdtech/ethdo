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

package walletcreate

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

type dataOut struct {
	mnemonic string
}

func output(_ context.Context, data *dataOut) (string, error) {
	if data == nil {
		return "", errors.New("no data")
	}
	if data.mnemonic != "" {
		return fmt.Sprintf(`The following phrase is your mnemonic for this wallet:

%s

Anyone with access to this mnemonic can recreate the accounts in this wallet, so please store this mnemonic safely.  More information about mnemonics can be found at https://support.mycrypto.com/general-knowledge/cryptography/how-do-mnemonic-phrases-work

Please note this mnemonic is not stored within the wallet, so cannot be retrieved or displayed again.  As such, this mnemonic should be stored securely, ideally offline, before proceeding.
`, data.mnemonic), nil
	}

	return "", nil
}
