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

package walletcreate

import (
	"context"
	"crypto/rand"
	"strings"

	"github.com/pkg/errors"
	bip39 "github.com/tyler-smith/go-bip39"
	"github.com/wealdtech/ethdo/util"
	distributed "github.com/wealdtech/go-eth2-wallet-distributed"
	keystorev4 "github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"
	hd "github.com/wealdtech/go-eth2-wallet-hd/v2"
	nd "github.com/wealdtech/go-eth2-wallet-nd/v2"
	"golang.org/x/text/unicode/norm"
)

func process(ctx context.Context, data *dataIn) (*dataOut, error) {
	if data == nil {
		return nil, errors.New("no data")
	}

	switch data.walletType {
	case "nd", "non-deterministic":
		return processND(ctx, data)
	case "hd", "hierarchical deterministic":
		return processHD(ctx, data)
	case "distributed":
		return processDistributed(ctx, data)
	default:
		return nil, errors.New("wallet type not supported")
	}
}

func processND(ctx context.Context, data *dataIn) (*dataOut, error) {
	if data == nil {
		return nil, errors.New("no data")
	}

	results := &dataOut{}

	if _, err := nd.CreateWallet(ctx, data.walletName, data.store, keystorev4.New()); err != nil {
		return nil, err
	}
	return results, nil
}

func processHD(ctx context.Context, data *dataIn) (*dataOut, error) {
	if data == nil {
		return nil, errors.New("no data")
	}
	if data.passphrase == "" {
		return nil, errors.New("wallet passphrase is required for hierarchical deterministic wallets")
	}
	if !util.AcceptablePassphrase(data.passphrase) {
		return nil, errors.New("supplied passphrase is weak; use a stronger one or run with the --allow-weak-passphrases flag")
	}
	if data.quiet {
		return nil, errors.New("creation of hierarchical deterministic wallets prints its mnemonic, so cannot be run with the --quiet flag")
	}

	results := &dataOut{}

	// Only show the mnemonic on output if we generate it.
	printMnemonic := data.mnemonic == ""
	mnemonicPassphrase := ""

	if data.mnemonic == "" {
		// Create a new random mnemonic.
		entropy := make([]byte, 32)
		_, err := rand.Read(entropy)
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate entropy for wallet mnemonic")
		}
		data.mnemonic, err = bip39.NewMnemonic(entropy)
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate wallet mnemonic")
		}
	} else {
		// We have an existing mnemonic.  If there are more than 24 words we treat the additional characters as the passphrase.
		mnemonicParts := strings.Split(data.mnemonic, " ")
		if len(mnemonicParts) > 24 {
			data.mnemonic = strings.Join(mnemonicParts[:24], " ")
			mnemonicPassphrase = strings.Join(mnemonicParts[24:], " ")
		}
	}
	// Normalise the input.
	data.mnemonic = string(norm.NFKD.Bytes([]byte(data.mnemonic)))
	mnemonicPassphrase = string(norm.NFKD.Bytes([]byte(mnemonicPassphrase)))

	// Ensure the mnemonic is valid
	if !bip39.IsMnemonicValid(data.mnemonic) {
		return nil, errors.New("mnemonic is not valid")
	}

	// Create seed from mnemonic and passphrase.
	seed := bip39.NewSeed(data.mnemonic, mnemonicPassphrase)

	if _, err := hd.CreateWallet(ctx, data.walletName, []byte(data.passphrase), data.store, keystorev4.New(), seed); err != nil {
		return nil, err
	}
	if printMnemonic {
		results.mnemonic = data.mnemonic
	}

	return results, nil
}

func processDistributed(ctx context.Context, data *dataIn) (*dataOut, error) {
	if data == nil {
		return nil, errors.New("no data")
	}

	results := &dataOut{}

	if _, err := distributed.CreateWallet(ctx, data.walletName, data.store, keystorev4.New()); err != nil {
		return nil, err
	}

	return results, nil
}
