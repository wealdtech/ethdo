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
	accountcreate "github.com/wealdtech/ethdo/cmd/account/create"
)

var accountCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an account",
	Long: `Create an account.  For example:

    ethdo account create --account="primary/operations" --passphrase="my secret"

In quiet mode this will return 0 if the account is created successfully, otherwise 1.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		res, err := accountcreate.Run(cmd)
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
	accountCmd.AddCommand(accountCreateCmd)
	accountFlags(accountCreateCmd)
	accountCreateCmd.Flags().Uint32("participants", 1, "Number of participants (1 for non-distributed accounts, >1 for distributed accounts)")
	accountCreateCmd.Flags().Uint32("signing-threshold", 1, "Signing threshold (1 for non-distributed accounts)")
}

func accountCreateBindings(cmd *cobra.Command) {
	if err := viper.BindPFlag("participants", cmd.Flags().Lookup("participants")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("signing-threshold", cmd.Flags().Lookup("signing-threshold")); err != nil {
		panic(err)
	}
}
