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
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	e2types "github.com/wealdtech/go-eth2-types/v2"
	keystorev4 "github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"
	nd "github.com/wealdtech/go-eth2-wallet-nd/v2"
	scratch "github.com/wealdtech/go-eth2-wallet-store-scratch"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

func TestProcess(t *testing.T) {
	require.NoError(t, e2types.InitBLS())

	testNDWallet, err := nd.CreateWallet(context.Background(),
		"Test",
		scratch.New(),
		keystorev4.New(),
	)
	require.NoError(t, err)
	require.NoError(t, testNDWallet.(e2wtypes.WalletLocker).Unlock(context.Background(), nil))
	viper.Set("passphrase", "pass")
	interop0, err := testNDWallet.(e2wtypes.WalletAccountImporter).ImportAccount(context.Background(),
		"Interop 0",
		hexToBytes("0x25295f0d1d592a90b333e26e85149708208e9f8e8bc18f6c77bd62f8ad7a6866"),
		[]byte("pass"),
	)
	require.NoError(t, err)

	tests := []struct {
		name   string
		dataIn *dataIn
		err    string
	}{
		{
			name: "Nil",
			err:  "no data",
		},
		{
			name: "PassphrasesMissing",
			dataIn: &dataIn{
				timeout: 5 * time.Second,
				account: interop0,
			},
			err: "passphrase is required",
		},
		{
			name: "Good",
			dataIn: &dataIn{
				timeout:     5 * time.Second,
				account:     interop0,
				passphrases: []string{"ce%NohGhah4ye5ra", "pass"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := process(context.Background(), test.dataIn)
			if test.err != "" {
				require.EqualError(t, err, test.err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, res)
				require.NotNil(t, res.key)
			}
		})
	}
}
