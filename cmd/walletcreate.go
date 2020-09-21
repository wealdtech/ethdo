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
	"crypto/rand"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	bip39 "github.com/tyler-smith/go-bip39"
	distributed "github.com/wealdtech/go-eth2-wallet-distributed"
	keystorev4 "github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"
	hd "github.com/wealdtech/go-eth2-wallet-hd/v2"
	nd "github.com/wealdtech/go-eth2-wallet-nd/v2"
	"golang.org/x/text/unicode/norm"
)

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
		assert(viper.GetString("type") != "", "--type is required")

		var err error
		switch strings.ToLower(viper.GetString("type")) {
		case "non-deterministic", "nd":
			assert(viper.GetString("mnemonic") == "", "--mnemonic is not allowed with non-deterministic wallets")
			err = walletCreateND(ctx, viper.GetString("wallet"))
		case "hierarchical deterministic", "hd":
			if quiet {
				fmt.Printf("Creation of hierarchical deterministic wallets prints its mnemonic, so cannot be run with the --quiet flag")
				os.Exit(_exitFailure)
			}
			assert(getWalletPassphrase() != "", "--walletpassphrase is required for hierarchical deterministic wallets")
			err = walletCreateHD(ctx, viper.GetString("wallet"), getWalletPassphrase(), viper.GetString("mnemonic"))
		case "distributed":
			assert(viper.GetString("mnemonic") == "", "--mnemonic is not allowed with distributed wallets")
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

// walletCreateHD creates a hierarchical-deterministic wallet.
func walletCreateHD(ctx context.Context, name string, passphrase string, mnemonic string) error {
	encryptor := keystorev4.New()

	printMnemonic := mnemonic == ""
	mnemonicPassphrase := ""

	if mnemonic == "" {
		// Create a new random mnemonic.
		entropy := make([]byte, 32)
		_, err := rand.Read(entropy)
		if err != nil {
			return errors.Wrap(err, "failed to generate entropy for wallet mnemonic")
		}
		mnemonic, err = bip39.NewMnemonic(entropy)
		if err != nil {
			return errors.Wrap(err, "failed to generate wallet mnemonic")
		}
	} else {
		// We have an existing mnemonic.  If there are more than 24 words we treat the additional characters as the passphrase.
		mnemonicParts := strings.Split(mnemonic, " ")
		if len(mnemonicParts) > 24 {
			mnemonic = strings.Join(mnemonicParts[:24], " ")
			mnemonicPassphrase = strings.Join(mnemonicParts[24:], " ")
		}
	}
	// Normalise the input.
	mnemonic = string(norm.NFKD.Bytes([]byte(mnemonic)))
	mnemonicPassphrase = string(norm.NFKD.Bytes([]byte(mnemonicPassphrase)))

	// Ensure the mnemonic is valid
	if !bip39.IsMnemonicValid(mnemonic) {
		return errors.New("mnemonic is not valid")
	}

	// Create seed from mnemonic and passphrase.
	seed := bip39.NewSeed(mnemonic, mnemonicPassphrase)

	_, err := hd.CreateWallet(ctx, name, []byte(passphrase), store, encryptor, seed)

	if printMnemonic {
		fmt.Printf(`The following phrase is your mnemonic for this wallet:

%s

Anyone with access to this mnemonic can recreate the accounts in this wallet, so please store this mnemonic safely.  More information about mnemonics can be found at https://support.mycrypto.com/general-knowledge/cryptography/how-do-mnemonic-phrases-work

Please note this mnemonic is not stored within the wallet, so cannot be retrieved or displayed again.  As such, this mnemonic should be written down or otherwise protected before proceeding.
`, mnemonic)
	}

	return err
}

func init() {
	walletCmd.AddCommand(walletCreateCmd)
	walletFlags(walletCreateCmd)
	walletCreateCmd.Flags().String("type", "non-deterministic", "Type of wallet to create (non-deterministic or hierarchical deterministic)")
	walletCreateCmd.Flags().String("mnemonic", "", "The 24-word mnemonic for a hierarchical deterministic wallet")
}

func walletCreateBindings() {
	if err := viper.BindPFlag("type", walletCreateCmd.Flags().Lookup("type")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("mnemonic", walletCreateCmd.Flags().Lookup("mnemonic")); err != nil {
		panic(err)
	}
}
