// Copyright © 2023 Weald Technology Trading.
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
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wealdtech/ethdo/util"
)

func bytesStr(input string) []byte {
	bytes, err := hex.DecodeString(strings.TrimPrefix(input, "0x"))
	if err != nil {
		panic(err)
	}
	return bytes
}

func TestSeedFromMnemonic(t *testing.T) {
	tests := []struct {
		name     string
		mnemonic string
		seed     []byte
		err      string
	}{
		{
			name: "Empty",
			err:  "mnemonic is invalid",
		},
		{
			name:     "Default",
			mnemonic: "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art",
			seed:     bytesStr("0x408b285c123836004f4b8842c89324c1f01382450c0d439af345ba7fc49acf705489c6fc77dbd4e3dc1dd8cc6bc9f043db8ada1e243c4a0eafb290d399480840"),
		},
		{
			name:     "English",
			mnemonic: "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art",
			seed:     bytesStr("0x408b285c123836004f4b8842c89324c1f01382450c0d439af345ba7fc49acf705489c6fc77dbd4e3dc1dd8cc6bc9f043db8ada1e243c4a0eafb290d399480840"),
		},
		{
			name:     "Spanish",
			mnemonic: "ábaco ábaco ábaco ábaco ábaco ábaco ábaco ábaco ábaco ábaco ábaco ábaco ábaco ábaco ábaco ábaco ábaco ábaco ábaco ábaco ábaco ábaco ábaco ancla",
			seed:     bytesStr("0x1e0de8aa97db3c7988f692d9c6151968be89debdbd71b1e34cab15d15ec10eed33412891129e1274fb84624565fd835f7e56df22a997439fca3da05c9c82a156"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			seed, err := util.SeedFromMnemonic(test.mnemonic)
			if test.err != "" {
				require.EqualError(t, err, test.err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.seed, seed)
			}
		})
	}
}
