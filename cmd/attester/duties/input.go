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
	"time"

	eth2client "github.com/attestantio/go-eth2-client"
	spec "github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/wealdtech/ethdo/util"
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
	account    string
	pubKey     string
	eth2Client eth2client.Service
	epoch      spec.Epoch
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

	// Account or pubkey.
	if viper.GetString("account") == "" && viper.GetString("pubkey") == "" {
		return nil, errors.New("account or pubkey is required")
	}
	data.account = viper.GetString("account")
	data.pubKey = viper.GetString("pubkey")

	// Ethereum 2 client.
	var err error
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

	return data, nil
}
