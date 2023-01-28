// Copyright © 2017-2019 Weald Technology Trading
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
	accountkey "github.com/wealdtech/ethdo/cmd/account/key"
)

// accountKeyCmd represents the account key command.
var accountKeyCmd = &cobra.Command{
	Use:   "key",
	Short: "Obtain the private key of an account.",
	Long: `Obtain the private key of an account.  For example:

    ethdo account key --account="Personal wallet/Operations" --passphrase="my account passphrase"

In quiet mode this will return 0 if the key can be obtained, otherwise 1.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		res, err := accountkey.Run(cmd)
		if err != nil {
			return err
		}
		if viper.GetBool("quiet") {
			return nil
		}
		if res != "" {
			fmt.Println(res)
		}
		return nil
	},
}

func init() {
	accountCmd.AddCommand(accountKeyCmd)
	accountFlags(accountKeyCmd)
}
