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
		res     string
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
			res: "Public key: 0x99b1f1d84d76185466d86c34bde1101316afddae76217aa86cd066979b19858c2c9d9e56eebc1e067ac54277a61790db",
		},
		{
			name: "PrivatKey",
			dataOut: &dataOut{
				key:     blsPrivateKey("0x068dce0c90cb428ab37a74af0191eac49648035f1aaef077734b91e05985ec55"),
				showKey: true,
			},
			res: "Private key: 0x068dce0c90cb428ab37a74af0191eac49648035f1aaef077734b91e05985ec55\nPublic key: 0x99b1f1d84d76185466d86c34bde1101316afddae76217aa86cd066979b19858c2c9d9e56eebc1e067ac54277a61790db",
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
