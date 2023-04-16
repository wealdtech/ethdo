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
	validatorkeycheck "github.com/wealdtech/ethdo/cmd/validator/keycheck"
)

var validatorKeycheckCmd = &cobra.Command{
	Use:   "keycheck",
	Short: "Check that the withdrawal credentials for a validator matches the given key.",
	Long: `Check that the withdrawal credentials for a validator matches the given key.  For example:

    ethdo validator keycheck --withdrawal-credentials=0x007e28dcf9029e8d92ca4b5d01c66c934e7f3110606f34ae3052cbf67bd3fc02 --private-key=0x1b46e61babc7a6a0fbfe8e416de3c71f85e367f24e0bfcb12e57adb11117662c

A mnemonic can be used in place of a private key, in which case the first 1,024 indices of the standard withdrawal key path will be scanned for a matching key.

In quiet mode this will return 0 if the withdrawal credentials match the key, otherwise 1.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		res, err := validatorkeycheck.Run(cmd)
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
	validatorCmd.AddCommand(validatorKeycheckCmd)
	validatorFlags(validatorKeycheckCmd)
	validatorKeycheckCmd.Flags().String("withdrawal-credentials", "", "Withdrawal credentials to check (can run offline)")
}

func validatorKeycheckBindings(cmd *cobra.Command) {
	if err := viper.BindPFlag("withdrawal-credentials", cmd.Flags().Lookup("withdrawal-credentials")); err != nil {
		panic(err)
	}
}
