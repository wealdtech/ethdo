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
	validatorcredentialsset "github.com/wealdtech/ethdo/cmd/validator/credentials/set"
)

var validatorCredentialsSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set withdrawal credentials for an Ethereum consensus validator",
	Long: `Set withdrawal credentials for an Ethereum consensus validator via a "change credentials" operation.  For example:

    ethdo validator credentials set --validator=primary/validator --withdrawal-address=0x00...13 --private-key=0x00...1f

The existing account can be specified in one of a number of ways:

  - mnemonic using --mnemonic; this will scan the mnemonic and generate all required operations
  - mnemonic and path to the validator key using --mnemonic and --path; this will generate a single operation
  - private key using --private-key; this will generate a single operation
  - account and passphrase using --account and --passphrase; this will generate a single operation

In quiet mode this will return 0 if the credentials operation has been generated (and successfully broadcast if online), otherwise 1.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		res, err := validatorcredentialsset.Run(cmd)
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
	validatorCredentialsCmd.AddCommand(validatorCredentialsSetCmd)
	validatorCredentialsFlags(validatorCredentialsSetCmd)
	validatorCredentialsSetCmd.Flags().Bool("prepare-offline", false, "Create files for offline use")
	validatorCredentialsSetCmd.Flags().String("validator", "", "Validator for which to set validator credentials")
	validatorCredentialsSetCmd.Flags().String("withdrawal-address", "", "Execution address to which to direct withdrawals")
	validatorCredentialsSetCmd.Flags().String("signed-operation", "", "Use pre-defined JSON signed operation as created by --json to transmit the credentials change operation")
	validatorCredentialsSetCmd.Flags().Bool("json", false, "Generate JSON data containing a signed operation rather than broadcast it to the network (implied when offline)")
	validatorCredentialsSetCmd.Flags().Bool("offline", false, "Do not attempt to connect to a beacon node to obtain information for the operation")
	validatorCredentialsSetCmd.Flags().String("fork-version", "", "Fork version to use for signing (offline only)")
	validatorCredentialsSetCmd.Flags().String("genesis-validators-root", "", "Genesis validators root to use for signing (offline only)")
}

func validatorCredentialsSetBindings() {
	if err := viper.BindPFlag("prepare-offline", validatorCredentialsSetCmd.Flags().Lookup("prepare-offline")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("validator", validatorCredentialsSetCmd.Flags().Lookup("validator")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("signed-operation", validatorCredentialsSetCmd.Flags().Lookup("signed-operation")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("withdrawal-address", validatorCredentialsSetCmd.Flags().Lookup("withdrawal-address")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("json", validatorCredentialsSetCmd.Flags().Lookup("json")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("offline", validatorCredentialsSetCmd.Flags().Lookup("offline")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("fork-version", validatorCredentialsSetCmd.Flags().Lookup("fork-version")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("genesis-validators-root", validatorCredentialsSetCmd.Flags().Lookup("genesis-validators-root")); err != nil {
		panic(err)
	}
}
