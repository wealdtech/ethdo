// Copyright © 2020 Weald Technology Trading
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
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

var walletDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a wallet",
	Long: `Delete a wallet.  For example:

    ethdo wallet delete --wallet=primary

In quiet mode this will return 0 if the wallet has been deleted, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		assert(!remote, "wallet delete not available with remote wallets")
		assert(walletWallet != "", "--wallet is required")

		wallet, err := walletFromPath(walletWallet)
		errCheck(err, "Failed to access wallet")

		storeProvider, ok := wallet.(wtypes.StoreProvider)
		assert(ok, "Cannot obtain store for the wallet")
		store := storeProvider.Store()
		storeLocationProvider, ok := store.(wtypes.StoreLocationProvider)
		assert(ok, "Cannot obtain store location for the wallet")
		walletLocation := filepath.Join(storeLocationProvider.Location(), wallet.ID().String())
		err = os.RemoveAll(walletLocation)
		errCheck(err, "Failed to delete wallet")

		os.Exit(_exitSuccess)
	},
}

func init() {
	walletCmd.AddCommand(walletDeleteCmd)
	walletFlags(walletDeleteCmd)
}
