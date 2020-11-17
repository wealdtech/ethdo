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

package util

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// DepositInfo is a generic deposit structure.
type DepositInfo struct {
	Name                  string
	Account               string
	PublicKey             []byte
	WithdrawalCredentials []byte
	Signature             []byte
	DepositDataRoot       []byte
	DepositMessageRoot    []byte
	ForkVersion           []byte
	Amount                uint64
	Version               uint64
}

// depositInfoV1 is an ethdo V1 deposit structure.
type depositInfoV1 struct {
	Name                  string `json:"name,omitempty"`
	Account               string `json:"account,omitempty"`
	PublicKey             string `json:"pubkey"`
	WithdrawalCredentials string `json:"withdrawal_credentials"`
	Signature             string `json:"signature"`
	DepositDataRoot       string `json:"deposit_data_root"`
	Value                 uint64 `json:"value"`
	Version               uint64 `json:"version"`
}

// depositInfoV3 is an ethdo V3 deposit structure.
type depositInfoV3 struct {
	Name                  string `json:"name,omitempty"`
	Account               string `json:"account,omitempty"`
	PublicKey             string `json:"pubkey"`
	WithdrawalCredentials string `json:"withdrawal_credentials"`
	Signature             string `json:"signature"`
	DepositDataRoot       string `json:"deposit_data_root"`
	DepositMessageRoot    string `json:"deposit_message_root"`
	ForkVersion           string `json:"fork_version"`
	Amount                uint64 `json:"amount"`
	Version               uint64 `json:"version"`
}

// depositInfoCLI is a deposit structure from the eth2 deposit CLI.
type depositInfoCLI struct {
	PublicKey             string `json:"pubkey"`
	WithdrawalCredentials string `json:"withdrawal_credentials"`
	Signature             string `json:"signature"`
	DepositDataRoot       string `json:"deposit_data_root"`
	DepositMessageRoot    string `json:"deposit_message_root"`
	ForkVersion           string `json:"fork_version"`
	Amount                uint64 `json:"amount"`
}

// DepositInfoFromJSON obtains deposit info from various possibly formx of JSON.
func DepositInfoFromJSON(input []byte) ([]*DepositInfo, error) {
	if len(input) == 0 {
		return nil, errors.New("no data supplied")
	}
	// Work out the type of data that we're dealing with, and decode it appropriately.
	depositInfo, err := tryRawTxData(input)
	if err != nil {
		// At this point, ensure we have brackets around the deposit info.
		if input[0] != '[' {
			input = []byte(fmt.Sprintf("[%s]", string(input)))
		}
		depositInfo, err = tryV3DepositInfoFromJSON(input)
		if err != nil {
			depositInfo, err = tryV1DepositInfoFromJSON(input)
			if err != nil {
				depositInfo, err = tryCLIDepositInfoFromJSON(input)
				if err != nil {
					// Give up
					return nil, errors.New("unknown deposit data format")
				}
			}
		}
	}

	if len(depositInfo) == 0 {
		return nil, errors.New("no deposits supplied")
	}

	return depositInfo, nil
}

func tryV3DepositInfoFromJSON(data []byte) ([]*DepositInfo, error) {
	var depositData []*depositInfoV3
	err := json.Unmarshal(data, &depositData)
	if err != nil {
		return nil, err
	}

	depositInfos := make([]*DepositInfo, len(depositData))
	for i, deposit := range depositData {
		if deposit.Version != 3 {
			return nil, errors.New("incorrect V3 deposit version")
		}
		if deposit.PublicKey == "" {
			return nil, errors.New("public key missing")
		}
		publicKey, err := hex.DecodeString(strings.TrimPrefix(deposit.PublicKey, "0x"))
		if err != nil {
			return nil, errors.New("public key invalid")
		}
		if deposit.WithdrawalCredentials == "" {
			return nil, errors.New("withdrawal credentials missing")
		}
		withdrawalCredentials, err := hex.DecodeString(strings.TrimPrefix(deposit.WithdrawalCredentials, "0x"))
		if err != nil {
			return nil, errors.New("withdrawal credentials invalid")
		}
		if deposit.Signature == "" {
			return nil, errors.New("signature missing")
		}
		signature, err := hex.DecodeString(strings.TrimPrefix(deposit.Signature, "0x"))
		if err != nil {
			return nil, errors.New("signature invalid")
		}
		if deposit.DepositDataRoot == "" {
			return nil, errors.New("deposit data root missing")
		}
		depositDataRoot, err := hex.DecodeString(strings.TrimPrefix(deposit.DepositDataRoot, "0x"))
		if err != nil {
			return nil, errors.New("deposit data root invalid")
		}
		if deposit.DepositMessageRoot == "" {
			return nil, errors.New("deposit message root missing")
		}
		depositMessageRoot, err := hex.DecodeString(strings.TrimPrefix(deposit.DepositMessageRoot, "0x"))
		if err != nil {
			return nil, errors.New("deposit message root invalid")
		}
		if deposit.ForkVersion == "" {
			return nil, errors.New("fork version missing")
		}
		forkVersion, err := hex.DecodeString(strings.TrimPrefix(deposit.ForkVersion, "0x"))
		if err != nil {
			return nil, errors.New("fork version invalid")
		}
		depositInfos[i] = &DepositInfo{
			Name:                  deposit.Name,
			Account:               deposit.Account,
			PublicKey:             publicKey,
			WithdrawalCredentials: withdrawalCredentials,
			Signature:             signature,
			DepositDataRoot:       depositDataRoot,
			DepositMessageRoot:    depositMessageRoot,
			ForkVersion:           forkVersion,
			Amount:                deposit.Amount,
			Version:               3,
		}
	}

	return depositInfos, nil
}

func tryCLIDepositInfoFromJSON(data []byte) ([]*DepositInfo, error) {
	var depositData []*depositInfoCLI
	err := json.Unmarshal(data, &depositData)
	if err != nil {
		return nil, err
	}

	depositInfos := make([]*DepositInfo, len(depositData))
	for i, deposit := range depositData {
		if deposit.PublicKey == "" {
			return nil, errors.New("public key missing")
		}
		publicKey, err := hex.DecodeString(strings.TrimPrefix(deposit.PublicKey, "0x"))
		if err != nil {
			return nil, errors.New("public key invalid")
		}
		if deposit.WithdrawalCredentials == "" {
			return nil, errors.New("withdrawal credentials missing")
		}
		withdrawalCredentials, err := hex.DecodeString(strings.TrimPrefix(deposit.WithdrawalCredentials, "0x"))
		if err != nil {
			return nil, errors.New("withdrawal credentials invalid")
		}
		if deposit.Signature == "" {
			return nil, errors.New("signature missing")
		}
		signature, err := hex.DecodeString(strings.TrimPrefix(deposit.Signature, "0x"))
		if err != nil {
			return nil, errors.New("signature invalid")
		}
		if deposit.DepositDataRoot == "" {
			return nil, errors.New("deposit data root missing")
		}
		depositDataRoot, err := hex.DecodeString(strings.TrimPrefix(deposit.DepositDataRoot, "0x"))
		if err != nil {
			return nil, errors.New("deposit data root invalid")
		}
		if deposit.DepositMessageRoot == "" {
			return nil, errors.New("deposit message root missing")
		}
		depositMessageRoot, err := hex.DecodeString(strings.TrimPrefix(deposit.DepositMessageRoot, "0x"))
		if err != nil {
			return nil, errors.New("deposit message root invalid")
		}
		if deposit.ForkVersion == "" {
			return nil, errors.New("fork version missing")
		}
		forkVersion, err := hex.DecodeString(strings.TrimPrefix(deposit.ForkVersion, "0x"))
		if err != nil {
			return nil, errors.New("fork version invalid")
		}
		depositInfos[i] = &DepositInfo{
			PublicKey:             publicKey,
			WithdrawalCredentials: withdrawalCredentials,
			Signature:             signature,
			DepositDataRoot:       depositDataRoot,
			DepositMessageRoot:    depositMessageRoot,
			ForkVersion:           forkVersion,
			Amount:                deposit.Amount,
			Version:               3,
		}
	}

	return depositInfos, nil
}

func tryV1DepositInfoFromJSON(data []byte) ([]*DepositInfo, error) {
	var depositData []*depositInfoV1
	err := json.Unmarshal(data, &depositData)
	if err != nil {
		return nil, err
	}

	depositInfos := make([]*DepositInfo, len(depositData))
	for i, deposit := range depositData {
		if deposit.Version < 1 || deposit.Version > 2 {
			return nil, errors.New("incorrect deposit version")
		}
		publicKey, err := hex.DecodeString(strings.TrimPrefix(deposit.PublicKey, "0x"))
		if err != nil {
			return nil, errors.New("public key invalid")
		}
		withdrawalCredentials, err := hex.DecodeString(strings.TrimPrefix(deposit.WithdrawalCredentials, "0x"))
		if err != nil {
			return nil, errors.New("withdrawal credentials invalid")
		}
		signature, err := hex.DecodeString(strings.TrimPrefix(deposit.Signature, "0x"))
		if err != nil {
			return nil, errors.New("signature invalid")
		}
		depositDataRoot, err := hex.DecodeString(strings.TrimPrefix(deposit.DepositDataRoot, "0x"))
		if err != nil {
			return nil, errors.New("deposit data root invalid")
		}
		depositInfos[i] = &DepositInfo{
			Name:                  deposit.Name,
			Account:               deposit.Account,
			PublicKey:             publicKey,
			WithdrawalCredentials: withdrawalCredentials,
			Signature:             signature,
			DepositDataRoot:       depositDataRoot,
			Amount:                deposit.Value,
			Version:               3,
		}
	}

	return depositInfos, nil
}

func tryRawTxData(data []byte) ([]*DepositInfo, error) {
	txData, err := hex.DecodeString(strings.TrimPrefix(string(data), "0x"))
	if err != nil {
		return nil, errors.New("public key invalid")
	}

	depositInfos := make([]*DepositInfo, 1)

	if len(txData) != 420 {
		return nil, errors.New("invalid transaction length")
	}
	if !bytes.Equal(txData[0:4], []byte{0x22, 0x89, 0x51, 0x18}) {
		return nil, errors.New("invalid function signature")
	}

	depositInfos[0] = &DepositInfo{
		PublicKey:             txData[164:212],
		WithdrawalCredentials: txData[260:292],
		Signature:             txData[324:420],
		DepositDataRoot:       txData[100:132],
	}

	return depositInfos, nil
}
