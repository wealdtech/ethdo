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
	// Handle situations where there may be a passphrase with the mnemonic.
	mnemonicParts := strings.Split(mnemonic, " ")
	mnemonicPassphrase := ""
	switch {
	case len(mnemonicParts) == 13:
		// Assume that passphrase is a single word here.
		mnemonic = strings.Join(mnemonicParts[:12], " ")
		mnemonicPassphrase = mnemonicParts[12]
	case len(mnemonicParts) == 19:
		// Assume that passphrase is a single word here.
		mnemonic = strings.Join(mnemonicParts[:18], " ")
		mnemonicPassphrase = mnemonicParts[18]
	case len(mnemonicParts) > 24:
		mnemonic = strings.Join(mnemonicParts[:24], " ")
		mnemonicPassphrase = strings.Join(mnemonicParts[24:], " ")
	}

	// Normalise the input.
	mnemonic = string(norm.NFKD.Bytes([]byte(mnemonic)))
	mnemonicPassphrase = string(norm.NFKD.Bytes([]byte(mnemonicPassphrase)))

	// Try with the various word lists.
	for _, wl := range mnemonicWordLists {
		bip39.SetWordList(wl)
		seed, err := bip39.NewSeedWithErrorChecking(expandMnemonic(mnemonic), mnemonicPassphrase)
		if err == nil {
			return seed, nil
		}
	}

	return nil, errors.New("mnemonic is invalid")
}

// expandMnmenonic expands mnemonics from their 4-letter versions.
func expandMnemonic(input string) string {
	wordList := bip39.GetWordList()
	truncatedWords := make(map[string]string, len(wordList))
	for _, word := range wordList {
		if len(word) > 4 {
			truncatedWords[firstFour(word)] = word
		}
	}
	mnemonicWords := strings.Split(input, " ")
	for i := range mnemonicWords {
		if fullWord, exists := truncatedWords[norm.NFKC.String(mnemonicWords[i])]; exists {
			mnemonicWords[i] = fullWord
		}
	}
	return strings.Join(mnemonicWords, " ")
}

// firstFour provides the first four letters for a potentially longer word.
func firstFour(s string) string {
	// Use NFKC here for composition, to avoid accents counting as their own characters.
	s = norm.NFKC.String(s)
	r := []rune(s)
	if len(r) > 4 {
		return string(r[:4])
	}
	return s
}
