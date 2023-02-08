// Copyright © 2020 - 2023 Weald Technology Trading
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

package util

import (
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/tyler-smith/go-bip39"
	"github.com/tyler-smith/go-bip39/wordlists"
	"golang.org/x/text/unicode/norm"
)

// hdPathRegex is the regular expression that matches an HD path.
var hdPathRegex = regexp.MustCompile("^m/[0-9]+/[0-9]+(/[0-9+])+")

var mnemonicWordLists = [][]string{
	wordlists.English,
	wordlists.ChineseSimplified,
	wordlists.ChineseTraditional,
	wordlists.Czech,
	wordlists.French,
	wordlists.Italian,
	wordlists.Japanese,
	wordlists.Korean,
	wordlists.Spanish,
}

// SeedFromMnemonic creates a seed from a mnemonic.
func SeedFromMnemonic(mnemonic string) ([]byte, error) {
	// If there are more than 24 words we treat the additional characters as the passphrase.
	mnemonicParts := strings.Split(mnemonic, " ")
	mnemonicPassphrase := ""
	if len(mnemonicParts) > 24 {
		mnemonic = strings.Join(mnemonicParts[:24], " ")
		mnemonicPassphrase = strings.Join(mnemonicParts[24:], " ")
	}
	// Normalise the input.
	mnemonic = string(norm.NFKD.Bytes([]byte(mnemonic)))
	mnemonicPassphrase = string(norm.NFKD.Bytes([]byte(mnemonicPassphrase)))

	// Try with the various word lists.
	for _, wl := range mnemonicWordLists {
		bip39.SetWordList(wl)
		seed, err := bip39.NewSeedWithErrorChecking(mnemonic, mnemonicPassphrase)
		if err == nil {
			return seed, nil
		}
	}

	return nil, errors.New("mnemonic is invalid")
}
