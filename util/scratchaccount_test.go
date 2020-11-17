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
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wealdtech/ethdo/testutil"
	"github.com/wealdtech/ethdo/util"
	e2types "github.com/wealdtech/go-eth2-types/v2"
)

func TestScratchAccountFromPrivKey(t *testing.T) {
	require.NoError(t, e2types.InitBLS())

	tests := []struct {
		name      string
		key       []byte
		err       string
		sigErr    string
		signature []byte
	}{
		{
			name: "Nil",
			err:  "public key must be 48 bytes",
		},
		{
			name: "KeyShort",
			key:  testutil.HexToBytes("0x295f0d1d592a90b333e26e85149708208e9f8e8bc18f6c77bd62f8ad7a6866"),
			err:  "private key must be 32 bytes",
		},
		{
			name: "KeyLong",
			key:  testutil.HexToBytes("0x2525295f0d1d592a90b333e26e85149708208e9f8e8bc18f6c77bd62f8ad7a6866"),
			err:  "private key must be 32 bytes",
		},
		{
			name:      "Good",
			key:       testutil.HexToBytes("0x25295f0d1d592a90b333e26e85149708208e9f8e8bc18f6c77bd62f8ad7a6866"),
			signature: testutil.HexToBytes("0x9004c971416fc1e48c0443d5650c4e998ab33b456223a1c3cd24da90e06174c0d66b80f492bc7b24d656a3c2d3051238020a3a4c0fd1fe98d61b97e9e5aa680841c965e8578425df4ce0b0a21270e330437931eadae1a9109336d415aeb420bb"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			account, err := util.NewScratchAccount(test.key, nil)
			if test.err == "" {
				require.NoError(t, err)
				require.NotNil(t, account.ID())
				require.Equal(t, "scratch", account.Name())
				require.Equal(t, "", account.Path())
				require.NotNil(t, account.PublicKey())
				unlocked, err := account.IsUnlocked(context.Background())
				require.NoError(t, err)
				require.False(t, unlocked)
				_, err = account.Sign(context.Background(), testutil.HexToBytes("0x000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"))
				require.EqualError(t, err, "locked")
				err = account.Unlock(context.Background(), nil)
				require.NoError(t, err)
				unlocked, err = account.IsUnlocked(context.Background())
				require.NoError(t, err)
				require.True(t, unlocked)
				signature, err := account.Sign(context.Background(), testutil.HexToBytes("0x000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"))
				if test.sigErr == "" {
					require.NoError(t, err)
					require.Equal(t, test.signature, signature.Marshal())
				} else {
					require.EqualError(t, err, test.sigErr)
				}
				require.NoError(t, account.Lock(context.Background()))
				unlocked, err = account.IsUnlocked(context.Background())
				require.NoError(t, err)
				require.False(t, unlocked)
			} else {
				require.EqualError(t, err, test.err)
			}
		})
	}
}

func TestScratchAccountFromPublicKey(t *testing.T) {
	require.NoError(t, e2types.InitBLS())

	tests := []struct {
		name   string
		pubKey []byte
		err    string
	}{
		{
			name: "Nil",
			err:  "public key must be 48 bytes",
		},
		{
			name:   "KeyShort",
			pubKey: testutil.HexToBytes("0x9a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c"),
			err:    "public key must be 48 bytes",
		},
		{
			name:   "KeyLong",
			pubKey: testutil.HexToBytes("0xa9a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c"),
			err:    "public key must be 48 bytes",
		},
		{
			name:   "Good",
			pubKey: testutil.HexToBytes("0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			account, err := util.NewScratchAccount(nil, test.pubKey)
			if test.err == "" {
				require.NoError(t, err)
				require.NotNil(t, account.ID())
				require.Equal(t, "scratch", account.Name())
				require.Equal(t, "", account.Path())
				unlocked, err := account.IsUnlocked(context.Background())
				require.NoError(t, err)
				require.False(t, unlocked)
				_, err = account.Sign(context.Background(), testutil.HexToBytes("0x000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"))
				require.EqualError(t, err, "locked")
				require.NoError(t, account.Unlock(context.Background(), nil))
				unlocked, err = account.IsUnlocked(context.Background())
				require.NoError(t, err)
				require.True(t, unlocked)
				_, err = account.Sign(context.Background(), testutil.HexToBytes("0x000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"))
				require.EqualError(t, err, "no private key")
				require.NoError(t, account.Lock(context.Background()))
				unlocked, err = account.IsUnlocked(context.Background())
				require.NoError(t, err)
				require.False(t, unlocked)
			} else {
				require.EqualError(t, err, test.err)
			}
		})
	}
}
