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
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

var walletInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Information about a wallet",
	Long: `Provide information about a wallet.  For example:

    ethdo wallet info --wallet=primary

In quiet mode this will return 0 if the wallet exists, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
		defer cancel()

		assert(viper.GetString("remote") == "", "wallet info not available with remote wallets")
		assert(viper.GetString("wallet") != "", "--wallet is required")

		wallet, err := walletFromPath(ctx, viper.GetString("wallet"))
		errCheck(err, "unknown wallet")

		if viper.GetBool("quiet") {
			os.Exit(0)
		}

		outputIf(viper.GetBool("verbose"), fmt.Sprintf("UUID: %v", wallet.ID()))
		fmt.Printf("Type: %s\n", wallet.Type())
		if viper.GetBool("verbose") {
			if storeProvider, ok := wallet.(wtypes.StoreProvider); ok {
				store := storeProvider.Store()
				fmt.Printf("Store: %s\n", store.Name())
				if storeLocationProvider, ok := store.(wtypes.StoreLocationProvider); ok {
					fmt.Printf("Location: %s\n", filepath.Join(storeLocationProvider.Location(), wallet.ID().String()))
				}
			}
		}

		// Count the accounts.
		accounts := 0
		for range wallet.Accounts(ctx) {
			accounts++
		}
		fmt.Printf("Accounts: %d\n", accounts)
	},
}

func init() {
	walletCmd.AddCommand(walletInfoCmd)
	walletFlags(walletInfoCmd)
}
