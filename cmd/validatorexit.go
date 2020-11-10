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

    ethdo validator exit --account=primary/validator --passphrase=secret

In quiet mode this will return 0 if the transaction has been generated, otherwise 1.`,
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
	validatorExitCmd.Flags().String("key", "", "Private key if validator not known by ethdo")
	validatorExitCmd.Flags().String("exit", "", "Use pre-defined JSON data as created by --json to exit")
	validatorExitCmd.Flags().Bool("json", false, "Generate JSON data for an exit; do not broadcast to network")
}

func validatorExitBindings() {
	if err := viper.BindPFlag("epoch", validatorExitCmd.Flags().Lookup("epoch")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("key", validatorExitCmd.Flags().Lookup("key")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("exit", validatorExitCmd.Flags().Lookup("exit")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("json", validatorExitCmd.Flags().Lookup("json")); err != nil {
		panic(err)
	}
}
