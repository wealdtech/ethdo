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
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wealdtech/ethdo/util"
	"github.com/wealdtech/go-bytesutil"
	e2wallet "github.com/wealdtech/go-eth2-wallet"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

var accountImportKey string

var accountImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import an account",
	Long: `Import an account from its private key.  For example:

    ethdo account import --account="primary/testing" --key="0x..." --passphrase="my secret"

In quiet mode this will return 0 if the account is imported successfully, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		assert(!remote, "account import not available with remote wallets")
		assert(viper.GetString("account") != "", "--account is required")
		assert(accountImportKey != "", "--key is required")
		assert(util.AcceptablePassphrase(getPassphrase()), "supplied passphrase is weak; use a stronger one or run with the --allow-weak-passphrases flag")

		ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
		defer cancel()

		key, err := bytesutil.FromHexString(accountImportKey)
		errCheck(err, "Invalid key")

		w, err := walletFromPath(ctx, viper.GetString("account"))
		errCheck(err, "Failed to access wallet")

		_, ok := w.(e2wtypes.WalletAccountImporter)
		assert(ok, fmt.Sprintf("wallets of type %q do not allow importing accounts", w.Type()))

		_, _, err = walletAndAccountFromPath(ctx, viper.GetString("account"))
		assert(err != nil, "Account already exists")

		locker, isLocker := w.(e2wtypes.WalletLocker)
		if isLocker {
			errCheck(locker.Unlock(ctx, []byte(getWalletPassphrase())), "Failed to unlock wallet")
		}

		_, accountName, err := e2wallet.WalletAndAccountNames(viper.GetString("account"))
		errCheck(err, "Failed to obtain account name")

		account, err := w.(e2wtypes.WalletAccountImporter).ImportAccount(ctx, accountName, key, []byte(getPassphrase()))
		errCheck(err, "Failed to create account")

		pubKey, err := bestPublicKey(account)
		if err == nil {
			outputIf(verbose, fmt.Sprintf("%#x", pubKey.Marshal()))
		}

		os.Exit(_exitSuccess)
	},
}

func init() {
	accountCmd.AddCommand(accountImportCmd)
	accountFlags(accountImportCmd)
	accountImportCmd.Flags().StringVar(&accountImportKey, "key", "", "Private key of the account to import (0x...)")
}
