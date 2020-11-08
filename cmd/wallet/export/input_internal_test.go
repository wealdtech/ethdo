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

package walletexport

import (
	"context"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	e2types "github.com/wealdtech/go-eth2-types/v2"
	e2wallet "github.com/wealdtech/go-eth2-wallet"
	keystorev4 "github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"
	nd "github.com/wealdtech/go-eth2-wallet-nd/v2"
	scratch "github.com/wealdtech/go-eth2-wallet-store-scratch"
)

func TestInput(t *testing.T) {
	require.NoError(t, e2types.InitBLS())

	store := scratch.New()
	require.NoError(t, e2wallet.UseStore(store))
	wallet, err := nd.CreateWallet(context.Background(), "Test wallet", store, keystorev4.New())
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
				"wallet": "Test wallet",
			},
			err: "timeout is required",
		},
		{
			name: "WalletMissing",
			vars: map[string]interface{}{
				"timeout": "5s",
			},
			err: "failed to access wallet: cannot determine wallet",
		},
		{
			name: "WalletUnknown",
			vars: map[string]interface{}{
				"timeout": "5s",
				"wallet":  "unknown",
			},
			err: "failed to access wallet: wallet not found",
		},
		{
			name: "Remote",
			vars: map[string]interface{}{
				"timeout": "5s",
				"remote":  "remoteaddress",
			},
			err: "wallet export not available for remote wallets",
		},
		{
			name: "PassphraseMissing",
			vars: map[string]interface{}{
				"timeout": "5s",
				"wallet":  "Test wallet",
			},
			err: "failed to obtain export passphrase: passphrase is required",
		},
		{
			name: "Good",
			vars: map[string]interface{}{
				"timeout":    "5s",
				"wallet":     "Test wallet",
				"passphrase": "export",
			},
			res: &dataIn{
				timeout: 5 * time.Second,
				wallet:  wallet,
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
				require.Equal(t, test.res.timeout, res.timeout)
				require.Equal(t, test.vars["wallet"], res.wallet.Name())
			}
		})
	}
}
