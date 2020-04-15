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
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	pb "github.com/wealdtech/eth2-signer-api/pb/v1"
)

var accountLockCmd = &cobra.Command{
	Use:   "lock",
	Short: "Lock a remote account",
	Long: `Lock a remote account.  For example:

    ethdo account lock --account="primary/my funds"

In quiet mode this will return 0 if the account is locked, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		assert(remote, "account lock only works with remote wallets")
		assert(rootAccount != "", "--account is required")

		client := pb.NewAccountManagerClient(remoteGRPCConn)
		lockReq := &pb.LockAccountRequest{
			Account: rootAccount,
		}
		ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
		defer cancel()
		resp, err := client.Lock(ctx, lockReq)
		errCheck(err, "Failed in attempt to lock account")
		switch resp.State {
		case pb.ResponseState_DENIED:
			die("Lock request denied")
		case pb.ResponseState_FAILED:
			die("Lock request failed")
		case pb.ResponseState_SUCCEEDED:
			outputIf(!quiet, "Lock request succeeded")
			os.Exit(_exitSuccess)
		}
	},
}

func init() {
	accountCmd.AddCommand(accountLockCmd)
	accountFlags(accountLockCmd)
}
