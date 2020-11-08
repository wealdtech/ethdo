// Copyright © 2020 Weald Technology Trading
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
	"fmt"

	"github.com/spf13/cobra"
	walletdelete "github.com/wealdtech/ethdo/cmd/wallet/delete"
)

var walletDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a wallet",
	Long: `Delete a wallet.  For example:

    ethdo wallet delete --wallet=primary

In quiet mode this will return 0 if the wallet has been deleted, otherwise 1.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		res, err := walletdelete.Run(cmd)
		if err != nil {
			return err
		}
		if res != "" {
			fmt.Println(res)
		}
		return nil
	},
}

func init() {
	walletCmd.AddCommand(walletDeleteCmd)
	walletFlags(walletDeleteCmd)
}
