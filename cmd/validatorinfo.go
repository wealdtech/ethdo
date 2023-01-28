// Copyright © 2020 - 2022 Weald Technology Trading
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
	"encoding/json"
	"fmt"
	"io"
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
	string2eth "github.com/wealdtech/go-string2eth"
)

var validatorInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Obtain information about a validator",
	Long: `Obtain information about validator.  For example:

    ethdo validator info --validator=primary/validator

In quiet mode this will return 0 if the validator information can be obtained, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		eth2Client, err := util.ConnectToBeaconNode(ctx,
			viper.GetString("connection"),
			viper.GetDuration("timeout"),
			viper.GetBool("allow-insecure-connections"),
		)
		errCheck(err, "Failed to connect to Ethereum 2 beacon node")

		if viper.GetString("validator") == "" {
			fmt.Println("validator is required")
			os.Exit(_exitFailure)
		}

		validator, err := util.ParseValidator(ctx, eth2Client.(eth2client.ValidatorsProvider), viper.GetString("validator"), "head")
		errCheck(err, "Failed to obtain validator")

		if verbose {
			network, err := util.Network(ctx, eth2Client)
			errCheck(err, "Failed to obtain network")
			outputIf(debug, fmt.Sprintf("Network is %s", network))
			pubKey, err := validator.PubKey(ctx)
			if err == nil {
				deposits, totalDeposited, err := graphData(network, pubKey[:])
				if err == nil && deposits > 0 {
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

// graphData returns data from the graph about number and amount of deposits.
func graphData(network string, validatorPubKey []byte) (uint64, spec.Gwei, error) {
	subgraph := ""
	if network == "Mainnet" {
		subgraph = "attestantio/eth2deposits"
	} else {
		subgraph = fmt.Sprintf("attestantio/eth2deposits-%s", strings.ToLower(network))
	}
	query := fmt.Sprintf(`{"query": "{deposits(where: {validatorPubKey:\"%#x\"}) { id amount withdrawalCredentials }}"}`, validatorPubKey)
	url := fmt.Sprintf("https://api.thegraph.com/subgraphs/name/%s", subgraph)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewBufferString(query))
	if err != nil {
		return 0, 0, errors.Wrap(err, "failed to start request")
	}
	req.Header.Set("Accept", "application/json")
	graphResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, 0, errors.Wrap(err, "failed to check if there is already a deposit for this validator")
	}
	defer graphResp.Body.Close()
	body, err := io.ReadAll(graphResp.Body)
	if err != nil {
		return 0, 0, errors.Wrap(err, "bad information returned from existing deposit check")
	}

	type graphDeposit struct {
		Index  string `json:"index"`
		Amount string `json:"amount"`
		// Using graph API JSON names in camel case.
		//nolint:tagliatelle
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
	validatorInfoCmd.Flags().String("validator", "", "Public key for which to obtain status")
	validatorFlags(validatorInfoCmd)
}

func validatorInfoBindings() {
	if err := viper.BindPFlag("validator", validatorInfoCmd.Flags().Lookup("validator")); err != nil {
		panic(err)
	}
}
