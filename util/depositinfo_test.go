// Copyright © 2020 Weald Technology Trading
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

package util_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wealdtech/ethdo/util"
	e2types "github.com/wealdtech/go-eth2-types/v2"
)

func TestDepositInfo(t *testing.T) {
	require.NoError(t, e2types.InitBLS())

	tests := []struct {
		name  string
		input []byte
		err   string
	}{
		{
			name: "Nil",
			err:  "no data supplied",
		},
		{
			name:  "Invalid",
			input: []byte("bad"),
			err:   "unknown deposit data format",
		},
		{
			name:  "Incorrect",
			input: []byte("{}"),
			err:   "unknown deposit data format",
		},
		{
			name:  "Empty",
			input: []byte("[]"),
			err:   "no deposits supplied",
		},
		{
			name:  "V2",
			input: []byte(`{"name":"Deposit for interop/00000","account":"interop/00000","pubkey":"0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","withdrawal_credentials":"0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b","signature":"0xb7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2","value":32000000000,"deposit_data_root":"0x9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554","version":2}`),
		},
		{
			name:  "V3",
			input: []byte(`[{"name":"Deposit for interop/00000","account":"interop/00000","pubkey":"0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","withdrawal_credentials":"0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b","signature":"0xb7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2","amount":32000000000,"deposit_data_root":"0x9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554","deposit_message_root":"0x139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6","fork_version":"0x01020304","version":3}]`),
		},
		{
			name:  "Launchpad",
			input: []byte(`[{"pubkey":"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","withdrawal_credentials":"00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b","amount":32000000000,"signature":"b7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2","deposit_message_root":"139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6","deposit_data_root":"9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554","fork_version":"01020304"}]`),
		},
		{
			name:  "Raw",
			input: []byte(`0x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000001209e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a35540000000000000000000000000000000000000000000000000000000000000030a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b0000000000000000000000000000000000000000000000000000000000000060b7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2`),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			depositInfo, err := util.DepositInfoFromJSON(test.input)
			if test.err == "" {
				require.NoError(t, err)
				require.NotNil(t, depositInfo)
			} else {
				require.EqualError(t, err, test.err)
			}
		})
	}
}
