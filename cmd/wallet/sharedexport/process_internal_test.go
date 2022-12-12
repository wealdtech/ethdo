// Copyright © 2021, 2022 Weald Technology Trading
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

package walletsharedexport

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	e2types "github.com/wealdtech/go-eth2-types/v2"
	e2wallet "github.com/wealdtech/go-eth2-wallet"
	keystorev4 "github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"
	nd "github.com/wealdtech/go-eth2-wallet-nd/v2"
	filesystem "github.com/wealdtech/go-eth2-wallet-store-filesystem"
)

func TestProcess(t *testing.T) {
	require.NoError(t, e2types.InitBLS())

	base, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	defer os.RemoveAll(base)
	store := filesystem.New(filesystem.WithLocation(base))
	require.NoError(t, e2wallet.UseStore(store))
	wallet, err := nd.CreateWallet(context.Background(), "Test wallet", store, keystorev4.New())
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
			name: "WalletMissing",
			dataIn: &dataIn{
				timeout: 5 * time.Second,
			},
			err: "wallet is required",
		},
		{
			name: "FileInvalid",
			dataIn: &dataIn{
				timeout:      5 * time.Second,
				wallet:       wallet,
				file:         "/bad/bad/bad/backup.dat",
				participants: 5,
				threshold:    3,
			},
			err: "failed to write export file: open /bad/bad/bad/backup.dat: no such file or directory",
		},
		{
			name: "Good",
			dataIn: &dataIn{
				timeout:      5 * time.Second,
				wallet:       wallet,
				file:         "test.dat",
				participants: 5,
				threshold:    3,
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
				os.Remove(test.dataIn.file)
				require.Len(t, res.shares, int(test.dataIn.participants))
			}
		})
	}
}
