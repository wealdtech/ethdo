// Copyright © 2021 Weald Technology Trading
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

package validatorkeycheck

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/tyler-smith/go-bip39"
	e2types "github.com/wealdtech/go-eth2-types/v2"
	util "github.com/wealdtech/go-eth2-util"
	"golang.org/x/text/unicode/norm"
)

func process(ctx context.Context, data *dataIn) (*dataOut, error) {
	if data == nil {
		return nil, errors.New("no data")
	}

	validatorWithdrawalCredentials, err := hex.DecodeString(strings.TrimPrefix(data.withdrawalCredentials, "0x"))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse withdrawal credentials")
	}

	match := false
	path := ""
	if data.privKey != "" {
		// Single private key to check.
		keyBytes, err := hex.DecodeString(strings.TrimPrefix(data.privKey, "0x"))
		if err != nil {
			return nil, err
		}
		key, err := e2types.BLSPrivateKeyFromBytes(keyBytes)
		if err != nil {
			return nil, err
		}

		match, err = checkPrivKey(ctx, validatorWithdrawalCredentials, key)
		if err != nil {
			return nil, err
		}
	} else {
		// Mnemonic to check.
		match, path, err = checkMnemonic(ctx, data.debug, validatorWithdrawalCredentials, data.mnemonic)
		if err != nil {
			return nil, err
		}
	}

	results := &dataOut{
		debug:   data.debug,
		quiet:   data.quiet,
		verbose: data.verbose,
		match:   match,
		path:    path,
	}

	return results, nil
}

func checkPrivKey(_ context.Context, validatorWithdrawalCredentials []byte, key *e2types.BLSPrivateKey) (bool, error) {
	pubKey := key.PublicKey()

	withdrawalCredentials := util.SHA256(pubKey.Marshal())
	withdrawalCredentials[0] = byte(0) // BLS_WITHDRAWAL_PREFIX

	return bytes.Equal(withdrawalCredentials, validatorWithdrawalCredentials), nil
}

func checkMnemonic(ctx context.Context, debug bool, validatorWithdrawalCredentials []byte, mnemonic string) (bool, string, error) {
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

	if !bip39.IsMnemonicValid(mnemonic) {
		return false, "", errors.New("mnemonic is invalid")
	}

	// Create seed from mnemonic and passphrase.
	seed := bip39.NewSeed(mnemonic, mnemonicPassphrase)
	// Check first 1024 indices.
	for i := 0; i < 1024; i++ {
		path := fmt.Sprintf("m/12381/3600/%d/0", i)
		if debug {
			fmt.Printf("Checking path %s\n", path)
		}
		key, err := util.PrivateKeyFromSeedAndPath(seed, path)
		if err != nil {
			return false, "", errors.Wrap(err, "failed to generate key")
		}
		match, err := checkPrivKey(ctx, validatorWithdrawalCredentials, key)
		if err != nil {
			return false, "", errors.Wrap(err, "failed to match key")
		}
		if match {
			return true, path, nil
		}
	}

	return false, "", nil
}
