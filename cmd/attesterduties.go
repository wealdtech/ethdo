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
	attesterduties "github.com/wealdtech/ethdo/cmd/attester/duties"
)

var attesterDutiesCmd = &cobra.Command{
	Use:   "duties",
	Short: "Obtain information about duties of an attester",
	Long: `Obtain information about dutes of an attester.  For example:

    ethdo attester duties --validator=Validators/00001 --epoch=12345

In quiet mode this will return 0 if a duty from the attester is found, otherwise 1.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		res, err := attesterduties.Run(cmd)
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
	attesterCmd.AddCommand(attesterDutiesCmd)
	attesterFlags(attesterDutiesCmd)
	attesterDutiesCmd.Flags().String("epoch", "head", "the epoch for which to obtain the duties")
	attesterDutiesCmd.Flags().String("validator", "", "the index, public key, or acount of the validator")
}

func attesterDutiesBindings(cmd *cobra.Command) {
	if err := viper.BindPFlag("epoch", cmd.Flags().Lookup("epoch")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("validator", cmd.Flags().Lookup("validator")); err != nil {
		panic(err)
	}
}
