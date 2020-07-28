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
	"time"

	"github.com/pkg/errors"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wealdtech/ethdo/grpc"
	"github.com/wealdtech/ethdo/util"
	e2wallet "github.com/wealdtech/go-eth2-wallet"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
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
		assert(viper.GetString("account") != "" || validatorInfoPubKey != "", "--account or --pubkey is required")

		err := connect()
		errCheck(err, "Failed to obtain connection to Ethereum 2 beacon chain node")

		account, err := validatorInfoAccount()
		errCheck(err, "Failed to obtain validator account")

		if verbose {
			network := network()
			outputIf(debug, fmt.Sprintf("Network is %s", network))
			pubKey, err := bestPublicKey(account)
			if err == nil {
				deposits, totalDeposited, err := graphData(network, pubKey.Marshal())
				if err == nil {
					fmt.Printf("Number of deposits: %d\n", deposits)
					fmt.Printf("Total deposited: %s\n", string2eth.GWeiToString(totalDeposited, true))
				}
			}
		}

		validatorInfo, err := grpc.FetchValidatorInfo(eth2GRPCConn, account)
		errCheck(err, "Failed to obtain validator information")
		validator, err := grpc.FetchValidator(eth2GRPCConn, account)
		if err != nil {
			// We can live with this.
			validator = nil
		}
		if validatorInfo.Status != ethpb.ValidatorStatus_DEPOSITED &&
			validatorInfo.Status != ethpb.ValidatorStatus_UNKNOWN_STATUS {
			errCheck(err, "Failed to obtain validator definition")
		}
		assert(validatorInfo.Status != ethpb.ValidatorStatus_UNKNOWN_STATUS, "Not known as a validator")

		if quiet {
			os.Exit(_exitSuccess)
		}

		outputIf(verbose, fmt.Sprintf("Epoch of data: %v", validatorInfo.Epoch))
		outputIf(verbose && validatorInfo.Status != ethpb.ValidatorStatus_DEPOSITED, fmt.Sprintf("Index: %v", validatorInfo.Index))
		outputIf(verbose, fmt.Sprintf("Public key: %#x", validatorInfo.PublicKey))
		fmt.Printf("Status: %s\n", strings.Title(strings.ToLower(validatorInfo.Status.String())))
		fmt.Printf("Balance: %s\n", string2eth.GWeiToString(validatorInfo.Balance, true))

		if validatorInfo.Status == ethpb.ValidatorStatus_ACTIVE ||
			validatorInfo.Status == ethpb.ValidatorStatus_EXITING ||
			validatorInfo.Status == ethpb.ValidatorStatus_SLASHING {
			fmt.Printf("Effective balance: %s\n", string2eth.GWeiToString(validatorInfo.EffectiveBalance, true))
		}

		if validator != nil {
			outputIf(verbose, fmt.Sprintf("Withdrawal credentials: %#x", validator.WithdrawalCredentials))
		}

		transition := time.Unix(int64(validatorInfo.TransitionTimestamp), 0)
		transitionPassed := int64(validatorInfo.TransitionTimestamp) <= time.Now().Unix()
		switch validatorInfo.Status {
		case ethpb.ValidatorStatus_DEPOSITED:
			if validatorInfo.TransitionTimestamp != 0 {
				fmt.Printf("Inclusion in chain: %s\n", transition)
			}
		case ethpb.ValidatorStatus_PENDING:
			fmt.Printf("Activation: %s\n", transition)
		case ethpb.ValidatorStatus_EXITING, ethpb.ValidatorStatus_SLASHING:
			fmt.Printf("Attesting finishes: %s\n", transition)
		case ethpb.ValidatorStatus_EXITED:
			if transitionPassed {
				fmt.Printf("Funds withdrawable: Now\n")
			} else {
				fmt.Printf("Funds withdrawable: %s\n", transition)
			}
		}

		os.Exit(_exitSuccess)
	},
}

// validatorInfoAccount obtains the account for the validator info command.
func validatorInfoAccount() (e2wtypes.Account, error) {
	var account e2wtypes.Account
	if viper.GetString("account") != "" {
		wallet, err := openWallet()
		if err != nil {
			return nil, errors.Wrap(err, "failed to open wallet")
		}
		_, accountName, err := e2wallet.WalletAndAccountNames(viper.GetString("account"))
		if err != nil {
			return nil, errors.Wrap(err, "failed to obtain account name")
		}

		if wallet.Type() == "hierarchical deterministic" && strings.HasPrefix(accountName, "m/") {
			assert(getWalletPassphrase() != "", "walletpassphrase is required to obtain information about validators with dynamically generated hierarchical deterministic accounts")
			locker, isLocker := wallet.(e2wtypes.WalletLocker)
			if isLocker {
				ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
				defer cancel()
				errCheck(locker.Unlock(ctx, []byte(getWalletPassphrase())), "Failed to unlock wallet")
			}
		}

		accountByNameProvider, isProvider := wallet.(e2wtypes.WalletAccountByNameProvider)
		if !isProvider {
			return nil, errors.New("failed to ask wallet for account by name")
		}
		ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
		defer cancel()
		account, err = accountByNameProvider.AccountByName(ctx, accountName)
		if err != nil {
			return nil, errors.Wrap(err, "failed to obtain account")
		}
	} else {
		pubKeyBytes, err := hex.DecodeString(strings.TrimPrefix(validatorInfoPubKey, "0x"))
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("failed to decode public key %s", validatorInfoPubKey))
		}
		account, err = util.NewScratchAccount(nil, pubKeyBytes)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("invalid public key %s", validatorInfoPubKey))
		}
	}
	return account, nil
}

// graphData returns data from the graph about number and amount of deposits
func graphData(network string, validatorPubKey []byte) (uint64, uint64, error) {
	subgraph := fmt.Sprintf("attestantio/eth2deposits-%s", strings.ToLower(network))
	query := fmt.Sprintf(`{"query": "{deposits(where: {validatorPubKey:\"%#x\"}) { id amount withdrawalCredentials }}"}`, validatorPubKey)
	url := fmt.Sprintf("https://api.thegraph.com/subgraphs/name/%s", subgraph)
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
	totalDeposited := uint64(0)
	if response.Data != nil && len(response.Data.Deposits) > 0 {
		for _, deposit := range response.Data.Deposits {
			deposits++
			depositAmount, err := strconv.ParseUint(deposit.Amount, 10, 64)
			if err != nil {
				return 0, 0, errors.Wrap(err, fmt.Sprintf("invalid deposit amount from pre-existing deposit %s", deposit.Amount))
			}
			totalDeposited += depositAmount
		}
	}
	return deposits, totalDeposited, nil
}

func init() {
	validatorCmd.AddCommand(validatorInfoCmd)
	validatorInfoCmd.Flags().StringVar(&validatorInfoPubKey, "pubkey", "", "Public key for which to obtain status")
	validatorFlags(validatorInfoCmd)
}
