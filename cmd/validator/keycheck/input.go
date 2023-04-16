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

package validatorkeycheck

import (
	"context"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type dataIn struct {
	// System.
	quiet   bool
	verbose bool
	debug   bool
	// Withdrawal credentials.
	withdrawalCredentials string
	// Operation.
	mnemonic string
	privKey  string
}

func input(_ context.Context) (*dataIn, error) {
	data := &dataIn{}

	data.quiet = viper.GetBool("quiet")
	data.verbose = viper.GetBool("verbose")
	data.debug = viper.GetBool("debug")

	// Withdrawal credentials.
	data.withdrawalCredentials = viper.GetString("withdrawal-credentials")
	if data.withdrawalCredentials == "" {
		return nil, errors.New("withdrawal credentials are required")
	}

	data.mnemonic = viper.GetString("mnemonic")
	data.privKey = viper.GetString("private-key")
	if data.mnemonic == "" && data.privKey == "" {
		return nil, errors.New("mnemonic or private key is required")
	}

	return data, nil
}
