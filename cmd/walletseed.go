// Copyright © 2019, 2020 Weald Technology Trading
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
	"bytes"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	bip39 "github.com/tyler-smith/go-bip39"
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
		assert(getWalletPassphrase() != "", "--walletpassphrase is required")

		wallet, err := walletFromPath(walletWallet)
		errCheck(err, "Failed to access wallet")
		_, ok := wallet.(types.WalletKeyProvider)
		assert(ok, fmt.Sprintf("wallets of type %q do not have a seed", wallet.Type()))

		err = wallet.Unlock([]byte(getWalletPassphrase()))
		errCheck(err, "Failed to unlock wallet")
		seed, err := wallet.(types.WalletKeyProvider).Key()
		errCheck(err, "Failed to obtain wallet key")
		outputIf(debug, fmt.Sprintf("Seed is %#x", seed))
		seedStr, err := bip39.NewMnemonic(seed)
		errCheck(err, "Failed to generate seed mnemonic")
		// Re-read mnemonimc to ensure correctness.
		recalcSeed, err := bip39.MnemonicToByteArray(seedStr)
		// Drop checksum (last byte).
		errCheck(err, "Failed to recalculate seed")
		recalcSeed = recalcSeed[:len(recalcSeed)-1]
		outputIf(debug, fmt.Sprintf("Recalc seed is %#x", recalcSeed))
		errCheck(err, "Failed to recalculate seed mnemonic")
		assert(bytes.Equal(recalcSeed, seed), "Generated invalid mnemonic")

		outputIf(!quiet, seedStr)
		os.Exit(_exitSuccess)
	},
}

func init() {
	walletCmd.AddCommand(walletSeedCmd)
	walletFlags(walletSeedCmd)
}
