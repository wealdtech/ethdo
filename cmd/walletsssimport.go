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
	walletsssimport "github.com/wealdtech/ethdo/cmd/wallet/sssimport"
)

var walletSSSImportCmd = &cobra.Command{
	Use:   "sssimport",
	Short: "Import a wallet using Shamir secret sharing",
	Long: `Import a wallet for backup of transfer using Shamir secret sharing.  For example:

	ethdo wallet sssimport --file=backup.dat --shares="1234 2345 3456"

In quiet mode this will return 0 if the wallet is imported successfully, otherwise 1.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		res, err := walletsssimport.Run(cmd)
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
	walletCmd.AddCommand(walletSSSImportCmd)
	walletFlags(walletSSSImportCmd)
	walletSSSImportCmd.Flags().String("file", "", "Name of the file that stores the export")
	walletSSSImportCmd.Flags().String("shares", "", "Shares required to decrypt the export, separated with spaces")
}

func walletSSSImportBindings() {
	if err := viper.BindPFlag("file", walletSSSImportCmd.Flags().Lookup("file")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("shares", walletSSSImportCmd.Flags().Lookup("shares")); err != nil {
		panic(err)
	}
}
