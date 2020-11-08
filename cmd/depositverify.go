// Copyright © 2019, 2020 Weald Technology Trading
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
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/spf13/cobra"
	"github.com/wealdtech/ethdo/util"
	e2types "github.com/wealdtech/go-eth2-types/v2"
	eth2util "github.com/wealdtech/go-eth2-util"
	string2eth "github.com/wealdtech/go-string2eth"
)

var depositVerifyData string
var depositVerifyWithdrawalPubKey string
var depositVerifyValidatorPubKey string
var depositVerifyDepositAmount string

var depositVerifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify deposit data matches the provided data",
	Long: `Verify deposit data matches the provided input data.  For example:

    ethdo deposit verify --data=depositdata.json --withdrawalaccount=primary/current --value="32 Ether"

The deposit data is compared to the supplied withdrawal account/public key, validator public key, and value to ensure they match.

In quiet mode this will return 0 if the the data is verified correctly, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		assert(depositVerifyData != "", "--data is required")
		var data []byte
		var err error
		// Input could be JSON or a path to JSON.
		switch {
		case strings.HasPrefix(depositVerifyData, "0x"):
			// Looks like raw binary.
			data = []byte(depositVerifyData)
		case strings.HasPrefix(depositVerifyData, "{"):
			// Looks like JSON.
			data = []byte("[" + depositVerifyData + "]")
		case strings.HasPrefix(depositVerifyData, "["):
			// Looks like JSON array.
			data = []byte(depositVerifyData)
		default:
			// Assume it's a path to JSON.
			data, err = ioutil.ReadFile(depositVerifyData)
			errCheck(err, "Failed to read deposit data file")
			if data[0] == '{' {
				data = []byte("[" + string(data) + "]")
			}
		}

		deposits, err := util.DepositInfoFromJSON(data)
		errCheck(err, "Failed to fetch deposit data")

		var withdrawalCredentials []byte
		if depositVerifyWithdrawalPubKey != "" {
			withdrawalPubKeyBytes, err := hex.DecodeString(strings.TrimPrefix(depositVerifyWithdrawalPubKey, "0x"))
			errCheck(err, "Invalid withdrawal public key")
			assert(len(withdrawalPubKeyBytes) == 48, "Public key should be 48 bytes")
			withdrawalPubKey, err := e2types.BLSPublicKeyFromBytes(withdrawalPubKeyBytes)
			errCheck(err, "Value supplied with --withdrawalpubkey is not a valid public key")
			withdrawalCredentials = eth2util.SHA256(withdrawalPubKey.Marshal())
			withdrawalCredentials[0] = 0 // BLS_WITHDRAWAL_PREFIX
		}
		outputIf(debug, fmt.Sprintf("Withdrawal credentials are %#x", withdrawalCredentials))

		depositAmount := uint64(0)
		if depositVerifyDepositAmount != "" {
			depositAmount, err = string2eth.StringToGWei(depositVerifyDepositAmount)
			errCheck(err, "Invalid value")
			assert(depositAmount >= 1000000000, "deposit amount must be at least 1 Ether") // MIN_DEPOSIT_AMOUNT
		}

		validatorPubKeys := make(map[[48]byte]bool)
		if depositVerifyValidatorPubKey != "" {
			validatorPubKeys, err = validatorPubKeysFromInput(depositVerifyValidatorPubKey)
			errCheck(err, "Failed to obtain validator public key(s))")
		}

		failures := false
		for i, deposit := range deposits {
			if deposit.Amount == 0 {
				deposit.Amount = depositAmount
			}
			verified, err := verifyDeposit(deposit, withdrawalCredentials, validatorPubKeys, depositAmount)
			errCheck(err, fmt.Sprintf("Error attempting to verify deposit %d", i))
			if !verified {
				failures = true
				outputIf(!quiet, fmt.Sprintf("Deposit %q failed verification", deposit.Name))
			} else {
				outputIf(quiet, fmt.Sprintf("Deposit %q verified", deposit.Name))
			}
		}

		if failures {
			os.Exit(_exitFailure)
		}
		os.Exit(_exitSuccess)
	},
}

func validatorPubKeysFromInput(input string) (map[[48]byte]bool, error) {
	pubKeys := make(map[[48]byte]bool)
	var err error
	var data []byte
	// Input could be a public key or a path to public keys.
	if strings.HasPrefix(input, "0x") {
		// Looks like a public key.
		pubKeyBytes, err := hex.DecodeString(strings.TrimPrefix(input, "0x"))
		if err != nil {
			return nil, errors.Wrap(err, "public key is not a hex string")
		}
		if len(pubKeyBytes) != 48 {
			return nil, errors.New("public key should be 48 bytes")
		}
		pubKey, err := e2types.BLSPublicKeyFromBytes(pubKeyBytes)
		if err != nil {
			return nil, errors.Wrap(err, "invalid public key")
		}
		var key [48]byte
		copy(key[:], pubKey.Marshal())
		pubKeys[key] = true
	} else {
		// Assume it's a path to a file of public keys.
		data, err = ioutil.ReadFile(input)
		if err != nil {
			return nil, errors.Wrap(err, "failed to find public key file")
		}
		lines := bytes.Split(bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n")), []byte("\n"))
		if len(lines) == 0 {
			return nil, errors.New("file has no public keys")
		}
		for _, line := range lines {
			if len(line) == 0 {
				continue
			}
			pubKeyBytes, err := hex.DecodeString(strings.TrimPrefix(string(line), "0x"))
			if err != nil {
				return nil, errors.Wrap(err, "public key is not a hex string")
			}
			if len(pubKeyBytes) != 48 {
				return nil, errors.New("public key should be 48 bytes")
			}
			pubKey, err := e2types.BLSPublicKeyFromBytes(pubKeyBytes)
			if err != nil {
				return nil, errors.Wrap(err, "invalid public key")
			}
			var key [48]byte
			copy(key[:], pubKey.Marshal())
			pubKeys[key] = true
		}
	}

	return pubKeys, nil
}

func verifyDeposit(deposit *util.DepositInfo, withdrawalCredentials []byte, validatorPubKeys map[[48]byte]bool, amount uint64) (bool, error) {
	if withdrawalCredentials != nil {
		if !bytes.Equal(deposit.WithdrawalCredentials, withdrawalCredentials) {
			return false, errors.New("withdrawal credentials incorrect")
		}
		outputIf(verbose, "Withdrawal credentials verified")
	}
	if amount != 0 {
		if deposit.Amount != amount {
			return false, errors.New("deposit value incorrect")
		}
		outputIf(verbose, "Amount verified")
	}

	if len(validatorPubKeys) != 0 {
		var key [48]byte
		copy(key[:], deposit.PublicKey)
		if _, exists := validatorPubKeys[key]; !exists {
			return false, errors.New("validator public key incorrect")
		}
		outputIf(verbose, "Validator public key verified")
	}

	depositData := &ethpb.Deposit_Data{
		PublicKey:             deposit.PublicKey,
		WithdrawalCredentials: deposit.WithdrawalCredentials,
		Amount:                deposit.Amount,
		Signature:             deposit.Signature,
	}
	depositDataRoot, err := depositData.HashTreeRoot()
	if err != nil {
		return false, errors.Wrap(err, "failed to generate deposit data root")
	}
	if !bytes.Equal(deposit.DepositDataRoot, depositDataRoot[:]) {
		return false, errors.New("deposit data root incorrect")
	}
	outputIf(debug, "Deposit data root verified")

	outputIf(verbose, "Deposit verified")

	return true, nil
}

func init() {
	depositCmd.AddCommand(depositVerifyCmd)
	depositFlags(depositVerifyCmd)
	depositVerifyCmd.Flags().StringVar(&depositVerifyData, "data", "", "JSON data, or path to JSON data")
	depositVerifyCmd.Flags().StringVar(&depositVerifyWithdrawalPubKey, "withdrawalpubkey", "", "Public key of the account to which the validator funds will be withdrawn")
	depositVerifyCmd.Flags().StringVar(&depositVerifyDepositAmount, "depositvalue", "", "Value of the amount to be deposited")
	depositVerifyCmd.Flags().StringVar(&depositVerifyValidatorPubKey, "validatorpubkey", "", "Public key(s) of the account(s) that will be carrying out validation")
}
