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

package depositdata

import (
	"context"
	"testing"

	spec "github.com/attestantio/go-eth2-client/spec/phase0"
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
	require.NoError(t, e2types.InitBLS())

	store := scratch.New()
	require.NoError(t, e2wallet.UseStore(store))
	testWallet, err := nd.CreateWallet(context.Background(), "Test", store, keystorev4.New())
	require.NoError(t, err)
	require.NoError(t, testWallet.(e2wtypes.WalletLocker).Unlock(context.Background(), nil))

	viper.Set("passphrase", "pass")
	interop0, err := testWallet.(e2wtypes.WalletAccountImporter).ImportAccount(context.Background(),
		"Interop 0",
		testutil.HexToBytes("0x25295f0d1d592a90b333e26e85149708208e9f8e8bc18f6c77bd62f8ad7a6866"),
		[]byte("pass"),
	)
	require.NoError(t, err)
	_, err = testWallet.(e2wtypes.WalletAccountImporter).ImportAccount(context.Background(),
		"Interop 1",
		testutil.HexToBytes("0x51d0b65185db6989ab0b560d6deed19c7ead0e24b9b6372cbecb1f26bdfad000"),
		[]byte("pass"),
	)
	require.NoError(t, err)

	var mainnetForkVersion *spec.Version
	{
		tmp := testutil.HexToVersion("0x00000000")
		mainnetForkVersion = &tmp
	}
	var mainnetDomain *spec.Domain
	{
		tmp := testutil.HexToDomain("0x03000000f5a5fd42d16a20302798ef6ed309979b43003d2320d9f0e8ea9831a9")
		mainnetDomain = &tmp
	}

	var forkVersion *spec.Version
	{
		tmp := testutil.HexToVersion("0x01020304")
		forkVersion = &tmp
	}
	var domain *spec.Domain
	{
		tmp := testutil.HexToDomain("0x03000000ffd2fc34e5796a643f749b0b2b908c4ca3ce58ce24a00c49329a2dc0")
		domain = &tmp
	}

	tests := []struct {
		name string
		vars map[string]interface{}
		res  *dataIn
		err  string
	}{
		{
			name: "Nil",
			err:  "validator account is required",
		},
		{
			name: "ValidatorAccountMissing",
			vars: map[string]interface{}{
				"withdrawalaccount": "Test/Interop 0",
				"depositvalue":      "32 Ether",
				"forkversion":       "0x01020304",
			},
			err: "validator account is required",
		},
		{
			name: "ValidatorAccountUnknown",
			vars: map[string]interface{}{
				"validatoraccount":  "Test/Unknown",
				"withdrawalaccount": "Test/Interop 0",
				"depositvalue":      "32 Ether",
				"forkversion":       "0x01020304",
			},
			err: "unknown validator account",
		},
		{
			name: "WithdrawalAccountMissing",
			vars: map[string]interface{}{
				"launchpad":        true,
				"validatoraccount": "Test/Interop 0",
				"depositvalue":     "32 Ether",
				"forkversion":      "0x01020304",
			},
			err: "withdrawalaccount or withdrawal public key is required",
		},
		{
			name: "WithdrawalAccountUnknown",
			vars: map[string]interface{}{
				"raw":               true,
				"validatoraccount":  "Test/Interop 0",
				"withdrawalaccount": "Test/Unknown",
				"depositvalue":      "32 Ether",
				"forkversion":       "0x01020304",
			},
			err: "failed to obtain withdrawal account: failed to obtain account: no account with name \"Unknown\"",
		},
		{
			name: "WithdrawalPubKeyInvalid",
			vars: map[string]interface{}{
				"validatoraccount": "Test/Interop 0",
				"withdrawalpubkey": "invalid",
				"depositvalue":     "32 Ether",
				"forkversion":      "0x01020304",
			},
			err: "failed to decode withdrawal public key: encoding/hex: invalid byte: U+0069 'i'",
		},
		{
			name: "WithdrawalPubKeyWrongLength",
			vars: map[string]interface{}{
				"validatoraccount": "Test/Interop 0",
				"withdrawalpubkey": "0xb89bebc699769726a318c8e9971bd3171297c61aea4a6578a7a4f94b547dcba5bac16a89108b6b6a1fe3695d1a874a0bff",
				"depositvalue":     "32 Ether",
				"forkversion":      "0x01020304",
			},
			err: "withdrawal public key must be exactly 48 bytes in length",
		},
		{
			name: "WithdrawalPubKeyNotPubKey",
			vars: map[string]interface{}{
				"validatoraccount": "Test/Interop 0",
				"withdrawalpubkey": "0x089bebc699769726a318c8e9971bd3171297c61aea4a6578a7a4f94b547dcba5bac16a89108b6b6a1fe3695d1a874a0b",
				"depositvalue":     "32 Ether",
				"forkversion":      "0x01020304",
			},
			err: "withdrawal public key is not valid: failed to deserialize public key: err blsPublicKeyDeserialize 089bebc699769726a318c8e9971bd3171297c61aea4a6578a7a4f94b547dcba5bac16a89108b6b6a1fe3695d1a874a0b",
		},
		{
			name: "DepositValueMissing",
			vars: map[string]interface{}{
				"validatoraccount":  "Test/Interop 0",
				"withdrawalaccount": "Test/Interop 0",
				"forkversion":       "0x01020304",
			},
			err: "deposit value is required",
		},
		{
			name: "DepositValueTooSmall",
			vars: map[string]interface{}{
				"validatoraccount":  "Test/Interop 0",
				"withdrawalaccount": "Test/Interop 0",
				"depositvalue":      "1000 Wei",
				"forkversion":       "0x01020304",
			},
			err: "deposit value must be at least 1 Ether",
		},
		{
			name: "DepositValueInvalid",
			vars: map[string]interface{}{
				"validatoraccount":  "Test/Interop 0",
				"withdrawalaccount": "Test/Interop 0",
				"depositvalue":      "1 groat",
				"forkversion":       "0x01020304",
			},
			err: "deposit value is invalid: failed to parse unit of 1 groat",
		},
		{
			name: "ForkVersionInvalid",
			vars: map[string]interface{}{
				"validatoraccount":  "Test/Interop 0",
				"withdrawalaccount": "Test/Interop 0",
				"depositvalue":      "32 Ether",
				"forkversion":       "invalid",
			},
			err: "failed to obtain fork version: failed to decode fork version: encoding/hex: invalid byte: U+0069 'i'",
		},
		{
			name: "Good",
			vars: map[string]interface{}{
				"validatoraccount":  "Test/Interop 0",
				"withdrawalaccount": "Test/Interop 0",
				"depositvalue":      "32 Ether",
			},
			res: &dataIn{
				format:                "json",
				withdrawalCredentials: testutil.HexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
				amount:                32000000000,
				validatorAccounts:     []e2wtypes.Account{interop0},
				forkVersion:           mainnetForkVersion,
				domain:                mainnetDomain,
			},
		},
		{
			name: "GoodForkVersionOverride",
			vars: map[string]interface{}{
				"validatoraccount":  "Test/Interop 0",
				"withdrawalaccount": "Test/Interop 0",
				"depositvalue":      "32 Ether",
				"forkversion":       "0x01020304",
			},
			res: &dataIn{
				format:                "json",
				withdrawalCredentials: testutil.HexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
				amount:                32000000000,
				validatorAccounts:     []e2wtypes.Account{interop0},
				forkVersion:           forkVersion,
				domain:                domain,
			},
		},
		{
			name: "GoodWithdrawalPubKey",
			vars: map[string]interface{}{
				"validatoraccount": "Test/Interop 0",
				"withdrawalpubkey": "0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c",
				"depositvalue":     "32 Ether",
				"forkversion":      "0x01020304",
			},
			res: &dataIn{
				format:                "json",
				withdrawalCredentials: testutil.HexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
				amount:                32000000000,
				validatorAccounts:     []e2wtypes.Account{interop0},
				forkVersion:           forkVersion,
				domain:                domain,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			viper.Reset()
			for k, v := range test.vars {
				viper.Set(k, v)
			}
			res, err := input()
			if test.err != "" {
				require.EqualError(t, err, test.err)
			} else {
				require.NoError(t, err)
				// Cannot compare accounts directly, so need to check each element individually.
				require.Equal(t, test.res.format, res.format)
				require.Equal(t, test.res.withdrawalCredentials, res.withdrawalCredentials)
				require.Equal(t, test.res.amount, res.amount)
				require.Equal(t, test.res.forkVersion, res.forkVersion)
				require.Equal(t, test.res.domain, res.domain)
				require.Equal(t, len(test.res.validatorAccounts), len(res.validatorAccounts))
				for i := range test.res.validatorAccounts {
					require.Equal(t, test.res.validatorAccounts[i].ID(), res.validatorAccounts[i].ID())
				}
			}
		})
	}
}
