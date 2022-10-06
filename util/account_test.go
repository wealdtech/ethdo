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

package util_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wealdtech/ethdo/util"
	e2types "github.com/wealdtech/go-eth2-types/v2"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

func TestParseAccount(t *testing.T) {
	ctx := context.Background()
	require.NoError(t, e2types.InitBLS())

	tests := []struct {
		name             string
		accountStr       string
		supplementary    []string
		unlock           bool
		err              string
		expectedPubkey   string
		expectedUnlocked bool
	}{
		{
			name: "Zero",
			err:  "no account specified",
		},
		{
			name:       "Bad",
			accountStr: "bad",
			err:        "unknown account specifier bad",
		},
		{
			name:           "PublicKey",
			accountStr:     "0x99b1f1d84d76185466d86c34bde1101316afddae76217aa86cd066979b19858c2c9d9e56eebc1e067ac54277a61790db",
			expectedPubkey: "0x99b1f1d84d76185466d86c34bde1101316afddae76217aa86cd066979b19858c2c9d9e56eebc1e067ac54277a61790db",
		},
		{
			name:       "PublicKeyUnlocked",
			accountStr: "0x99b1f1d84d76185466d86c34bde1101316afddae76217aa86cd066979b19858c2c9d9e56eebc1e067ac54277a61790db",
			unlock:     true,
			err:        "cannot unlock an account specified by its public key",
		},
		{
			name:           "PrivateKey",
			accountStr:     "0x068dce0c90cb428ab37a74af0191eac49648035f1aaef077734b91e05985ec55",
			expectedPubkey: "0x99b1f1d84d76185466d86c34bde1101316afddae76217aa86cd066979b19858c2c9d9e56eebc1e067ac54277a61790db",
		},
		{
			name:             "PrivateKeyUnlocked",
			accountStr:       "0x068dce0c90cb428ab37a74af0191eac49648035f1aaef077734b91e05985ec55",
			expectedPubkey:   "0x99b1f1d84d76185466d86c34bde1101316afddae76217aa86cd066979b19858c2c9d9e56eebc1e067ac54277a61790db",
			unlock:           true,
			expectedUnlocked: true,
		},
		{
			name:           "Mnemonic",
			accountStr:     "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art",
			supplementary:  []string{"m/12381/3600/0/0"},
			expectedPubkey: "0x99b1f1d84d76185466d86c34bde1101316afddae76217aa86cd066979b19858c2c9d9e56eebc1e067ac54277a61790db",
		},
		{
			name:             "MnemonicUnlocked",
			accountStr:       "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art",
			supplementary:    []string{"m/12381/3600/0/0"},
			expectedPubkey:   "0x99b1f1d84d76185466d86c34bde1101316afddae76217aa86cd066979b19858c2c9d9e56eebc1e067ac54277a61790db",
			unlock:           true,
			expectedUnlocked: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			account, err := util.ParseAccount(ctx, test.accountStr, test.supplementary, test.unlock)
			if test.err == "" {
				require.NoError(t, err)
				require.NotNil(t, account)
				unlocked, err := account.(e2wtypes.AccountLocker).IsUnlocked(ctx)
				require.NoError(t, err)
				require.Equal(t, test.expectedPubkey, fmt.Sprintf("%#x", account.PublicKey().Marshal()))
				require.Equal(t, test.expectedUnlocked, unlocked)
			} else {
				require.EqualError(t, err, test.err)
			}
		})
	}
}
