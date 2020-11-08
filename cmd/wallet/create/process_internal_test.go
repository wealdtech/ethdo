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

package walletcreate

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	e2types "github.com/wealdtech/go-eth2-types/v2"
	scratch "github.com/wealdtech/go-eth2-wallet-store-scratch"
)

func TestProcess(t *testing.T) {
	require.NoError(t, e2types.InitBLS())

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
			name: "TypeUnknown",
			dataIn: &dataIn{
				timeout:    5 * time.Second,
				store:      scratch.New(),
				walletType: "unknown",
				walletName: "Test wallet",
			},
			err: "wallet type not supported",
		},
		{
			name: "NDGood",
			dataIn: &dataIn{
				timeout:    5 * time.Second,
				store:      scratch.New(),
				walletType: "nd",
				walletName: "Test wallet",
			},
		},
		{
			name: "HDPassphraseMissing",
			dataIn: &dataIn{
				timeout:    5 * time.Second,
				store:      scratch.New(),
				walletType: "hd",
				walletName: "Test wallet",
			},
			err: "wallet passphrase is required for hierarchical deterministic wallets",
		},
		{
			name: "HDPassphraseWeak",
			dataIn: &dataIn{
				timeout:    5 * time.Second,
				store:      scratch.New(),
				walletType: "hd",
				walletName: "Test wallet",
				passphrase: "weak",
			},
			err: "supplied passphrase is weak; use a stronger one or run with the --allow-weak-passphrases flag",
		},
		{
			name: "HDQuiet",
			dataIn: &dataIn{
				timeout:    5 * time.Second,
				quiet:      true,
				store:      scratch.New(),
				walletType: "hd",
				walletName: "Test wallet",
				passphrase: "ce%NohGhah4ye5ra",
			},
			err: "creation of hierarchical deterministic wallets prints its mnemonic, so cannot be run with the --quiet flag",
		},
		{
			name: "HDMnemonic",
			dataIn: &dataIn{
				timeout:    5 * time.Second,
				store:      scratch.New(),
				walletType: "hd",
				walletName: "Test wallet",
				passphrase: "ce%NohGhah4ye5ra",
				mnemonic:   "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art",
			},
		},
		{
			name: "HDMnemonicExtra",
			dataIn: &dataIn{
				timeout:    5 * time.Second,
				store:      scratch.New(),
				walletType: "hd",
				walletName: "Test wallet",
				passphrase: "ce%NohGhah4ye5ra",
				mnemonic:   "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art extra",
			},
		},
		{
			name: "HDGood",
			dataIn: &dataIn{
				timeout:    5 * time.Second,
				store:      scratch.New(),
				walletType: "hd",
				walletName: "Test wallet",
				passphrase: "ce%NohGhah4ye5ra",
			},
		},
		{
			name: "DistributedGood",
			dataIn: &dataIn{
				timeout:    5 * time.Second,
				store:      scratch.New(),
				walletType: "distributed",
				walletName: "Test wallet",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := process(context.Background(), test.dataIn)
			if test.err != "" {
				require.EqualError(t, err, test.err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNilData(t *testing.T) {
	_, err := processND(context.Background(), nil)
	require.EqualError(t, err, "no data")
	_, err = processHD(context.Background(), nil)
	require.EqualError(t, err, "no data")
	_, err = processDistributed(context.Background(), nil)
	require.EqualError(t, err, "no data")
}
