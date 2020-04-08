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
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/spf13/cobra"
	"github.com/wealdtech/ethdo/grpc"
	"github.com/wealdtech/ethdo/util"
	types "github.com/wealdtech/go-eth2-wallet-types/v2"
	string2eth "github.com/wealdtech/go-string2eth"
)

var validatorInfoPubKey string

var validatorInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Obtain information about a validator",
	Long: `Obtain information about validator.  For example:

    ethdo validator info --account=primary/validator

In quiet mode this will return 0 if the validator information can be obtained, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Sanity checking and setup.
		assert(rootAccount != "" || validatorInfoPubKey != "", "--account or --pubkey is required")

		err := connect()
		errCheck(err, "Failed to obtain connection to Ethereum 2 beacon chain node")

		var account types.Account
		if rootAccount != "" {
			account, err = accountFromPath(rootAccount)
			errCheck(err, "Failed to access account")
		} else {
			pubKeyBytes, err := hex.DecodeString(strings.TrimPrefix(validatorInfoPubKey, "0x"))
			errCheck(err, fmt.Sprintf("Failed to decode public key %s", validatorInfoPubKey))
			account, err = util.NewScratchAccount(pubKeyBytes)
			errCheck(err, fmt.Sprintf("Invalid public key %s", validatorInfoPubKey))
		}

		validatorInfo, err := grpc.FetchValidatorInfo(eth2GRPCConn, account)
		errCheck(err, "Failed to obtain validator information")
		validatorDef, err := grpc.FetchValidator(eth2GRPCConn, account)
		if validatorInfo.Status != ethpb.ValidatorStatus_DEPOSITED &&
			validatorInfo.Status != ethpb.ValidatorStatus_UNKNOWN_STATUS {
			errCheck(err, "Failed to obtain validator definition")
		}

		assert(validatorInfo.Status != ethpb.ValidatorStatus_UNKNOWN_STATUS, "Not known as a validator")

		if quiet {
			os.Exit(_exitSuccess)
		}

		outputIf(verbose, fmt.Sprintf("Epoch of data:\t\t%v", validatorInfo.Epoch))
		outputIf(verbose && validatorInfo.Status != ethpb.ValidatorStatus_DEPOSITED, fmt.Sprintf("Index:\t\t\t%v", validatorInfo.Index))
		outputIf(verbose, fmt.Sprintf("Public key:\t\t%#x", validatorInfo.PublicKey))
		fmt.Printf("Status:\t\t\t%s\n", strings.Title(strings.ToLower(validatorInfo.Status.String())))
		fmt.Printf("Balance:\t\t%s\n", string2eth.GWeiToString(validatorInfo.Balance, true))
		if validatorInfo.Status == ethpb.ValidatorStatus_ACTIVE ||
			validatorInfo.Status == ethpb.ValidatorStatus_EXITING ||
			validatorInfo.Status == ethpb.ValidatorStatus_SLASHING {
			fmt.Printf("Effective balance:\t%s\n", string2eth.GWeiToString(validatorInfo.EffectiveBalance, true))
		}
		if validatorDef != nil {
			outputIf(verbose, fmt.Sprintf("Withdrawal credentials:\t%#x", validatorDef.WithdrawalCredentials))
		}
		transition := time.Unix(int64(validatorInfo.TransitionTimestamp), 0)
		transitionPassed := int64(validatorInfo.TransitionTimestamp) <= time.Now().Unix()
		switch validatorInfo.Status {
		case ethpb.ValidatorStatus_DEPOSITED:
			if validatorInfo.TransitionTimestamp != 0 {
				fmt.Printf("Inclusion in chain:\t%s\n", transition)
			}
		case ethpb.ValidatorStatus_PENDING:
			fmt.Printf("Activation:\t\t%s\n", transition)
		case ethpb.ValidatorStatus_EXITING, ethpb.ValidatorStatus_SLASHING:
			fmt.Printf("Attesting finishes:\t%s\n", transition)
		case ethpb.ValidatorStatus_EXITED:
			if transitionPassed {
				fmt.Printf("Funds withdrawable:\tNow\n")
			} else {
				fmt.Printf("Funds withdrawable:\t%s\n", transition)
			}
		}

		os.Exit(_exitSuccess)
	},
}

func init() {
	validatorCmd.AddCommand(validatorInfoCmd)
	validatorInfoCmd.Flags().StringVar(&validatorInfoPubKey, "pubkey", "", "Public key for which to obtain status")
	validatorFlags(validatorInfoCmd)
}
