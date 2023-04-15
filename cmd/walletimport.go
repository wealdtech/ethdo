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
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	walletimport "github.com/wealdtech/ethdo/cmd/wallet/import"
)

var walletImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import a wallet",
	Long: `Import a wallet.  For example:

    ethdo wallet import --data=primary --passphrase="my export secret"

In quiet mode this will return 0 if the wallet is imported successfully, otherwise 1.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		res, err := walletimport.Run(cmd)
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
	walletCmd.AddCommand(walletImportCmd)
	walletFlags(walletImportCmd)
	walletImportCmd.Flags().String("data", "", "The data to import, or the name of a data import file")
	walletImportCmd.Flags().Bool("verify", false, "Verify the wallet can be imported, but do not import it")
}

func walletImportBindings(cmd *cobra.Command) {
	if err := viper.BindPFlag("data", cmd.Flags().Lookup("data")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("verify", cmd.Flags().Lookup("verify")); err != nil {
		panic(err)
	}
}
