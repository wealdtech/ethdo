// Copyright © 2020, 2022 Weald Technology Trading
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
	"context"
	"fmt"
	"os"
	"time"

	eth2client "github.com/attestantio/go-eth2-client"
	"github.com/attestantio/go-eth2-client/api"
	spec "github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wealdtech/ethdo/util"
)

var chainInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Obtain information about a chain",
	Long: `Obtain information about a chain.  For example:

    ethdo chain info

In quiet mode this will return 0 if the chain information can be obtained, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		eth2Client, err := util.ConnectToBeaconNode(ctx, &util.ConnectOpts{
			Address:       viper.GetString("connection"),
			Timeout:       viper.GetDuration("timeout"),
			AllowInsecure: viper.GetBool("allow-insecure-connections"),
			LogFallback:   !viper.GetBool("quiet"),
		})
		errCheck(err, "Failed to connect to Ethereum 2 beacon node")

		specResponse, err := eth2Client.(eth2client.SpecProvider).Spec(ctx, &api.SpecOpts{})
		errCheck(err, "Failed to obtain beacon chain specification")

		genesisResponse, err := eth2Client.(eth2client.GenesisProvider).Genesis(ctx, &api.GenesisOpts{})
		errCheck(err, "Failed to obtain beacon chain genesis")

		forkResponse, err := eth2Client.(eth2client.ForkProvider).Fork(ctx, &api.ForkOpts{State: "head"})
		errCheck(err, "Failed to obtain current fork")

		if viper.GetBool("quiet") {
			os.Exit(_exitSuccess)
		}

		if genesisResponse.Data.GenesisTime.Unix() == 0 {
			fmt.Println("Genesis time: undefined")
		} else {
			fmt.Printf("Genesis time: %s\n", genesisResponse.Data.GenesisTime.Format(time.UnixDate))
			outputIf(viper.GetBool("verbose"), fmt.Sprintf("Genesis timestamp: %v", genesisResponse.Data.GenesisTime.Unix()))
		}
		fmt.Printf("Genesis validators root: %#x\n", genesisResponse.Data.GenesisValidatorsRoot)
		fmt.Printf("Genesis fork version: %#x\n", specResponse.Data["GENESIS_FORK_VERSION"].(spec.Version))
		fmt.Printf("Current fork version: %#x\n", forkResponse.Data.CurrentVersion)
		if viper.GetBool("verbose") {
			forkData := &spec.ForkData{
				CurrentVersion:        forkResponse.Data.CurrentVersion,
				GenesisValidatorsRoot: genesisResponse.Data.GenesisValidatorsRoot,
			}
			forkDataRoot, err := forkData.HashTreeRoot()
			if err == nil {
				var forkDigest spec.ForkDigest
				copy(forkDigest[:], forkDataRoot[:])
				fmt.Printf("Fork digest: %#x\n", forkDigest)
			}
		}
		fmt.Printf("Seconds per slot: %d\n", int(specResponse.Data["SECONDS_PER_SLOT"].(time.Duration).Seconds()))
		fmt.Printf("Slots per epoch: %d\n", specResponse.Data["SLOTS_PER_EPOCH"].(uint64))

		os.Exit(_exitSuccess)
	},
}

func init() {
	chainCmd.AddCommand(chainInfoCmd)
	chainFlags(chainInfoCmd)
}

func chainInfoBindings(_ *cobra.Command) {
}
