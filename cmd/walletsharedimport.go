// Copyright © 2021 Weald Technology Trading
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
	walletsharedimport "github.com/wealdtech/ethdo/cmd/wallet/sharedimport"
)

var walletSharedImportCmd = &cobra.Command{
	Use:   "sharedimport",
	Short: "Import a wallet using Shamir secret sharing",
	Long: `Import a wallet for backup of transfer using Shamir secret sharing.  For example:

	ethdo wallet sharedimport --file=backup.dat --shares="1234 2345 3456"

In quiet mode this will return 0 if the wallet is imported successfully, otherwise 1.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		res, err := walletsharedimport.Run(cmd)
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
	walletCmd.AddCommand(walletSharedImportCmd)
	walletFlags(walletSharedImportCmd)
	walletSharedImportCmd.Flags().String("file", "", "Name of the file that stores the export")
	walletSharedImportCmd.Flags().String("shares", "", "Shares required to decrypt the export, separated with spaces")
}

func walletSharedImportBindings(cmd *cobra.Command) {
	if err := viper.BindPFlag("file", cmd.Flags().Lookup("file")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("shares", cmd.Flags().Lookup("shares")); err != nil {
		panic(err)
	}
}
