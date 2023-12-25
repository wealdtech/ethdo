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

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	synccommitteemembers "github.com/wealdtech/ethdo/cmd/synccommittee/members"
)

var synccommitteeMembersCmd = &cobra.Command{
	Use:   "members",
	Short: "Obtain information about members of a synccommittee",
	Long: `Obtain information about members of a synccommittee.  For example:

    ethdo synccommittee members --epoch=12345

In quiet mode this will return 0 if the synccommittee members are found, otherwise 1.

epoch can be a specific epoch.  period can be 'current' for the current sync period or 'next' for the next sync period`,
	RunE: func(cmd *cobra.Command, args []string) error {
		res, err := synccommitteemembers.Run(cmd)
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
	synccommitteeCmd.AddCommand(synccommitteeMembersCmd)
	synccommitteeFlags(synccommitteeMembersCmd)
	synccommitteeMembersCmd.Flags().String("epoch", "", "the epoch for which to fetch sync committees")
	synccommitteeMembersCmd.Flags().String("period", "", "the sync committee period for which to fetch sync committees ('current', 'next')")
}

func synccommitteeMembersBindings(cmd *cobra.Command) {
	if err := viper.BindPFlag("epoch", cmd.Flags().Lookup("epoch")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("period", cmd.Flags().Lookup("period")); err != nil {
		panic(err)
	}
}
