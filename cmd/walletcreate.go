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
	"context"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	bip39 "github.com/tyler-smith/go-bip39"
	distributed "github.com/wealdtech/go-eth2-wallet-distributed"
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
		ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
		defer cancel()

		assert(viper.GetString("remote") == "", "wallet create not available with remote wallets")
		assert(viper.GetString("wallet") != "", "--wallet is required")
		assert(walletCreateType != "", "--type is required")

		var err error
		switch strings.ToLower(walletCreateType) {
		case "non-deterministic", "nd":
			assert(walletCreateSeed == "", "--seed is not allowed with non-deterministic wallets")
			err = walletCreateND(ctx, viper.GetString("wallet"))
		case "hierarchical deterministic", "hd":
			assert(getWalletPassphrase() != "", "--walletpassphrase is required for hierarchical deterministic wallets")
			err = walletCreateHD(ctx, viper.GetString("wallet"), getWalletPassphrase(), walletCreateSeed)
		case "distributed":
			assert(walletCreateSeed == "", "--seed is not allowed with distributed wallets")
			err = walletCreateDistributed(ctx, viper.GetString("wallet"))
		default:
			die("unknown wallet type")
		}
		errCheck(err, "Failed to create wallet")
	},
}

// walletCreateND creates a non-deterministic wallet.
func walletCreateND(ctx context.Context, name string) error {
	_, err := nd.CreateWallet(ctx, name, store, keystorev4.New())
	return err
}

// walletCreateDistributed creates a distributed wallet.
func walletCreateDistributed(ctx context.Context, name string) error {
	_, err := distributed.CreateWallet(ctx, name, store, keystorev4.New())
	return err
}

// walletCreateND creates a hierarchical-deterministic wallet.
func walletCreateHD(ctx context.Context, name string, passphrase string, seedPhrase string) error {
	encryptor := keystorev4.New()
	if seedPhrase != "" {
		// Create wallet from a user-supplied seed.
		var seed []byte
		seed, err := bip39.MnemonicToByteArray(seedPhrase)
		errCheck(err, "Failed to decode seed")
		// Strip checksum; last byte.
		seed = seed[:len(seed)-1]
		assert(len(seed) == 32, "Seed must have 24 words")
		_, err = hd.CreateWalletFromSeed(ctx, name, []byte(passphrase), store, encryptor, seed)
		return err
	}
	// Create wallet with a random seed.
	_, err := hd.CreateWallet(ctx, name, []byte(passphrase), store, encryptor)
	return err
}

func init() {
	walletCmd.AddCommand(walletCreateCmd)
	walletFlags(walletCreateCmd)
	walletCreateCmd.Flags().StringVar(&walletCreateType, "type", "non-deterministic", "Type of wallet to create (non-deterministic or hierarchical deterministic)")
	walletCreateCmd.Flags().StringVar(&walletCreateSeed, "seed", "", "The 24-word seed phrase for a hierarchical deterministic wallet")
}
