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
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	util "github.com/wealdtech/go-eth2-util"
)

var accountWithdrawalCredentialsCmd = &cobra.Command{
	Use:   "withdrawalcredentials",
	Short: "Provide withdrawal credentials for an account",
	Long: `Provide withdrawal credentials for an account.  For example:

    ethdo account withdrawalcredentials --account="Validators/1"

In quiet mode this will return 0 if the account exists, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
		defer cancel()

		assert(viper.GetString("account") != "" || viper.GetString("pubkey") != "", "account or pubkey is required")

		var pubKey []byte
		if viper.GetString("pubkey") != "" {
			var err error
			pubKey, err = hex.DecodeString(strings.TrimPrefix(viper.GetString("pubkey"), "0x"))
			errCheck(err, "Failed to decode supplied public key")
		} else {
			_, account, err := walletAndAccountFromInput(ctx)
			errCheck(err, "Failed to obtain account")

			key, err := bestPublicKey(account)
			errCheck(err, "Account does not provide a public key")
			pubKey = key.Marshal()
		}

		if quiet {
			os.Exit(_exitSuccess)
		}

		withdrawalCredentials := util.SHA256(pubKey)
		withdrawalCredentials[0] = byte(0) // BLS_WITHDRAWAL_PREFIX
		fmt.Printf("%#x\n", withdrawalCredentials)
	},
}

func init() {
	accountCmd.AddCommand(accountWithdrawalCredentialsCmd)
	accountFlags(accountWithdrawalCredentialsCmd)
	accountWithdrawalCredentialsCmd.Flags().String("pubkey", "", "Public key (overrides account)")
	if err := viper.BindPFlag("pubkey", accountCreateCmd.Flags().Lookup("pubkey")); err != nil {
		panic(err)
	}
}
