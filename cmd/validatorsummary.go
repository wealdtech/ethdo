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
	validatorsummary "github.com/wealdtech/ethdo/cmd/validator/summary"
)

var validatorSummaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Obtain summary information about validator(s) in an epoch",
	Long: `Obtain summary information about one or more validators in an epoch.  For example:

    ethdo validator summary --validators=1,2,3 --epoch=12345

In quiet mode this will return 0 if information for the epoch is found, otherwise 1.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		res, err := validatorsummary.Run(cmd)
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
	validatorCmd.AddCommand(validatorSummaryCmd)
	validatorFlags(validatorSummaryCmd)
	validatorSummaryCmd.Flags().String("epoch", "", "the epoch for which to obtain information ()")
	validatorSummaryCmd.Flags().StringSlice("validators", nil, "the list of validators for which to obtain information")
}

func validatorSummaryBindings(cmd *cobra.Command) {
	validatorBindings()
	if err := viper.BindPFlag("epoch", cmd.Flags().Lookup("epoch")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("validators", cmd.Flags().Lookup("validators")); err != nil {
		panic(err)
	}
}
