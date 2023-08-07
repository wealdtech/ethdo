// Copyright © 2023 Weald Technology Trading.
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
	"github.com/spf13/viper"
	walletbatch "github.com/wealdtech/ethdo/cmd/wallet/batch"
)

var walletBatchCmd = &cobra.Command{
	Use:   "batch",
	Short: "Batch a wallet",
	Long: `Batch a wallet.  For example:

    ethdo wallet batch --wallet="Primary wallet" --passphrase=accounts-secret --batch-passphrase=batch-secret

In quiet mode this will return 0 if the wallet is batched successfully, otherwise 1.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		res, err := walletbatch.Run(cmd)
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
	walletCmd.AddCommand(walletBatchCmd)
	walletFlags(walletBatchCmd)
	walletBatchCmd.Flags().String("batch-passphrase", "", "The passphrase to use for the batch")
}

func walletBatchBindings(cmd *cobra.Command) {
	if err := viper.BindPFlag("batch-passphrase", cmd.Flags().Lookup("batch-passphrase")); err != nil {
		panic(err)
	}
}
