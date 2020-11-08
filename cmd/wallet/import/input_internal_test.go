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
	"fmt"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	e2types "github.com/wealdtech/go-eth2-types/v2"
	e2wallet "github.com/wealdtech/go-eth2-wallet"
	keystorev4 "github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"
	nd "github.com/wealdtech/go-eth2-wallet-nd/v2"
	scratch "github.com/wealdtech/go-eth2-wallet-store-scratch"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

func TestInput(t *testing.T) {
	require.NoError(t, e2types.InitBLS())

	store := scratch.New()
	require.NoError(t, e2wallet.UseStore(store))
	wallet, err := nd.CreateWallet(context.Background(), "Test wallet", store, keystorev4.New())
	require.NoError(t, err)
	data, err := wallet.(e2wtypes.WalletExporter).Export(context.Background(), []byte("ce%NohGhah4ye5ra"))
	require.NoError(t, err)
	require.NoError(t, e2wallet.UseStore(scratch.New()))

	tests := []struct {
		name string
		vars map[string]interface{}
		res  *dataIn
		err  string
	}{
		{
			name: "TimeoutMissing",
			vars: map[string]interface{}{
				"data": fmt.Sprintf("%#x", data),
			},
			err: "timeout is required",
		},
		{
			name: "DataMissing",
			vars: map[string]interface{}{
				"timeout": "5s",
			},
			err: "data is required",
		},
		{
			name: "DataInvalid",
			vars: map[string]interface{}{
				"timeout": "5s",
				"data":    "0xinvalid",
			},
			err: "data is invalid: encoding/hex: invalid byte: U+0069 'i'",
		},
		{
			name: "DataFileMissing",
			vars: map[string]interface{}{
				"timeout": "5s",
				"data":    "missing",
			},
			err: "failed to read wallet import data: open missing: no such file or directory",
		},
		{
			name: "Remote",
			vars: map[string]interface{}{
				"timeout":    "5s",
				"remote":     "remoteaddress",
				"data":       fmt.Sprintf("%#x", data),
				"passphrase": "export",
			},
			err: "wallet import not available for remote wallets",
		},
		{
			name: "PassphraseMissing",
			vars: map[string]interface{}{
				"timeout": "5s",
				"data":    fmt.Sprintf("%#x", data),
			},
			err: "failed to obtain import passphrase: passphrase is required",
		},
		{
			name: "Good",
			vars: map[string]interface{}{
				"timeout":    "5s",
				"data":       fmt.Sprintf("%#x", data),
				"passphrase": "export",
			},
			res: &dataIn{
				timeout: 5 * time.Second,
			},
		},
		{
			name: "Verify",
			vars: map[string]interface{}{
				"timeout":    "5s",
				"data":       fmt.Sprintf("%#x", data),
				"passphrase": "export",
				"verify":     true,
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
				require.NotNil(t, res)
			}
		})
	}
}
