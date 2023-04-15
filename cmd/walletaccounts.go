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
	"sort"
	"strconv"
	"strings"

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

		wallet, err := walletFromInput(ctx)
		errCheck(err, "Failed to obtain wallet")

		accounts := make([]e2wtypes.Account, 0, 128)
		for account := range wallet.Accounts(ctx) {
			accounts = append(accounts, account)
		}
		assert(len(accounts) > 0, "")

		if _, isPathProvider := accounts[0].(e2wtypes.AccountPathProvider); isPathProvider {
			// Order accounts by their path components.
			sort.Slice(accounts, func(i int, j int) bool {
				iBits := strings.Split(accounts[i].(e2wtypes.AccountPathProvider).Path(), "/")
				jBits := strings.Split(accounts[j].(e2wtypes.AccountPathProvider).Path(), "/")
				for index := range iBits {
					if iBits[index] == "m" && jBits[index] == "m" {
						continue
					}
					if len(jBits) <= index {
						return false
					}
					iBit, err := strconv.ParseUint(iBits[index], 10, 64)
					if err != nil {
						return true
					}
					jBit, err := strconv.ParseUint(jBits[index], 10, 64)
					if err != nil {
						return false
					}
					if iBit < jBit {
						return true
					}
					if iBit > jBit {
						return false
					}
				}
				return len(jBits) > len(iBits)
			})
		} else {
			// Order accounts by their name.
			sort.Slice(accounts, func(i int, j int) bool {
				return strings.Compare(accounts[i].Name(), accounts[j].Name()) < 0
			})
		}

		for _, account := range accounts {
			outputIf(!viper.GetBool("quiet"), account.Name())
			if viper.GetBool("verbose") {
				fmt.Printf(" UUID: %v\n", account.ID())
				if pathProvider, isProvider := account.(e2wtypes.AccountPathProvider); isProvider {
					if pathProvider.Path() != "" {
						fmt.Printf("Path: %s\n", pathProvider.Path())
					}
				}
				if pubKeyProvider, isProvider := account.(e2wtypes.AccountPublicKeyProvider); isProvider {
					fmt.Printf(" Public key: %#x\n", pubKeyProvider.PublicKey().Marshal())
				}
				if compositePubKeyProvider, isProvider := account.(e2wtypes.AccountCompositePublicKeyProvider); isProvider {
					fmt.Printf(" Composite public key: %#x\n", compositePubKeyProvider.CompositePublicKey().Marshal())
				}
			}
		}
		os.Exit(_exitSuccess)
	},
}

func init() {
	walletCmd.AddCommand(walletAccountsCmd)
	walletFlags(walletAccountsCmd)
}
