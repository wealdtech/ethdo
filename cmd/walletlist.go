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
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	e2wallet "github.com/wealdtech/go-eth2-wallet"
)

var walletListCmd = &cobra.Command{
	Use:   "list",
	Short: "List known wallets",
	Long: `Provide information about local wallets.  For example:

    ethdo wallet list

In quiet mode this will return 0 if any wallets are found, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		assert(viper.GetString("remote") == "", "wallet list not available with remote wallets")
		assert(viper.GetString("wallet") == "", "wallet list does not take a --wallet parameter")

		walletsFound := false
		for w := range e2wallet.Wallets() {
			walletsFound = true
			outputIf(!viper.GetBool("quiet") && !viper.GetBool("verbose"), w.Name())
			outputIf(viper.GetBool("verbose"), fmt.Sprintf("%s\n UUID: %s", w.Name(), w.ID().String()))
		}

		if !walletsFound {
			os.Exit(_exitFailure)
		}
		os.Exit(_exitSuccess)
	},
}

func init() {
	walletCmd.AddCommand(walletListCmd)
	walletFlags(walletListCmd)
}
