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
	"strings"

	"github.com/spf13/cobra"
	wallet "github.com/wealdtech/go-eth2-wallet"
)

var walletCreateType string

var walletCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a wallet",
	Long: `Create a wallet.  For example:

    ethdo wallet create --wallet="Primary wallet" --type=non-deterministic

In quiet mode this will return 0 if the wallet is created successfully, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		assert(walletWallet != "", "--wallet is required")
		assert(walletCreateType != "", "--type is required")

		var err error
		switch strings.ToLower(walletCreateType) {
		case "non-deterministic", "nd":
			_, err = wallet.CreateWallet(walletWallet, wallet.WithType("nd"))
		case "hierarchical deterministic", "hd":
			assert(rootWalletPassphrase != "", "--walletpassphrase is required for hierarchical deterministic wallets")
			_, err = wallet.CreateWallet(walletWallet, wallet.WithType("hd"), wallet.WithPassphrase([]byte(rootWalletPassphrase)))
		default:
			die("unknown wallet type")
		}
		errCheck(err, "Failed to create wallet")
	},
}

func init() {
	walletCmd.AddCommand(walletCreateCmd)
	walletFlags(walletCreateCmd)
	walletCreateCmd.Flags().StringVar(&walletCreateType, "type", "non-deterministic", "Type of wallet to create (non-deterministic or hierarchical deterministic)")
}
