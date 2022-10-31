// Copyright © 2022 Weald Technology Trading.
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

package validatorcredentialsset

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
)

type chainInfo struct {
	Version               uint64
	Validators            []*validatorInfo
	GenesisValidatorsRoot phase0.Root
	Epoch                 phase0.Epoch
	ForkVersion           phase0.Version
	Domain                phase0.Domain
}

type chainInfoJSON struct {
	Version               string           `json:"version"`
	Validators            []*validatorInfo `json:"validators"`
	GenesisValidatorsRoot string           `json:"genesis_validators_root"`
	Epoch                 string           `json:"epoch"`
	ForkVersion           string           `json:"fork_version"`
	Domain                string           `json:"domain"`
}

// MarshalJSON implements json.Marshaler.
func (v *chainInfo) MarshalJSON() ([]byte, error) {
	return json.Marshal(&chainInfoJSON{
		Version:               fmt.Sprintf("%d", v.Version),
		Validators:            v.Validators,
		GenesisValidatorsRoot: fmt.Sprintf("%#x", v.GenesisValidatorsRoot),
		Epoch:                 fmt.Sprintf("%d", v.Epoch),
		ForkVersion:           fmt.Sprintf("%#x", v.ForkVersion),
		Domain:                fmt.Sprintf("%#x", v.Domain),
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (v *chainInfo) UnmarshalJSON(input []byte) error {
	var data chainInfoJSON
	if err := json.Unmarshal(input, &data); err != nil {
		return errors.Wrap(err, "invalid JSON")
	}

	if data.Version == "" {
		// Default to 1.
		v.Version = 1
	} else {
		version, err := strconv.ParseUint(data.Version, 10, 64)
		if err != nil {
			return errors.Wrap(err, "version invalid")
		}
		v.Version = version
	}

	if len(data.Validators) == 0 {
		return errors.New("validators missing")
	}
	v.Validators = data.Validators

	if data.GenesisValidatorsRoot == "" {
		return errors.New("genesis validators root missing")
	}

	genesisValidatorsRootBytes, err := hex.DecodeString(strings.TrimPrefix(data.GenesisValidatorsRoot, "0x"))
	if err != nil {
		return errors.Wrap(err, "genesis validators root invalid")
	}
	if len(genesisValidatorsRootBytes) != phase0.RootLength {
		return errors.New("genesis validators root incorrect length")
	}
	copy(v.GenesisValidatorsRoot[:], genesisValidatorsRootBytes)

	if data.Epoch == "" {
		return errors.New("epoch missing")
	}
	epoch, err := strconv.ParseUint(data.Epoch, 10, 64)
	if err != nil {
		return errors.Wrap(err, "epoch invalid")
	}
	v.Epoch = phase0.Epoch(epoch)

	if data.ForkVersion == "" {
		return errors.New("fork version missing")
	}
	forkVersionBytes, err := hex.DecodeString(strings.TrimPrefix(data.ForkVersion, "0x"))
	if err != nil {
		return errors.Wrap(err, "fork version invalid")
	}
	if len(forkVersionBytes) != phase0.ForkVersionLength {
		return errors.New("fork version incorrect length")
	}
	copy(v.ForkVersion[:], forkVersionBytes)

	if data.Domain == "" {
		return errors.New("domain missing")
	}
	domainBytes, err := hex.DecodeString(strings.TrimPrefix(data.Domain, "0x"))
	if err != nil {
		return errors.Wrap(err, "domain invalid")
	}
	if len(domainBytes) != phase0.DomainLength {
		return errors.New("domain incorrect length")
	}
	copy(v.Domain[:], domainBytes)

	return nil
}
