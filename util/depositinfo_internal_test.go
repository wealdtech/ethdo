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

package util

import (
	"testing"

	"github.com/stretchr/testify/require"
	e2types "github.com/wealdtech/go-eth2-types/v2"
)

func TestRawDepositInfo(t *testing.T) {
	require.NoError(t, e2types.InitBLS())

	tests := []struct {
		name  string
		input []byte
		err   string
	}{
		{
			name: "Nil",
			err:  "invalid transaction length",
		},
		{
			name:  "Invalid",
			input: []byte("invalid"),
			err:   "public key invalid",
		},
		{
			name:  "IncorrectSignature",
			input: []byte(`0x02895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000001209e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a35540000000000000000000000000000000000000000000000000000000000000030a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b0000000000000000000000000000000000000000000000000000000000000060b7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2`),
			err:   "invalid function signature",
		},
		{
			name:  "IncorrectSize",
			input: []byte(`0x22895118`),
			err:   "invalid transaction length",
		},
		{
			name:  "Good",
			input: []byte(`0x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000001209e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a35540000000000000000000000000000000000000000000000000000000000000030a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b0000000000000000000000000000000000000000000000000000000000000060b7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2`),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			depositInfo, err := tryRawTxData(test.input)
			if test.err == "" {
				require.NoError(t, err)
				require.NotNil(t, depositInfo)
			} else {
				require.EqualError(t, err, test.err)
			}
		})
	}
}

func TestCLIDepositInfo(t *testing.T) {
	require.NoError(t, e2types.InitBLS())

	tests := []struct {
		name  string
		input []byte
		err   string
	}{
		{
			name: "Nil",
			err:  "unexpected end of JSON input",
		},
		{
			name:  "Invalid",
			input: []byte("invalid"),
			err:   "invalid character 'i' looking for beginning of value",
		},
		{
			name:  "PubKeyMissing",
			input: []byte(`[{"withdrawal_credentials":"00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b","amount":32000000000,"signature":"b7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2","deposit_message_root":"139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6","deposit_data_root":"9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554","fork_version":"01020304"}]`),
			err:   "public key missing",
		},
		{
			name:  "PubKeyInvalid",
			input: []byte(`[{"pubkey":"invalid","withdrawal_credentials":"00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b","amount":32000000000,"signature":"b7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2","deposit_message_root":"139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6","deposit_data_root":"9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554","fork_version":"01020304"}]`),
			err:   "public key invalid",
		},
		{
			name:  "WithdrawalCredentialsMissing",
			input: []byte(`[{"pubkey":"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","amount":32000000000,"signature":"b7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2","deposit_message_root":"139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6","deposit_data_root":"9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554","fork_version":"01020304"}]`),
			err:   "withdrawal credentials missing",
		},
		{
			name:  "WithdrawalCredentialsInvalid",
			input: []byte(`[{"pubkey":"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","withdrawal_credentials":"invalid","amount":32000000000,"signature":"b7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2","deposit_message_root":"139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6","deposit_data_root":"9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554","fork_version":"01020304"}]`),
			err:   "withdrawal credentials invalid",
		},
		{
			name:  "SignatureMissing",
			input: []byte(`[{"pubkey":"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","withdrawal_credentials":"00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b","amount":32000000000,"deposit_message_root":"139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6","deposit_data_root":"9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554","fork_version":"01020304"}]`),
			err:   "signature missing",
		},
		{
			name:  "SignatureInvalid",
			input: []byte(`[{"pubkey":"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","withdrawal_credentials":"00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b","amount":32000000000,"signature":"invalid","deposit_message_root":"139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6","deposit_data_root":"9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554","fork_version":"01020304"}]`),
			err:   "signature invalid",
		},
		{
			name:  "DepositMessageRootMissing",
			input: []byte(`[{"pubkey":"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","withdrawal_credentials":"00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b","amount":32000000000,"signature":"b7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2","deposit_data_root":"9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554","fork_version":"01020304"}]`),
			err:   "deposit message root missing",
		},
		{
			name:  "DepositMessageRootInvalid",
			input: []byte(`[{"pubkey":"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","withdrawal_credentials":"00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b","amount":32000000000,"signature":"b7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2","deposit_message_root":"invalid","deposit_data_root":"9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554","fork_version":"01020304"}]`),
			err:   "deposit message root invalid",
		},
		{
			name:  "DepositDataRootMissing",
			input: []byte(`[{"pubkey":"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","withdrawal_credentials":"00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b","amount":32000000000,"signature":"b7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2","deposit_message_root":"139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6","fork_version":"01020304"}]`),
			err:   "deposit data root missing",
		},
		{
			name:  "DepositDataRootInvalid",
			input: []byte(`[{"pubkey":"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","withdrawal_credentials":"00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b","amount":32000000000,"signature":"b7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2","deposit_message_root":"139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6","deposit_data_root":"invalid","fork_version":"01020304"}]`),
			err:   "deposit data root invalid",
		},
		{
			name:  "ForkVersionMissing",
			input: []byte(`[{"pubkey":"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","withdrawal_credentials":"00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b","amount":32000000000,"signature":"b7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2","deposit_message_root":"139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6","deposit_data_root":"9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554"}]`),
			err:   "fork version missing",
		},
		{
			name:  "ForkVersionInvalid",
			input: []byte(`[{"pubkey":"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","withdrawal_credentials":"00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b","amount":32000000000,"signature":"b7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2","deposit_message_root":"139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6","deposit_data_root":"9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554","fork_version":"invalid"}]`),
			err:   "fork version invalid",
		},
		{
			name:  "Good",
			input: []byte(`[{"pubkey":"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","withdrawal_credentials":"00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b","amount":32000000000,"signature":"b7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2","deposit_message_root":"139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6","deposit_data_root":"9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554","fork_version":"01020304"}]`),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			depositInfo, err := tryCLIDepositInfoFromJSON(test.input)
			if test.err == "" {
				require.NoError(t, err)
				require.NotNil(t, depositInfo)
			} else {
				require.EqualError(t, err, test.err)
			}
		})
	}
}

func TestV3DepositInfo(t *testing.T) {
	require.NoError(t, e2types.InitBLS())

	tests := []struct {
		name  string
		input []byte
		err   string
	}{
		{
			name: "Nil",
			err:  "unexpected end of JSON input",
		},
		{
			name:  "Invalid",
			input: []byte("invalid"),
			err:   "invalid character 'i' looking for beginning of value",
		},
		{
			name:  "PubKeyMissing",
			input: []byte(`[{"name":"Deposit for interop/00000","account":"interop/00000","withdrawal_credentials":"0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b","value":32000000000,"signature":"0xb7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2","deposit_message_root":"0x139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6","deposit_data_root":"0x9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554","fork_version":"0x01020304","version":3}]`),
			err:   "public key missing",
		},
		{
			name:  "PubKeyInvalid",
			input: []byte(`[{"name":"Deposit for interop/00000","account":"interop/00000","pubkey":"invalid","withdrawal_credentials":"0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b","value":32000000000,"signature":"0xb7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2","deposit_message_root":"0x139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6","deposit_data_root":"0x9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554","fork_version":"0x01020304","version":3}]`),
			err:   "public key invalid",
		},
		{
			name:  "WithdrawalCredentialsMissing",
			input: []byte(`[{"name":"Deposit for interop/00000","account":"interop/00000","pubkey":"0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","value":32000000000,"signature":"0xb7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2","deposit_message_root":"0x139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6","deposit_data_root":"0x9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554","fork_version":"0x01020304","version":3}]`),
			err:   "withdrawal credentials missing",
		},
		{
			name:  "WithdrawalCredentialsInvalid",
			input: []byte(`[{"name":"Deposit for interop/00000","account":"interop/00000","pubkey":"0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","withdrawal_credentials":"invalid","value":32000000000,"signature":"0xb7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2","deposit_message_root":"0x139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6","deposit_data_root":"0x9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554","fork_version":"0x01020304","version":3}]`),
			err:   "withdrawal credentials invalid",
		},
		{
			name:  "SignatureMissing",
			input: []byte(`[{"name":"Deposit for interop/00000","account":"interop/00000","pubkey":"0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","withdrawal_credentials":"0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b","value":32000000000,"deposit_message_root":"0x139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6","deposit_data_root":"0x9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554","fork_version":"0x01020304","version":3}]`),
			err:   "signature missing",
		},
		{
			name:  "SignatureInvalid",
			input: []byte(`[{"name":"Deposit for interop/00000","account":"interop/00000","pubkey":"0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","withdrawal_credentials":"0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b","value":32000000000,"signature":"invalid","deposit_message_root":"0x139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6","deposit_data_root":"0x9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554","fork_version":"0x01020304","version":3}]`),
			err:   "signature invalid",
		},
		{
			name:  "DepositMessageRootMissing",
			input: []byte(`[{"name":"Deposit for interop/00000","account":"interop/00000","pubkey":"0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","withdrawal_credentials":"0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b","value":32000000000,"signature":"0xb7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2","deposit_data_root":"0x9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554","fork_version":"0x01020304","version":3}]`),
			err:   "deposit message root missing",
		},
		{
			name:  "DepositMessageRootInvalid",
			input: []byte(`[{"name":"Deposit for interop/00000","account":"interop/00000","pubkey":"0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","withdrawal_credentials":"0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b","value":32000000000,"signature":"0xb7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2","deposit_message_root":"invalid","deposit_data_root":"0x9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554","fork_version":"0x01020304","version":3}]`),
			err:   "deposit message root invalid",
		},
		{
			name:  "DepositDataRootMissing",
			input: []byte(`[{"name":"Deposit for interop/00000","account":"interop/00000","pubkey":"0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","withdrawal_credentials":"0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b","value":32000000000,"signature":"0xb7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2","deposit_message_root":"0x139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6","fork_version":"0x01020304","version":3}]`),
			err:   "deposit data root missing",
		},
		{
			name:  "DepositDataRootInvalid",
			input: []byte(`[{"name":"Deposit for interop/00000","account":"interop/00000","pubkey":"0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","withdrawal_credentials":"0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b","value":32000000000,"signature":"0xb7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2","deposit_message_root":"0x139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6","deposit_data_root":"invalid","fork_version":"0x01020304","version":3}]`),
			err:   "deposit data root invalid",
		},
		{
			name:  "ForkVersionMissing",
			input: []byte(`[{"name":"Deposit for interop/00000","account":"interop/00000","pubkey":"0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","withdrawal_credentials":"0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b","value":32000000000,"signature":"0xb7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2","deposit_message_root":"0x139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6","deposit_data_root":"0x9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554","version":3}]`),
			err:   "fork version missing",
		},
		{
			name:  "ForkVersionInvalid",
			input: []byte(`[{"name":"Deposit for interop/00000","account":"interop/00000","pubkey":"0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","withdrawal_credentials":"0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b","value":32000000000,"signature":"0xb7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2","deposit_message_root":"0x139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6","deposit_data_root":"0x9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554","fork_version":"invalid","version":3}]`),
			err:   "fork version invalid",
		},
		{
			name:  "Good",
			input: []byte(`[{"name":"Deposit for interop/00000","account":"interop/00000","pubkey":"0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","withdrawal_credentials":"0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b","value":32000000000,"signature":"0xb7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2","deposit_message_root":"0x139b510ea7f2788ab82da1f427d6cbe1db147c15a053db738ad5500cd83754a6","deposit_data_root":"0x9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554","fork_version":"0x01020304","version":3}]`),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			depositInfo, err := tryV3DepositInfoFromJSON(test.input)
			if test.err == "" {
				require.NoError(t, err)
				require.NotNil(t, depositInfo)
			} else {
				require.EqualError(t, err, test.err)
			}
		})
	}
}

func TestV1DepositInfo(t *testing.T) {
	require.NoError(t, e2types.InitBLS())

	tests := []struct {
		name  string
		input []byte
		err   string
	}{
		{
			name: "Nil",
			err:  "unexpected end of JSON input",
		},
		{
			name:  "Invalid",
			input: []byte("invalid"),
			err:   "invalid character 'i' looking for beginning of value",
		},
		{
			name:  "PubKeyInvalid",
			input: []byte(`[{"name":"Deposit for interop/00000","account":"interop/00000","pubkey":"invalid","withdrawal_credentials":"0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b","signature":"0xb7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2","value":32000000000,"deposit_data_root":"0x9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554","version":2}]`),
			err:   "public key invalid",
		},
		{
			name:  "WithdrawalCredentialsInvalid",
			input: []byte(`[{"name":"Deposit for interop/00000","account":"interop/00000","pubkey":"0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","withdrawal_credentials":"invalid","signature":"0xb7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2","value":32000000000,"deposit_data_root":"0x9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554","version":2}]`),
			err:   "withdrawal credentials invalid",
		},
		{
			name:  "SignatureInvalid",
			input: []byte(`[{"name":"Deposit for interop/00000","account":"interop/00000","pubkey":"0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","withdrawal_credentials":"0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b","signature":"invalid","value":32000000000,"deposit_data_root":"0x9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554","version":2}]`),
			err:   "signature invalid",
		},
		{
			name:  "DepositDataRootInvalid",
			input: []byte(`[{"name":"Deposit for interop/00000","account":"interop/00000","pubkey":"0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","withdrawal_credentials":"0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b","signature":"0xb7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2","value":32000000000,"deposit_data_root":"invalid","version":2}]`),
			err:   "deposit data root invalid",
		},
		{
			name:  "Good",
			input: []byte(`[{"name":"Deposit for interop/00000","account":"interop/00000","pubkey":"0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","withdrawal_credentials":"0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b","signature":"0xb7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2","value":32000000000,"deposit_data_root":"0x9e51b386f4271c18149dd0f73297a26a4a8c15c3622c44af79c92446f44a3554","version":2}]`),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			depositInfo, err := tryV1DepositInfoFromJSON(test.input)
			if test.err == "" {
				require.NoError(t, err)
				require.NotNil(t, depositInfo)
			} else {
				require.EqualError(t, err, test.err)
			}
		})
	}
}
