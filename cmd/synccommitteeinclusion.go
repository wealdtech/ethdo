// Copyright © 2022 Weald Technology Trading.
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
	synccommitteeinclusion "github.com/wealdtech/ethdo/cmd/synccommittee/inclusion"
)

var synccommitteeInclusionCmd = &cobra.Command{
	Use:   "inclusion",
	Short: "Obtain sync committee inclusion data for a validator",
	Long: `Obtain sync committee inclusion data for a validator.  For example:

    ethdo synccommittee inclusion --epoch=12345 --index=11111

In quiet mode this will return 0 if the validator was in the sync committee, otherwise 1.

epoch can be a specific epoch; If not supplied all slots for the current sync committee period will be provided`,
	RunE: func(cmd *cobra.Command, args []string) error {
		res, err := synccommitteeinclusion.Run(cmd)
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
	synccommitteeCmd.AddCommand(synccommitteeInclusionCmd)
	synccommitteeFlags(synccommitteeInclusionCmd)
	synccommitteeInclusionCmd.Flags().String("epoch", "", "the epoch for which to fetch sync committee inclusion")
	synccommitteeInclusionCmd.Flags().String("validator", "", "the index, public key, or acount of the validator")
}

func synccommitteeInclusionBindings(cmd *cobra.Command) {
	if err := viper.BindPFlag("epoch", cmd.Flags().Lookup("epoch")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("validator", cmd.Flags().Lookup("validator")); err != nil {
		panic(err)
	}
}
