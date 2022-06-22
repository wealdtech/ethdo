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

package validatorexit

import (
	"context"
	"os"
	"testing"
	"time"

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
				"account":           "Test wallet",
				"wallet-passphrase": "ce%NohGhah4ye5ra",
				"type":              "nd",
			},
			err: "timeout is required",
		},
		{
			name: "NoMethod",
			vars: map[string]interface{}{
				"timeout": "5s",
			},
			err: "must supply account, key, or pre-constructed JSON",
		},
		{
			name: "KeyInvalid",
			vars: map[string]interface{}{
				"timeout": "5s",
				"key":     "0xinvalid",
			},
			err: "failed to decode key: encoding/hex: invalid byte: U+0069 'i'",
		},
		{
			name: "KeyBad",
			vars: map[string]interface{}{
				"timeout": "5s",
				"key":     "0x00",
			},
			err: "failed to create acount from key: private key must be 32 bytes",
		},
		{
			name: "KeyGood",
			vars: map[string]interface{}{
				"connection":                 os.Getenv("ETHDO_TEST_CONNECTION"),
				"allow-insecure-connections": true,
				"timeout":                    "5s",
				"key":                        "0x25295f0d1d592a90b333e26e85149708208e9f8e8bc18f6c77bd62f8ad7a6866",
			},
			res: &dataIn{
				timeout: 5 * time.Second,
			},
		},
		{
			name: "AccountUnknown",
			vars: map[string]interface{}{
				"connection":                 os.Getenv("ETHDO_TEST_CONNECTION"),
				"allow-insecure-connections": true,
				"timeout":                    "5s",
				"account":                    "Test wallet/unknown",
			},
			res: &dataIn{
				timeout: 5 * time.Second,
			},
			err: "failed to obtain acount: failed to obtain account: no account with name \"unknown\"",
		},
		{
			name: "AccountGood",
			vars: map[string]interface{}{
				"connection":                 os.Getenv("ETHDO_TEST_CONNECTION"),
				"allow-insecure-connections": true,
				"timeout":                    "5s",
				"account":                    "Test wallet/Interop 0",
			},
			res: &dataIn{
				timeout: 5 * time.Second,
			},
		},
		{
			name: "JSONInvalid",
			vars: map[string]interface{}{
				"connection":                 os.Getenv("ETHDO_TEST_CONNECTION"),
				"allow-insecure-connections": true,
				"timeout":                    "5s",
				"exit":                       `invalid`,
			},
			res: &dataIn{
				timeout: 5 * time.Second,
			},
			err: "invalid character 'i' looking for beginning of value",
		},
		{
			name: "JSONGood",
			vars: map[string]interface{}{
				"connection":                 os.Getenv("ETHDO_TEST_CONNECTION"),
				"allow-insecure-connections": true,
				"timeout":                    "5s",
				"exit":                       `{"exit":{"message":{"epoch":"123","validator_index":"456"},"signature":"0x000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f202122232425262728292a2b2c2d2e2f303132333435363738393a3b3c3d3e3f404142434445464748494a4b4c4d4e4f505152535455565758595a5b5c5d5e5f"},"fork_version":"0x00002009"}`,
			},
			res: &dataIn{
				timeout: 5 * time.Second,
			},
		},
		{
			name: "ClientBad",
			vars: map[string]interface{}{
				"connection":                 "localhost:1",
				"allow-insecure-connections": true,
				"timeout":                    "5s",
				"key":                        "0x25295f0d1d592a90b333e26e85149708208e9f8e8bc18f6c77bd62f8ad7a6866",
			},
			err: "failed to connect to Ethereum 2 beacon node: failed to connect to beacon node: failed to confirm node connection: failed to fetch genesis: failed to request genesis: failed to call GET endpoint: Get \"http://localhost:1/eth/v1/beacon/genesis\": dial tcp 127.0.0.1:1: connect: connection refused",
		},
		{
			name: "EpochProvided",
			vars: map[string]interface{}{
				"connection":                 os.Getenv("ETHDO_TEST_CONNECTION"),
				"allow-insecure-connections": true,
				"timeout":                    "5s",
				"key":                        "0x25295f0d1d592a90b333e26e85149708208e9f8e8bc18f6c77bd62f8ad7a6866",
				"epoch":                      "123",
			},
			res: &dataIn{
				timeout: 5 * time.Second,
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
