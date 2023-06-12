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
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	validatoryield "github.com/wealdtech/ethdo/cmd/validator/yield"
)

var validatorYieldCmd = &cobra.Command{
	Use:   "yield",
	Short: "Calculate yield for validators",
	Long: `Calculate yield for validators.  For example:

    ethdo validator yield

It is important to understand the yield is both probabilistic and dependent on network conditions.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		res, err := validatoryield.Run(cmd)
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
	validatorCmd.AddCommand(validatorYieldCmd)
	validatorFlags(validatorYieldCmd)
	validatorYieldCmd.Flags().String("validators", "", "Number of active validators (default fetches from chain)")
	validatorYieldCmd.Flags().String("epoch", "", "Epoch at which to calculate yield")
}

func validatorYieldBindings(cmd *cobra.Command) {
	if err := viper.BindPFlag("validators", cmd.Flags().Lookup("validators")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("epoch", cmd.Flags().Lookup("epoch")); err != nil {
		panic(err)
	}
}
