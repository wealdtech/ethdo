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
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

var walletAccountsCmd = &cobra.Command{
	Use:   "accounts",
	Short: "List accounts in a wallet",
	Long: `List accounts in a wallet.  For example:

    ethdo wallet accounts --wallet=primary

In quiet mode this will return 0 if the wallet holds any addresses, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
		defer cancel()

		assert(viper.GetString("wallet") != "", "wallet is required")

		wallet, err := openWallet()
		errCheck(err, "Failed to access wallet")

		hasAccounts := false
		for account := range wallet.Accounts(ctx) {
			hasAccounts = true
			outputIf(!quiet, account.Name())
			if verbose {
				fmt.Printf(" UUID: %v\n", account.ID())
				pubKeyProvider, isProvider := account.(e2wtypes.AccountPublicKeyProvider)
				if isProvider {
					fmt.Printf(" Public key: %#x\n", pubKeyProvider.PublicKey().Marshal())
				}
				compositePubKeyProvider, isProvider := account.(e2wtypes.AccountCompositePublicKeyProvider)
				if isProvider {
					fmt.Printf(" Composite public key: %#x\n", compositePubKeyProvider.CompositePublicKey().Marshal())
				}
			}
		}

		if hasAccounts {
			os.Exit(_exitSuccess)
		}
		os.Exit(_exitFailure)
	},
}

func init() {
	walletCmd.AddCommand(walletAccountsCmd)
	walletFlags(walletAccountsCmd)
}
