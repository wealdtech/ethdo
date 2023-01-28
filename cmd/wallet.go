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
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// walletCmd represents the wallet command.
var walletCmd = &cobra.Command{
	Use:   "wallet",
	Short: "Manage wallets",
	Long:  `Create and manage wallets.`,
}

func init() {
	RootCmd.AddCommand(walletCmd)
}

var walletFlag *pflag.Flag

func walletFlags(cmd *cobra.Command) {
	if walletFlag == nil {
		cmd.Flags().String("wallet", "", "Name of the wallet")
		walletFlag = cmd.Flags().Lookup("wallet")
		if err := viper.BindPFlag("wallet", walletFlag); err != nil {
			panic(err)
		}
	} else {
		cmd.Flags().AddFlag(walletFlag)
	}
}
