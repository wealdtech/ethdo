// Copyright © 2022 Weald Technology Trading
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
	validatorcredentialsget "github.com/wealdtech/ethdo/cmd/validator/credentials/get"
)

var validatorCredentialsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Obtain withdrawal credentials for an Ethereum consensus validator",
	Long: `Obtain withdrawal credentials for an Ethereum consensus validator.  For example:

    ethdo validator credentials get --validator=primary/validator

In quiet mode this will return 0 if the validator exists, otherwise 1.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		res, err := validatorcredentialsget.Run(cmd)
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
	validatorCredentialsCmd.AddCommand(validatorCredentialsGetCmd)
	validatorCredentialsFlags(validatorCredentialsGetCmd)
	validatorCredentialsGetCmd.Flags().String("validator", "", "Validator for which to get validator credentials")
}

func validatorCredentialsGetBindings(cmd *cobra.Command) {
	if err := viper.BindPFlag("validator", cmd.Flags().Lookup("validator")); err != nil {
		panic(err)
	}
}
