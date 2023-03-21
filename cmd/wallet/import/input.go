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

package walletimport

import (
	"context"
	"encoding/hex"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/wealdtech/ethdo/util"
)

type dataIn struct {
	// System.
	timeout    time.Duration
	quiet      bool
	verbose    bool
	debug      bool
	data       []byte
	passphrase string
	verify     bool
}

func input(_ context.Context) (*dataIn, error) {
	var err error
	data := &dataIn{}

	if viper.GetString("remote") != "" {
		return nil, errors.New("wallet import not available for remote wallets")
	}

	if viper.GetDuration("timeout") == 0 {
		return nil, errors.New("timeout is required")
	}
	data.timeout = viper.GetDuration("timeout")
	data.quiet = viper.GetBool("quiet")
	data.verbose = viper.GetBool("verbose")
	data.debug = viper.GetBool("debug")

	// Data.
	if viper.GetString("data") == "" {
		return nil, errors.New("data is required")
	}
	if !strings.HasPrefix(viper.GetString("data"), "0x") {
		// Assume this is a path; read the file and replace the path with its contents.
		fileData, err := os.ReadFile(viper.GetString("data"))
		if err != nil {
			return nil, errors.Wrap(err, "failed to read wallet import data")
		}
		viper.Set("data", strings.TrimSpace(string(fileData)))
	}
	data.data, err = hex.DecodeString(strings.TrimPrefix(viper.GetString("data"), "0x"))
	if err != nil {
		return nil, errors.Wrap(err, "data is invalid")
	}

	// Passphrase.
	data.passphrase, err = util.GetPassphrase()
	if err != nil {
		return nil, errors.Wrap(err, "failed to obtain import passphrase")
	}

	// Verify.
	data.verify = viper.GetBool("verify")

	return data, nil
}
