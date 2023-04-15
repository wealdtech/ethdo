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
	epochsummary "github.com/wealdtech/ethdo/cmd/epoch/summary"
)

var epochSummaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Obtain summary information about an epoch",
	Long: `Obtain summary information about an epoch.  For example:

    ethdo epoch summary --epoch=12345

In quiet mode this will return 0 if information for the epoch is found, otherwise 1.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		res, err := epochsummary.Run(cmd)
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
	epochCmd.AddCommand(epochSummaryCmd)
	epochFlags(epochSummaryCmd)
}

func epochSummaryBindings(cmd *cobra.Command) {
	epochBindings(cmd)
}
