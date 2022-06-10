// Copyright © 2020, 2021 Weald Technology Trading
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
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	eth2client "github.com/attestantio/go-eth2-client"
	api "github.com/attestantio/go-eth2-client/api/v1"
	spec "github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wealdtech/ethdo/util"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
	string2eth "github.com/wealdtech/go-string2eth"
)

var validatorInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Obtain information about a validator",
	Long: `Obtain information about validator.  For example:

    ethdo validator info --account=primary/validator

In quiet mode this will return 0 if the validator information can be obtained, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		eth2Client, err := util.ConnectToBeaconNode(ctx,
			viper.GetString("connection"),
			viper.GetDuration("timeout"),
			viper.GetBool("allow-insecure-connections"),
		)
		errCheck(err, "Failed to connect to Ethereum 2 beacon node")

		account, err := validatorInfoAccount(ctx, eth2Client)
		errCheck(err, "Failed to obtain validator account")

		pubKeys := make([]spec.BLSPubKey, 1)
		pubKey, err := util.BestPublicKey(account)
		errCheck(err, "Failed to obtain validator public key")
		copy(pubKeys[0][:], pubKey.Marshal())
		validators, err := eth2Client.(eth2client.ValidatorsProvider).ValidatorsByPubKey(ctx, "head", pubKeys)
		errCheck(err, "Failed to obtain validator information")
		if len(validators) == 0 {
			fmt.Println("Validator not known by beacon node")
			os.Exit(_exitSuccess)
		}

		var validator *api.Validator
		for _, v := range validators {
			validator = v
		}

		if verbose {
			network, err := util.Network(ctx, eth2Client)
			errCheck(err, "Failed to obtain network")
			outputIf(debug, fmt.Sprintf("Network is %s", network))
			pubKey, err := bestPublicKey(account)
			if err == nil {
				deposits, totalDeposited, err := graphData(network, pubKey.Marshal())
				if err == nil {
					fmt.Printf("Number of deposits: %d\n", deposits)
					fmt.Printf("Total deposited: %s\n", string2eth.GWeiToString(uint64(totalDeposited), true))
				}
			}
		}

		if quiet {
			os.Exit(_exitSuccess)
		}

		if validator.Status.IsPending() || validator.Status.HasActivated() {
			fmt.Printf("Index: %d\n", validator.Index)
		}
		if verbose {
			if validator.Status.IsPending() {
				fmt.Printf("Activation eligibility epoch: %d\n", validator.Validator.ActivationEligibilityEpoch)
			}
			if validator.Status.HasActivated() {
				fmt.Printf("Activation epoch: %d\n", validator.Validator.ActivationEpoch)
			}
			fmt.Printf("Public key: %#x\n", validator.Validator.PublicKey)
		}
		fmt.Printf("Status: %v\n", validator.Status)
		switch validator.Status {
		case api.ValidatorStateActiveExiting, api.ValidatorStateActiveSlashed:
			fmt.Printf("Exit epoch: %d\n", validator.Validator.ExitEpoch)
		case api.ValidatorStateExitedUnslashed, api.ValidatorStateExitedSlashed:
			fmt.Printf("Withdrawable epoch: %d\n", validator.Validator.WithdrawableEpoch)
		}
		fmt.Printf("Balance: %s\n", string2eth.GWeiToString(uint64(validator.Balance), true))
		if validator.Status.IsActive() {
			fmt.Printf("Effective balance: %s\n", string2eth.GWeiToString(uint64(validator.Validator.EffectiveBalance), true))
		}
		if verbose {
			fmt.Printf("Withdrawal credentials: %#x\n", validator.Validator.WithdrawalCredentials)
		}

		os.Exit(_exitSuccess)
	},
}

// validatorInfoAccount obtains the account for the validator info command.
func validatorInfoAccount(ctx context.Context, eth2Client eth2client.Service) (e2wtypes.Account, error) {
	var account e2wtypes.Account
	var err error
	switch {
	case viper.GetString("account") != "":
		ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
		defer cancel()
		_, account, err = walletAndAccountFromPath(ctx, viper.GetString("account"))
		if err != nil {
			return nil, errors.Wrap(err, "failed to obtain account")
		}
	case viper.GetString("pubkey") != "":
		pubKeyBytes, err := hex.DecodeString(strings.TrimPrefix(viper.GetString("pubkey"), "0x"))
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("failed to decode public key %s", viper.GetString("pubkey")))
		}
		account, err = util.NewScratchAccount(nil, pubKeyBytes)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("invalid public key %s", viper.GetString("pubkey")))
		}
	case viper.GetInt64("index") != -1:
		validatorsProvider, isValidatorsProvider := eth2Client.(eth2client.ValidatorsProvider)
		if !isValidatorsProvider {
			return nil, errors.New("client does not provide validator information")
		}
		index := spec.ValidatorIndex(viper.GetInt64("index"))
		validators, err := validatorsProvider.Validators(ctx, "head", []spec.ValidatorIndex{
			index,
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to obtain validator information")
		}
		if len(validators) == 0 {
			return nil, errors.New("unknown validator index")
		}
		pubKeyBytes := make([]byte, 48)
		copy(pubKeyBytes, validators[index].Validator.PublicKey[:])
		account, err = util.NewScratchAccount(nil, pubKeyBytes)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("invalid public key %s", viper.GetString("pubkey")))
		}
	default:
		return nil, errors.New("neither account nor public key supplied")
	}
	return account, nil
}

// graphData returns data from the graph about number and amount of deposits
func graphData(network string, validatorPubKey []byte) (uint64, spec.Gwei, error) {
	subgraph := ""
	if network == "Mainnet" {
		subgraph = "attestantio/eth2deposits"
	} else {
		subgraph = fmt.Sprintf("attestantio/eth2deposits-%s", strings.ToLower(network))
	}
	query := fmt.Sprintf(`{"query": "{deposits(where: {validatorPubKey:\"%#x\"}) { id amount withdrawalCredentials }}"}`, validatorPubKey)
	url := fmt.Sprintf("https://api.thegraph.com/subgraphs/name/%s", subgraph)
	// #nosec G107
	graphResp, err := http.Post(url, "application/json", bytes.NewBufferString(query))
	if err != nil {
		return 0, 0, errors.Wrap(err, "failed to check if there is already a deposit for this validator")
	}
	defer graphResp.Body.Close()
	body, err := ioutil.ReadAll(graphResp.Body)
	if err != nil {
		return 0, 0, errors.Wrap(err, "bad information returned from existing deposit check")
	}

	type graphDeposit struct {
		Index                 string `json:"index"`
		Amount                string `json:"amount"`
		WithdrawalCredentials string `json:"withdrawalCredentials"`
	}
	type graphData struct {
		Deposits []*graphDeposit `json:"deposits,omitempty"`
	}
	type graphResponse struct {
		Data *graphData `json:"data,omitempty"`
	}

	var response graphResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return 0, 0, errors.Wrap(err, "invalid data returned from existing deposit check")
	}
	deposits := uint64(0)
	totalDeposited := spec.Gwei(0)
	if response.Data != nil && len(response.Data.Deposits) > 0 {
		for _, deposit := range response.Data.Deposits {
			deposits++
			depositAmount, err := strconv.ParseUint(deposit.Amount, 10, 64)
			if err != nil {
				return 0, 0, errors.Wrap(err, fmt.Sprintf("invalid deposit amount from pre-existing deposit %s", deposit.Amount))
			}
			totalDeposited += spec.Gwei(depositAmount)
		}
	}
	return deposits, totalDeposited, nil
}

func init() {
	validatorCmd.AddCommand(validatorInfoCmd)
	validatorInfoCmd.Flags().String("pubkey", "", "Public key for which to obtain status")
	validatorInfoCmd.Flags().Int64("index", -1, "Index for which to obtain status")
	validatorFlags(validatorInfoCmd)
}

func validatorInfoBindings() {
	if err := viper.BindPFlag("pubkey", validatorInfoCmd.Flags().Lookup("pubkey")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("index", validatorInfoCmd.Flags().Lookup("index")); err != nil {
		panic(err)
	}
}
