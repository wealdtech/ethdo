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

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	e2types "github.com/wealdtech/go-eth2-types/v2"
	util "github.com/wealdtech/go-eth2-util"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

var accountInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Information about an account",
	Long: `Obtain information about an account.  For example:

    ethdo account info --account="primary/my funds"

In quiet mode this will return 0 if the account exists, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
		defer cancel()

		assert(viper.GetString("account") != "", "--account is required")
		wallet, account, err := walletAndAccountFromInput(ctx)
		errCheck(err, "Failed to obtain account")

		// Disallow wildcards (for now)
		assert(fmt.Sprintf("%s/%s", wallet.Name(), account.Name()) == viper.GetString("account"), "Mismatched account name")

		if viper.GetBool("quiet") {
			os.Exit(_exitSuccess)
		}

		outputIf(viper.GetBool("verbose"), fmt.Sprintf("UUID: %v", account.ID()))
		var withdrawalPubKey e2types.PublicKey
		if pubKeyProvider, ok := account.(e2wtypes.AccountPublicKeyProvider); ok {
			fmt.Printf("Public key: %#x\n", pubKeyProvider.PublicKey().Marshal())
			// May be overwritten later, but grab it for now.
			withdrawalPubKey = pubKeyProvider.PublicKey()
		}
		if distributedAccount, ok := account.(e2wtypes.DistributedAccount); ok {
			fmt.Printf("Composite public key: %#x\n", distributedAccount.CompositePublicKey().Marshal())
			fmt.Printf("Signing threshold: %d/%d\n", distributedAccount.SigningThreshold(), len(distributedAccount.Participants()))
			if viper.GetBool("verbose") {
				fmt.Printf("Participants:\n")
				for k, v := range distributedAccount.Participants() {
					fmt.Printf(" %d: %s\n", k, v)
				}
			}

			withdrawalPubKey = distributedAccount.CompositePublicKey()
		}
		if viper.GetBool("verbose") {
			withdrawalCredentials := util.SHA256(withdrawalPubKey.Marshal())
			withdrawalCredentials[0] = byte(0) // BLS_WITHDRAWAL_PREFIX
			fmt.Printf("Withdrawal credentials: %#x\n", withdrawalCredentials)
		}
		if pathProvider, ok := account.(e2wtypes.AccountPathProvider); ok {
			if pathProvider.Path() != "" {
				fmt.Printf("Path: %s\n", pathProvider.Path())
			}
		}

		os.Exit(_exitSuccess)
	},
}

func init() {
	accountCmd.AddCommand(accountInfoCmd)
	accountFlags(accountInfoCmd)
}
