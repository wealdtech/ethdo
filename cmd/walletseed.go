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

	bip39 "github.com/FactomProject/go-bip39"
	"github.com/spf13/cobra"
	types "github.com/wealdtech/go-eth2-wallet-types/v2"
)

var walletSeedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Display the seed of a wallet",
	Long: `Display the seed for an hierarchical deterministic wallet.  For example:

    ethdo wallet seed --wallet=primary

In quiet mode this will return 0 if the wallet is a hierarchical deterministic wallet, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		assert(!remote, "wallet seed not available with remote wallets")
		assert(walletWallet != "", "--wallet is required")
		assert(rootWalletPassphrase != "", "--walletpassphrase is required")

		wallet, err := walletFromPath(walletWallet)
		errCheck(err, "Failed to access wallet")
		_, ok := wallet.(types.WalletKeyProvider)
		assert(ok, fmt.Sprintf("wallets of type %q do not provide keys", wallet.Type()))

		err = wallet.Unlock([]byte(rootWalletPassphrase))
		errCheck(err, "Failed to unlock wallet")
		seed, err := wallet.(types.WalletKeyProvider).Key()
		errCheck(err, "Failed to obtain wallet key")
		outputIf(debug, fmt.Sprintf("Seed is %#0x", seed))
		seedStr, err := bip39.NewMnemonic(seed)
		errCheck(err, "Failed to generate seed mnemonic")

		outputIf(!quiet, seedStr)
		os.Exit(_exitSuccess)
	},
}

func init() {
	walletCmd.AddCommand(walletSeedCmd)
	walletFlags(walletSeedCmd)
}
