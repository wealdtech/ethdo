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

package validatorexit

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	eth2client "github.com/attestantio/go-eth2-client"
	spec "github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/wealdtech/ethdo/core"
	"github.com/wealdtech/ethdo/util"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

type dataIn struct {
	// System.
	timeout time.Duration
	quiet   bool
	verbose bool
	debug   bool
	// Operation.
	eth2Client eth2client.Service
	jsonOutput bool
	// Chain information.
	fork         *spec.Fork
	currentEpoch spec.Epoch
	// Exit information.
	account             e2wtypes.Account
	passphrases         []string
	epoch               spec.Epoch
	domain              spec.Domain
	signedVoluntaryExit *spec.SignedVoluntaryExit
}

func input(ctx context.Context) (*dataIn, error) {
	data := &dataIn{}

	if viper.GetDuration("timeout") == 0 {
		return nil, errors.New("timeout is required")
	}
	data.timeout = viper.GetDuration("timeout")
	data.quiet = viper.GetBool("quiet")
	data.verbose = viper.GetBool("verbose")
	data.debug = viper.GetBool("debug")
	data.passphrases = util.GetPassphrases()
	data.jsonOutput = viper.GetBool("json")

	switch {
	case viper.GetString("exit") != "":
		return inputJSON(ctx, data)
	case viper.GetString("account") != "":
		return inputAccount(ctx, data)
	case viper.GetString("key") != "":
		return inputKey(ctx, data)
	default:
		return nil, errors.New("must supply account, key, or pre-constructed JSON")
	}
}

func inputJSON(ctx context.Context, data *dataIn) (*dataIn, error) {
	validatorData := &util.ValidatorExitData{}
	err := json.Unmarshal([]byte(viper.GetString("exit")), validatorData)
	if err != nil {
		return nil, err
	}
	data.signedVoluntaryExit = validatorData.Data
	return inputChainData(ctx, data)
}

func inputAccount(ctx context.Context, data *dataIn) (*dataIn, error) {
	var err error
	_, data.account, err = core.WalletAndAccountFromInput(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to obtain acount")
	}
	return inputChainData(ctx, data)
}

func inputKey(ctx context.Context, data *dataIn) (*dataIn, error) {
	privKeyBytes, err := hex.DecodeString(strings.TrimPrefix(viper.GetString("key"), "0x"))
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode key")
	}
	data.account, err = util.NewScratchAccount(privKeyBytes, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create acount from key")
	}
	return inputChainData(ctx, data)
}

func inputChainData(ctx context.Context, data *dataIn) (*dataIn, error) {
	var err error
	data.eth2Client, err = util.ConnectToBeaconNode(ctx, viper.GetString("connection"), viper.GetDuration("timeout"), viper.GetBool("allow-insecure-connections"))
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to Ethereum 2 beacon node")
	}

	// Current fork.
	data.fork, err = data.eth2Client.(eth2client.ForkProvider).Fork(ctx, "head")
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to obtain fork information")
	}

	// Calculate current epoch.
	config, err := data.eth2Client.(eth2client.SpecProvider).Spec(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to obtain configuration information")
	}
	genesis, err := data.eth2Client.(eth2client.GenesisProvider).Genesis(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to obtain genesis information")
	}
	data.currentEpoch = spec.Epoch(uint64(time.Since(genesis.GenesisTime).Seconds()) / (uint64(config["SECONDS_PER_SLOT"].(time.Duration).Seconds()) * config["SLOTS_PER_EPOCH"].(uint64)))

	// Epoch.
	if viper.GetInt64("epoch") == -1 {
		data.epoch = data.currentEpoch
	} else {
		data.epoch = spec.Epoch(viper.GetUint64("epoch"))
	}

	// Domain.
	domain, err := data.eth2Client.(eth2client.DomainProvider).Domain(ctx, config["DOMAIN_VOLUNTARY_EXIT"].(spec.DomainType), data.epoch)
	if err != nil {
		return nil, errors.New("failed to calculate domain")
	}
	data.domain = domain

	return data, nil
}
