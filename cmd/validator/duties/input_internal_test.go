// Copyright © 2020 Weald Technology Trading
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

package validatorduties

import (
	"context"
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"github.com/wealdtech/ethdo/testutil"
	e2types "github.com/wealdtech/go-eth2-types/v2"
	e2wallet "github.com/wealdtech/go-eth2-wallet"
	keystorev4 "github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"
	nd "github.com/wealdtech/go-eth2-wallet-nd/v2"
	scratch "github.com/wealdtech/go-eth2-wallet-store-scratch"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

func TestInput(t *testing.T) {
	if os.Getenv("ETHDO_TEST_CONNECTION") == "" {
		t.Skip("ETHDO_TEST_CONNECTION not configured; cannot run tests")
	}

	require.NoError(t, e2types.InitBLS())

	store := scratch.New()
	require.NoError(t, e2wallet.UseStore(store))
	testWallet, err := nd.CreateWallet(context.Background(), "Test wallet", store, keystorev4.New())
	require.NoError(t, err)
	require.NoError(t, testWallet.(e2wtypes.WalletLocker).Unlock(context.Background(), nil))
	viper.Set("passphrase", "pass")
	_, err = testWallet.(e2wtypes.WalletAccountImporter).ImportAccount(context.Background(),
		"Interop 0",
		testutil.HexToBytes("0x25295f0d1d592a90b333e26e85149708208e9f8e8bc18f6c77bd62f8ad7a6866"),
		[]byte("pass"),
	)
	require.NoError(t, err)

	tests := []struct {
		name string
		vars map[string]interface{}
		res  *dataIn
		err  string
	}{
		{
			name: "TimeoutMissing",
			vars: map[string]interface{}{
				"connection": "http://locahost:4000",
				"pubkey":     "0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c",
			},
			err: "timeout is required",
		},
		{
			name: "AccountMissing",
			vars: map[string]interface{}{
				"timeout":    "5s",
				"connection": "http://locahost:4000",
			},
			err: "account, pubkey or index required",
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
				require.Equal(t, test.res.timeout, res.timeout)
			}
		})
	}
}
