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

var walletAccountsCmd = &cobra.Command{
	Use:   "accounts",
	Short: "List accounts in a wallet",
	Long: `List accounts in a wallet.  For example:

    ethdo wallet accounts --wallet=primary

In quiet mode this will return 0 if the wallet holds any addresses, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		assert(walletWallet != "", "wallet is required")

		wallet, err := walletFromPath(walletWallet)
		errCheck(err, "Failed to access wallet")

		hasAccounts := false
		for account := range wallet.Accounts() {
			hasAccounts = true
			if verbose {
				fmt.Printf("%s\n\tUUID:\t\t%s\n\tPublic key:\t0x%048x\n", account.Name(), account.ID(), account.PublicKey().Marshal())
			} else if !quiet {
				fmt.Printf("%s\n", account.Name())
			}
		}

		if quiet {
			if hasAccounts {
				os.Exit(_exit_success)
			}
			os.Exit(_exit_failure)
		}
	},
}

func init() {
	walletCmd.AddCommand(walletAccountsCmd)
	walletFlags(walletAccountsCmd)
}
