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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	e2types "github.com/wealdtech/go-eth2-types/v2"
	util "github.com/wealdtech/go-eth2-util"
	string2eth "github.com/wealdtech/go-string2eth"
)

type depositData struct {
	Name                  string `json:"name,omitempty"`
	Account               string `json:"account,omitempty"`
	PublicKey             string `json:"pubkey"`
	WithdrawalCredentials string `json:"withdrawal_credentials"`
	Signature             string `json:"signature"`
	DepositDataRoot       string `json:"deposit_data_root"`
	Value                 uint64 `json:"value"`
	Version               uint64 `json:"version"`
}

var depositVerifyData string
var depositVerifyWithdrawalPubKey string
var depositVerifyValidatorPubKey string
var depositVerifyDepositValue string

var depositVerifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify deposit data matches requirements",
	Long: `Verify deposit data matches requirements.  For example:

    ethdo deposit verify --data=depositdata.json --withdrawalaccount=primary/current --value="32 Ether"

The information generated can be passed to ethereal to create a deposit from the Ethereum 1 chain.

In quiet mode this will return 0 if the the data can be generated correctly, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		assert(depositVerifyData != "", "--data is required")
		deposits, err := depositDataFromJSON(depositVerifyData)
		errCheck(err, "Failed to fetch deposit data")

		withdrawalCredentials := ""
		if depositVerifyWithdrawalPubKey != "" {
			withdrawalPubKeyBytes, err := hex.DecodeString(strings.TrimPrefix(depositVerifyWithdrawalPubKey, "0x"))
			errCheck(err, "Invalid withdrawal public key")
			assert(len(withdrawalPubKeyBytes) == 48, "Public key should be 48 bytes")
			withdrawalPubKey, err := e2types.BLSPublicKeyFromBytes(withdrawalPubKeyBytes)
			errCheck(err, "Value supplied with --withdrawalpubkey is not a valid public key")
			withdrawalBytes := util.SHA256(withdrawalPubKey.Marshal())
			withdrawalBytes[0] = 0 // BLS_WITHDRAWAL_PREFIX
			withdrawalCredentials = fmt.Sprintf("%x", withdrawalBytes)
		}
		outputIf(debug, fmt.Sprintf("Withdrawal credentials are %s", withdrawalCredentials))

		depositValue := uint64(0)
		if depositVerifyDepositValue != "" {
			depositValue, err = string2eth.StringToGWei(depositVerifyDepositValue)
			errCheck(err, "Invalid value")
			// This is hard-coded, to allow deposit data to be generated without a connection to the beacon node.
			assert(depositValue >= 1000000000, "deposit value must be at least 1 Ether") // MIN_DEPOSIT_AMOUNT
		}

		validatorPubKeys := make(map[string]bool)
		if depositVerifyValidatorPubKey != "" {
			validatorPubKeys, err = validatorPubKeysFromInput(depositVerifyValidatorPubKey)
			errCheck(err, "Failed to obtain validator public key(s))")
		}

		failures := false
		for i, deposit := range deposits {
			if withdrawalCredentials != "" {
				if deposit.WithdrawalCredentials != withdrawalCredentials {
					outputIf(!quiet, fmt.Sprintf("Invalid withdrawal credentials for deposit %d", i))
					failures = true
				}
			}
			if depositValue != 0 {
				if deposit.Value != depositValue {
					outputIf(!quiet, fmt.Sprintf("Invalid deposit value for deposit %d", i))
					failures = true
				}
			}
			if len(validatorPubKeys) != 0 {
				if _, exists := validatorPubKeys[deposit.PublicKey]; !exists {
					outputIf(!quiet, fmt.Sprintf("Unknown validator public key for deposit %d", i))
					failures = true
				}
			}
			outputIf(!quiet, fmt.Sprintf("Deposit %q verified", deposit.Name))
		}

		if failures {
			os.Exit(_exitFailure)
		}
		os.Exit(_exitSuccess)
	},
}

func validatorPubKeysFromInput(input string) (map[string]bool, error) {
	pubKeys := make(map[string]bool)
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
		pubKeys[fmt.Sprintf("%x", pubKey.Marshal())] = true
	} else {
		// Assume it's a path to a file of public keys.
		data, err = ioutil.ReadFile(input)
		if err != nil {
			return nil, errors.Wrap(err, "failed to find public key file")
		}
		lines := bytes.Split(bytes.Replace(data, []byte("\r\n"), []byte("\n"), -1), []byte("\n"))
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
			pubKeys[fmt.Sprintf("%x", pubKey.Marshal())] = true
		}
	}

	return pubKeys, nil
}

func depositDataFromJSON(input string) ([]*depositData, error) {
	var err error
	var data []byte
	// Input could be JSON or a path to JSON
	switch {
	case strings.HasPrefix(input, "{"):
		// Looks like JSON
		data = []byte("[" + input + "]")
	case strings.HasPrefix(input, "["):
		// Looks like JSON array
		data = []byte(input)
	default:
		// Assume it's a path to JSON
		data, err = ioutil.ReadFile(input)
		if err != nil {
			return nil, errors.Wrap(err, "failed to find deposit data file")
		}
		if data[0] == '{' {
			data = []byte("[" + string(data) + "]")
		}
	}
	var depositData []*depositData
	err = json.Unmarshal(data, &depositData)
	if err != nil {
		return nil, errors.Wrap(err, "data is not valid JSON")
	}
	if len(depositData) == 0 {
		return nil, errors.New("no deposits supplied")
	}
	minVersion := depositData[0].Version
	maxVersion := depositData[0].Version
	for i := range depositData {
		if depositData[i].PublicKey == "" {
			return nil, fmt.Errorf("no public key for deposit %d", i)
		}
		if depositData[i].DepositDataRoot == "" {
			return nil, fmt.Errorf("no data root for deposit %d", i)
		}
		if depositData[i].Signature == "" {
			return nil, fmt.Errorf("no signature for deposit %d", i)
		}
		if depositData[i].WithdrawalCredentials == "" {
			return nil, fmt.Errorf("no withdrawal credentials for deposit %d", i)
		}
		if depositData[i].Value < 1000000000 {
			return nil, fmt.Errorf("Deposit amount too small for deposit %d", i)
		}
		if depositData[i].Version > maxVersion {
			maxVersion = depositData[i].Version
		}
		if depositData[i].Version < minVersion {
			minVersion = depositData[i].Version
		}
	}
	return depositData, nil
}

func init() {
	depositCmd.AddCommand(depositVerifyCmd)
	depositFlags(depositVerifyCmd)
	depositVerifyCmd.Flags().StringVar(&depositVerifyData, "data", "", "JSON data, or path to JSON data")
	depositVerifyCmd.Flags().StringVar(&depositVerifyWithdrawalPubKey, "withdrawalpubkey", "", "Public key of the account to which the validator funds will be withdrawn")
	depositVerifyCmd.Flags().StringVar(&depositVerifyDepositValue, "depositvalue", "", "Value of the amount to be deposited")
	depositVerifyCmd.Flags().StringVar(&depositVerifyValidatorPubKey, "validatorpubkey", "", "Public key(s) of the account(s) that will be carrying out validation")
}
