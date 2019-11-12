// Copyright © 2019 Weald Technology Trading
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

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wealdtech/go-bytesutil"
	types "github.com/wealdtech/go-eth2-wallet-types"
)

var accountImportKey string

var accountImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import an account",
	Long: `Import an account from its private key.  For example:

    ethdo account import --account="primary/testing" --key="0x..." --passphrase="my secret"

In quiet mode this will return 0 if the account is imported successfully, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		assert(rootAccount != "", "--account is required")
		assert(rootAccountPassphrase != "", "--passphrase is required")
		assert(accountImportKey != "", "--key is required")

		key, err := bytesutil.FromHexString(accountImportKey)
		errCheck(err, "Invalid key")

		w, err := walletFromPath(rootAccount)
		errCheck(err, "Failed to access wallet")

		_, ok := w.(types.WalletAccountImporter)
		assert(ok, fmt.Sprintf("wallets of type %q do not allow importing accounts", w.Type()))

		_, err = accountFromPath(rootAccount)
		assert(err != nil, "Account already exists")

		err = w.Unlock([]byte(rootWalletPassphrase))
		errCheck(err, "Failed to unlock wallet")

		_, accountName, err := walletAndAccountNamesFromPath(rootAccount)
		errCheck(err, "Failed to obtain accout name")

		account, err := w.(types.WalletAccountImporter).ImportAccount(accountName, key, []byte(rootAccountPassphrase))
		errCheck(err, "Failed to create account")

		outputIf(verbose, fmt.Sprintf("0x%048x", account.PublicKey().Marshal()))
		os.Exit(_exit_success)
	},
}

func init() {
	accountCmd.AddCommand(accountImportCmd)
	accountFlags(accountImportCmd)
	accountImportCmd.Flags().StringVar(&accountImportKey, "key", "", "Private key of the account to import (0x...)")
}
