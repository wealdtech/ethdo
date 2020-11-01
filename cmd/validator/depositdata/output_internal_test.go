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
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func hexToBytes(input string) []byte {
	res, err := hex.DecodeString(strings.TrimPrefix(input, "0x"))
	if err != nil {
		panic(err)
	}
	return res
}

func TestOutputJSON(t *testing.T) {
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
					validatorPubKey:       hexToBytes("0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c"),
					withdrawalCredentials: hexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
					amount:                32000000000,
					signature:             hexToBytes("0xb7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2"),
					forkVersion:           hexToBytes("0x01020304"),
					depositDataRoot:       hexToBytes("0x9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554"),
					depositMessageRoot:    hexToBytes("0x139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6"),
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
					withdrawalCredentials: hexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
					amount:                32000000000,
					signature:             hexToBytes("0xb7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2"),
					forkVersion:           hexToBytes("0x01020304"),
					depositDataRoot:       hexToBytes("0x9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554"),
					depositMessageRoot:    hexToBytes("0x139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6"),
				},
			},
			err: "validator public key must be 48 bytes",
		},
		{
			name: "MissingWithdrawalCredentials",
			dataOut: []*dataOut{
				{
					format:             "json",
					account:            "interop/00000",
					validatorPubKey:    hexToBytes("0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c"),
					amount:             32000000000,
					signature:          hexToBytes("0xb7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2"),
					forkVersion:        hexToBytes("0x01020304"),
					depositDataRoot:    hexToBytes("0x9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554"),
					depositMessageRoot: hexToBytes("0x139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6"),
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
					validatorPubKey:       hexToBytes("0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c"),
					withdrawalCredentials: hexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
					amount:                32000000000,
					forkVersion:           hexToBytes("0x01020304"),
					depositDataRoot:       hexToBytes("0x9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554"),
					depositMessageRoot:    hexToBytes("0x139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6"),
				},
			},
			err: "signature must be 96 bytes",
		},
		{
			name: "AmountMissing",
			dataOut: []*dataOut{
				{
					format:                "json",
					account:               "interop/00000",
					validatorPubKey:       hexToBytes("0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c"),
					withdrawalCredentials: hexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
					signature:             hexToBytes("0xb7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2"),
					forkVersion:           hexToBytes("0x01020304"),
					depositDataRoot:       hexToBytes("0x9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554"),
					depositMessageRoot:    hexToBytes("0x139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6"),
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
					validatorPubKey:       hexToBytes("0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c"),
					withdrawalCredentials: hexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
					amount:                32000000000,
					signature:             hexToBytes("0xb7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2"),
					forkVersion:           hexToBytes("0x01020304"),
					depositMessageRoot:    hexToBytes("0x139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6"),
				},
			},
			err: "deposit data root must be 32 bytes",
		},
		{
			name: "Single",
			dataOut: []*dataOut{
				{
					format:                "json",
					account:               "interop/00000",
					validatorPubKey:       hexToBytes("0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c"),
					withdrawalCredentials: hexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
					amount:                32000000000,
					signature:             hexToBytes("0xb7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2"),
					forkVersion:           hexToBytes("0x01020304"),
					depositDataRoot:       hexToBytes("0x9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554"),
					depositMessageRoot:    hexToBytes("0x139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6"),
				},
			},
			res: `[{"name":"Deposit for interop/00000","account":"interop/00000","pubkey":"0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","withdrawal_credentials":"0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b","signature":"0xb7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2","value":32000000000,"deposit_data_root":"0x9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554","version":2}]`,
		},
		{
			name: "Double",
			dataOut: []*dataOut{
				{
					format:                "json",
					account:               "interop/00000",
					validatorPubKey:       hexToBytes("0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c"),
					withdrawalCredentials: hexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
					amount:                32000000000,
					signature:             hexToBytes("0xb7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2"),
					forkVersion:           hexToBytes("0x01020304"),
					depositDataRoot:       hexToBytes("0x9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554"),
					depositMessageRoot:    hexToBytes("0x139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6"),
				},
				{
					format:                "json",
					account:               "interop/00001",
					validatorPubKey:       hexToBytes("0xb89bebc699769726a318c8e9971bd3171297c61aea4a6578a7a4f94b547dcba5bac16a89108b6b6a1fe3695d1a874a0b"),
					withdrawalCredentials: hexToBytes("0x00ec7ef7780c9d151597924036262dd28dc60e1228f4da6fecf9d402cb3f3594"),
					amount:                32000000000,
					signature:             hexToBytes("0x911fe0766e8b79d711dde46bc2142eb51e35be99e5f7da505af9eaad85707bbb8013f0dea35e30403b3e57bb13054c1d0d389aceeba1d4160a148026212c7e017044e3ea69cd96fbd23b6aa9fd1e6f7e82494fbd5f8fc75856711a6b8998926e"),
					forkVersion:           hexToBytes("0x01020304"),
					depositDataRoot:       hexToBytes("0x3b51670e9f266d44c879682a230d60f0d534c64ab25ee68700fe3adb17ddfcab"),
					depositMessageRoot:    hexToBytes("0xbb4b6184b25873cdf430df3838c8d3e3d16cf3dc3b214e2f3ab7df9e6d5a9b52"),
				},
			},
			res: `[{"name":"Deposit for interop/00000","account":"interop/00000","pubkey":"0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","withdrawal_credentials":"0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b","signature":"0xb7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2","value":32000000000,"deposit_data_root":"0x9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554","version":2},{"name":"Deposit for interop/00001","account":"interop/00001","pubkey":"0xb89bebc699769726a318c8e9971bd3171297c61aea4a6578a7a4f94b547dcba5bac16a89108b6b6a1fe3695d1a874a0b","withdrawal_credentials":"0x00ec7ef7780c9d151597924036262dd28dc60e1228f4da6fecf9d402cb3f3594","signature":"0x911fe0766e8b79d711dde46bc2142eb51e35be99e5f7da505af9eaad85707bbb8013f0dea35e30403b3e57bb13054c1d0d389aceeba1d4160a148026212c7e017044e3ea69cd96fbd23b6aa9fd1e6f7e82494fbd5f8fc75856711a6b8998926e","value":32000000000,"deposit_data_root":"0x3b51670e9f266d44c879682a230d60f0d534c64ab25ee68700fe3adb17ddfcab","version":2}]`,
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
					withdrawalCredentials: hexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
					amount:                32000000000,
					signature:             hexToBytes("0xb7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2"),
					forkVersion:           hexToBytes("0x01020304"),
					depositDataRoot:       hexToBytes("0x9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554"),
					depositMessageRoot:    hexToBytes("0x139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6"),
				},
			},
			err: "validator public key must be 48 bytes",
		},
		{
			name: "MissingWithdrawalCredentials",
			dataOut: []*dataOut{
				{
					format:             "launchpad",
					account:            "interop/00000",
					validatorPubKey:    hexToBytes("0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c"),
					amount:             32000000000,
					signature:          hexToBytes("0xb7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2"),
					forkVersion:        hexToBytes("0x01020304"),
					depositDataRoot:    hexToBytes("0x9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554"),
					depositMessageRoot: hexToBytes("0x139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6"),
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
					validatorPubKey:       hexToBytes("0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c"),
					withdrawalCredentials: hexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
					amount:                32000000000,
					forkVersion:           hexToBytes("0x01020304"),
					depositDataRoot:       hexToBytes("0x9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554"),
					depositMessageRoot:    hexToBytes("0x139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6"),
				},
			},
			err: "signature must be 96 bytes",
		},
		{
			name: "AmountMissing",
			dataOut: []*dataOut{
				{
					format:                "launchpad",
					account:               "interop/00000",
					validatorPubKey:       hexToBytes("0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c"),
					withdrawalCredentials: hexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
					signature:             hexToBytes("0xb7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2"),
					forkVersion:           hexToBytes("0x01020304"),
					depositDataRoot:       hexToBytes("0x9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554"),
					depositMessageRoot:    hexToBytes("0x139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6"),
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
					validatorPubKey:       hexToBytes("0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c"),
					withdrawalCredentials: hexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
					amount:                32000000000,
					signature:             hexToBytes("0xb7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2"),
					forkVersion:           hexToBytes("0x01020304"),
					depositMessageRoot:    hexToBytes("0x139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6"),
				},
			},
			err: "deposit data root must be 32 bytes",
		},
		{
			name: "DepositMessageRootMissing",
			dataOut: []*dataOut{
				{
					format:                "launchpad",
					account:               "interop/00000",
					validatorPubKey:       hexToBytes("0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c"),
					withdrawalCredentials: hexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
					amount:                32000000000,
					signature:             hexToBytes("0xb7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2"),
					forkVersion:           hexToBytes("0x01020304"),
					depositDataRoot:       hexToBytes("0x9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554"),
				},
			},
			err: "deposit message root must be 32 bytes",
		},
		{
			name: "Single",
			dataOut: []*dataOut{
				{
					format:                "launchpad",
					account:               "interop/00000",
					validatorPubKey:       hexToBytes("0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c"),
					withdrawalCredentials: hexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
					amount:                32000000000,
					signature:             hexToBytes("0xb7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2"),
					forkVersion:           hexToBytes("0x01020304"),
					depositDataRoot:       hexToBytes("0x9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554"),
					depositMessageRoot:    hexToBytes("0x139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6"),
				},
			},
			res: `[{"pubkey":"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","withdrawal_credentials":"00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b","amount":32000000000,"signature":"b7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2","deposit_message_root":"139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6","deposit_data_root":"9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554","fork_version":"01020304"}]`,
		},
		{
			name: "Double",
			dataOut: []*dataOut{
				{
					format:                "launchpad",
					account:               "interop/00000",
					validatorPubKey:       hexToBytes("0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c"),
					withdrawalCredentials: hexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
					amount:                32000000000,
					signature:             hexToBytes("0xb7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2"),
					forkVersion:           hexToBytes("0x01020304"),
					depositDataRoot:       hexToBytes("0x9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554"),
					depositMessageRoot:    hexToBytes("0x139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6"),
				},
				{
					format:                "launchpad",
					account:               "interop/00001",
					validatorPubKey:       hexToBytes("0xb89bebc699769726a318c8e9971bd3171297c61aea4a6578a7a4f94b547dcba5bac16a89108b6b6a1fe3695d1a874a0b"),
					withdrawalCredentials: hexToBytes("0x00ec7ef7780c9d151597924036262dd28dc60e1228f4da6fecf9d402cb3f3594"),
					amount:                32000000000,
					signature:             hexToBytes("0x911fe0766e8b79d711dde46bc2142eb51e35be99e5f7da505af9eaad85707bbb8013f0dea35e30403b3e57bb13054c1d0d389aceeba1d4160a148026212c7e017044e3ea69cd96fbd23b6aa9fd1e6f7e82494fbd5f8fc75856711a6b8998926e"),
					forkVersion:           hexToBytes("0x01020304"),
					depositDataRoot:       hexToBytes("0x3b51670e9f266d44c879682a230d60f0d534c64ab25ee68700fe3adb17ddfcab"),
					depositMessageRoot:    hexToBytes("0xbb4b6184b25873cdf430df3838c8d3e3d16cf3dc3b214e2f3ab7df9e6d5a9b52"),
				},
			},
			res: `[{"pubkey":"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","withdrawal_credentials":"00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b","amount":32000000000,"signature":"b7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2","deposit_message_root":"139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6","deposit_data_root":"9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554","fork_version":"01020304"},{"pubkey":"b89bebc699769726a318c8e9971bd3171297c61aea4a6578a7a4f94b547dcba5bac16a89108b6b6a1fe3695d1a874a0b","withdrawal_credentials":"00ec7ef7780c9d151597924036262dd28dc60e1228f4da6fecf9d402cb3f3594","amount":32000000000,"signature":"911fe0766e8b79d711dde46bc2142eb51e35be99e5f7da505af9eaad85707bbb8013f0dea35e30403b3e57bb13054c1d0d389aceeba1d4160a148026212c7e017044e3ea69cd96fbd23b6aa9fd1e6f7e82494fbd5f8fc75856711a6b8998926e","deposit_message_root":"bb4b6184b25873cdf430df3838c8d3e3d16cf3dc3b214e2f3ab7df9e6d5a9b52","deposit_data_root":"3b51670e9f266d44c879682a230d60f0d534c64ab25ee68700fe3adb17ddfcab","fork_version":"01020304"}]`,
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
					withdrawalCredentials: hexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
					amount:                32000000000,
					signature:             hexToBytes("0xb7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2"),
					forkVersion:           hexToBytes("0x01020304"),
					depositDataRoot:       hexToBytes("0x9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554"),
					depositMessageRoot:    hexToBytes("0x139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6"),
				},
			},
			err: "validator public key must be 48 bytes",
		},
		{
			name: "MissingWithdrawalCredentials",
			dataOut: []*dataOut{
				{
					format:             "raw",
					account:            "interop/00000",
					validatorPubKey:    hexToBytes("0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c"),
					amount:             32000000000,
					signature:          hexToBytes("0xb7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2"),
					forkVersion:        hexToBytes("0x01020304"),
					depositDataRoot:    hexToBytes("0x9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554"),
					depositMessageRoot: hexToBytes("0x139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6"),
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
					validatorPubKey:       hexToBytes("0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c"),
					withdrawalCredentials: hexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
					amount:                32000000000,
					forkVersion:           hexToBytes("0x01020304"),
					depositDataRoot:       hexToBytes("0x9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554"),
					depositMessageRoot:    hexToBytes("0x139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6"),
				},
			},
			err: "signature must be 96 bytes",
		},
		{
			name: "AmountMissing",
			dataOut: []*dataOut{
				{
					format:                "raw",
					account:               "interop/00000",
					validatorPubKey:       hexToBytes("0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c"),
					withdrawalCredentials: hexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
					signature:             hexToBytes("0xb7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2"),
					forkVersion:           hexToBytes("0x01020304"),
					depositDataRoot:       hexToBytes("0x9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554"),
					depositMessageRoot:    hexToBytes("0x139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6"),
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
					validatorPubKey:       hexToBytes("0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c"),
					withdrawalCredentials: hexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
					amount:                32000000000,
					signature:             hexToBytes("0xb7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2"),
					forkVersion:           hexToBytes("0x01020304"),
					depositMessageRoot:    hexToBytes("0x139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6"),
				},
			},
			err: "deposit data root must be 32 bytes",
		},
		{
			name: "DepositMessageRootMissing",
			dataOut: []*dataOut{
				{
					format:                "raw",
					account:               "interop/00000",
					validatorPubKey:       hexToBytes("0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c"),
					withdrawalCredentials: hexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
					amount:                32000000000,
					signature:             hexToBytes("0xb7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2"),
					forkVersion:           hexToBytes("0x01020304"),
					depositDataRoot:       hexToBytes("0x9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554"),
				},
			},
			err: "deposit message root must be 32 bytes",
		},
		{
			name: "Single",
			dataOut: []*dataOut{
				{
					format:                "raw",
					account:               "interop/00000",
					validatorPubKey:       hexToBytes("0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c"),
					withdrawalCredentials: hexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
					amount:                32000000000,
					signature:             hexToBytes("0xb7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2"),
					forkVersion:           hexToBytes("0x01020304"),
					depositDataRoot:       hexToBytes("0x9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554"),
					depositMessageRoot:    hexToBytes("0x139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6"),
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
					validatorPubKey:       hexToBytes("0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c"),
					withdrawalCredentials: hexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
					amount:                32000000000,
					signature:             hexToBytes("0xb7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2"),
					forkVersion:           hexToBytes("0x01020304"),
					depositDataRoot:       hexToBytes("0x9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554"),
					depositMessageRoot:    hexToBytes("0x139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6"),
				},
				{
					format:                "raw",
					account:               "interop/00001",
					validatorPubKey:       hexToBytes("0xb89bebc699769726a318c8e9971bd3171297c61aea4a6578a7a4f94b547dcba5bac16a89108b6b6a1fe3695d1a874a0b"),
					withdrawalCredentials: hexToBytes("0x00ec7ef7780c9d151597924036262dd28dc60e1228f4da6fecf9d402cb3f3594"),
					amount:                32000000000,
					signature:             hexToBytes("0x911fe0766e8b79d711dde46bc2142eb51e35be99e5f7da505af9eaad85707bbb8013f0dea35e30403b3e57bb13054c1d0d389aceeba1d4160a148026212c7e017044e3ea69cd96fbd23b6aa9fd1e6f7e82494fbd5f8fc75856711a6b8998926e"),
					forkVersion:           hexToBytes("0x01020304"),
					depositDataRoot:       hexToBytes("0x3b51670e9f266d44c879682a230d60f0d534c64ab25ee68700fe3adb17ddfcab"),
					depositMessageRoot:    hexToBytes("0xbb4b6184b25873cdf430df3838c8d3e3d16cf3dc3b214e2f3ab7df9e6d5a9b52"),
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
