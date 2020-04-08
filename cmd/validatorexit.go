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
	"context"
	"fmt"
	"os"
	"time"

	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	ssz "github.com/prysmaticlabs/go-ssz"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wealdtech/ethdo/grpc"
)

var validatorExitEpoch int64

var validatorExitCmd = &cobra.Command{
	Use:   "exit",
	Short: "Send an exit request for a validator",
	Long: `Send an exit request for a validator.  For example:

    ethdo validator exit --account=primary/validator --passphrase=secret

In quiet mode this will return 0 if the transaction has been sent, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Sanity checking and setup.
		assert(rootAccount != "", "--account is required")
		account, err := accountFromPath(rootAccount)
		errCheck(err, "Failed to access account")
		err = connect()
		errCheck(err, "Failed to obtain connect to Ethereum 2 beacon chain node")

		// Beacon chain config required for later work.
		config, err := grpc.FetchChainConfig(eth2GRPCConn)
		errCheck(err, "Failed to obtain beacon chain configuration")

		// Fetch the validator's index.
		index, err := grpc.FetchValidatorIndex(eth2GRPCConn, account)
		errCheck(err, "Failed to obtain validator index")
		outputIf(debug, fmt.Sprintf("Validator index is %d", index))

		// Ensure the validator is active.
		state, err := grpc.FetchValidatorState(eth2GRPCConn, account)
		errCheck(err, "Failed to obtain validator state")
		outputIf(debug, fmt.Sprintf("Validator state is %v", state))
		assert(state == ethpb.ValidatorStatus_ACTIVE, "Validator must be active to exit")

		// Ensure the validator has been active long enough to exit.
		validator, err := grpc.FetchValidator(eth2GRPCConn, account)
		errCheck(err, "Failed to obtain validator information")
		outputIf(debug, fmt.Sprintf("Activation epoch is %v", validator.ActivationEpoch))
		earliestExitEpoch := validator.ActivationEpoch + config["PersistentCommitteePeriod"].(uint64)

		secondsPerEpoch := config["SecondsPerSlot"].(uint64) * config["SlotsPerEpoch"].(uint64)
		genesisTime, err := grpc.FetchGenesis(eth2GRPCConn)
		errCheck(err, "Failed to obtain genesis time")

		currentEpoch := uint64(time.Since(genesisTime).Seconds()) / secondsPerEpoch
		assert(currentEpoch >= earliestExitEpoch, fmt.Sprintf("Validator cannot exit until %s ( epoch %d)", genesisTime.Add(time.Duration(secondsPerEpoch*earliestExitEpoch)*time.Second).Format(time.UnixDate), earliestExitEpoch))
		outputIf(verbose, "Validator confirmed to be in a suitable state")

		// Set up the transaction.
		exit := &ethpb.VoluntaryExit{
			Epoch:          currentEpoch,
			ValidatorIndex: index,
		}
		root, err := ssz.HashTreeRoot(exit)
		errCheck(err, "Failed to generate exit proposal root")
		// TODO fetch current fork version from config (currently using genesis fork version)
		// currentForkVersion := config["GenesisForkVersion"].([]byte)
		// domain := types.Domain(types.DomainVoluntaryExit, currentForkVersion)

		err = account.Unlock([]byte(rootAccountPassphrase))
		errCheck(err, "Failed to unlock account; please confirm passphrase is correct")
		// TODO supply domain
		signature, err := sign(account, root[:], []byte{})
		errCheck(err, "Failed to sign exit proposal")

		proposal := &ethpb.SignedVoluntaryExit{
			Exit:      exit,
			Signature: signature.Marshal(),
		}

		validatorClient := ethpb.NewBeaconNodeValidatorClient(eth2GRPCConn)
		ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
		defer cancel()
		_, err = validatorClient.ProposeExit(ctx, proposal)
		errCheck(err, "Failed to propose exit")

		outputIf(!quiet, "Validator exit transaction sent")
		os.Exit(_exitSuccess)
	},
}

func init() {
	validatorCmd.AddCommand(validatorExitCmd)
	validatorFlags(validatorExitCmd)
	validatorExitCmd.Flags().Int64Var(&validatorExitEpoch, "epoch", -1, "Epoch at which to exit (defaults to now)")
	addTransactionFlags(validatorExitCmd)
}
