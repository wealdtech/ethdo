// Copyright © 2021 Weald Technology Trading
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

package attesterduties

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	eth2client "github.com/attestantio/go-eth2-client"
	api "github.com/attestantio/go-eth2-client/api/v1"
	spec "github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/wealdtech/ethdo/util"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

type dataIn struct {
	// System.
	timeout time.Duration
	quiet   bool
	verbose bool
	debug   bool
	json    bool
	// Chain information.
	slotsPerEpoch uint64
	// Operation.
	validator  *api.Validator
	eth2Client eth2client.Service
	epoch      spec.Epoch
	account    e2wtypes.Account
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
	data.json = viper.GetBool("json")

	// Account.
	var err error
	data.account, err = attesterDutiesAccount()
	if err != nil {
		return nil, errors.Wrap(err, "failed to obtain account")
	}

	// Ethereum 2 client.
	data.eth2Client, err = util.ConnectToBeaconNode(ctx, viper.GetString("connection"), viper.GetDuration("timeout"), viper.GetBool("allow-insecure-connections"))
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to Ethereum 2 beacon node")
	}

	// Epoch
	epoch := viper.GetInt64("epoch")
	if epoch == -1 {
		config, err := data.eth2Client.(eth2client.SpecProvider).Spec(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "failed to obtain beacon chain configuration")
		}
		data.slotsPerEpoch = config["SLOTS_PER_EPOCH"].(uint64)
		slotDuration := config["SECONDS_PER_SLOT"].(time.Duration)
		genesis, err := data.eth2Client.(eth2client.GenesisProvider).Genesis(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "failed to obtain genesis data")
		}
		epoch = int64(time.Since(genesis.GenesisTime).Seconds()) / (int64(slotDuration.Seconds()) * int64(data.slotsPerEpoch))
	}
	data.epoch = spec.Epoch(epoch)

	pubKeys := make([]spec.BLSPubKey, 1)
	pubKey, err := util.BestPublicKey(data.account)
	if err != nil {
		return nil, errors.Wrap(err, "failed to obtain public key for account")
	}
	copy(pubKeys[0][:], pubKey.Marshal())
	validators, err := data.eth2Client.(eth2client.ValidatorsProvider).ValidatorsByPubKey(ctx, fmt.Sprintf("%d", uint64(data.epoch)*data.slotsPerEpoch), pubKeys)
	if err != nil {
		return nil, errors.New("failed to obtain validator information")
	}
	if len(validators) == 0 {
		return nil, errors.New("validator is not known")
	}
	data.validator = validators[0]

	return data, nil
}

// attesterDutiesAccount obtains the account for the attester duties command.
func attesterDutiesAccount() (e2wtypes.Account, error) {
	var account e2wtypes.Account
	var err error
	if viper.GetString("account") != "" {
		ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
		defer cancel()
		_, account, err = util.WalletAndAccountFromPath(ctx, viper.GetString("account"))
		if err != nil {
			return nil, errors.Wrap(err, "failed to obtain account")
		}
	} else {
		pubKey := viper.GetString("pubkey")
		pubKeyBytes, err := hex.DecodeString(strings.TrimPrefix(pubKey, "0x"))
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("failed to decode public key %s", pubKey))
		}
		account, err = util.NewScratchAccount(nil, pubKeyBytes)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("invalid public key %s", pubKey))
		}
	}
	return account, nil
}
