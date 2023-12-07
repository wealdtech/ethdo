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
	"strconv"
	"strings"
	"time"

	eth2client "github.com/attestantio/go-eth2-client"
	"github.com/attestantio/go-eth2-client/api"
	apiv1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	standardchaintime "github.com/wealdtech/ethdo/services/chaintime/standard"
	"github.com/wealdtech/ethdo/util"
	string2eth "github.com/wealdtech/go-string2eth"
)

var chainStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Obtain status about a chain",
	Long: `Obtain status about a chain.  For example:

    ethdo chain status

In quiet mode this will return 0 if the chain status can be obtained, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		eth2Client, err := util.ConnectToBeaconNode(ctx, &util.ConnectOpts{
			Address:       viper.GetString("connection"),
			Timeout:       viper.GetDuration("timeout"),
			AllowInsecure: viper.GetBool("allow-insecure-connections"),
			LogFallback:   !viper.GetBool("quiet"),
		})
		errCheck(err, "Failed to connect to Ethereum 2 beacon node")

		chainTime, err := standardchaintime.New(ctx,
			standardchaintime.WithGenesisProvider(eth2Client.(eth2client.GenesisProvider)),
			standardchaintime.WithSpecProvider(eth2Client.(eth2client.SpecProvider)),
		)
		errCheck(err, "Failed to configure chaintime service")

		finalityProvider, isProvider := eth2Client.(eth2client.FinalityProvider)
		assert(isProvider, "beacon node does not provide finality; cannot report on chain status")
		finalityResponse, err := finalityProvider.Finality(ctx, &api.FinalityOpts{
			State: "head",
		})
		errCheck(err, "Failed to obtain finality information")
		finality := finalityResponse.Data

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

		if viper.GetBool("verbose") {
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
		if viper.GetBool("verbose") {
			distance := epoch - finality.Justified.Epoch
			res.WriteString("Justified epoch distance: ")
			res.WriteString(fmt.Sprintf("%d", distance))
			res.WriteString("\n")
		}

		res.WriteString("Finalized epoch: ")
		res.WriteString(fmt.Sprintf("%d", finality.Finalized.Epoch))
		res.WriteString("\n")
		if viper.GetBool("verbose") {
			distance := epoch - finality.Finalized.Epoch
			res.WriteString("Finalized epoch distance: ")
			res.WriteString(fmt.Sprintf("%d", distance))
			res.WriteString("\n")
		}

		if viper.GetBool("verbose") {
			validatorsProvider, isProvider := eth2Client.(eth2client.ValidatorsProvider)
			if isProvider {
				validatorsResponse, err := validatorsProvider.Validators(ctx, &api.ValidatorsOpts{State: "head"})
				errCheck(err, "Failed to obtain validators information")
				// Stats of inteest.
				totalBalance := phase0.Gwei(0)
				activeEffectiveBalance := phase0.Gwei(0)
				validatorCount := make(map[apiv1.ValidatorState]int)
				for _, validator := range validatorsResponse.Data {
					validatorCount[validator.Status]++
					totalBalance += validator.Balance
					if validator.Status.IsActive() {
						activeEffectiveBalance += validator.Validator.EffectiveBalance
					}
				}
				res.WriteString(fmt.Sprintf("Total balance: %s\n", string2eth.GWeiToString(uint64(totalBalance), true)))
				res.WriteString(fmt.Sprintf("Active effective balance: %s\n", string2eth.GWeiToString(uint64(activeEffectiveBalance), true)))
				res.WriteString("Validator states:\n")
				res.WriteString(fmt.Sprintf("  Pending: %d\n", validatorCount[apiv1.ValidatorStatePendingInitialized]))
				res.WriteString(fmt.Sprintf("  Activating: %d\n", validatorCount[apiv1.ValidatorStatePendingQueued]))
				res.WriteString(fmt.Sprintf("  Active: %d\n", validatorCount[apiv1.ValidatorStateActiveOngoing]+validatorCount[apiv1.ValidatorStateActiveSlashed]))
				res.WriteString(fmt.Sprintf("  Exiting: %d\n", validatorCount[apiv1.ValidatorStateActiveExiting]))
				res.WriteString(fmt.Sprintf("  Exited: %d\n", validatorCount[apiv1.ValidatorStateExitedUnslashed]+validatorCount[apiv1.ValidatorStateExitedSlashed]+validatorCount[apiv1.ValidatorStateWithdrawalPossible]+validatorCount[apiv1.ValidatorStateWithdrawalDone]))
				res.WriteString(fmt.Sprintf("  Unknown: %d\n", validatorCount[apiv1.ValidatorStateUnknown]))
			}
		}

		if epoch >= chainTime.AltairInitialEpoch() {
			period := chainTime.SlotToSyncCommitteePeriod(slot)
			periodStartEpoch := chainTime.FirstEpochOfSyncPeriod(period)
			periodStartSlot := chainTime.FirstSlotOfEpoch(periodStartEpoch)
			nextPeriod := period + 1
			nextPeriodStartEpoch := chainTime.FirstEpochOfSyncPeriod(nextPeriod)
			periodEndEpoch := nextPeriodStartEpoch - 1
			periodEndSlot := chainTime.FirstSlotOfEpoch(periodEndEpoch+1) - 1
			nextPeriodTimestamp := chainTime.StartOfEpoch(nextPeriodStartEpoch)

			res.WriteString("Sync committee period: ")
			res.WriteString(strconv.FormatUint(period, 10))
			res.WriteString("\n")

			if viper.GetBool("verbose") {
				res.WriteString("Sync committee epochs: ")
				res.WriteString(fmt.Sprintf("%d", periodStartEpoch))
				res.WriteString("-")
				res.WriteString(fmt.Sprintf("%d", periodEndEpoch))
				res.WriteString("\n")

				res.WriteString("Sync committee slots: ")
				res.WriteString(fmt.Sprintf("%d", periodStartSlot))
				res.WriteString("-")
				res.WriteString(fmt.Sprintf("%d", periodEndSlot))
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
