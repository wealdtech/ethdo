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

package accountderive

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wealdtech/ethdo/testutil"
	e2types "github.com/wealdtech/go-eth2-types/v2"
)

func TestProcess(t *testing.T) {
	require.NoError(t, e2types.InitBLS())

	tests := []struct {
		name    string
		dataIn  *dataIn
		privKey []byte
		err     string
	}{
		{
			name: "Nil",
			err:  "no data",
		},
		{
			name: "MnemonicMissing",
			dataIn: &dataIn{
				path: "m/12381/3600/0/0",
			},
			err: "failed to derive account: no account specified",
		},
		{
			name: "MnemonicInvalid",
			dataIn: &dataIn{
				mnemonic: "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art",
				path:     "m/12381/3600/0/0",
			},
			err: "failed to derive account: mnemonic is invalid",
		},
		{
			name: "PathMissing",
			dataIn: &dataIn{
				mnemonic: "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art",
			},
			err: "failed to derive account: path does not match expected format m/…",
		},
		{
			name: "PathInvalid",
			dataIn: &dataIn{
				mnemonic: "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art",
				path:     "n/12381/3600/0/0",
			},
			err: "failed to derive account: path does not match expected format m/…",
		},
		{
			name: "Good",
			dataIn: &dataIn{
				mnemonic: "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art",
				path:     "m/12381/3600/0/0",
			},
			privKey: testutil.HexToBytes("0x068dce0c90cb428ab37a74af0191eac49648035f1aaef077734b91e05985ec55"),
		},
		{
			name: "Extended",
			dataIn: &dataIn{
				mnemonic: "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art extended",
				path:     "m/12381/3600/0/0",
			},
			privKey: testutil.HexToBytes("0x58c8b280ae035de0452797b52fb62555f27f78541ea2f04b23e7bb0fcd0fc2d6"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := process(context.Background(), test.dataIn)
			if test.err != "" {
				require.EqualError(t, err, test.err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.privKey, res.key.Marshal())
			}
		})
	}
}
