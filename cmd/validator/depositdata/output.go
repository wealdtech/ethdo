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

package depositdata

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

type dataOut struct {
	format                string
	account               string
	validatorPubKey       []byte
	withdrawalCredentials []byte
	amount                uint64
	signature             []byte
	forkVersion           []byte
	depositDataRoot       []byte
	depositMessageRoot    []byte
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
	if len(datum.validatorPubKey) != 48 {
		return "", errors.New("validator public key must be 48 bytes")
	}
	if len(datum.withdrawalCredentials) != 32 {
		return "", errors.New("withdrawal credentials must be 32 bytes")
	}
	if datum.amount == 0 {
		return "", errors.New("missing amount")
	}
	if len(datum.signature) != 96 {
		return "", errors.New("signature must be 96 bytes")
	}
	if len(datum.depositMessageRoot) != 32 {
		return "", errors.New("deposit message root must be 32 bytes")
	}
	if len(datum.depositDataRoot) != 32 {
		return "", errors.New("deposit data root must be 32 bytes")
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
		datum.depositDataRoot,
		datum.validatorPubKey,
		datum.withdrawalCredentials,
		datum.signature,
	)
	return output, nil
}

func validatorDepositDataOutputLaunchpad(datum *dataOut) (string, error) {
	if len(datum.validatorPubKey) != 48 {
		return "", errors.New("validator public key must be 48 bytes")
	}
	if len(datum.withdrawalCredentials) != 32 {
		return "", errors.New("withdrawal credentials must be 32 bytes")
	}
	if datum.amount == 0 {
		return "", errors.New("missing amount")
	}
	if len(datum.signature) != 96 {
		return "", errors.New("signature must be 96 bytes")
	}
	if len(datum.depositMessageRoot) != 32 {
		return "", errors.New("deposit message root must be 32 bytes")
	}
	if len(datum.depositDataRoot) != 32 {
		return "", errors.New("deposit data root must be 32 bytes")
	}

	output := fmt.Sprintf(`{"pubkey":"%x","withdrawal_credentials":"%x","amount":%d,"signature":"%x","deposit_message_root":"%x","deposit_data_root":"%x","fork_version":"%x"}`,
		datum.validatorPubKey,
		datum.withdrawalCredentials,
		datum.amount,
		datum.signature,
		datum.depositMessageRoot,
		datum.depositDataRoot,
		datum.forkVersion,
	)
	return output, nil
}

func validatorDepositDataOutputJSON(datum *dataOut) (string, error) {
	if datum.account == "" {
		return "", errors.New("missing account")
	}
	if len(datum.validatorPubKey) != 48 {
		return "", errors.New("validator public key must be 48 bytes")
	}
	if len(datum.withdrawalCredentials) != 32 {
		return "", errors.New("withdrawal credentials must be 32 bytes")
	}
	if len(datum.signature) != 96 {
		return "", errors.New("signature must be 96 bytes")
	}
	if datum.amount == 0 {
		return "", errors.New("missing amount")
	}
	if len(datum.depositDataRoot) != 32 {
		return "", errors.New("deposit data root must be 32 bytes")
	}

	output := fmt.Sprintf(`{"name":"Deposit for %s","account":"%s","pubkey":"%#x","withdrawal_credentials":"%#x","signature":"%#x","value":%d,"deposit_data_root":"%#x","version":2}`,
		datum.account,
		datum.account,
		datum.validatorPubKey,
		datum.withdrawalCredentials,
		datum.signature,
		datum.amount,
		datum.depositDataRoot,
	)
	return output, nil
}
