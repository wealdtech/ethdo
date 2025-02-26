// Copyright © 2025 Weald Technology Trading
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
	blocktrail "github.com/wealdtech/ethdo/cmd/block/trail"
)

var blockTrailCmd = &cobra.Command{
	Use:   "trail",
	Short: "Trail back in the chain from a given block.",
	Long: `Trail back in the chain for a given block.  For example:

    ethdo block trail --blockid=12345 --target=finalized

In quiet mode this will return 0 if the block trail ends up at the finalized state, otherwise 1.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		res, err := blocktrail.Run(cmd)
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
	blockCmd.AddCommand(blockTrailCmd)
	blockFlags(blockTrailCmd)
	blockTrailCmd.Flags().String("blockid", "head", "the ID of the block to fetch")
	blockTrailCmd.Flags().String("target", "justified", "the target block (block number, hash, justified or finalized)")
	blockTrailCmd.Flags().Int("max-blocks", 16384, "the maximum number of blocks to look at before halting")
}

func blockTrailBindings(cmd *cobra.Command) {
	if err := viper.BindPFlag("blockid", cmd.Flags().Lookup("blockid")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("target", cmd.Flags().Lookup("target")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("max-blocks", cmd.Flags().Lookup("max-blocks")); err != nil {
		panic(err)
	}
}
