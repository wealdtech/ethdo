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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// chainVerifyCmd represents the chain verify command.
var chainVerifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify a beacon chain signature",
	Long:  "Verify the signature for a given beacon chain structure is correct",
}

func init() {
	chainCmd.AddCommand(chainVerifyCmd)
}

func chainVerifyFlags(cmd *cobra.Command) {
	chainFlags(cmd)
	cmd.Flags().String("validator", "", "The account, public key or index of the validator")
	cmd.Flags().String("data", "", "The data to verify, as a JSON structure")
}

func chainVerifyBindings(cmd *cobra.Command) {
	if err := viper.BindPFlag("validator", cmd.Flags().Lookup("validator")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("data", cmd.Flags().Lookup("data")); err != nil {
		panic(err)
	}
}
