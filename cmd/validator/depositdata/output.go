// Copyright © 2019 - 2022 Weald Technology Trading.
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

package depositdata

import (
	"fmt"
	"strings"

	spec "github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
)

type dataOut struct {
	format                string
	account               string
	validatorPubKey       *spec.BLSPubKey
	withdrawalCredentials []byte
	amount                spec.Gwei
	signature             *spec.BLSSignature
	forkVersion           *spec.Version
	depositDataRoot       *spec.Root
	depositMessageRoot    *spec.Root
}

func output(data []*dataOut) (string, error) {
	outputs := make([]string, 0)
	for _, datum := range data {
		if datum == nil {
			continue
		}
		var output string
		var err error
		switch datum.format {
		case "raw":
			output, err = validatorDepositDataOutputRaw(datum)
		case "launchpad":
			output, err = validatorDepositDataOutputLaunchpad(datum)
		default:
			output, err = validatorDepositDataOutputJSON(datum)
		}
		if err != nil {
			return "", err
		}
		outputs = append(outputs, output)
	}
	return fmt.Sprintf("[%s]", strings.Join(outputs, ",")), nil
}

func validatorDepositDataOutputRaw(datum *dataOut) (string, error) {
	if datum.validatorPubKey == nil {
		return "", errors.New("validator public key required")
	}
	if len(datum.withdrawalCredentials) != 32 {
		return "", errors.New("withdrawal credentials must be 32 bytes")
	}
	if datum.amount == 0 {
		return "", errors.New("missing amount")
	}
	if datum.signature == nil {
		return "", errors.New("signature required")
	}
	if datum.depositDataRoot == nil {
		return "", errors.New("deposit data root required")
	}

	output := fmt.Sprintf(
		`"`+
			// Function signature.
			"0x22895118"+
			// Pointer to validator public key.
			"0000000000000000000000000000000000000000000000000000000000000080"+
			// Pointer to withdrawal credentials.
			"00000000000000000000000000000000000000000000000000000000000000e0"+
			// Pointer to validator signature.
			"0000000000000000000000000000000000000000000000000000000000000120"+
			// Deposit data root.
			"%x"+
			// Validator public key (padded).
			"0000000000000000000000000000000000000000000000000000000000000030"+
			"%x00000000000000000000000000000000"+
			// Withdrawal credentials.
			"0000000000000000000000000000000000000000000000000000000000000020"+
			"%x"+
			// Deposit signature.
			"0000000000000000000000000000000000000000000000000000000000000060"+
			"%x"+
			`"`,
		*datum.depositDataRoot,
		*datum.validatorPubKey,
		datum.withdrawalCredentials,
		*datum.signature,
	)
	return output, nil
}

func validatorDepositDataOutputLaunchpad(datum *dataOut) (string, error) {
	// Map of fork version to network name.
	forkVersionMap := map[spec.Version]string{
		[4]byte{0x00, 0x00, 0x00, 0x00}: "mainnet",
		[4]byte{0x00, 0x00, 0x20, 0x09}: "pyrmont",
		[4]byte{0x00, 0x00, 0x10, 0x20}: "goerli",
		[4]byte{0x80, 0x00, 0x00, 0x69}: "ropsten",
		[4]byte{0x90, 0x00, 0x00, 0x69}: "sepolia",
		[4]byte{0x00, 0x01, 0x70, 0x00}: "holesky",
	}

	if datum.validatorPubKey == nil {
		return "", errors.New("validator public key required")
	}
	if len(datum.withdrawalCredentials) != 32 {
		return "", errors.New("withdrawal credentials must be 32 bytes")
	}
	if datum.amount == 0 {
		return "", errors.New("missing amount")
	}
	if datum.signature == nil {
		return "", errors.New("signature required")
	}
	if datum.depositMessageRoot == nil {
		return "", errors.New("deposit message root required")
	}
	if datum.depositDataRoot == nil {
		return "", errors.New("deposit data root required")
	}

	networkName := "unknown"
	if network, exists := forkVersionMap[*datum.forkVersion]; exists {
		networkName = network
	}
	output := fmt.Sprintf(`{"pubkey":"%x","withdrawal_credentials":"%x","amount":%d,"signature":"%x","deposit_message_root":"%x","deposit_data_root":"%x","fork_version":"%x","eth2_network_name":"%s","deposit_cli_version":"2.5.0"}`,
		*datum.validatorPubKey,
		datum.withdrawalCredentials,
		datum.amount,
		*datum.signature,
		*datum.depositMessageRoot,
		*datum.depositDataRoot,
		*datum.forkVersion,
		networkName,
	)
	return output, nil
}

func validatorDepositDataOutputJSON(datum *dataOut) (string, error) {
	if datum.account == "" {
		return "", errors.New("missing account")
	}
	if datum.validatorPubKey == nil {
		return "", errors.New("validator public key required")
	}
	if len(datum.withdrawalCredentials) != 32 {
		return "", errors.New("withdrawal credentials must be 32 bytes")
	}
	if datum.signature == nil {
		return "", errors.New("signature required")
	}
	if datum.amount == 0 {
		return "", errors.New("missing amount")
	}
	if datum.depositDataRoot == nil {
		return "", errors.New("deposit data root required")
	}
	if datum.depositDataRoot == nil {
		return "", errors.New("deposit message root required")
	}
	if datum.forkVersion == nil {
		return "", errors.New("fork version required")
	}

	output := fmt.Sprintf(`{"name":"Deposit for %s","account":"%s","pubkey":"%#x","withdrawal_credentials":"%#x","signature":"%#x","amount":%d,"deposit_data_root":"%#x","deposit_message_root":"%#x","fork_version":"%#x","version":3}`,
		datum.account,
		datum.account,
		*datum.validatorPubKey,
		datum.withdrawalCredentials,
		*datum.signature,
		datum.amount,
		*datum.depositDataRoot,
		*datum.depositMessageRoot,
		*datum.forkVersion,
	)
	return output, nil
}
