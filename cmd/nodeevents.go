// Copyright © 2019 Weald Technology Trading
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
	nodeevents "github.com/wealdtech/ethdo/cmd/node/events"
)

var nodeEventsCmd = &cobra.Command{
	Use:   "events",
	Short: "Report events from a node",
	Long: `Report events from a node.  For example:

    ethdo node events --events=head,chain_reorg.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		res, err := nodeevents.Run(cmd)
		if err != nil {
			return err
		}
		if res != "" {
			fmt.Println(res)
		}
		return nil
	},
}

func init() {
	nodeCmd.AddCommand(nodeEventsCmd)
	nodeFlags(nodeEventsCmd)
	nodeEventsCmd.Flags().StringSlice("topics", nil, "The topics of events for which to listen (attestation,block,chain_reorg,finalized_checkpoint,head,voluntary_exit)")
}

func nodeEventsBindings(cmd *cobra.Command) {
	if err := viper.BindPFlag("topics", cmd.Flags().Lookup("topics")); err != nil {
		panic(err)
	}
}
