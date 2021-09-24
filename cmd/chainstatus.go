// Copyright © 2020, 2021 Weald Technology Trading.
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
	"strings"
	"time"

	eth2client "github.com/attestantio/go-eth2-client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	standardchaintime "github.com/wealdtech/ethdo/services/chaintime/standard"
	"github.com/wealdtech/ethdo/util"
)

var chainStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Obtain status about a chain",
	Long: `Obtain status about a chain.  For example:

    ethdo chain status

In quiet mode this will return 0 if the chain status can be obtained, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		eth2Client, err := util.ConnectToBeaconNode(ctx, viper.GetString("connection"), viper.GetDuration("timeout"), viper.GetBool("allow-insecure-connections"))
		errCheck(err, "Failed to connect to Ethereum 2 beacon node")

		chainTime, err := standardchaintime.New(ctx,
			standardchaintime.WithGenesisTimeProvider(eth2Client.(eth2client.GenesisTimeProvider)),
			standardchaintime.WithForkScheduleProvider(eth2Client.(eth2client.ForkScheduleProvider)),
			standardchaintime.WithSpecProvider(eth2Client.(eth2client.SpecProvider)),
		)
		errCheck(err, "Failed to configure chaintime service")

		finalityProvider, isProvider := eth2Client.(eth2client.FinalityProvider)
		assert(isProvider, "beacon node does not provide finality; cannot report on chain status")
		finality, err := finalityProvider.Finality(ctx, "head")
		errCheck(err, "Failed to obtain finality information")

		slot := chainTime.CurrentSlot()

		nextSlot := slot + 1
		nextSlotTimestamp := chainTime.StartOfSlot(nextSlot)

		epoch := chainTime.CurrentEpoch()
		epochStartSlot := chainTime.FirstSlotOfEpoch(epoch)
		epochEndSlot := chainTime.FirstSlotOfEpoch(epoch+1) - 1

		nextEpoch := epoch + 1
		nextEpochStartSlot := chainTime.FirstSlotOfEpoch(nextEpoch)
		nextEpochTimestamp := chainTime.StartOfEpoch(nextEpoch)

		res := strings.Builder{}

		res.WriteString("Current slot: ")
		res.WriteString(fmt.Sprintf("%d", slot))
		res.WriteString("\n")

		res.WriteString("Current epoch: ")
		res.WriteString(fmt.Sprintf("%d", epoch))
		res.WriteString("\n")

		if verbose {
			res.WriteString("Epoch slots: ")
			res.WriteString(fmt.Sprintf("%d", epochStartSlot))
			res.WriteString("-")
			res.WriteString(fmt.Sprintf("%d", epochEndSlot))
			res.WriteString("\n")
		}

		res.WriteString("Time until next slot: ")
		res.WriteString(time.Until(nextSlotTimestamp).Round(time.Second).String())
		res.WriteString("\n")

		res.WriteString("Time until next epoch: ")
		res.WriteString(time.Until(nextEpochTimestamp).Round(time.Second).String())
		res.WriteString("\n")

		res.WriteString("Slots until next epoch: ")
		res.WriteString(fmt.Sprintf("%d", nextEpochStartSlot-slot))
		res.WriteString("\n")

		res.WriteString("Justified epoch: ")
		res.WriteString(fmt.Sprintf("%d", finality.Justified.Epoch))
		res.WriteString("\n")
		if verbose {
			distance := epoch - finality.Justified.Epoch
			res.WriteString("Justified epoch distance: ")
			res.WriteString(fmt.Sprintf("%d", distance))
			res.WriteString("\n")
		}

		res.WriteString("Finalized epoch: ")
		res.WriteString(fmt.Sprintf("%d", finality.Finalized.Epoch))
		res.WriteString("\n")
		if verbose {
			distance := epoch - finality.Finalized.Epoch
			res.WriteString("Finalized epoch distance: ")
			res.WriteString(fmt.Sprintf("%d", distance))
			res.WriteString("\n")
		}

		if epoch >= chainTime.AltairInitialEpoch() {
			period := chainTime.SlotToSyncCommitteePeriod(slot)
			periodStartEpoch := chainTime.FirstEpochOfSyncPeriod(period)
			nextPeriod := period + 1
			nextPeriodStartEpoch := chainTime.FirstEpochOfSyncPeriod(nextPeriod)
			periodEndEpoch := nextPeriodStartEpoch - 1
			nextPeriodTimestamp := chainTime.StartOfEpoch(nextPeriodStartEpoch)

			res.WriteString("Sync committee period: ")
			res.WriteString(fmt.Sprintf("%d", period))
			res.WriteString("\n")

			if verbose {
				res.WriteString("Sync committee epochs: ")
				res.WriteString(fmt.Sprintf("%d", periodStartEpoch))
				res.WriteString("-")
				res.WriteString(fmt.Sprintf("%d", periodEndEpoch))
				res.WriteString("\n")

				res.WriteString("Time until next sync committee period: ")
				res.WriteString(time.Until(nextPeriodTimestamp).Round(time.Second).String())
				res.WriteString("\n")
			}
		}

		fmt.Print(res.String())

		os.Exit(_exitSuccess)
	},
}

func init() {
	chainCmd.AddCommand(chainStatusCmd)
	chainFlags(chainStatusCmd)
}
