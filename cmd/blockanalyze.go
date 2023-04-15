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
	blockanalyze "github.com/wealdtech/ethdo/cmd/block/analyze"
)

var blockAnalyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze a block",
	Long: `Analyze the contents of a block.  For example:

    ethdo block analyze --blockid=12345

In quiet mode this will return 0 if the block information is present and not skipped, otherwise 1.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		res, err := blockanalyze.Run(cmd)
		if err != nil {
			return err
		}
		if viper.GetBool("quiet") {
			return nil
		}
		if res != "" {
			fmt.Print(res)
		}
		return nil
	},
}

func init() {
	blockCmd.AddCommand(blockAnalyzeCmd)
	blockFlags(blockAnalyzeCmd)
	blockAnalyzeCmd.Flags().String("blockid", "head", "the ID of the block to fetch")
	blockAnalyzeCmd.Flags().Bool("stream", false, "continually stream blocks as they arrive")
}

func blockAnalyzeBindings(cmd *cobra.Command) {
	if err := viper.BindPFlag("blockid", cmd.Flags().Lookup("blockid")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("stream", cmd.Flags().Lookup("stream")); err != nil {
		panic(err)
	}
}
