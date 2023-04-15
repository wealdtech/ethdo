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
	validatorduties "github.com/wealdtech/ethdo/cmd/validator/duties"
)

var validatorDutiesCmd = &cobra.Command{
	Use:   "duties",
	Short: "List known duties for a validator",
	Long: `List known duties for a validator. For example:

    ethdo validator duties --account=Validators/One

Attester duties are known for the current and next epoch.  Proposer duties are known for the current epoch.

In quiet mode this will return 0 if the duties have been obtained, otherwise 1.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		res, err := validatorduties.Run(cmd)
		if err != nil {
			return err
		}
		if viper.GetBool("quiet") {
			return nil
		}
		fmt.Print(res)
		return nil
	},
}

func init() {
	validatorCmd.AddCommand(validatorDutiesCmd)
	validatorFlags(validatorDutiesCmd)
	validatorDutiesCmd.Flags().String("pubkey", "", "validator public key for duties")
	validatorDutiesCmd.Flags().String("index", "", "validator index for duties")
}

func validatorDutiesBindings(cmd *cobra.Command) {
	if err := viper.BindPFlag("pubkey", cmd.Flags().Lookup("pubkey")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("index", cmd.Flags().Lookup("index")); err != nil {
		panic(err)
	}
}
