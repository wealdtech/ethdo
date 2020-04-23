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
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wealdtech/ethdo/grpc"
	"github.com/wealdtech/ethdo/util"
	e2types "github.com/wealdtech/go-eth2-types/v2"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

var validatorExitEpoch int64
var validatorExitKey string
var validatorExitJSON string
var validatorExitJSONOutput bool

var validatorExitCmd = &cobra.Command{
	Use:   "exit",
	Short: "Send an exit request for a validator",
	Long: `Send an exit request for a validator.  For example:

    ethdo validator exit --account=primary/validator --passphrase=secret

In quiet mode this will return 0 if the transaction has been generated, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := connect()
		errCheck(err, "Failed to obtain connect to Ethereum 2 beacon chain node")

		exit, signature := validatorExitHandleInput()
		validatorExitHandleExit(exit, signature)
		os.Exit(_exitSuccess)
	},
}

func validatorExitHandleInput() (*ethpb.VoluntaryExit, e2types.Signature) {
	if validatorExitJSON != "" {
		return validatorExitHandleJSONInput(validatorExitJSON)
	}
	if rootAccount != "" {
		account, err := accountFromPath(rootAccount)
		errCheck(err, "Failed to access account")
		return validatorExitHandleAccountInput(account)
	}
	if validatorExitKey != "" {
		privKeyBytes, err := hex.DecodeString(strings.TrimPrefix(validatorExitKey, "0x"))
		errCheck(err, fmt.Sprintf("Failed to decode key %s", validatorExitKey))
		account, err := util.NewScratchAccount(privKeyBytes, nil)
		errCheck(err, "Invalid private key")
		return validatorExitHandleAccountInput(account)
	}
	die("one of --json, --account or --key is required")
	return nil, nil
}

func validatorExitHandleJSONInput(input string) (*ethpb.VoluntaryExit, e2types.Signature) {
	data := &validatorExitData{}
	err := json.Unmarshal([]byte(input), data)
	errCheck(err, "Invalid JSON input")
	exit := &ethpb.VoluntaryExit{
		Epoch:          data.Epoch,
		ValidatorIndex: data.ValidatorIndex,
	}
	signature, err := e2types.BLSSignatureFromBytes(data.Signature)
	errCheck(err, "Invalid signature")
	return exit, signature
}

func validatorExitHandleAccountInput(account e2wtypes.Account) (*ethpb.VoluntaryExit, e2types.Signature) {
	exit := &ethpb.VoluntaryExit{}

	// Beacon chain config required for later work.
	config, err := grpc.FetchChainConfig(eth2GRPCConn)
	errCheck(err, "Failed to obtain beacon chain configuration")
	secondsPerEpoch := config["SecondsPerSlot"].(uint64) * config["SlotsPerEpoch"].(uint64)

	// Fetch the validator's index.
	index, err := grpc.FetchValidatorIndex(eth2GRPCConn, account)
	errCheck(err, "Failed to obtain validator index")
	outputIf(debug, fmt.Sprintf("Validator index is %d", index))
	exit.ValidatorIndex = index

	// Ensure the validator is active.
	state, err := grpc.FetchValidatorState(eth2GRPCConn, account)
	errCheck(err, "Failed to obtain validator state")
	outputIf(debug, fmt.Sprintf("Validator state is %v", state))
	assert(state == ethpb.ValidatorStatus_ACTIVE, "Validator must be active to exit")

	if validatorExitEpoch < 0 {
		// Ensure the validator has been active long enough to exit.
		validator, err := grpc.FetchValidator(eth2GRPCConn, account)
		errCheck(err, "Failed to obtain validator information")
		outputIf(debug, fmt.Sprintf("Activation epoch is %v", validator.ActivationEpoch))
		earliestExitEpoch := validator.ActivationEpoch + config["PersistentCommitteePeriod"].(uint64)

		genesisTime, err := grpc.FetchGenesis(eth2GRPCConn)
		errCheck(err, "Failed to obtain genesis time")

		currentEpoch := uint64(time.Since(genesisTime).Seconds()) / secondsPerEpoch
		assert(currentEpoch >= earliestExitEpoch, fmt.Sprintf("Validator cannot exit until %s ( epoch %d)", genesisTime.Add(time.Duration(secondsPerEpoch*earliestExitEpoch)*time.Second).Format(time.UnixDate), earliestExitEpoch))
		outputIf(verbose, "Validator confirmed to be in a suitable state")
		exit.Epoch = currentEpoch
	} else {
		// User-specified epoch; no checks.
		exit.Epoch = uint64(validatorExitEpoch)
	}

	// TODO fetch current fork version from config (currently using genesis fork version)
	currentForkVersion := config["GenesisForkVersion"].([]byte)
	// TODO fetch genesis validators root from API.
	genesisValidatorsRoot := []byte{
		0x55, 0x13, 0x8e, 0x46, 0xa2, 0x44, 0x2d, 0x2f,
		0xfd, 0x89, 0x55, 0x0a, 0x0f, 0x30, 0x56, 0x21,
		0x27, 0xbc, 0x56, 0xe6, 0x24, 0x4d, 0x0f, 0xa2,
		0xb5, 0x18, 0xa3, 0xf4, 0xce, 0x19, 0x33, 0x7e,
	}
	domain := e2types.Domain(e2types.DomainVoluntaryExit, currentForkVersion, genesisValidatorsRoot)

	err = account.Unlock([]byte(rootAccountPassphrase))
	errCheck(err, "Failed to unlock account; please confirm passphrase is correct")
	signature, err := signStruct(account, exit, domain)
	errCheck(err, "Failed to sign exit proposal")

	return exit, signature
}

// validatorExitHandleExit handles the exit request.
func validatorExitHandleExit(exit *ethpb.VoluntaryExit, signature e2types.Signature) {
	if validatorExitJSONOutput {
		data := &validatorExitData{
			Epoch:          exit.Epoch,
			ValidatorIndex: exit.ValidatorIndex,
			Signature:      signature.Marshal(),
		}
		res, err := json.Marshal(data)
		errCheck(err, "Failed to generate JSON")
		outputIf(!quiet, string(res))
	} else {
		proposal := &ethpb.SignedVoluntaryExit{
			Exit:      exit,
			Signature: signature.Marshal(),
		}

		validatorClient := ethpb.NewBeaconNodeValidatorClient(eth2GRPCConn)
		ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
		defer cancel()
		_, err := validatorClient.ProposeExit(ctx, proposal)
		errCheck(err, "Failed to propose exit")
		outputIf(!quiet, "Validator exit transaction sent")
	}
}

func init() {
	validatorCmd.AddCommand(validatorExitCmd)
	validatorFlags(validatorExitCmd)
	validatorExitCmd.Flags().Int64Var(&validatorExitEpoch, "epoch", -1, "Epoch at which to exit (defaults to current epoch)")
	validatorExitCmd.Flags().StringVar(&validatorExitKey, "key", "", "Private key if account not known by ethdo")
	validatorExitCmd.Flags().BoolVar(&validatorExitJSONOutput, "json-output", false, "Print JSON transaction; do not broadcast to network")
	validatorExitCmd.Flags().StringVar(&validatorExitJSON, "json", "", "Use JSON as created by --json-output to exit")
	addTransactionFlags(validatorExitCmd)
}

type validatorExitData struct {
	Epoch          uint64 `json:"epoch"`
	ValidatorIndex uint64 `json:"validator_index"`
	Signature      []byte `json:"signature"`
}

// MarshalJSON implements custom JSON marshaller.
func (d *validatorExitData) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`{"epoch":%d,"validator_index":%d,"signature":"%#x"}`, d.Epoch, d.ValidatorIndex, d.Signature)), nil
}

// UnmarshalJSON implements custom JSON unmarshaller.
func (d *validatorExitData) UnmarshalJSON(data []byte) error {
	var v map[string]interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	if val, exists := v["epoch"]; exists {
		var ok bool
		epoch, ok := val.(float64)
		if !ok {
			return errors.New("epoch invalid")
		}
		d.Epoch = uint64(epoch)
	} else {
		return errors.New("epoch missing")
	}

	if val, exists := v["validator_index"]; exists {
		var ok bool
		validatorIndex, ok := val.(float64)
		if !ok {
			return errors.New("validator_index invalid")
		}
		d.ValidatorIndex = uint64(validatorIndex)
	} else {
		return errors.New("validator_index missing")
	}

	if val, exists := v["signature"]; exists {
		signatureBytes, ok := val.(string)
		if !ok {
			return errors.New("signature invalid")
		}
		signature, err := hex.DecodeString(strings.TrimPrefix(signatureBytes, "0x"))
		if err != nil {
			return errors.Wrap(err, "signature invalid")
		}
		d.Signature = signature
	} else {
		return errors.New("signature missing")
	}

	return nil
}
