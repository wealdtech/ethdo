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
	chaineth1votes "github.com/wealdtech/ethdo/cmd/chain/eth1votes"
)

var chainEth1VotesCmd = &cobra.Command{
	Use:   "eth1votes",
	Short: "Show chain execution votes",
	Long: `Show beacon chain execution votes.  For example:

    ethdo chain eth1votes

Note that this will fetch the votes made in blocks up to the end of the provided epoch.

In quiet mode this will return 0 if there is a majority for the votes, otherwise 1.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		res, err := chaineth1votes.Run(cmd)
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
	chainCmd.AddCommand(chainEth1VotesCmd)
	chainFlags(chainEth1VotesCmd)
	chainEth1VotesCmd.Flags().String("epoch", "", "epoch for which to fetch the votes")
	chainEth1VotesCmd.Flags().String("period", "", "period for which to fetch the votes")
}

func chainEth1VotesBindings(cmd *cobra.Command) {
	if err := viper.BindPFlag("epoch", cmd.Flags().Lookup("epoch")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("period", cmd.Flags().Lookup("period")); err != nil {
		panic(err)
	}
}
