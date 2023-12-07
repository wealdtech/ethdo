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

package attesterinclusion

import (
	"context"
	"time"

	eth2client "github.com/attestantio/go-eth2-client"
	spec "github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/wealdtech/ethdo/services/chaintime"
	standardchaintime "github.com/wealdtech/ethdo/services/chaintime/standard"
	"github.com/wealdtech/ethdo/util"
)

type dataIn struct {
	// System.
	timeout time.Duration
	quiet   bool
	verbose bool
	debug   bool
	// Operation.
	eth2Client eth2client.Service
	chainTime  chaintime.Service
	epoch      spec.Epoch
	validator  string
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

	data.validator = viper.GetString("validator")
	if data.validator == "" {
		return nil, errors.New("validator is required")
	}

	// Ethereum 2 client.
	var err error
	data.eth2Client, err = util.ConnectToBeaconNode(ctx, &util.ConnectOpts{
		Address:       viper.GetString("connection"),
		Timeout:       viper.GetDuration("timeout"),
		AllowInsecure: viper.GetBool("allow-insecure-connections"),
		LogFallback:   !data.quiet,
	})
	if err != nil {
		return nil, err
	}

	data.chainTime, err = standardchaintime.New(ctx,
		standardchaintime.WithSpecProvider(data.eth2Client.(eth2client.SpecProvider)),
		standardchaintime.WithGenesisProvider(data.eth2Client.(eth2client.GenesisProvider)),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to set up chaintime service")
	}

	// Epoch.
	data.epoch, err = util.ParseEpoch(ctx, data.chainTime, viper.GetString("epoch"))
	if err != nil {
		return nil, err
	}

	return data, nil
}
