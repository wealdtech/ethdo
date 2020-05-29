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
	"strings"

	"github.com/spf13/cobra"
	bip39 "github.com/tyler-smith/go-bip39"
	keystorev4 "github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"
	hd "github.com/wealdtech/go-eth2-wallet-hd/v2"
	nd "github.com/wealdtech/go-eth2-wallet-nd/v2"
)

var walletCreateType string
var walletCreateSeed string

var walletCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a wallet",
	Long: `Create a wallet.  For example:

    ethdo wallet create --wallet="Primary wallet" --type=non-deterministic

In quiet mode this will return 0 if the wallet is created successfully, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		assert(!remote, "wallet create not available with remote wallets")
		assert(walletWallet != "", "--wallet is required")
		assert(walletCreateType != "", "--type is required")

		var err error
		switch strings.ToLower(walletCreateType) {
		case "non-deterministic", "nd":
			assert(walletCreateSeed == "", "--seed is not allowed with non-deterministic wallets")
			err = walletCreateND(walletWallet)
		case "hierarchical deterministic", "hd":
			assert(rootWalletPassphrase != "", "--walletpassphrase is required for hierarchical deterministic wallets")
			err = walletCreateHD(walletWallet, rootWalletPassphrase, walletCreateSeed)
		default:
			die("unknown wallet type")
		}
		errCheck(err, "Failed to create wallet")
	},
}

// walletCreateND creates a non-deterministic wallet.
func walletCreateND(name string) error {
	_, err := nd.CreateWallet(name, store, keystorev4.New())
	return err
}

// walletCreateND creates a hierarchical-deterministic wallet.
func walletCreateHD(name string, passphrase string, seedPhrase string) error {
	encryptor := keystorev4.New()
	if seedPhrase != "" {
		// Create wallet from a user-supplied seed.
		var seed []byte
		seed, err := bip39.MnemonicToByteArray(seedPhrase)
		errCheck(err, "Failed to decode seed")
		// Strip checksum; last byte.
		seed = seed[:len(seed)-1]
		assert(len(seed) == 32, "Seed must have 24 words")
		_, err = hd.CreateWalletFromSeed(name, []byte(passphrase), store, encryptor, seed)
		return err
	}
	// Create wallet with a random seed.
	_, err := hd.CreateWallet(name, []byte(passphrase), store, encryptor)
	return err
}

func init() {
	walletCmd.AddCommand(walletCreateCmd)
	walletFlags(walletCreateCmd)
	walletCreateCmd.Flags().StringVar(&walletCreateType, "type", "non-deterministic", "Type of wallet to create (non-deterministic or hierarchical deterministic)")
	walletCreateCmd.Flags().StringVar(&walletCreateSeed, "seed", "", "The 24-word seed phrase for a hierarchical deterministic wallet")
}
