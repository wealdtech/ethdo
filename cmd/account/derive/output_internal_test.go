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
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	e2types "github.com/wealdtech/go-eth2-types/v2"
)

func blsPrivateKey(input string) *e2types.BLSPrivateKey {
	data, err := hex.DecodeString(strings.TrimPrefix(input, "0x"))
	if err != nil {
		panic(err)
	}
	key, err := e2types.BLSPrivateKeyFromBytes(data)
	if err != nil {
		panic(err)
	}
	return key
}

func TestOutput(t *testing.T) {
	tests := []struct {
		name    string
		dataOut *dataOut
		needs   []string
		err     string
	}{
		{
			name: "Nil",
			err:  "no data",
		},
		{
			name:    "KeyMissing",
			dataOut: &dataOut{},
			err:     "no key",
		},
		{
			name: "Good",
			dataOut: &dataOut{
				key: blsPrivateKey("0x068dce0c90cb428ab37a74af0191eac49648035f1aaef077734b91e05985ec55"),
			},
			needs: []string{"Public key"},
		},
		{
			name: "PrivatKey",
			dataOut: &dataOut{
				key:            blsPrivateKey("0x068dce0c90cb428ab37a74af0191eac49648035f1aaef077734b91e05985ec55"),
				showPrivateKey: true,
			},
			needs: []string{"Private key"},
		},
		{
			name: "WithdrawalCredentials",
			dataOut: &dataOut{
				key:                       blsPrivateKey("0x068dce0c90cb428ab37a74af0191eac49648035f1aaef077734b91e05985ec55"),
				showWithdrawalCredentials: true,
			},
			needs: []string{"Withdrawal credentials"},
		},
		{
			name: "All",
			dataOut: &dataOut{
				key:                       blsPrivateKey("0x068dce0c90cb428ab37a74af0191eac49648035f1aaef077734b91e05985ec55"),
				showPrivateKey:            true,
				showWithdrawalCredentials: true,
			},
			needs: []string{"Private key", "Withdrawal credentials"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := output(context.Background(), test.dataOut)
			if test.err != "" {
				require.EqualError(t, err, test.err)
			} else {
				require.NoError(t, err)
				for _, need := range test.needs {
					require.Contains(t, res, need)
				}
			}
		})
	}
}
