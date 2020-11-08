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

package walletcreate

import (
	"context"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	e2types "github.com/wealdtech/go-eth2-types/v2"
	e2wallet "github.com/wealdtech/go-eth2-wallet"
	scratch "github.com/wealdtech/go-eth2-wallet-store-scratch"
)

func TestInput(t *testing.T) {
	require.NoError(t, e2types.InitBLS())

	store := scratch.New()
	require.NoError(t, e2wallet.UseStore(store))

	tests := []struct {
		name string
		vars map[string]interface{}
		res  *dataIn
		err  string
	}{
		{
			name: "TimeoutMissing",
			vars: map[string]interface{}{
				"account":           "Test wallet",
				"wallet-passphrase": "ce%NohGhah4ye5ra",
				"type":              "nd",
			},
			err: "timeout is required",
		},
		{
			name: "StoreMissing",
			vars: map[string]interface{}{
				"timeout":           "5s",
				"wallet":            "Test wallet",
				"type":              "nd",
				"wallet-passphrase": "ce%NohGhah4ye5ra",
			},
			err: "store is required",
		},
		{
			name: "WalletMissing",
			vars: map[string]interface{}{
				"timeout":           "5s",
				"store":             store,
				"type":              "nd",
				"wallet-passphrase": "ce%NohGhah4ye5ra",
			},
			err: "wallet is required",
		},
		{
			name: "WalletInvalid",
			vars: map[string]interface{}{
				"timeout":           "5s",
				"store":             store,
				"wallet":            "/",
				"type":              "nd",
				"wallet-passphrase": "ce%NohGhah4ye5ra",
			},
			err: "failed to obtain wallet name: invalid account format",
		},
		{
			name: "TypeMissing",
			vars: map[string]interface{}{
				"timeout":           "5s",
				"store":             store,
				"wallet":            "Test wallet",
				"wallet-passphrase": "ce%NohGhah4ye5ra",
			},
			err: "wallet type is required",
		},
		{
			name: "Good",
			vars: map[string]interface{}{
				"timeout":           "5s",
				"store":             store,
				"wallet":            "Test wallet",
				"type":              "nd",
				"wallet-passphrase": "ce%NohGhah4ye5ra",
			},
			res: &dataIn{
				timeout:    5 * time.Second,
				store:      store,
				walletName: "Test account",
				walletType: "nd",
				passphrase: "ce%NohGhah4ye5ra",
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
			}
		})
	}
}
