// Copyright © 2023 Weald Technology Trading.
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

package beacon

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	apiv1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
)

type ValidatorInfo struct {
	Index                 phase0.ValidatorIndex
	Pubkey                phase0.BLSPubKey
	State                 apiv1.ValidatorState
	WithdrawalCredentials []byte
}

type validatorInfoJSON struct {
	Index                 string               `json:"index"`
	Pubkey                string               `json:"pubkey"`
	State                 apiv1.ValidatorState `json:"state"`
	WithdrawalCredentials string               `json:"withdrawal_credentials"`
}

// MarshalJSON implements json.Marshaler.
func (v *ValidatorInfo) MarshalJSON() ([]byte, error) {
	return json.Marshal(&validatorInfoJSON{
		Index:                 fmt.Sprintf("%d", v.Index),
		Pubkey:                fmt.Sprintf("%#x", v.Pubkey),
		State:                 v.State,
		WithdrawalCredentials: fmt.Sprintf("%#x", v.WithdrawalCredentials),
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (v *ValidatorInfo) UnmarshalJSON(input []byte) error {
	var data validatorInfoJSON
	if err := json.Unmarshal(input, &data); err != nil {
		return errors.Wrap(err, "invalid JSON")
	}

	if data.Index == "" {
		return errors.New("index missing")
	}
	index, err := strconv.ParseUint(data.Index, 10, 64)
	if err != nil {
		return errors.Wrap(err, "invalid value for index")
	}
	v.Index = phase0.ValidatorIndex(index)

	if data.Pubkey == "" {
		return errors.New("public key missing")
	}
	pubkey, err := hex.DecodeString(strings.TrimPrefix(data.Pubkey, "0x"))
	if err != nil {
		return errors.Wrap(err, "invalid value for public key")
	}
	if len(pubkey) != phase0.PublicKeyLength {
		return fmt.Errorf("incorrect length %d for public key", len(pubkey))
	}
	copy(v.Pubkey[:], pubkey)

	if data.State == apiv1.ValidatorStateUnknown {
		return errors.New("state unknown")
	}
	v.State = data.State

	if data.WithdrawalCredentials == "" {
		return errors.New("withdrawal credentials missing")
	}
	v.WithdrawalCredentials, err = hex.DecodeString(strings.TrimPrefix(data.WithdrawalCredentials, "0x"))
	if err != nil {
		return errors.Wrap(err, "invalid value for withdrawal credentials")
	}
	if len(v.WithdrawalCredentials) != phase0.HashLength {
		return fmt.Errorf("incorrect length %d for withdrawal credentials", len(v.WithdrawalCredentials))
	}

	return nil
}

// String implements the Stringer interface.
func (v *ValidatorInfo) String() string {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("Err: %v\n", err)
	}
	return string(data)
}
