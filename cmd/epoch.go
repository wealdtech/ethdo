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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// epochCmd represents the epoch command.
var epochCmd = &cobra.Command{
	Use:   "epoch",
	Short: "Obtain information about Ethereum 2 epochs",
	Long:  "Obtain information about Ethereum 2 epochs",
}

func init() {
	RootCmd.AddCommand(epochCmd)
}

func epochFlags(_ *cobra.Command) {
	epochSummaryCmd.Flags().String("epoch", "", "the epoch for which to obtain information (default current, can be 'current', 'last' or a number)")
}

func epochBindings(cmd *cobra.Command) {
	if err := viper.BindPFlag("epoch", cmd.Flags().Lookup("epoch")); err != nil {
		panic(err)
	}
}
