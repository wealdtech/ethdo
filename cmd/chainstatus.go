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

var chainStatusSlot bool

var chainStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Obtain status about a chain",
	Long: `Obtain status about a chain.  For example:

    ethdo chain status

In quiet mode this will return 0 if the chain status can be obtained, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := connect()
		errCheck(err, "Failed to obtain connection to Ethereum 2 beacon chain node")
		config, err := grpc.FetchChainConfig(eth2GRPCConn)
		errCheck(err, "Failed to obtain beacon chain configuration")

		genesisTime, err := grpc.FetchGenesis(eth2GRPCConn)
		errCheck(err, "Failed to obtain genesis time")

		info, err := grpc.FetchChainInfo(eth2GRPCConn)
		errCheck(err, "Failed to obtain chain info")

		if quiet {
			os.Exit(_exitSuccess)
		}

		slot := timestampToSlot(genesisTime.Unix(), time.Now().Unix(), config["SecondsPerSlot"].(uint64))
		if chainStatusSlot {
			fmt.Printf("Current slot:\t\t%d\n", slot)
			fmt.Printf("Justified slot:\t\t%d", info.GetJustifiedSlot())
			if verbose {
				distance := slot - info.GetJustifiedSlot()
				fmt.Printf(" (%d)", distance)
			}
			fmt.Printf("\n")
			fmt.Printf("Finalized slot:\t\t%d", info.GetFinalizedSlot())
			if verbose {
				distance := slot - info.GetFinalizedSlot()
				fmt.Printf(" (%d)", distance)
			}
			fmt.Printf("\n")
			outputIf(verbose, fmt.Sprintf("Prior justified slot:\t%v (%d)", info.GetPreviousJustifiedSlot(), slot-info.GetPreviousJustifiedSlot()))
		} else {
			slotsPerEpoch := config["SlotsPerEpoch"].(uint64)
			fmt.Printf("Current epoch:\t\t%d\n", slot/slotsPerEpoch)
			fmt.Printf("Justified epoch:\t%d", info.GetJustifiedSlot()/slotsPerEpoch)
			if verbose {
				distance := (slot - info.GetJustifiedSlot()) / slotsPerEpoch
				fmt.Printf(" (%d)", distance)
			}
			fmt.Printf("\n")
			fmt.Printf("Finalized epoch:\t%d", info.GetFinalizedSlot()/slotsPerEpoch)
			if verbose {
				distance := (slot - info.GetFinalizedSlot()) / slotsPerEpoch
				fmt.Printf(" (%d)", distance)
			}
			fmt.Printf("\n")
			outputIf(verbose, fmt.Sprintf("Prior justified epoch:\t%v (%d)", info.GetPreviousJustifiedSlot()/slotsPerEpoch, (slot-info.GetPreviousJustifiedSlot())/slotsPerEpoch))
		}

		os.Exit(_exitSuccess)
	},
}

func init() {
	chainCmd.AddCommand(chainStatusCmd)
	chainFlags(chainStatusCmd)
	chainStatusCmd.Flags().BoolVar(&chainStatusSlot, "slot", false, "Print slot-based values")

}
