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
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wealdtech/ethdo/grpc"
	"github.com/wealdtech/ethdo/util"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

var attesterInclusionCmd = &cobra.Command{
	Use:   "inclusion",
	Short: "Obtain information about attester inclusion",
	Long: `Obtain information about attester inclusion.  For example:

    ethdo attester inclusion --account=Validators/00001 --epoch=12345

In quiet mode this will return 0 if an attestation from the attester is found on the block fo the given epoch, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := connect()
		errCheck(err, "Failed to obtain connection to Ethereum 2 beacon chain block")

		// Obtain the epoch.
		epoch := viper.GetInt64("epoch")
		if epoch == -1 {
			outputIf(debug, "No epoch supplied; fetching current epoch")
			config, err := grpc.FetchChainConfig(eth2GRPCConn)
			errCheck(err, "Failed to obtain beacon chain configuration")
			slotsPerEpoch := config["SlotsPerEpoch"].(uint64)
			secondsPerSlot := config["SecondsPerSlot"].(uint64)
			genesisTime, err := grpc.FetchGenesisTime(eth2GRPCConn)
			errCheck(err, "Failed to obtain beacon chain genesis")
			epoch = int64(time.Since(genesisTime).Seconds()) / int64(secondsPerSlot*slotsPerEpoch)
		}
		outputIf(debug, fmt.Sprintf("Epoch is %d", epoch))

		// Obtain the validator.
		account, err := attesterInclusionAccount()
		errCheck(err, "Failed to obtain account")
		validatorIndex, err := grpc.FetchValidatorIndex(eth2GRPCConn, account)
		errCheck(err, "Failed to obtain validator")

		// Find the attesting slot for the given epoch.
		committees, err := grpc.FetchValidatorCommittees(eth2GRPCConn, uint64(epoch))
		errCheck(err, "Failed to obtain validator committees")

		slot := uint64(0)
		committeeIndex := uint64(0)
		validatorPositionInCommittee := uint64(0)
		found := false
		for searchSlot, committee := range committees {
			for searchCommitteeIndex, committeeValidatorIndices := range committee {
				for position, committeeValidatorIndex := range committeeValidatorIndices {
					if validatorIndex == committeeValidatorIndex {
						outputIf(debug, fmt.Sprintf("Validator %d attesting at slot %d for epoch %d: entry %d in committee %d", validatorIndex, searchSlot, epoch, position, searchCommitteeIndex))
						slot = searchSlot
						committeeIndex = uint64(searchCommitteeIndex)
						validatorPositionInCommittee = uint64(position)
						found = true
						break
					}
				}
			}
		}
		assert(found, "Failed to find attester duty for validator in the given epoch")

		startSlot := slot + 1
		endSlot := startSlot + 32
		for curSlot := startSlot; curSlot < endSlot; curSlot++ {
			signedBlock, err := grpc.FetchBlock(eth2GRPCConn, curSlot)
			errCheck(err, "Failed to obtain block")
			if signedBlock == nil {
				outputIf(debug, fmt.Sprintf("No block at slot %d", curSlot))
				continue
			}
			outputIf(debug, fmt.Sprintf("Fetched block %d", curSlot))
			for i, attestation := range signedBlock.Block.Body.Attestations {
				outputIf(debug, fmt.Sprintf("Attestation %d is for slot %d and committee %d", i, attestation.Data.Slot, attestation.Data.CommitteeIndex))
				if attestation.Data.Slot == slot &&
					attestation.Data.CommitteeIndex == committeeIndex &&
					attestation.AggregationBits.BitAt(validatorPositionInCommittee) {
					if verbose {
						fmt.Printf("Attestation included in block %d, attestation %d (inclusion delay %d)\n", curSlot, i, curSlot-slot)
					} else if !quiet {
						fmt.Printf("Attestation included in block %d (inclusion delay %d)\n", curSlot, curSlot-slot)
					}
					os.Exit(_exitSuccess)
				}
			}
		}
		os.Exit(_exitFailure)
	},
}

// attesterInclusionAccount obtains the account for the attester inclusion command.
func attesterInclusionAccount() (e2wtypes.Account, error) {
	var account e2wtypes.Account
	var err error
	if viper.GetString("account") != "" {
		ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
		defer cancel()
		_, account, err = walletAndAccountFromPath(ctx, viper.GetString("account"))
		if err != nil {
			return nil, errors.Wrap(err, "failed to obtain account")
		}
	} else {
		pubKey := viper.GetString("pubkey")
		pubKeyBytes, err := hex.DecodeString(strings.TrimPrefix(pubKey, "0x"))
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("failed to decode public key %s", pubKey))
		}
		account, err = util.NewScratchAccount(nil, pubKeyBytes)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("invalid public key %s", pubKey))
		}
	}
	return account, nil
}

func init() {
	attesterCmd.AddCommand(attesterInclusionCmd)
	attesterFlags(attesterInclusionCmd)
	attesterInclusionCmd.Flags().Int64("epoch", -1, "the current epoch")
	attesterInclusionCmd.Flags().String("pubkey", "", "the public key of the attester")
}

func attesterInclusionBindings() {
	if err := viper.BindPFlag("epoch", attesterInclusionCmd.Flags().Lookup("epoch")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("pubkey", attesterInclusionCmd.Flags().Lookup("pubkey")); err != nil {
		panic(err)
	}
}
