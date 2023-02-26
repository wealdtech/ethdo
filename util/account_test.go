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
		{
			name:           "Keystore",
			accountStr:     `{"crypto": {"kdf": {"function": "scrypt", "params": {"dklen": 32, "n": 262144, "r": 8, "p": 1, "salt": "d27e392342918fa1912dadb171d90683c81146ba7ad36c0c22936d7fe3528300"}, "message": ""}, "checksum": {"function": "sha256", "params": {}, "message": "6f60216a8eda37426d3103f9fa608fe474944c4e287e09f416aad6bfe3983283"}, "cipher": {"function": "aes-128-ctr", "params": {"iv": "8b542e5a71fbde321407ba3d1ae098f6"}, "message": "a6bb744433adf9b7474b3793a09b71b451be1d595d031dba39adaaf6b9d6a67a"}}, "description": "", "pubkey": "91a4e10c877569f930e8800b745d4cb8fd03fd52dc17e87b49a55b548813275145e77ae01d56423becb5572f2632be5a", "path": "m/12381/3600/0/0/0", "uuid": "7858f402-cb53-4898-9193-b38bbf8fec12", "version": 4}`,
			expectedPubkey: "0x91a4e10c877569f930e8800b745d4cb8fd03fd52dc17e87b49a55b548813275145e77ae01d56423becb5572f2632be5a",
		},
		{
			name:             "KeystoreUnlocked",
			accountStr:       `{"crypto": {"kdf": {"function": "scrypt", "params": {"dklen": 32, "n": 262144, "r": 8, "p": 1, "salt": "d27e392342918fa1912dadb171d90683c81146ba7ad36c0c22936d7fe3528300"}, "message": ""}, "checksum": {"function": "sha256", "params": {}, "message": "6f60216a8eda37426d3103f9fa608fe474944c4e287e09f416aad6bfe3983283"}, "cipher": {"function": "aes-128-ctr", "params": {"iv": "8b542e5a71fbde321407ba3d1ae098f6"}, "message": "a6bb744433adf9b7474b3793a09b71b451be1d595d031dba39adaaf6b9d6a67a"}}, "description": "", "pubkey": "91a4e10c877569f930e8800b745d4cb8fd03fd52dc17e87b49a55b548813275145e77ae01d56423becb5572f2632be5a", "path": "m/12381/3600/0/0/0", "uuid": "7858f402-cb53-4898-9193-b38bbf8fec12", "version": 4}`,
			supplementary:    []string{"testtest"},
			expectedPubkey:   "0x91a4e10c877569f930e8800b745d4cb8fd03fd52dc17e87b49a55b548813275145e77ae01d56423becb5572f2632be5a",
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
