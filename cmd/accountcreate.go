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
)

var accountCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an account",
	Long: `Create an account.  For example:

    ethdo account create --account="primary/operations" --passphrase="my secret"

In quiet mode this will return 0 if the account is created successfully, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		assert(rootAccount != "", "--account is required")
		assert(rootAccountPassphrase != "", "--passphrase is required")

		w, err := walletFromPath(rootAccount)
		errCheck(err, "Failed to access wallet")

		if w.Type() == "hierarchical deterministic" {
			assert(rootWalletPassphrase != "", "--walletpassphrase is required to create new accounts with hierarchical deterministic wallets")
		}
		_, err = accountFromPath(rootAccount)
		assert(err != nil, "Account already exists")

		err = w.Unlock([]byte(rootWalletPassphrase))
		errCheck(err, "Failed to unlock wallet")

		_, accountName, err := walletAndAccountNamesFromPath(rootAccount)
		errCheck(err, "Failed to obtain accout name")

		account, err := w.CreateAccount(accountName, []byte(rootAccountPassphrase))
		errCheck(err, "Failed to create account")

		outputIf(verbose, fmt.Sprintf("0x%048x", account.PublicKey().Marshal()))
		os.Exit(_exit_success)
	},
}

func init() {
	accountCmd.AddCommand(accountCreateCmd)
	accountFlags(accountCreateCmd)
}
