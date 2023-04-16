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
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestInput(t *testing.T) {
	tests := []struct {
		name string
		vars map[string]interface{}
		res  *dataIn
		err  string
	}{
		{
			name: "WithdrawalCredentialsMissing",
			vars: map[string]interface{}{},
			err:  "withdrawal credentials are required",
		},
		{
			name: "MnemonicAndPrivateKeyMissing",
			vars: map[string]interface{}{
				"withdrawal-credentials": "0x007e28dcf9029e8d92ca4b5d01c66c934e7f3110606f34ae3052cbf67bd3fc02",
			},
			err: "mnemonic or private key is required",
		},
		{
			name: "GoodWithMnemonic",
			vars: map[string]interface{}{
				"withdrawal-credentials": "0x007e28dcf9029e8d92ca4b5d01c66c934e7f3110606f34ae3052cbf67bd3fc02",
				"mnemonic":               "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art",
			},
			res: &dataIn{
				withdrawalCredentials: "0x007e28dcf9029e8d92ca4b5d01c66c934e7f3110606f34ae3052cbf67bd3fc02",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			viper.Reset()

			for k, v := range test.vars {
				viper.Set(k, v)
			}
			res, err := input(context.Background())
			if test.err != "" {
				require.EqualError(t, err, test.err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.res.withdrawalCredentials, res.withdrawalCredentials)
			}
		})
	}
}
