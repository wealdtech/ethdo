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

package accountkey

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/wealdtech/ethdo/util"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

type dataIn struct {
	timeout     time.Duration
	account     e2wtypes.Account
	passphrases []string
}

func input(ctx context.Context) (*dataIn, error) {
	var err error
	data := &dataIn{}

	if viper.GetDuration("timeout") == 0 {
		return nil, errors.New("timeout is required")
	}
	data.timeout = viper.GetDuration("timeout")

	// Account.
	_, data.account, err = util.WalletAndAccountFromInput(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to obtain acount")
	}

	// Passphrases.
	data.passphrases = util.GetPassphrases()

	return data, nil
}
