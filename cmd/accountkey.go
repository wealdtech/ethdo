// Copyright © 2017-2019 Weald Technology Trading
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
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

// accountKeyCmd represents the account key command
var accountKeyCmd = &cobra.Command{
	Use:   "key",
	Short: "Obtain the private key of an account.",
	Long: `Obtain the private key of an account.  For example:

    ethdo account key --account="Personal wallet/Operations" --passphrase="my account passphrase"

In quiet mode this will return 0 if the key can be obtained, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
		defer cancel()

		assert(!remote, "account keys not available with remote wallets")
		assert(viper.GetString("account") != "", "--account is required")

		_, account, err := walletAndAccountFromInput(ctx)
		errCheck(err, "Failed to obtain account")

		privateKeyProvider, isPrivateKeyProvider := account.(e2wtypes.AccountPrivateKeyProvider)
		assert(isPrivateKeyProvider, fmt.Sprintf("account %q does not provide its private key", viper.GetString("account")))

		if locker, isLocker := account.(e2wtypes.AccountLocker); isLocker {
			unlocked, err := locker.IsUnlocked(ctx)
			errCheck(err, "Failed to find out if account is locked")
			if !unlocked {
				for _, passphrase := range getPassphrases() {
					err = locker.Unlock(ctx, []byte(passphrase))
					if err == nil {
						unlocked = true
						break
					}
				}
			}
			assert(unlocked, "Failed to unlock account to obtain private key")
			defer relockAccount(locker)
		}
		privateKey, err := privateKeyProvider.PrivateKey(ctx)
		errCheck(err, "Failed to obtain private key")

		outputIf(!quiet, fmt.Sprintf("%#x", privateKey.Marshal()))
		os.Exit(_exitSuccess)
	},
}

func init() {
	accountCmd.AddCommand(accountKeyCmd)
	accountFlags(accountKeyCmd)
}
