// Copyright © 2019-2021 Weald Technology Limited.
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
	"encoding/hex"
	"strings"
	"testing"

	spec "github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"github.com/wealdtech/ethdo/testutil"
	e2types "github.com/wealdtech/go-eth2-types/v2"
	keystorev4 "github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"
	nd "github.com/wealdtech/go-eth2-wallet-nd/v2"
	scratch "github.com/wealdtech/go-eth2-wallet-store-scratch"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

func TestProcess(t *testing.T) {
	require.NoError(t, e2types.InitBLS())

	testWallet, err := nd.CreateWallet(context.Background(), "Test", scratch.New(), keystorev4.New())
	require.NoError(t, err)
	require.NoError(t, testWallet.(e2wtypes.WalletLocker).Unlock(context.Background(), nil))

	viper.Set("passphrase", "pass")
	interop0, err := testWallet.(e2wtypes.WalletAccountImporter).ImportAccount(context.Background(),
		"Interop 0",
		testutil.HexToBytes("0x25295f0d1d592a90b333e26e85149708208e9f8e8bc18f6c77bd62f8ad7a6866"),
		[]byte("pass"),
	)
	require.NoError(t, err)
	interop1, err := testWallet.(e2wtypes.WalletAccountImporter).ImportAccount(context.Background(),
		"Interop 1",
		testutil.HexToBytes("0x51d0b65185db6989ab0b560d6deed19c7ead0e24b9b6372cbecb1f26bdfad000"),
		[]byte("pass"),
	)
	require.NoError(t, err)

	withdrawalAccount := "Test/Interop 0"
	withdrawalPubKey := "0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c"
	withdrawalAddress := "0x30C99930617B7b793beaB603ecEB08691005f2E5"

	var validatorPubKey *spec.BLSPubKey
	{
		tmp := testutil.HexToPubKey("0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c")
		validatorPubKey = &tmp
	}
	var signature *spec.BLSSignature
	{
		tmp := testutil.HexToSignature("0xb7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2")
		signature = &tmp
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
	var depositDataRoot *spec.Root
	{
		tmp := testutil.HexToRoot("0x9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554")
		depositDataRoot = &tmp
	}
	var depositMessageRoot *spec.Root
	{
		tmp := testutil.HexToRoot("0x139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6")
		depositMessageRoot = &tmp
	}

	var validatorPubKey2 *spec.BLSPubKey
	{
		tmp := testutil.HexToPubKey("0xb89bebc699769726a318c8e9971bd3171297c61aea4a6578a7a4f94b547dcba5bac16a89108b6b6a1fe3695d1a874a0b")
		validatorPubKey2 = &tmp
	}
	var signature2 *spec.BLSSignature
	{
		tmp := testutil.HexToSignature("0x939aedb76236c971c21227189c6a3a40d07909d19999798490294d284130a913b6f91d41d875768fb3e2ea4dcec672a316e5951272378f5df80a7c34fadb9a4d8462ee817faf50fe8b1c33e72d884fb17e71e665724f9e17bdf11f48eb6e9bfd")
		signature2 = &tmp
	}
	var depositDataRoot2 *spec.Root
	{
		tmp := testutil.HexToRoot("0x182c7708aad7027bea2f6251eddf62431fae4876ee3e55339082219ae7014443")
		depositDataRoot2 = &tmp
	}
	var depositMessageRoot2 *spec.Root
	{
		tmp := testutil.HexToRoot("0x1dc5053486d74f5c91fa90e1e86d718d3fb42bb92e5cfdce98e994eb2bff2c46")
		depositMessageRoot2 = &tmp
	}

	var depositDataRoot3 *spec.Root
	{
		tmp := testutil.HexToRoot("0x489500535b03dd9deffa0f00cb38d82346111856fb58a9541fe1f01a1a97429c")
		depositDataRoot3 = &tmp
	}
	var depositMessageRoot3 *spec.Root
	{
		tmp := testutil.HexToRoot("0x7b8ee5694e4338cf2bfe5a4d2f46540f0ade85ebd30713673cf5783c4e925681")
		depositMessageRoot3 = &tmp
	}
	var signature3 *spec.BLSSignature
	{
		tmp := testutil.HexToSignature("0xba0019d5c421f205d845782f52a87ab95cd489fbef2911f8a1f9cf7c14b4ce59eefa82641e770a4cb405534b7776d0f801b0a8b178c1b71b718c104e89f4e633da10a398c7919a00c403d58f3f4b827af8adb263b192e7a45b0ed1926dff5f66")
		signature3 = &tmp
	}

	tests := []struct {
		name   string
		dataIn *dataIn
		res    []*dataOut
		err    string
	}{
		{
			name: "Nil",
			err:  "no data",
		},
		{
			name: "WithdrawalDetailsMissing",
			dataIn: &dataIn{
				format:            "raw",
				passphrases:       []string{"pass"},
				amount:            32000000000,
				validatorAccounts: []e2wtypes.Account{interop0},
				forkVersion:       forkVersion,
				domain:            domain,
			},
			err: "withdrawal account, public key or address is required",
		},
		{
			name: "WithdrawalAccountUnknown",
			dataIn: &dataIn{
				format:            "raw",
				passphrases:       []string{"pass"},
				withdrawalAccount: "Unknown",
				amount:            32000000000,
				validatorAccounts: []e2wtypes.Account{interop0},
				forkVersion:       forkVersion,
				domain:            domain,
			},
			err: "failed to obtain withdrawal account: failed to open wallet for account: wallet not found",
		},
		{
			name: "WithdrawalPubKeyInvalid",
			dataIn: &dataIn{
				format:            "raw",
				passphrases:       []string{"pass"},
				withdrawalPubKey:  "invalid",
				amount:            32000000000,
				validatorAccounts: []e2wtypes.Account{interop0},
				forkVersion:       forkVersion,
				domain:            domain,
			},
			err: "failed to decode withdrawal public key: encoding/hex: invalid byte: U+0069 'i'",
		},
		{
			name: "WithdrawalPubKeyWrongLength",
			dataIn: &dataIn{
				format:            "raw",
				passphrases:       []string{"pass"},
				withdrawalPubKey:  "0xb89bebc699769726a318c8e9971bd3171297c61aea4a6578a7a4f94b547dcba5bac16a89108b6b6a1fe3695d1a874a0bff",
				amount:            32000000000,
				validatorAccounts: []e2wtypes.Account{interop0},
				forkVersion:       forkVersion,
				domain:            domain,
			},
			err: "withdrawal public key must be exactly 48 bytes in length",
		},
		{
			name: "WithdrawalPubKeyNotPubKey",
			dataIn: &dataIn{
				format:            "raw",
				passphrases:       []string{"pass"},
				withdrawalPubKey:  "0x089bebc699769726a318c8e9971bd3171297c61aea4a6578a7a4f94b547dcba5bac16a89108b6b6a1fe3695d1a874a0b",
				amount:            32000000000,
				validatorAccounts: []e2wtypes.Account{interop0},
				forkVersion:       forkVersion,
				domain:            domain,
			},
			err: "withdrawal public key is not valid: failed to deserialize public key: err blsPublicKeyDeserialize 089bebc699769726a318c8e9971bd3171297c61aea4a6578a7a4f94b547dcba5bac16a89108b6b6a1fe3695d1a874a0b",
		},
		{
			name: "WithdrawalAddressInvalid",
			dataIn: &dataIn{
				format:            "raw",
				passphrases:       []string{"pass"},
				withdrawalAddress: "invalid",
				amount:            32000000000,
				validatorAccounts: []e2wtypes.Account{interop0},
				forkVersion:       forkVersion,
				domain:            domain,
			},
			err: "failed to decode withdrawal address: encoding/hex: invalid byte: U+0069 'i'",
		},
		{
			name: "WithdrawalAddressWrongLength",
			dataIn: &dataIn{
				format:            "raw",
				passphrases:       []string{"pass"},
				withdrawalAddress: "0x30C99930617B7b793beaB603ecEB08691005f2",
				amount:            32000000000,
				validatorAccounts: []e2wtypes.Account{interop0},
				forkVersion:       forkVersion,
				domain:            domain,
			},
			err: "withdrawal address must be exactly 20 bytes in length",
		},
		{
			name: "WithdrawalAddressIncorrectChecksum",
			dataIn: &dataIn{
				format:            "raw",
				passphrases:       []string{"pass"},
				withdrawalAddress: "0x30c99930617b7b793beab603eceb08691005f2e5",
				amount:            32000000000,
				validatorAccounts: []e2wtypes.Account{interop0},
				forkVersion:       forkVersion,
				domain:            domain,
			},
			err: "withdrawal address checksum does not match (expected 0x30C99930617B7b793beaB603ecEB08691005f2E5)",
		},
		{
			name: "Single",
			dataIn: &dataIn{
				format:            "raw",
				passphrases:       []string{"pass"},
				withdrawalAccount: withdrawalAccount,
				amount:            32000000000,
				validatorAccounts: []e2wtypes.Account{interop0},
				forkVersion:       forkVersion,
				domain:            domain,
			},
			res: []*dataOut{
				{
					format:                "raw",
					account:               "Test/Interop 0",
					validatorPubKey:       validatorPubKey,
					amount:                32000000000,
					withdrawalCredentials: testutil.HexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
					signature:             signature,
					forkVersion:           forkVersion,
					depositDataRoot:       depositDataRoot,
					depositMessageRoot:    depositMessageRoot,
				},
			},
		},
		{
			name: "Double",
			dataIn: &dataIn{
				format:            "raw",
				passphrases:       []string{"pass"},
				withdrawalPubKey:  withdrawalPubKey,
				amount:            32000000000,
				validatorAccounts: []e2wtypes.Account{interop0, interop1},
				forkVersion:       forkVersion,
				domain:            domain,
			},
			res: []*dataOut{
				{
					format:                "raw",
					account:               "Test/Interop 0",
					validatorPubKey:       validatorPubKey,
					amount:                32000000000,
					withdrawalCredentials: testutil.HexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
					signature:             signature,
					forkVersion:           forkVersion,
					depositDataRoot:       depositDataRoot,
					depositMessageRoot:    depositMessageRoot,
				},
				{
					format:                "raw",
					account:               "Test/Interop 1",
					validatorPubKey:       validatorPubKey2,
					amount:                32000000000,
					withdrawalCredentials: testutil.HexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
					signature:             signature2,
					forkVersion:           forkVersion,
					depositDataRoot:       depositDataRoot2,
					depositMessageRoot:    depositMessageRoot2,
				},
			},
		},
		{
			name: "WithdrawalAddress",
			dataIn: &dataIn{
				format:            "raw",
				passphrases:       []string{"pass"},
				withdrawalAddress: withdrawalAddress,
				amount:            32000000000,
				validatorAccounts: []e2wtypes.Account{interop0},
				forkVersion:       forkVersion,
				domain:            domain,
			},
			res: []*dataOut{
				{
					format:                "raw",
					account:               "Test/Interop 0",
					validatorPubKey:       validatorPubKey,
					amount:                32000000000,
					withdrawalCredentials: testutil.HexToBytes("0x01000000000000000000000030C99930617B7b793beaB603ecEB08691005f2E5"),
					signature:             signature3,
					forkVersion:           forkVersion,
					depositDataRoot:       depositDataRoot3,
					depositMessageRoot:    depositMessageRoot3,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := process(test.dataIn)
			if test.err != "" {
				require.EqualError(t, err, test.err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.res, res)
			}
		})
	}
}

func TestAddressBytesToEIP55(t *testing.T) {
	tests := []string{
		"0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed",
		"0xfB6916095ca1df60bB79Ce92cE3Ea74c37c5d359",
		"0xdbF03B407c01E7cD3CBea99509d93f8DDDC8C6FB",
		"0xD1220A0cf47c7B9Be7A2E6BA89F429762e7b9aDb",
	}

	for _, test := range tests {
		bytes, err := hex.DecodeString(strings.TrimPrefix(test, "0x"))
		require.NoError(t, err)
		require.Equal(t, addressBytesToEIP55(bytes), test)
	}
}
