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
	e2wallet "github.com/wealdtech/go-eth2-wallet"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

var accountCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an account",
	Long: `Create an account.  For example:

    ethdo account create --account="primary/operations" --passphrase="my secret"

In quiet mode this will return 0 if the account is created successfully, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		assert(!remote, "account create not available with remote wallets")
		assert(rootAccount != "", "--account is required")

		wallet, err := walletFromPath(rootAccount)
		errCheck(err, "Failed to access wallet")

		if wallet.Type() == "hierarchical deterministic" {
			assert(getWalletPassphrase() != "", "--walletpassphrase is required to create new accounts with hierarchical deterministic wallets")
		}
		_, err = accountFromPath(rootAccount)
		assert(err != nil, "Account already exists")

		err = wallet.Unlock([]byte(getWalletPassphrase()))
		errCheck(err, "Failed to unlock wallet")

		_, accountName, err := e2wallet.WalletAndAccountNames(rootAccount)
		errCheck(err, "Failed to obtain account name")

		walletAccountCreator, ok := wallet.(e2wtypes.WalletAccountCreator)
		assert(ok, "wallet does not allow account creation")

		account, err := walletAccountCreator.CreateAccount(accountName, []byte(getPassphrase()))
		errCheck(err, "Failed to create account")

		outputIf(verbose, fmt.Sprintf("0x%048x", account.PublicKey().Marshal()))
		os.Exit(_exitSuccess)
	},
}

func init() {
	accountCmd.AddCommand(accountCreateCmd)
	accountFlags(accountCreateCmd)
}
