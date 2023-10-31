// Copyright © 2019-2021 Weald Technology Limited.
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

package depositdata

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	spec "github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
	"github.com/wealdtech/ethdo/signing"
	ethdoutil "github.com/wealdtech/ethdo/util"
	e2types "github.com/wealdtech/go-eth2-types/v2"
	util "github.com/wealdtech/go-eth2-util"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

func process(data *dataIn) ([]*dataOut, error) {
	if data == nil {
		return nil, errors.New("no data")
	}

	results := make([]*dataOut, 0)

	withdrawalCredentials, err := createWithdrawalCredentials(data)
	if err != nil {
		return nil, err
	}

	for _, validatorAccount := range data.validatorAccounts {
		validatorPubKey, err := ethdoutil.BestPublicKey(validatorAccount)
		if err != nil {
			return nil, errors.Wrap(err, "validator account does not provide a public key")
		}

		var pubKey spec.BLSPubKey
		copy(pubKey[:], validatorPubKey.Marshal())
		depositMessage := &spec.DepositMessage{
			PublicKey:             pubKey,
			WithdrawalCredentials: withdrawalCredentials,
			Amount:                data.amount,
		}
		root, err := depositMessage.HashTreeRoot()
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate deposit message root")
		}
		var depositMessageRoot spec.Root
		copy(depositMessageRoot[:], root[:])

		sig, err := signing.SignRoot(context.Background(), validatorAccount, data.passphrases, depositMessageRoot, *data.domain)
		if err != nil {
			return nil, errors.Wrap(err, "failed to sign deposit message")
		}

		depositData := &spec.DepositData{
			PublicKey:             pubKey,
			WithdrawalCredentials: withdrawalCredentials,
			Amount:                data.amount,
			Signature:             sig,
		}

		root, err = depositData.HashTreeRoot()
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate deposit data root")
		}
		var depositDataRoot spec.Root
		copy(depositDataRoot[:], root[:])

		validatorWallet := validatorAccount.(e2wtypes.AccountWalletProvider).Wallet()
		results = append(results, &dataOut{
			format:                data.format,
			account:               fmt.Sprintf("%s/%s", validatorWallet.Name(), validatorAccount.Name()),
			validatorPubKey:       &pubKey,
			withdrawalCredentials: withdrawalCredentials,
			amount:                data.amount,
			signature:             &sig,
			forkVersion:           data.forkVersion,
			depositMessageRoot:    &depositMessageRoot,
			depositDataRoot:       &depositDataRoot,
		})
	}
	return results, nil
}

// createWithdrawalCredentials creates withdrawal credentials given an account, public key or Ethereum 1 address.
func createWithdrawalCredentials(data *dataIn) ([]byte, error) {
	var withdrawalCredentials []byte

	switch {
	case data.withdrawalAccount != "":
		ctx, cancel := context.WithTimeout(context.Background(), data.timeout)
		defer cancel()
		_, withdrawalAccount, err := ethdoutil.WalletAndAccountFromPath(ctx, data.withdrawalAccount)
		if err != nil {
			return nil, errors.Wrap(err, "failed to obtain withdrawal account")
		}
		pubKey, err := ethdoutil.BestPublicKey(withdrawalAccount)
		if err != nil {
			return nil, errors.Wrap(err, "failed to obtain public key for withdrawal account")
		}
		withdrawalCredentials = util.SHA256(pubKey.Marshal())
		// This is hard-coded, to allow deposit data to be generated without a connection to the beacon node.
		withdrawalCredentials[0] = byte(0) // BLS_WITHDRAWAL_PREFIX
	case data.withdrawalPubKey != "":
		withdrawalPubKeyBytes, err := hex.DecodeString(strings.TrimPrefix(data.withdrawalPubKey, "0x"))
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode withdrawal public key")
		}
		if len(withdrawalPubKeyBytes) != 48 {
			return nil, errors.New("withdrawal public key must be exactly 48 bytes in length")
		}
		pubKey, err := e2types.BLSPublicKeyFromBytes(withdrawalPubKeyBytes)
		if err != nil {
			return nil, errors.Wrap(err, "withdrawal public key is not valid")
		}
		withdrawalCredentials = util.SHA256(pubKey.Marshal())
		// This is hard-coded, to allow deposit data to be generated without a connection to the beacon node.
		withdrawalCredentials[0] = byte(0) // BLS_WITHDRAWAL_PREFIX
	case data.withdrawalAddress != "":
		withdrawalAddressBytes, err := hex.DecodeString(strings.TrimPrefix(data.withdrawalAddress, "0x"))
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode withdrawal address")
		}
		if len(withdrawalAddressBytes) != 20 {
			return nil, errors.New("withdrawal address must be exactly 20 bytes in length")
		}
		// Ensure the address is properly checksummed.
		checksummedAddress := addressBytesToEIP55(withdrawalAddressBytes)
		if checksummedAddress != data.withdrawalAddress {
			return nil, fmt.Errorf("withdrawal address checksum does not match (expected %s)", checksummedAddress)
		}
		withdrawalCredentials = make([]byte, 32)
		copy(withdrawalCredentials[12:32], withdrawalAddressBytes)
		// This is hard-coded, to allow deposit data to be generated without a connection to the beacon node.
		withdrawalCredentials[0] = byte(1) // ETH1_ADDRESS_WITHDRAWAL_PREFIX
	default:
		return nil, errors.New("withdrawal account, public key or address is required")
	}

	return withdrawalCredentials, nil
}

// addressBytesToEIP55 converts a byte array in to an EIP-55 string format.
func addressBytesToEIP55(address []byte) string {
	bytes := []byte(hex.EncodeToString(address))
	hash := util.Keccak256(bytes)
	for i := 0; i < len(bytes); i++ {
		hashByte := hash[i/2]
		if i%2 == 0 {
			hashByte >>= 4
		} else {
			hashByte &= 0xf
		}
		if bytes[i] > '9' && hashByte > 7 {
			bytes[i] -= 32
		}
	}

	return fmt.Sprintf("0x%s", string(bytes))
}
