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
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	validatorexpectation "github.com/wealdtech/ethdo/cmd/validator/expectation"
)

var validatorExpectationCmd = &cobra.Command{
	Use:   "expectation",
	Short: "Calculate expectation for individual validators",
	Long: `Calculate expectation for individual validators.  For example:

    ethdo validator expectation`,
	RunE: func(cmd *cobra.Command, args []string) error {
		res, err := validatorexpectation.Run(cmd)
		if err != nil {
			return err
		}
		if viper.GetBool("quiet") {
			return nil
		}
		res = strings.TrimRight(res, "\n")
		fmt.Println(res)
		return nil
	},
}

func init() {
	validatorCmd.AddCommand(validatorExpectationCmd)
	validatorFlags(validatorExpectationCmd)
	validatorExpectationCmd.Flags().Int64("validators", 1, "Number of validators")
}

func validatorExpectationBindings(cmd *cobra.Command) {
	if err := viper.BindPFlag("validators", cmd.Flags().Lookup("validators")); err != nil {
		panic(err)
	}
}
