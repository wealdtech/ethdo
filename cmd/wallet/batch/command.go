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

package walletbatch

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/wealdtech/ethdo/util"
)

type command struct {
	quiet   bool
	verbose bool
	debug   bool

	timeout time.Duration

	// Operation.
	walletName      string
	passphrases     []string
	batchPassphrase string
}

func newCommand(_ context.Context) (*command, error) {
	c := &command{
		quiet:           viper.GetBool("quiet"),
		verbose:         viper.GetBool("verbose"),
		debug:           viper.GetBool("debug"),
		timeout:         viper.GetDuration("timeout"),
		walletName:      viper.GetString("wallet"),
		passphrases:     util.GetPassphrases(),
		batchPassphrase: viper.GetString("batch-passphrase"),
	}

	if c.timeout == 0 {
		return nil, errors.New("timeout is required")
	}

	if c.walletName == "" {
		return nil, errors.New("wallet is required")
	}

	if c.batchPassphrase == "" {
		return nil, errors.New("batch passphrase is required")
	}

	return c, nil
}
