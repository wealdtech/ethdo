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
	"sort"

	"github.com/spf13/cobra"
	types "github.com/wealdtech/go-eth2-types"
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

		// List the accounts.  They come to us in random order and we want them in name order, so store them in an array and sort
		output := make([]addressListResult, 0)
		for account := range wallet.Accounts() {
			output = append(output, addressListResult{name: account.Name(), pubkey: account.PublicKey()})
		}

		if quiet {
			if len(output) == 0 {
				os.Exit(1)
			}
			os.Exit(0)
		}

		sort.Slice(output, func(i, j int) bool {
			return output[i].name < output[j].name
		})
		for _, out := range output {
			if verbose {
				fmt.Printf("%s: 0x%048x\n", out.name, out.pubkey.Marshal())
			} else if !quiet {
				fmt.Printf("%s\n", out.name)
			}
		}
	},
}

func init() {
	walletCmd.AddCommand(walletAccountsCmd)
	walletFlags(walletAccountsCmd)
}

type addressListResult struct {
	name   string
	pubkey types.PublicKey
}
