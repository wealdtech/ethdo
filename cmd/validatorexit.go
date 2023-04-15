// Copyright © 2020 Weald Technology Trading
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
	validatorexit "github.com/wealdtech/ethdo/cmd/validator/exit"
)

var validatorExitCmd = &cobra.Command{
	Use:   "exit",
	Short: "Send an exit request for a validator",
	Long: `Send an exit request for a validator.  For example:

    ethdo validator exit --validator=12345

The validator and key can be specified in one of a number of ways:

  - mnemonic and path to the validator using --mnemonic and --path
  - mnemonic and validator index or public key using --mnemonic and --validator
  - validator private key using --private-key
  - validator account using --validator

In quiet mode this will return 0 if the exit operation has been generated (and successfully broadcast if online), otherwise 1.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		res, err := validatorexit.Run(cmd)
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
	validatorCmd.AddCommand(validatorExitCmd)
	validatorFlags(validatorExitCmd)
	validatorExitCmd.Flags().Int64("epoch", -1, "Epoch at which to exit (defaults to current epoch)")
	validatorExitCmd.Flags().Bool("prepare-offline", false, "Create files for offline use")
	validatorExitCmd.Flags().String("validator", "", "Validator to exit")
	validatorExitCmd.Flags().String("signed-operation", "", "Use pre-defined JSON signed operation as created by --json to transmit the exit operation (reads from exit-operation.json if not present)")
	validatorExitCmd.Flags().Bool("offline", false, "Do not attempt to connect to a beacon node to obtain information for the operation")
	validatorExitCmd.Flags().String("fork-version", "", "Fork version to use for signing (overrides fetching from beacon node)")
	validatorExitCmd.Flags().String("genesis-validators-root", "", "Genesis validators root to use for signing (overrides fetching from beacon node)")
}

func validatorExitBindings() {
	if err := viper.BindPFlag("epoch", validatorExitCmd.Flags().Lookup("epoch")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("prepare-offline", validatorExitCmd.Flags().Lookup("prepare-offline")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("validator", validatorExitCmd.Flags().Lookup("validator")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("signed-operation", validatorExitCmd.Flags().Lookup("signed-operation")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("offline", validatorExitCmd.Flags().Lookup("offline")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("fork-version", validatorExitCmd.Flags().Lookup("fork-version")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("genesis-validators-root", validatorExitCmd.Flags().Lookup("genesis-validators-root")); err != nil {
		panic(err)
	}
}
