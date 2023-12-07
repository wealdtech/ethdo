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
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	consensusclient "github.com/attestantio/go-eth2-client"
	"github.com/attestantio/go-eth2-client/api"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
	"github.com/wealdtech/ethdo/services/chaintime"
	"github.com/wealdtech/ethdo/util"
)

type ChainInfo struct {
	Version                        uint64
	Validators                     []*ValidatorInfo
	GenesisValidatorsRoot          phase0.Root
	Epoch                          phase0.Epoch
	GenesisForkVersion             phase0.Version
	ExitForkVersion                phase0.Version
	CurrentForkVersion             phase0.Version
	BLSToExecutionChangeDomainType phase0.DomainType
	VoluntaryExitDomainType        phase0.DomainType
}

type chainInfoJSON struct {
	Version                        string           `json:"version"`
	Validators                     []*ValidatorInfo `json:"validators"`
	GenesisValidatorsRoot          string           `json:"genesis_validators_root"`
	Epoch                          string           `json:"epoch"`
	GenesisForkVersion             string           `json:"genesis_fork_version"`
	ExitForkVersion                string           `json:"exit_fork_version"`
	CurrentForkVersion             string           `json:"current_fork_version"`
	BLSToExecutionChangeDomainType string           `json:"bls_to_execution_change_domain_type"`
	VoluntaryExitDomainType        string           `json:"voluntary_exit_domain_type"`
}

type chainInfoVersionJSON struct {
	Version string `json:"version"`
}

// MarshalJSON implements json.Marshaler.
func (c *ChainInfo) MarshalJSON() ([]byte, error) {
	return json.Marshal(&chainInfoJSON{
		Version:                        strconv.FormatUint(c.Version, 10),
		Validators:                     c.Validators,
		GenesisValidatorsRoot:          fmt.Sprintf("%#x", c.GenesisValidatorsRoot),
		Epoch:                          fmt.Sprintf("%d", c.Epoch),
		GenesisForkVersion:             fmt.Sprintf("%#x", c.GenesisForkVersion),
		ExitForkVersion:                fmt.Sprintf("%#x", c.ExitForkVersion),
		CurrentForkVersion:             fmt.Sprintf("%#x", c.CurrentForkVersion),
		BLSToExecutionChangeDomainType: fmt.Sprintf("%#x", c.BLSToExecutionChangeDomainType),
		VoluntaryExitDomainType:        fmt.Sprintf("%#x", c.VoluntaryExitDomainType),
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *ChainInfo) UnmarshalJSON(input []byte) error {
	// See which version we are dealing with.
	var metadata chainInfoVersionJSON
	if err := json.Unmarshal(input, &metadata); err != nil {
		return errors.Wrap(err, "invalid JSON")
	}
	if metadata.Version == "" {
		return errors.New("version missing")
	}
	version, err := strconv.ParseUint(metadata.Version, 10, 64)
	if err != nil {
		return errors.Wrap(err, "version invalid")
	}
	if version < 3 {
		return errors.New("outdated version; please regenerate your offline data")
	}
	c.Version = version

	var data chainInfoJSON
	if err := json.Unmarshal(input, &data); err != nil {
		return errors.Wrap(err, "invalid JSON")
	}

	if len(data.Validators) == 0 {
		return errors.New("validators missing")
	}
	c.Validators = data.Validators

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
	copy(c.GenesisValidatorsRoot[:], genesisValidatorsRootBytes)

	if data.Epoch == "" {
		return errors.New("epoch missing")
	}
	epoch, err := strconv.ParseUint(data.Epoch, 10, 64)
	if err != nil {
		return errors.Wrap(err, "epoch invalid")
	}
	c.Epoch = phase0.Epoch(epoch)

	if data.GenesisForkVersion == "" {
		return errors.New("genesis fork version missing")
	}
	genesisForkVersionBytes, err := hex.DecodeString(strings.TrimPrefix(data.GenesisForkVersion, "0x"))
	if err != nil {
		return errors.Wrap(err, "genesis fork version invalid")
	}
	if len(genesisForkVersionBytes) != phase0.ForkVersionLength {
		return errors.New("genesis fork version incorrect length")
	}
	copy(c.GenesisForkVersion[:], genesisForkVersionBytes)

	if data.ExitForkVersion == "" {
		return errors.New("exit fork version missing")
	}
	exitForkVersionBytes, err := hex.DecodeString(strings.TrimPrefix(data.ExitForkVersion, "0x"))
	if err != nil {
		return errors.Wrap(err, "exit fork version invalid")
	}
	if len(exitForkVersionBytes) != phase0.ForkVersionLength {
		return errors.New("exit fork version incorrect length")
	}
	copy(c.ExitForkVersion[:], exitForkVersionBytes)

	if data.CurrentForkVersion == "" {
		return errors.New("current fork version missing")
	}
	currentForkVersionBytes, err := hex.DecodeString(strings.TrimPrefix(data.CurrentForkVersion, "0x"))
	if err != nil {
		return errors.Wrap(err, "current fork version invalid")
	}
	if len(currentForkVersionBytes) != phase0.ForkVersionLength {
		return errors.New("current fork version incorrect length")
	}
	copy(c.CurrentForkVersion[:], currentForkVersionBytes)

	if data.BLSToExecutionChangeDomainType == "" {
		return errors.New("bls to execution domain type missing")
	}
	blsToExecutionChangeDomainType, err := hex.DecodeString(strings.TrimPrefix(data.BLSToExecutionChangeDomainType, "0x"))
	if err != nil {
		return errors.Wrap(err, "bls to execution domain type invalid")
	}
	if len(blsToExecutionChangeDomainType) != phase0.DomainTypeLength {
		return errors.New("bls to execution domain type incorrect length")
	}
	copy(c.BLSToExecutionChangeDomainType[:], blsToExecutionChangeDomainType)

	if data.VoluntaryExitDomainType == "" {
		return errors.New("voluntary exit domain type missing")
	}
	voluntaryExitDomainType, err := hex.DecodeString(strings.TrimPrefix(data.VoluntaryExitDomainType, "0x"))
	if err != nil {
		return errors.Wrap(err, "voluntary exit domain type invalid")
	}
	if len(voluntaryExitDomainType) != phase0.DomainTypeLength {
		return errors.New("voluntary exit domain type incorrect length")
	}
	copy(c.VoluntaryExitDomainType[:], voluntaryExitDomainType)

	return nil
}

// FetchValidatorInfo fetches validator info given a validator identifier.
func (c *ChainInfo) FetchValidatorInfo(ctx context.Context, id string) (*ValidatorInfo, error) {
	var validatorInfo *ValidatorInfo
	switch {
	case id == "":
		return nil, errors.New("no validator specified")
	case strings.HasPrefix(id, "0x"):
		// ID is a public key.
		// Check that the key is the correct length.
		if len(id) != 98 {
			return nil, errors.New("invalid public key: incorrect length")
		}
		for _, validator := range c.Validators {
			if strings.EqualFold(id, fmt.Sprintf("%#x", validator.Pubkey)) {
				validatorInfo = validator
				break
			}
		}
	case strings.Contains(id, "/"):
		// An account.
		_, account, err := util.WalletAndAccountFromPath(ctx, id)
		if err != nil {
			return nil, errors.Wrap(err, "unable to obtain account")
		}
		accPubKey, err := util.BestPublicKey(account)
		if err != nil {
			return nil, errors.Wrap(err, "unable to obtain public key for account")
		}
		pubkey := fmt.Sprintf("%#x", accPubKey.Marshal())
		for _, validator := range c.Validators {
			if strings.EqualFold(pubkey, fmt.Sprintf("%#x", validator.Pubkey)) {
				validatorInfo = validator
				break
			}
		}
	default:
		// An index.
		index, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse validator index")
		}
		validatorIndex := phase0.ValidatorIndex(index)
		for _, validator := range c.Validators {
			if validator.Index == validatorIndex {
				validatorInfo = validator
				break
			}
		}
	}

	if validatorInfo == nil {
		return nil, errors.New("unknown validator")
	}

	return validatorInfo, nil
}

// ObtainChainInfoFromNode obtains the chain information from a node.
func ObtainChainInfoFromNode(ctx context.Context,
	consensusClient consensusclient.Service,
	chainTime chaintime.Service,
) (
	*ChainInfo,
	error,
) {
	res := &ChainInfo{
		Version:    3,
		Validators: make([]*ValidatorInfo, 0),
		Epoch:      chainTime.CurrentEpoch(),
	}

	// Obtain validators.
	validatorsResponse, err := consensusClient.(consensusclient.ValidatorsProvider).Validators(ctx, &api.ValidatorsOpts{State: "head"})
	if err != nil {
		return nil, errors.Wrap(err, "failed to obtain validators")
	}

	for _, validator := range validatorsResponse.Data {
		res.Validators = append(res.Validators, &ValidatorInfo{
			Index:                 validator.Index,
			Pubkey:                validator.Validator.PublicKey,
			WithdrawalCredentials: validator.Validator.WithdrawalCredentials,
			State:                 validator.Status,
		})
	}

	// Genesis validators root obtained from beacon node.
	genesisResponse, err := consensusClient.(consensusclient.GenesisProvider).Genesis(ctx, &api.GenesisOpts{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to obtain genesis information")
	}
	res.GenesisValidatorsRoot = genesisResponse.Data.GenesisValidatorsRoot

	// Fetch the genesis fork version from the specification.
	specResponse, err := consensusClient.(consensusclient.SpecProvider).Spec(ctx, &api.SpecOpts{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to obtain spec")
	}
	tmp, exists := specResponse.Data["GENESIS_FORK_VERSION"]
	if !exists {
		return nil, errors.New("genesis fork version not known by chain")
	}
	var isForkVersion bool
	res.GenesisForkVersion, isForkVersion = tmp.(phase0.Version)
	if !isForkVersion {
		return nil, errors.New("could not obtain GENESIS_FORK_VERSION")
	}

	// Fetch the exit fork version (Capella) from the specification.
	tmp, exists = specResponse.Data["CAPELLA_FORK_VERSION"]
	if !exists {
		return nil, errors.New("capella fork version not known by chain")
	}
	res.ExitForkVersion, isForkVersion = tmp.(phase0.Version)
	if !isForkVersion {
		return nil, errors.New("could not obtain CAPELLA_FORK_VERSION")
	}

	// Fetch the current fork version from the fork schedule.
	forkScheduleResponse, err := consensusClient.(consensusclient.ForkScheduleProvider).ForkSchedule(ctx, &api.ForkScheduleOpts{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to obtain fork schedule")
	}
	for i := range forkScheduleResponse.Data {
		if forkScheduleResponse.Data[i].Epoch <= res.Epoch {
			res.CurrentForkVersion = forkScheduleResponse.Data[i].CurrentVersion
		}
	}

	blsToExecutionChangeDomainType, exists := specResponse.Data["DOMAIN_BLS_TO_EXECUTION_CHANGE"].(phase0.DomainType)
	if !exists {
		return nil, errors.New("failed to obtain DOMAIN_BLS_TO_EXECUTION_CHANGE")
	}
	copy(res.BLSToExecutionChangeDomainType[:], blsToExecutionChangeDomainType[:])

	voluntaryExitDomainType, exists := specResponse.Data["DOMAIN_VOLUNTARY_EXIT"].(phase0.DomainType)
	if !exists {
		return nil, errors.New("failed to obtain DOMAIN_VOLUNTARY_EXIT")
	}
	copy(res.VoluntaryExitDomainType[:], voluntaryExitDomainType[:])

	return res, nil
}
