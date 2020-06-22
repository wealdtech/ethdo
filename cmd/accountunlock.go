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

var accountUnlockCmd = &cobra.Command{
	Use:   "unlock",
	Short: "Unlock a remote account",
	Long: `Unlock a remote account.  For example:

    ethdo account unlock --account="primary/my funds" --passphrase="secret"

In quiet mode this will return 0 if the account is unlocked, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		assert(remote, "account unlock only works with remote wallets")
		assert(rootAccount != "", "--account is required")

		client := pb.NewAccountManagerClient(remoteGRPCConn)
		unlocked := false
		ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
		defer cancel()
		for _, passphrase := range getPassphrases() {
			unlockReq := &pb.UnlockAccountRequest{
				Account:    rootAccount,
				Passphrase: []byte(passphrase),
			}
			resp, err := client.Unlock(ctx, unlockReq)
			errCheck(err, "Failed in attempt to unlock account")
			if resp.State == pb.ResponseState_SUCCEEDED {
				unlocked = true
				break
			}
		}
		assert(unlocked, "Failed to unlock account")
		os.Exit(_exitSuccess)
	},
}

func init() {
	accountCmd.AddCommand(accountUnlockCmd)
	accountFlags(accountUnlockCmd)
}
