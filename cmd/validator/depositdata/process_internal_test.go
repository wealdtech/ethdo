// Copyright © 2019, 2020 eald Technology Trading
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
			name: "Single",
			dataIn: &dataIn{
				format:                "raw",
				passphrases:           []string{"pass"},
				withdrawalCredentials: testutil.HexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
				amount:                32000000000,
				validatorAccounts:     []e2wtypes.Account{interop0},
				forkVersion:           forkVersion,
				domain:                domain,
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
				format:                "raw",
				passphrases:           []string{"pass"},
				withdrawalCredentials: testutil.HexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
				amount:                32000000000,
				validatorAccounts:     []e2wtypes.Account{interop0, interop1},
				forkVersion:           forkVersion,
				domain:                domain,
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
