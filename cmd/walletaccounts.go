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
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	pb "github.com/wealdtech/eth2-signer-api/pb/v1"
)

var walletAccountsCmd = &cobra.Command{
	Use:   "accounts",
	Short: "List accounts in a wallet",
	Long: `List accounts in a wallet.  For example:

    ethdo wallet accounts --wallet=primary

In quiet mode this will return 0 if the wallet holds any addresses, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		assert(walletWallet != "", "--wallet is required")

		hasAccounts := false

		if remote {
			listerClient := pb.NewListerClient(remoteGRPCConn)
			listAccountsReq := &pb.ListAccountsRequest{
				Paths: []string{
					walletWallet,
				},
			}
			ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
			defer cancel()
			accountsResp, err := listerClient.ListAccounts(ctx, listAccountsReq)
			errCheck(err, "Failed to access wallet")
			assert(accountsResp.State == pb.ResponseState_SUCCEEDED, "Request to list wallet accounts failed")
			walletPrefixLen := len(walletWallet) + 1
			for _, account := range accountsResp.Accounts {
				hasAccounts = true
				if verbose {
					fmt.Printf("%s\n", account.Name[walletPrefixLen:])
					fmt.Printf("\tPublic key: %#048x\n", account.PublicKey)
				} else if !quiet {
					fmt.Printf("%s\n", account.Name[walletPrefixLen:])
				}
			}
		} else {
			wallet, err := walletFromPath(walletWallet)
			errCheck(err, "Failed to access wallet")

			for account := range wallet.Accounts() {
				hasAccounts = true
				if verbose {
					fmt.Printf("%s\n\tUUID:\t\t%s\n\tPublic key:\t0x%048x\n", account.Name(), account.ID(), account.PublicKey().Marshal())
				} else if !quiet {
					fmt.Printf("%s\n", account.Name())
				}
			}
		}

		if quiet {
			if hasAccounts {
				os.Exit(_exitSuccess)
			}
			os.Exit(_exitFailure)
		}
	},
}

func init() {
	walletCmd.AddCommand(walletAccountsCmd)
	walletFlags(walletAccountsCmd)
}
