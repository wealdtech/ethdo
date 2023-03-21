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

package chaintime

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type dataIn struct {
	// System.
	timeout time.Duration
	quiet   bool
	verbose bool
	debug   bool
	json    bool
	// Input
	connection               string
	allowInsecureConnections bool
	timestamp                string
	slot                     string
	epoch                    string
}

func input(_ context.Context) (*dataIn, error) {
	data := &dataIn{}

	if viper.GetDuration("timeout") == 0 {
		return nil, errors.New("timeout is required")
	}
	data.timeout = viper.GetDuration("timeout")
	data.quiet = viper.GetBool("quiet")
	data.verbose = viper.GetBool("verbose")
	data.debug = viper.GetBool("debug")
	data.json = viper.GetBool("json")

	haveInput := false
	if viper.GetString("timestamp") != "" {
		data.timestamp = viper.GetString("timestamp")
		haveInput = true
	}
	if viper.GetString("slot") != "" {
		if haveInput {
			return nil, errors.New("only one of timestamp, slot and epoch allowed")
		}
		data.slot = viper.GetString("slot")
		haveInput = true
	}
	if viper.GetString("epoch") != "" {
		if haveInput {
			return nil, errors.New("only one of timestamp, slot and epoch allowed")
		}
		data.epoch = viper.GetString("epoch")
		haveInput = true
	}
	if !haveInput {
		return nil, errors.New("one of timestamp, slot or epoch required")
	}

	data.connection = viper.GetString("connection")
	data.allowInsecureConnections = viper.GetBool("allow-insecure-connections")

	return data, nil
}
