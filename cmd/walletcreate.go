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
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	walletcreate "github.com/wealdtech/ethdo/cmd/wallet/create"
)

var walletCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a wallet",
	Long: `Create a wallet.  For example:

    ethdo wallet create --wallet="Primary wallet" --type=non-deterministic

In quiet mode this will return 0 if the wallet is created successfully, otherwise 1.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		res, err := walletcreate.Run(cmd)
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
	walletCmd.AddCommand(walletCreateCmd)
	walletFlags(walletCreateCmd)
	walletCreateCmd.Flags().String("type", "non-deterministic", "Type of wallet to create (non-deterministic or hierarchical deterministic)")
}

func walletCreateBindings(cmd *cobra.Command) {
	if err := viper.BindPFlag("type", cmd.Flags().Lookup("type")); err != nil {
		panic(err)
	}
}
