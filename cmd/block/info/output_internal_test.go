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

package blockinfo

import (
	"context"
	"testing"

	spec "github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/stretchr/testify/require"
	"github.com/wealdtech/ethdo/testutil"
)

func TestOutput(t *testing.T) {
	tests := []struct {
		name    string
		dataOut *dataOut
		res     string
		err     string
	}{
		{
			name: "Nil",
			err:  "no data",
		},
		{
			name:    "Good",
			dataOut: &dataOut{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := output(context.Background(), test.dataOut)
			if test.err != "" {
				require.EqualError(t, err, test.err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.res, res)
			}
		})
	}
}

// func TestOutputBlockText(t *testing.T) {
// 	tests := []struct {
// 		name              string
// 		dataOut           *dataOut
// 		signedBeaconBlock *spec.SignedBeaconBlock
// 		err               string
// 	}{
// 		{
// 			name: "Nil",
// 			err:  "no data",
// 		},
// 		{
// 			name:    "Good",
// 			dataOut: &dataOut{},
// 		},
// 	}
//
// 	for _, test := range tests {
// 		t.Run(test.name, func(t *testing.T) {
// 			res := outputBlockText(context.Background(), test.dataOut, test.signedBeaconBlock)
// 			if test.err != "" {
// 				require.EqualError(t, err, test.err)
// 			} else {
// 				require.NoError(t, err)
// 				require.Equal(t, test.res, res)
// 			}
// 		})
// 	}
// }

func TestOutputBlockDeposits(t *testing.T) {
	tests := []struct {
		name     string
		dataOut  *dataOut
		verbose  bool
		deposits []*spec.Deposit
		res      string
		err      string
	}{
		{
			name: "Nil",
			res:  "Deposits: 0\n",
		},
		{
			name: "Empty",
			res:  "Deposits: 0\n",
		},
		{
			name: "Single",
			deposits: []*spec.Deposit{
				{
					Data: &spec.DepositData{
						PublicKey:             testutil.HexToPubKey("0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c"),
						WithdrawalCredentials: testutil.HexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
						Amount:                spec.Gwei(32000000000),
						Signature:             testutil.HexToSignature("0xb7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2"),
					},
				},
			},
			res: "Deposits: 1\n",
		},
		{
			name: "SingleVerbose",
			deposits: []*spec.Deposit{
				{
					Data: &spec.DepositData{
						PublicKey:             testutil.HexToPubKey("0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c"),
						WithdrawalCredentials: testutil.HexToBytes("0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b"),
						Amount:                spec.Gwei(32000000000),
						Signature:             testutil.HexToSignature("0xb7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2"),
					},
				},
			},
			verbose: true,
			res:     "Deposits: 1\n  0:\n    Public key: 0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c\n    Amount: 32 Ether\n    Withdrawal credentials: 0x00fad2a6bfb0e7f1f0f45460944fbd8dfa7f37da06a4d13b3983cc90bb46963b\n    Signature: 0xb7a757a4c506ac6ac5f2d23e065de7d00dc9f5a6a3f9610a8b60b65f166379139ae382c91ecbbf5c9fabc34b1cd2cf8f0211488d50d8754716d8e72e17c1a00b5d9b37cc73767946790ebe66cf9669abfc5c25c67e1e2d1c2e11429d149c25a2\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := outputBlockDeposits(context.Background(), test.verbose, test.deposits)
			if test.err != "" {
				require.EqualError(t, err, test.err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.res, res)
			}
		})
	}
}

func TestOutputBlockETH1Data(t *testing.T) {
	tests := []struct {
		name     string
		dataOut  *dataOut
		verbose  bool
		eth1Data *spec.ETH1Data
		res      string
		err      string
	}{
		{
			name: "Good",
			eth1Data: &spec.ETH1Data{
				DepositRoot:  testutil.HexToRoot("0x92407b66d7daf4f30beb84820caae2cbba51add1c4648584101ff3c32151eb83"),
				DepositCount: 109936,
				BlockHash:    testutil.HexToBytes("0x77b03ebaf0f2835b491cbd99a7f4649a03a6e7999678603030a014a3c48b32a4"),
			},
			res: "Ethereum 1 deposit count: 109936\nEthereum 1 deposit root: 0x92407b66d7daf4f30beb84820caae2cbba51add1c4648584101ff3c32151eb83\nEthereum 1 block hash: 0x77b03ebaf0f2835b491cbd99a7f4649a03a6e7999678603030a014a3c48b32a4\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := outputBlockETH1Data(context.Background(), test.eth1Data)
			if test.err != "" {
				require.EqualError(t, err, test.err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.res, res)
			}
		})
	}
}
