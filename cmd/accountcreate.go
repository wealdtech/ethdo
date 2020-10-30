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
	"fmt"
	"os"
	"regexp"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wealdtech/ethdo/util"
	e2wallet "github.com/wealdtech/go-eth2-wallet"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

var accountCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an account",
	Long: `Create an account.  For example:

    ethdo account create --account="primary/operations" --passphrase="my secret"

In quiet mode this will return 0 if the account is created successfully, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
		defer cancel()

		assert(viper.GetString("account") != "", "--account is required")

		wallet, err := walletFromInput(ctx)
		errCheck(err, "Failed to access wallet")
		outputIf(debug, fmt.Sprintf("Opened wallet %q of type %s", wallet.Name(), wallet.Type()))
		if wallet.Type() == "hierarchical deterministic" {
			assert(getWalletPassphrase() != "", "walletpassphrase is required to create new accounts with hierarchical deterministic wallets")
		}
		locker, isLocker := wallet.(e2wtypes.WalletLocker)
		if isLocker {
			errCheck(locker.Unlock(ctx, []byte(getWalletPassphrase())), "Failed to unlock wallet")
		}

		_, accountName, err := e2wallet.WalletAndAccountNames(viper.GetString("account"))
		errCheck(err, "Failed to obtain account name")

		var account e2wtypes.Account
		if viper.GetUint("participants") > 0 {
			// Want a distributed account.
			distributedCreator, isDistributedCreator := wallet.(e2wtypes.WalletDistributedAccountCreator)
			assert(isDistributedCreator, "Wallet does not support distributed account creation")
			outputIf(debug, fmt.Sprintf("Distributed account has %d/%d threshold", viper.GetUint32("signing-threshold"), viper.GetUint32("participants")))
			ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
			defer cancel()
			if getOptionalPassphrase() != "" {
				assert(util.AcceptablePassphrase(getPassphrase()), "supplied passphrase is weak; use a stronger one or run with the --allow-weak-passphrases flag")
			}
			account, err = distributedCreator.CreateDistributedAccount(ctx, accountName, viper.GetUint32("participants"), viper.GetUint32("signing-threshold"), []byte(getOptionalPassphrase()))
		} else {
			assert(util.AcceptablePassphrase(getPassphrase()), "supplied passphrase is weak; use a stronger one or run with the --allow-weak-passphrases flag")
			if viper.GetString("path") != "" {
				// Want a pathed account
				creator, isCreator := wallet.(e2wtypes.WalletPathedAccountCreator)
				assert(isCreator, "Wallet does not support account creation with an explicit path")
				var match bool
				match, err = regexp.Match("^m/[0-9]+/[0-9]+(/[0-9+])+", []byte(viper.GetString("path")))
				errCheck(err, "Unable to match path to regular expression")
				assert(match, "Path does not match expected format m/...")
				ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
				defer cancel()
				account, err = creator.CreatePathedAccount(ctx, viper.GetString("path"), accountName, []byte(getPassphrase()))
			} else {
				// Want a standard account.
				creator, isCreator := wallet.(e2wtypes.WalletAccountCreator)
				assert(isCreator, "Wallet does not support account creation")
				ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
				defer cancel()
				account, err = creator.CreateAccount(ctx, accountName, []byte(getPassphrase()))
			}
		}
		errCheck(err, "Failed to create account")

		if pubKeyProvider, ok := account.(e2wtypes.AccountCompositePublicKeyProvider); ok {
			outputIf(verbose, fmt.Sprintf("%#x", pubKeyProvider.CompositePublicKey().Marshal()))
		} else if pubKeyProvider, ok := account.(e2wtypes.AccountPublicKeyProvider); ok {
			outputIf(verbose, fmt.Sprintf("%#x", pubKeyProvider.PublicKey().Marshal()))
		}

		os.Exit(_exitSuccess)
	},
}

func init() {
	accountCmd.AddCommand(accountCreateCmd)
	accountFlags(accountCreateCmd)
	accountCreateCmd.Flags().Uint32("participants", 0, "Number of participants (for distributed accounts)")
	accountCreateCmd.Flags().Uint32("signing-threshold", 0, "Signing threshold (for distributed accounts)")
	accountCreateCmd.Flags().String("path", "", "path of account (for hierarchical deterministic accounts)")
}

func accountCreateBindings() {
	if err := viper.BindPFlag("participants", accountCreateCmd.Flags().Lookup("participants")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("signing-threshold", accountCreateCmd.Flags().Lookup("signing-threshold")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("path", accountCreateCmd.Flags().Lookup("path")); err != nil {
		panic(err)
	}
}
