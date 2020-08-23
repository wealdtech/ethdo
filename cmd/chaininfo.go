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

var chainInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Obtain information about a chain",
	Long: `Obtain information about a chain.  For example:

    ethdo chain info

In quiet mode this will return 0 if the chain information can be obtained, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := connect()
		errCheck(err, "Failed to obtain connection to Ethereum 2 beacon chain node")
		config, err := grpc.FetchChainConfig(eth2GRPCConn)
		errCheck(err, "Failed to obtain beacon chain configuration")

		genesisTime, err := grpc.FetchGenesisTime(eth2GRPCConn)
		errCheck(err, "Failed to obtain genesis time")

		genesisValidatorsRoot, err := grpc.FetchGenesisValidatorsRoot(eth2GRPCConn)
		errCheck(err, "Failed to obtain genesis validators root")

		if quiet {
			os.Exit(_exitSuccess)
		}

		if genesisTime.Unix() == 0 {
			fmt.Println("Genesis time: undefined")
		} else {
			fmt.Printf("Genesis time: %s\n", genesisTime.Format(time.UnixDate))
			outputIf(verbose, fmt.Sprintf("Genesis timestamp: %v", genesisTime.Unix()))
		}
		fmt.Printf("Genesis validators root: %#x\n", genesisValidatorsRoot)
		fmt.Printf("Genesis fork version: %#x\n", config["GenesisForkVersion"].([]byte))
		fmt.Printf("Seconds per slot: %d\n", config["SecondsPerSlot"].(uint64))
		fmt.Printf("Slots per epoch: %d\n", config["SlotsPerEpoch"].(uint64))

		os.Exit(_exitSuccess)
	},
}

func init() {
	chainCmd.AddCommand(chainInfoCmd)
	chainFlags(chainInfoCmd)
}

func timestampToSlot(genesis int64, timestamp int64, secondsPerSlot uint64) uint64 {
	if timestamp < genesis {
		return 0
	}
	return uint64(timestamp-genesis) / secondsPerSlot
}

func slotToTimestamp(genesis int64, slot uint64, secondsPerSlot uint64) int64 {
	return genesis + int64(slot*secondsPerSlot)
}

func epochToTimestamp(genesis int64, slot uint64, secondsPerSlot uint64, slotsPerEpoch uint64) int64 {
	return genesis + int64(slot*secondsPerSlot*slotsPerEpoch)
}
