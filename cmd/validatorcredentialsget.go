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

    ethdo validator credentials get --account=primary/validator

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
	validatorCredentialsGetCmd.Flags().String("account", "", "Account for which to fetch validator credentials")
	validatorCredentialsGetCmd.Flags().String("index", "", "Validator index for which to fetch validator credentials")
	validatorCredentialsGetCmd.Flags().String("pubkey", "", "Validator public key for which to fetch validator credentials")
}

func validatorCredentialsGetBindings() {
	if err := viper.BindPFlag("account", validatorCredentialsGetCmd.Flags().Lookup("account")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("index", validatorCredentialsGetCmd.Flags().Lookup("index")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("pubkey", validatorCredentialsGetCmd.Flags().Lookup("pubkey")); err != nil {
		panic(err)
	}
}
