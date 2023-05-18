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
			name:     "Twelve",
			mnemonic: "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
			seed:     bytesStr("0x5eb00bbddcf069084889a8ab9155568165f5c453ccb85e70811aaed6f6da5fc19a5ac40b389cd370d086206dec8aa6c43daea6690f20ad3d8d48b2d2ce9e38e4"),
		},
		{
			name:     "TwelvePlusPassphrase",
			mnemonic: "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about passphrase",
			seed:     bytesStr("0x4865438d10636e1453b2d3c06444c669b80fb1ae77111f1f91b64278ed4d493465276d2e00f93be2a8e82c2f72555370a4bf31bcf1f9addaf0a31499a3baeeae"),
		},
		{
			name:     "Eighteen",
			mnemonic: "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon agent",
			seed:     bytesStr("0x4975bb3d1faf5308c86a30893ee903a976296609db223fd717e227da5a813a34dc1428b71c84a787fc51f3b9f9dc28e9459f48c08bd9578e9d1b170f2d7ea506"),
		},
		{
			name:     "EighteenPlusPassphrase",
			mnemonic: "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon agent passphrase",
			seed:     bytesStr("0xbea1dd48440f3a8a7c02d0f7977fe03ba1dd409dda1ce971e80adc38f750c51d0959bd15c48cca2649cbcba8160d8a6c4026f2ee22dd387aa9b005041a5b8ea2"),
		},
		{
			name:     "TwentyFour",
			mnemonic: "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art",
			seed:     bytesStr("0x408b285c123836004f4b8842c89324c1f01382450c0d439af345ba7fc49acf705489c6fc77dbd4e3dc1dd8cc6bc9f043db8ada1e243c4a0eafb290d399480840"),
		},
		{
			name:     "TwentyFourPlusPassphrase",
			mnemonic: "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art passphrase",
			seed:     bytesStr("0x3b9096d658962052e9e778a18e7fddb8f530cbf783f38b26cf3e89fff6bf385728028ea0e906d47c24f88b666d61a59bdb88a7fc11b9e302ae75482c9562c282"),
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
