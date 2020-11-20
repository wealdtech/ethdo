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

package accountderive

import (
	"context"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/tyler-smith/go-bip39"
	util "github.com/wealdtech/go-eth2-util"
	"golang.org/x/text/unicode/norm"
)

// pathRegex is the regular expression that matches an HD path.
var pathRegex = regexp.MustCompile("^m/[0-9]+/[0-9]+(/[0-9+])+")

func process(ctx context.Context, data *dataIn) (*dataOut, error) {
	if data == nil {
		return nil, errors.New("no data")
	}

	// If there are more than 24 words we treat the additional characters as the passphrase.
	mnemonicParts := strings.Split(data.mnemonic, " ")
	mnemonicPassphrase := ""
	if len(mnemonicParts) > 24 {
		data.mnemonic = strings.Join(mnemonicParts[:24], " ")
		mnemonicPassphrase = strings.Join(mnemonicParts[24:], " ")
	}
	// Normalise the input.
	data.mnemonic = string(norm.NFKD.Bytes([]byte(data.mnemonic)))
	mnemonicPassphrase = string(norm.NFKD.Bytes([]byte(mnemonicPassphrase)))

	if !bip39.IsMnemonicValid(data.mnemonic) {
		return nil, errors.New("mnemonic is invalid")
	}

	// Create seed from mnemonic and passphrase.
	seed := bip39.NewSeed(data.mnemonic, mnemonicPassphrase)

	// Ensure the path is valid.
	match := pathRegex.Match([]byte(data.path))
	if !match {
		return nil, errors.New("path does not match expected format m/…")
	}

	// Derive private key from seed and path.
	key, err := util.PrivateKeyFromSeedAndPath(seed, data.path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate key")
	}

	results := &dataOut{
		showKey: data.showKey,
		key:     key,
	}

	return results, nil
}
