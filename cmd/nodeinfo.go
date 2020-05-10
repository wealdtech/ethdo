// Copyright © 2020 Weald Technology Trading
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
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/wealdtech/ethdo/grpc"
)

var nodeInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Obtain information about a node",
	Long: `Obtain information about a node.  For example:

    ethdo node info

In quiet mode this will return 0 if the node information can be obtained, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := connect()
		errCheck(err, "Failed to obtain connection to Ethereum 2 beacon chain node")
		config, err := grpc.FetchChainConfig(eth2GRPCConn)
		errCheck(err, "Failed to obtain beacon chain configuration")

		genesisTime, err := grpc.FetchGenesisTime(eth2GRPCConn)
		errCheck(err, "Failed to obtain genesis time")

		if quiet {
			os.Exit(_exitSuccess)
		}

		if verbose {
			version, metadata, err := grpc.FetchVersion(eth2GRPCConn)
			errCheck(err, "Failed to obtain version")
			fmt.Printf("Version: %s\n", version)
			if metadata != "" {
				fmt.Printf("Metadata: %s\n", metadata)
			}
		}
		syncing, err := grpc.FetchSyncing(eth2GRPCConn)
		errCheck(err, "Failed to obtain syncing state")
		fmt.Printf("Syncing: %v\n", syncing)

		slot := timestampToSlot(genesisTime.Unix(), time.Now().Unix(), config["SecondsPerSlot"].(uint64))
		fmt.Printf("Current slot: %d\n", slot)
		fmt.Printf("Current epoch: %d\n", slot/config["SlotsPerEpoch"].(uint64))
		outputIf(verbose, fmt.Sprintf("Genesis timestamp: %v", genesisTime.Unix()))

		os.Exit(_exitSuccess)
	},
}

func init() {
	nodeCmd.AddCommand(nodeInfoCmd)
	nodeFlags(nodeInfoCmd)
}
