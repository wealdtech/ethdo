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
	pb "github.com/wealdtech/eth2-signer-api/pb/v1"
	util "github.com/wealdtech/go-eth2-util"
)

var accountInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Information about an account",
	Long: `Obtain information about an account.  For example:

    ethdo account info --account="primary/my funds"

In quiet mode this will return 0 if the account exists, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		assert(rootAccount != "", "--account is required")

		var withdrawalCredentials []byte
		if remote {
			listerClient := pb.NewListerClient(remoteGRPCConn)
			listAccountsReq := &pb.ListAccountsRequest{
				Paths: []string{
					rootAccount,
				},
			}
			resp, err := listerClient.ListAccounts(context.Background(), listAccountsReq)
			errCheck(err, "Failed to access account")
			assert(resp.State == pb.ResponseState_SUCCEEDED, "No such account")
			assert(len(resp.Accounts) == 1, "No such account")
			fmt.Printf("Public key: %#x\n", resp.Accounts[0].PublicKey)
			withdrawalCredentials = util.SHA256(resp.Accounts[0].PublicKey)
			withdrawalCredentials[0] = byte(0) // BLS_WITHDRAWAL_PREFIX
			outputIf(verbose, fmt.Sprintf("Withdrawal credentials: %#x", withdrawalCredentials))
		} else {
			account, err := accountFromPath(rootAccount)
			errCheck(err, "Failed to access wallet")
			outputIf(verbose, fmt.Sprintf("UUID: %v", account.ID()))
			outputIf(!quiet, fmt.Sprintf("Public key: %#x", account.PublicKey().Marshal()))
			withdrawalCredentials = util.SHA256(account.PublicKey().Marshal())
			withdrawalCredentials[0] = byte(0) // BLS_WITHDRAWAL_PREFIX
			outputIf(verbose, fmt.Sprintf("Withdrawal credentials: %#x", withdrawalCredentials))
			outputIf(verbose && account.Path() != "", fmt.Sprintf("Path: %s", account.Path()))
		}

		os.Exit(_exitSuccess)
	},
}

func init() {
	accountCmd.AddCommand(accountInfoCmd)
	accountFlags(accountInfoCmd)
}
