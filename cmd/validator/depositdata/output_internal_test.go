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
	"testing"

	spec "github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/stretchr/testify/require"
	"github.com/wealdtech/ethdo/testutil"
)

func TestOutputJSON(t *testing.T) {
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
		tmp := testutil.HexToSignature("0x911fe0766e8b79d711dde46bc2142eb51e35be99e5f7da505af9eaad85707bbb8013f0dea35e30403b3e57bb13054c1d0d389aceeba1d4160a148026212c7e017044e3ea69cd96fbd23b6aa9fd1e6f7e82494fbd5f8fc75856711a6b8998926e")
		signature2 = &tmp
	}
	var depositDataRoot2 *spec.Root
	{
		tmp := testutil.HexToRoot("0x3b51670e9f266d44c879682a230d60f0d534c64ab25ee68700fe3adb17ddfcab")
		depositDataRoot2 = &tmp
	}
	var depositMessageRoot2 *spec.Root
	{
		tmp := testutil.HexToRoot("0xbb4b6184b25873cdf430df3838c8d3e3d16cf3dc3b214e2f3ab7df9e6d5a9b52")
		depositMessageRoot2 = &tmp
	}

	tests := []struct {
		name    string
		dataOut []*dataOut
		res     string
		err     string
	}{
		{
			name: "Nil",
			res:  "[]",
		},
		{
			name: "NilDatum",
			dataOut: []*dataOut{
				nil,
			},
			res: "[]",
		},
		{
			name: "AccountMissing",
			dataOut: []*dataOut{
				{
					format:                "json",
					validatorPubKey:       validatorPubKey,
					withdrawalCredentials: testutil.HexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
					amount:                32000000000,
					signature:             signature,
					forkVersion:           forkVersion,
					depositDataRoot:       depositDataRoot,
					depositMessageRoot:    depositMessageRoot,
				},
			},
			err: "missing account",
		},
		{
			name: "MissingValidatorPubKey",
			dataOut: []*dataOut{
				{
					format:                "json",
					account:               "interop/00000",
					withdrawalCredentials: testutil.HexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
					amount:                32000000000,
					signature:             signature,
					forkVersion:           forkVersion,
					depositDataRoot:       depositDataRoot,
					depositMessageRoot:    depositMessageRoot,
				},
			},
			err: "validator public key required",
		},
		{
			name: "MissingWithdrawalCredentials",
			dataOut: []*dataOut{
				{
					format:             "json",
					account:            "interop/00000",
					validatorPubKey:    validatorPubKey,
					amount:             32000000000,
					signature:          signature,
					forkVersion:        forkVersion,
					depositDataRoot:    depositDataRoot,
					depositMessageRoot: depositMessageRoot,
				},
			},
			err: "withdrawal credentials must be 32 bytes",
		},
		{
			name: "SignatureMissing",
			dataOut: []*dataOut{
				{
					format:                "json",
					account:               "interop/00000",
					validatorPubKey:       validatorPubKey,
					withdrawalCredentials: testutil.HexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
					amount:                32000000000,
					forkVersion:           forkVersion,
					depositDataRoot:       depositDataRoot,
					depositMessageRoot:    depositMessageRoot,
				},
			},
			err: "signature required",
		},
		{
			name: "AmountMissing",
			dataOut: []*dataOut{
				{
					format:                "json",
					account:               "interop/00000",
					validatorPubKey:       validatorPubKey,
					withdrawalCredentials: testutil.HexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
					signature:             signature,
					forkVersion:           forkVersion,
					depositDataRoot:       depositDataRoot,
					depositMessageRoot:    depositMessageRoot,
				},
			},
			err: "missing amount",
		},
		{
			name: "DepositDataRootMissing",
			dataOut: []*dataOut{
				{
					format:                "json",
					account:               "interop/00000",
					validatorPubKey:       validatorPubKey,
					withdrawalCredentials: testutil.HexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
					amount:                32000000000,
					signature:             signature,
					forkVersion:           forkVersion,
					depositMessageRoot:    depositMessageRoot,
				},
			},
			err: "deposit data root required",
		},
		{
			name: "Single",
			dataOut: []*dataOut{
				{
					format:                "json",
					account:               "interop/00000",
					validatorPubKey:       validatorPubKey,
					withdrawalCredentials: testutil.HexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
					amount:                32000000000,
					signature:             signature,
					forkVersion:           forkVersion,
					depositDataRoot:       depositDataRoot,
					depositMessageRoot:    depositMessageRoot,
				},
			},
			res: `[{"name":"Deposit for interop/00000","account":"interop/00000","pubkey":"0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","withdrawal_credentials":"0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b","signature":"0xb7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2","amount":32000000000,"deposit_data_root":"0x9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554","deposit_message_root":"0x139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6","fork_version":"0x01020304","version":3}]`,
		},
		{
			name: "Double",
			dataOut: []*dataOut{
				{
					format:                "json",
					account:               "interop/00000",
					validatorPubKey:       validatorPubKey,
					withdrawalCredentials: testutil.HexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
					amount:                32000000000,
					signature:             signature,
					forkVersion:           forkVersion,
					depositDataRoot:       depositDataRoot,
					depositMessageRoot:    depositMessageRoot,
				},
				{
					format:                "json",
					account:               "interop/00001",
					validatorPubKey:       validatorPubKey2,
					withdrawalCredentials: testutil.HexToBytes("0x00ec7ef7780c9d151597924036262dd28dc60e1228f4da6fecf9d402cb3f3594"),
					amount:                32000000000,
					signature:             signature2,
					forkVersion:           forkVersion,
					depositDataRoot:       depositDataRoot2,
					depositMessageRoot:    depositMessageRoot2,
				},
			},
			res: `[{"name":"Deposit for interop/00000","account":"interop/00000","pubkey":"0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","withdrawal_credentials":"0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b","signature":"0xb7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2","amount":32000000000,"deposit_data_root":"0x9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554","deposit_message_root":"0x139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6","fork_version":"0x01020304","version":3},{"name":"Deposit for interop/00001","account":"interop/00001","pubkey":"0xb89bebc699769726a318c8e9971bd3171297c61aea4a6578a7a4f94b547dcba5bac16a89108b6b6a1fe3695d1a874a0b","withdrawal_credentials":"0x00ec7ef7780c9d151597924036262dd28dc60e1228f4da6fecf9d402cb3f3594","signature":"0x911fe0766e8b79d711dde46bc2142eb51e35be99e5f7da505af9eaad85707bbb8013f0dea35e30403b3e57bb13054c1d0d389aceeba1d4160a148026212c7e017044e3ea69cd96fbd23b6aa9fd1e6f7e82494fbd5f8fc75856711a6b8998926e","amount":32000000000,"deposit_data_root":"0x3b51670e9f266d44c879682a230d60f0d534c64ab25ee68700fe3adb17ddfcab","deposit_message_root":"0xbb4b6184b25873cdf430df3838c8d3e3d16cf3dc3b214e2f3ab7df9e6d5a9b52","fork_version":"0x01020304","version":3}]`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := output(test.dataOut)
			if test.err != "" {
				require.EqualError(t, err, test.err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.res, res)
			}
		})
	}
}

func TestOutputLaunchpad(t *testing.T) {
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
	var forkVersionPyrmont *spec.Version
	{
		tmp := testutil.HexToVersion("0x00002009")
		forkVersionPyrmont = &tmp
	}
	var forkVersionPrater *spec.Version
	{
		tmp := testutil.HexToVersion("0x00001020")
		forkVersionPrater = &tmp
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
		tmp := testutil.HexToSignature("0x911fe0766e8b79d711dde46bc2142eb51e35be99e5f7da505af9eaad85707bbb8013f0dea35e30403b3e57bb13054c1d0d389aceeba1d4160a148026212c7e017044e3ea69cd96fbd23b6aa9fd1e6f7e82494fbd5f8fc75856711a6b8998926e")
		signature2 = &tmp
	}
	var depositDataRoot2 *spec.Root
	{
		tmp := testutil.HexToRoot("0x3b51670e9f266d44c879682a230d60f0d534c64ab25ee68700fe3adb17ddfcab")
		depositDataRoot2 = &tmp
	}
	var depositMessageRoot2 *spec.Root
	{
		tmp := testutil.HexToRoot("0xbb4b6184b25873cdf430df3838c8d3e3d16cf3dc3b214e2f3ab7df9e6d5a9b52")
		depositMessageRoot2 = &tmp
	}

	tests := []struct {
		name    string
		dataOut []*dataOut
		res     string
		err     string
	}{
		{
			name: "Nil",
			res:  "[]",
		},
		{
			name: "NilDatum",
			dataOut: []*dataOut{
				nil,
			},
			res: "[]",
		},
		{
			name: "MissingValidatorPubKey",
			dataOut: []*dataOut{
				{
					format:                "launchpad",
					account:               "interop/00000",
					withdrawalCredentials: testutil.HexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
					amount:                32000000000,
					signature:             signature,
					forkVersion:           forkVersionPyrmont,
					depositDataRoot:       depositDataRoot,
					depositMessageRoot:    depositMessageRoot,
				},
			},
			err: "validator public key required",
		},
		{
			name: "MissingWithdrawalCredentials",
			dataOut: []*dataOut{
				{
					format:             "launchpad",
					account:            "interop/00000",
					validatorPubKey:    validatorPubKey,
					amount:             32000000000,
					signature:          signature,
					forkVersion:        forkVersionPyrmont,
					depositDataRoot:    depositDataRoot,
					depositMessageRoot: depositMessageRoot,
				},
			},
			err: "withdrawal credentials must be 32 bytes",
		},
		{
			name: "SignatureMissing",
			dataOut: []*dataOut{
				{
					format:                "launchpad",
					account:               "interop/00000",
					validatorPubKey:       validatorPubKey,
					withdrawalCredentials: testutil.HexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
					amount:                32000000000,
					forkVersion:           forkVersionPyrmont,
					depositDataRoot:       depositDataRoot,
					depositMessageRoot:    depositMessageRoot,
				},
			},
			err: "signature required",
		},
		{
			name: "AmountMissing",
			dataOut: []*dataOut{
				{
					format:                "launchpad",
					account:               "interop/00000",
					validatorPubKey:       validatorPubKey,
					withdrawalCredentials: testutil.HexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
					signature:             signature,
					forkVersion:           forkVersionPyrmont,
					depositDataRoot:       depositDataRoot,
					depositMessageRoot:    depositMessageRoot,
				},
			},
			err: "missing amount",
		},
		{
			name: "DepositDataRootMissing",
			dataOut: []*dataOut{
				{
					format:                "launchpad",
					account:               "interop/00000",
					validatorPubKey:       validatorPubKey,
					withdrawalCredentials: testutil.HexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
					amount:                32000000000,
					signature:             signature,
					forkVersion:           forkVersionPyrmont,
					depositMessageRoot:    depositMessageRoot,
				},
			},
			err: "deposit data root required",
		},
		{
			name: "DepositMessageRootMissing",
			dataOut: []*dataOut{
				{
					format:                "launchpad",
					account:               "interop/00000",
					validatorPubKey:       validatorPubKey,
					withdrawalCredentials: testutil.HexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
					amount:                32000000000,
					signature:             signature,
					forkVersion:           forkVersionPyrmont,
					depositDataRoot:       depositDataRoot,
				},
			},
			err: "deposit message root required",
		},
		{
			name: "SinglePyrmont",
			dataOut: []*dataOut{
				{
					format:                "launchpad",
					account:               "interop/00000",
					validatorPubKey:       validatorPubKey,
					withdrawalCredentials: testutil.HexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
					amount:                32000000000,
					signature:             signature,
					forkVersion:           forkVersionPyrmont,
					depositDataRoot:       depositDataRoot,
					depositMessageRoot:    depositMessageRoot,
				},
			},
			res: `[{"pubkey":"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","withdrawal_credentials":"00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b","amount":32000000000,"signature":"b7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2","deposit_message_root":"139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6","deposit_data_root":"9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554","fork_version":"00002009","eth2_network_name":"pyrmont","deposit_cli_version":"2.5.0"}]`,
		},
		{
			name: "SinglePrater",
			dataOut: []*dataOut{
				{
					format:                "launchpad",
					account:               "interop/00000",
					validatorPubKey:       validatorPubKey,
					withdrawalCredentials: testutil.HexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
					amount:                32000000000,
					signature:             signature,
					forkVersion:           forkVersionPrater,
					depositDataRoot:       depositDataRoot,
					depositMessageRoot:    depositMessageRoot,
				},
			},
			res: `[{"pubkey":"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","withdrawal_credentials":"00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b","amount":32000000000,"signature":"b7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2","deposit_message_root":"139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6","deposit_data_root":"9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554","fork_version":"00001020","eth2_network_name":"goerli","deposit_cli_version":"2.5.0"}]`,
		},
		{
			name: "Double",
			dataOut: []*dataOut{
				{
					format:                "launchpad",
					account:               "interop/00000",
					validatorPubKey:       validatorPubKey,
					withdrawalCredentials: testutil.HexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
					amount:                32000000000,
					signature:             signature,
					forkVersion:           forkVersionPyrmont,
					depositDataRoot:       depositDataRoot,
					depositMessageRoot:    depositMessageRoot,
				},
				{
					format:                "launchpad",
					account:               "interop/00001",
					validatorPubKey:       validatorPubKey2,
					withdrawalCredentials: testutil.HexToBytes("0x00ec7ef7780c9d151597924036262dd28dc60e1228f4da6fecf9d402cb3f3594"),
					amount:                32000000000,
					signature:             signature2,
					forkVersion:           forkVersionPyrmont,
					depositDataRoot:       depositDataRoot2,
					depositMessageRoot:    depositMessageRoot2,
				},
			},
			res: `[{"pubkey":"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","withdrawal_credentials":"00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b","amount":32000000000,"signature":"b7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2","deposit_message_root":"139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6","deposit_data_root":"9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554","fork_version":"00002009","eth2_network_name":"pyrmont","deposit_cli_version":"2.5.0"},{"pubkey":"b89bebc699769726a318c8e9971bd3171297c61aea4a6578a7a4f94b547dcba5bac16a89108b6b6a1fe3695d1a874a0b","withdrawal_credentials":"00ec7ef7780c9d151597924036262dd28dc60e1228f4da6fecf9d402cb3f3594","amount":32000000000,"signature":"911fe0766e8b79d711dde46bc2142eb51e35be99e5f7da505af9eaad85707bbb8013f0dea35e30403b3e57bb13054c1d0d389aceeba1d4160a148026212c7e017044e3ea69cd96fbd23b6aa9fd1e6f7e82494fbd5f8fc75856711a6b8998926e","deposit_message_root":"bb4b6184b25873cdf430df3838c8d3e3d16cf3dc3b214e2f3ab7df9e6d5a9b52","deposit_data_root":"3b51670e9f266d44c879682a230d60f0d534c64ab25ee68700fe3adb17ddfcab","fork_version":"00002009","eth2_network_name":"pyrmont","deposit_cli_version":"2.5.0"}]`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := output(test.dataOut)
			if test.err != "" {
				require.EqualError(t, err, test.err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.res, res)
			}
		})
	}
}

func TestOutputRaw(t *testing.T) {
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
		tmp := testutil.HexToSignature("0x911fe0766e8b79d711dde46bc2142eb51e35be99e5f7da505af9eaad85707bbb8013f0dea35e30403b3e57bb13054c1d0d389aceeba1d4160a148026212c7e017044e3ea69cd96fbd23b6aa9fd1e6f7e82494fbd5f8fc75856711a6b8998926e")
		signature2 = &tmp
	}
	var depositDataRoot2 *spec.Root
	{
		tmp := testutil.HexToRoot("0x3b51670e9f266d44c879682a230d60f0d534c64ab25ee68700fe3adb17ddfcab")
		depositDataRoot2 = &tmp
	}
	var depositMessageRoot2 *spec.Root
	{
		tmp := testutil.HexToRoot("0xbb4b6184b25873cdf430df3838c8d3e3d16cf3dc3b214e2f3ab7df9e6d5a9b52")
		depositMessageRoot2 = &tmp
	}

	tests := []struct {
		name    string
		dataOut []*dataOut
		res     string
		err     string
	}{
		{
			name: "Nil",
			res:  "[]",
		},
		{
			name: "NilDatum",
			dataOut: []*dataOut{
				nil,
			},
			res: "[]",
		},
		{
			name: "MissingValidatorPubKey",
			dataOut: []*dataOut{
				{
					format:                "raw",
					account:               "interop/00000",
					withdrawalCredentials: testutil.HexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
					amount:                32000000000,
					signature:             signature,
					forkVersion:           forkVersion,
					depositDataRoot:       depositDataRoot,
					depositMessageRoot:    depositMessageRoot,
				},
			},
			err: "validator public key required",
		},
		{
			name: "MissingWithdrawalCredentials",
			dataOut: []*dataOut{
				{
					format:             "raw",
					account:            "interop/00000",
					validatorPubKey:    validatorPubKey,
					amount:             32000000000,
					signature:          signature,
					forkVersion:        forkVersion,
					depositDataRoot:    depositDataRoot,
					depositMessageRoot: depositMessageRoot,
				},
			},
			err: "withdrawal credentials must be 32 bytes",
		},
		{
			name: "SignatureMissing",
			dataOut: []*dataOut{
				{
					format:                "raw",
					account:               "interop/00000",
					validatorPubKey:       validatorPubKey,
					withdrawalCredentials: testutil.HexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
					amount:                32000000000,
					forkVersion:           forkVersion,
					depositDataRoot:       depositDataRoot,
					depositMessageRoot:    depositMessageRoot,
				},
			},
			err: "signature required",
		},
		{
			name: "AmountMissing",
			dataOut: []*dataOut{
				{
					format:                "raw",
					account:               "interop/00000",
					validatorPubKey:       validatorPubKey,
					withdrawalCredentials: testutil.HexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
					signature:             signature,
					forkVersion:           forkVersion,
					depositDataRoot:       depositDataRoot,
					depositMessageRoot:    depositMessageRoot,
				},
			},
			err: "missing amount",
		},
		{
			name: "DepositDataRootMissing",
			dataOut: []*dataOut{
				{
					format:                "raw",
					account:               "interop/00000",
					validatorPubKey:       validatorPubKey,
					withdrawalCredentials: testutil.HexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
					amount:                32000000000,
					signature:             signature,
					forkVersion:           forkVersion,
					depositMessageRoot:    depositMessageRoot,
				},
			},
			err: "deposit data root required",
		},
		{
			name: "Single",
			dataOut: []*dataOut{
				{
					format:                "raw",
					account:               "interop/00000",
					validatorPubKey:       validatorPubKey,
					withdrawalCredentials: testutil.HexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
					amount:                32000000000,
					signature:             signature,
					forkVersion:           forkVersion,
					depositDataRoot:       depositDataRoot,
					depositMessageRoot:    depositMessageRoot,
				},
			},
			res: `["0x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000001209e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a35540000000000000000000000000000000000000000000000000000000000000030a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b0000000000000000000000000000000000000000000000000000000000000060b7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2"]`,
		},
		{
			name: "Double",
			dataOut: []*dataOut{
				{
					format:                "raw",
					account:               "interop/00000",
					validatorPubKey:       validatorPubKey,
					withdrawalCredentials: testutil.HexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
					amount:                32000000000,
					signature:             signature,
					forkVersion:           forkVersion,
					depositDataRoot:       depositDataRoot,
					depositMessageRoot:    depositMessageRoot,
				},
				{
					format:                "raw",
					account:               "interop/00001",
					validatorPubKey:       validatorPubKey2,
					withdrawalCredentials: testutil.HexToBytes("0x00ec7ef7780c9d151597924036262dd28dc60e1228f4da6fecf9d402cb3f3594"),
					amount:                32000000000,
					signature:             signature2,
					forkVersion:           forkVersion,
					depositDataRoot:       depositDataRoot2,
					depositMessageRoot:    depositMessageRoot2,
				},
			},
			res: `["0x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000001209e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a35540000000000000000000000000000000000000000000000000000000000000030a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b0000000000000000000000000000000000000000000000000000000000000060b7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2","0x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000001203b51670e9f266d44c879682a230d60f0d534c64ab25ee68700fe3adb17ddfcab0000000000000000000000000000000000000000000000000000000000000030b89bebc699769726a318c8e9971bd3171297c61aea4a6578a7a4f94b547dcba5bac16a89108b6b6a1fe3695d1a874a0b00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000ec7ef7780c9d151597924036262dd28dc60e1228f4da6fecf9d402cb3f35940000000000000000000000000000000000000000000000000000000000000060911fe0766e8b79d711dde46bc2142eb51e35be99e5f7da505af9eaad85707bbb8013f0dea35e30403b3e57bb13054c1d0d389aceeba1d4160a148026212c7e017044e3ea69cd96fbd23b6aa9fd1e6f7e82494fbd5f8fc75856711a6b8998926e"]`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := output(test.dataOut)
			if test.err != "" {
				require.EqualError(t, err, test.err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.res, res)
			}
		})
	}
}
