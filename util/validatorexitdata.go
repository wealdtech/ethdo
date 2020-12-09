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

package util

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	spec "github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
)

// ValidatorExitData contains data for a validator exit.
type ValidatorExitData struct {
	Exit        *spec.SignedVoluntaryExit
	ForkVersion spec.Version
}

type validatorExitJSON struct {
	Exit        *spec.SignedVoluntaryExit `json:"exit"`
	ForkVersion string                    `json:"fork_version"`
}

// MarshalJSON implements custom JSON marshaller.
func (d *ValidatorExitData) MarshalJSON() ([]byte, error) {
	validatorExitJSON := &validatorExitJSON{
		Exit:        d.Exit,
		ForkVersion: fmt.Sprintf("%#x", d.ForkVersion),
	}
	return json.Marshal(validatorExitJSON)
}

// UnmarshalJSON implements custom JSON unmarshaller.
func (d *ValidatorExitData) UnmarshalJSON(data []byte) error {
	validatorExitJSON := &validatorExitJSON{}

	if err := json.Unmarshal(data, validatorExitJSON); err != nil {
		return errors.Wrap(err, "failed to unmarshal JSON")
	}

	if validatorExitJSON.Exit == nil {
		return errors.New("exit missing")
	}
	d.Exit = validatorExitJSON.Exit

	if validatorExitJSON.ForkVersion == "" {
		return errors.New("fork version missing")
	}
	forkVersion, err := hex.DecodeString(strings.TrimPrefix(validatorExitJSON.ForkVersion, "0x"))
	if err != nil {
		return errors.Wrap(err, "fork version invalid")
	}
	copy(d.ForkVersion[:], forkVersion)

	return nil
}
