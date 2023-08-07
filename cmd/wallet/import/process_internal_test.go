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
	"time"

	"github.com/stretchr/testify/require"
	e2types "github.com/wealdtech/go-eth2-types/v2"
	e2wallet "github.com/wealdtech/go-eth2-wallet"
	keystorev4 "github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"
	nd "github.com/wealdtech/go-eth2-wallet-nd/v2"
	scratch "github.com/wealdtech/go-eth2-wallet-store-scratch"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

func TestProcess(t *testing.T) {
	require.NoError(t, e2types.InitBLS())

	store := scratch.New()
	require.NoError(t, e2wallet.UseStore(store))
	wallet, err := nd.CreateWallet(context.Background(), "Test wallet", store, keystorev4.New())
	require.NoError(t, err)
	data, err := wallet.(e2wtypes.WalletExporter).Export(context.Background(), []byte("ce%NohGhah4ye5ra"))
	require.NoError(t, err)
	require.NoError(t, e2wallet.UseStore(scratch.New()))

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
			name: "DataMissing",
			dataIn: &dataIn{
				timeout:    5 * time.Second,
				passphrase: "ce%NohGhah4ye5ra",
			},
			err: "import data is required",
		},
		{
			name: "DataBad",
			dataIn: &dataIn{
				timeout:    5 * time.Second,
				data:       append([]byte{0x00}, data...),
				passphrase: "ce%NohGhah4ye5ra",
			},
			err: "failed to import wallet: failed to decrypt wallet: unhandled version 0x00",
		},
		{
			name: "PassphraseMissing",
			dataIn: &dataIn{
				timeout: 5 * time.Second,
				data:    data,
			},
			err: "failed to import wallet: failed to decrypt wallet: invalid key",
		},
		{
			name: "PassphraseIncorrect",
			dataIn: &dataIn{
				timeout:    5 * time.Second,
				data:       data,
				passphrase: "weak",
			},
			err: "failed to import wallet: failed to decrypt wallet: invalid key",
		},
		{
			name: "Good",
			dataIn: &dataIn{
				timeout:    5 * time.Second,
				data:       data,
				passphrase: "ce%NohGhah4ye5ra",
			},
		},
		{
			name: "VerifyDataBad",
			dataIn: &dataIn{
				timeout:    5 * time.Second,
				verify:     true,
				data:       append([]byte{0x00}, data...),
				passphrase: "ce%NohGhah4ye5ra",
			},
			err: "failed to decrypt export: unhandled version 0x00",
		},
		{
			name: "VerifyPassphraseMissing",
			dataIn: &dataIn{
				timeout: 5 * time.Second,
				verify:  true,
				data:    data,
			},
			err: "failed to decrypt export: invalid key",
		},
		{
			name: "Verify",
			dataIn: &dataIn{
				timeout:    5 * time.Second,
				verify:     true,
				data:       data,
				passphrase: "ce%NohGhah4ye5ra",
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
