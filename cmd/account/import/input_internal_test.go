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

package accountimport

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
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

func TestInput(t *testing.T) {
	require.NoError(t, e2types.InitBLS())

	store := scratch.New()
	require.NoError(t, e2wallet.UseStore(store))
	testWallet, err := nd.CreateWallet(context.Background(), "Test wallet", store, keystorev4.New())
	require.NoError(t, err)
	require.NoError(t, testWallet.(e2wtypes.WalletLocker).Unlock(context.Background(), nil))

	tests := []struct {
		name string
		vars map[string]interface{}
		res  *dataIn
		err  string
	}{
		{
			name: "TimeoutMissing",
			vars: map[string]interface{}{
				"account":    "Test wallet/Test account",
				"passphrase": "ce%NohGhah4ye5ra",
			},
			err: "timeout is required",
		},
		{
			name: "WalletUnknown",
			vars: map[string]interface{}{
				"timeout":    "5s",
				"account":    "Unknown/Test account",
				"passphrase": "ce%NohGhah4ye5ra",
			},
			err: "failed to obtain wallet: wallet not found",
		},
		{
			name: "AccountMissing",
			vars: map[string]interface{}{
				"timeout":    "5s",
				"passphrase": "ce%NohGhah4ye5ra",
			},
			err: "account is required",
		},
		{
			name: "AccountWalletOnly",
			vars: map[string]interface{}{
				"timeout":    "5s",
				"passphrase": "ce%NohGhah4ye5ra",
				"account":    "Test wallet/",
			},
			err: "account name is required",
		},
		{
			name: "AccountMalformed",
			vars: map[string]interface{}{
				"timeout":    "5s",
				"account":    "//",
				"passphrase": "ce%NohGhah4ye5ra",
			},
			err: "failed to obtain account name: invalid account format",
		},
		{
			name: "MultiplePassphrases",
			vars: map[string]interface{}{
				"timeout":    "5s",
				"account":    "Test wallet/Test account",
				"passphrase": []string{"ce%NohGhah4ye5ra", "other"},
			},
			err: "failed to obtain passphrase: multiple passphrases supplied",
		},
		{
			name: "KeyMissing",
			vars: map[string]interface{}{
				"timeout":    "5s",
				"account":    "Test wallet/Test account",
				"passphrase": "ce%NohGhah4ye5ra",
			},
			err: "key or keystore is required",
		},
		{
			name: "KeyMalformed",
			vars: map[string]interface{}{
				"timeout":    "5s",
				"account":    "Test wallet/Test account",
				"passphrase": "ce%NohGhah4ye5ra",
				"key":        "invalid",
			},
			err: "key is malformed: encoding/hex: invalid byte: U+0069 'i'",
		},
		{
			name: "KeyandKeystore",
			vars: map[string]interface{}{
				"timeout":    "5s",
				"account":    "Test wallet/Test account",
				"passphrase": "ce%NohGhah4ye5ra",
				"key":        "0x25295f0d1d592a90b333e26e85149708208e9f8e8bc18f6c77bd62f8ad7a6866",
				"keystore":   "{}",
			},
			err: "only one of key and keystore is required",
		},
		{
			name: "KeystoreNoKeystorePassphrase",
			vars: map[string]interface{}{
				"timeout":  "5s",
				"account":  "Test wallet/Test account",
				"keystore": "{}",
			},
			err: "must supply keystore passphrase with keystore-passphrase when supplying keystore",
		},
		{
			name: "Good",
			vars: map[string]interface{}{
				"timeout":    "5s",
				"account":    "Test wallet/Test account",
				"passphrase": "ce%NohGhah4ye5ra",
				"key":        "0x25295f0d1d592a90b333e26e85149708208e9f8e8bc18f6c77bd62f8ad7a6866",
			},
			res: &dataIn{
				timeout:     5 * time.Second,
				accountName: "Test account",
				passphrase:  "ce%NohGhah4ye5ra",
				key:         hexToBytes("0x25295f0d1d592a90b333e26e85149708208e9f8e8bc18f6c77bd62f8ad7a6866"),
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
				// Cannot compare accounts directly, so need to check each element individually.
				require.Equal(t, test.res.timeout, res.timeout)
				require.Equal(t, test.res.accountName, res.accountName)
				require.Equal(t, test.res.passphrase, res.passphrase)
			}
		})
	}
}
