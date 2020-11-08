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

package accountimport

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	e2types "github.com/wealdtech/go-eth2-types/v2"
	keystorev4 "github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"
	nd "github.com/wealdtech/go-eth2-wallet-nd/v2"
	scratch "github.com/wealdtech/go-eth2-wallet-store-scratch"
)

func TestProcess(t *testing.T) {
	require.NoError(t, e2types.InitBLS())

	testNDWallet, err := nd.CreateWallet(context.Background(),
		"Test",
		scratch.New(),
		keystorev4.New(),
	)
	require.NoError(t, err)

	tests := []struct {
		name   string
		dataIn *dataIn
		err    string
	}{
		{
			name: "Nil",
			err:  "no data",
		},
		{
			name: "PassphraseMissing",
			dataIn: &dataIn{
				timeout:          5 * time.Second,
				wallet:           testNDWallet,
				accountName:      "Good",
				passphrase:       "",
				walletPassphrase: "pass",
				key:              hexToBytes("0x25295f0d1d592a90b333e26e85149708208e9f8e8bc18f6c77bd62f8ad7a6866"),
			},
			err: "passphrase is required",
		},
		{
			name: "PassphraseWeak",
			dataIn: &dataIn{
				timeout:          5 * time.Second,
				wallet:           testNDWallet,
				accountName:      "Good",
				passphrase:       "poor",
				walletPassphrase: "pass",
				key:              hexToBytes("0x25295f0d1d592a90b333e26e85149708208e9f8e8bc18f6c77bd62f8ad7a6866"),
			},
			err: "supplied passphrase is weak; use a stronger one or run with the --allow-weak-passphrases flag",
		},
		{
			name: "Good",
			dataIn: &dataIn{
				timeout:          5 * time.Second,
				wallet:           testNDWallet,
				accountName:      "Good",
				passphrase:       "ce%NohGhah4ye5ra",
				walletPassphrase: "pass",
				key:              hexToBytes("0x25295f0d1d592a90b333e26e85149708208e9f8e8bc18f6c77bd62f8ad7a6866"),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := process(context.Background(), test.dataIn)
			if test.err != "" {
				require.EqualError(t, err, test.err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.dataIn.accountName, res.account.Name())
			}
		})
	}
}
