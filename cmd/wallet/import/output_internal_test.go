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

package walletimport

import (
	"context"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
)

func TestOutput(t *testing.T) {
	export := &export{
		Wallet: &walletInfo{
			ID: uuid.FromBytesOrNil([]byte{
				0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
			}),
			Name: "Test wallet",
			Type: "non-deterministic",
		},
		Accounts: []*accountInfo{
			{
				Name: "Account 1",
			},
			{
				Name: "Account 2",
			},
		},
	}

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
			name: "Good",
			dataOut: &dataOut{
				export: export,
			},
			res: "",
		},
		{
			name: "Verify",
			dataOut: &dataOut{
				verify: true,
				export: export,
			},
			res: "Wallet name: Test wallet\nWallet type: non-deterministic\nWallet UUID: 00010203-0405-0607-0809-0a0b0c0d0e0f\nWallet accounts: 2",
		},
		{
			name: "VerifyVerbose",
			dataOut: &dataOut{
				verify:  true,
				verbose: true,
				export:  export,
			},
			res: "Wallet name: Test wallet\nWallet type: non-deterministic\nWallet UUID: 00010203-0405-0607-0809-0a0b0c0d0e0f\nWallet accounts: 2\n  Account 1\n  Account 2",
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
