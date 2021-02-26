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

package depositdata

import (
	"context"
	"encoding/hex"
	"strings"

	spec "github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	ethdoutil "github.com/wealdtech/ethdo/util"
	e2types "github.com/wealdtech/go-eth2-types/v2"
	util "github.com/wealdtech/go-eth2-util"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
	string2eth "github.com/wealdtech/go-string2eth"
)

type dataIn struct {
	format                string
	withdrawalCredentials []byte
	amount                spec.Gwei
	validatorAccounts     []e2wtypes.Account
	forkVersion           *spec.Version
	domain                *spec.Domain
	passphrases           []string
}

func input() (*dataIn, error) {
	var err error
	data := &dataIn{
		forkVersion: &spec.Version{},
		domain:      &spec.Domain{},
	}

	if viper.GetString("validatoraccount") == "" {
		return nil, errors.New("validator account is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
	defer cancel()
	_, data.validatorAccounts, err = ethdoutil.WalletAndAccountsFromPath(ctx, viper.GetString("validatoraccount"))
	if err != nil {
		return nil, errors.New("failed to obtain validator account")
	}
	if len(data.validatorAccounts) == 0 {
		return nil, errors.New("unknown validator account")
	}

	switch {
	case viper.GetBool("launchpad"):
		data.format = "launchpad"
	case viper.GetBool("raw"):
		data.format = "raw"
	default:
		data.format = "json"
	}

	data.passphrases = ethdoutil.GetPassphrases()

	switch {
	case viper.GetString("withdrawalaccount") != "":
		ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
		defer cancel()
		_, withdrawalAccount, err := ethdoutil.WalletAndAccountFromPath(ctx, viper.GetString("withdrawalaccount"))
		if err != nil {
			return nil, errors.Wrap(err, "failed to obtain withdrawal account")
		}
		pubKey, err := ethdoutil.BestPublicKey(withdrawalAccount)
		if err != nil {
			return nil, errors.Wrap(err, "failed to obtain public key for withdrawal account")
		}
		data.withdrawalCredentials = util.SHA256(pubKey.Marshal())
		// This is hard-coded, to allow deposit data to be generated without a connection to the beacon node.
		data.withdrawalCredentials[0] = byte(0) // BLS_WITHDRAWAL_PREFIX
	case viper.GetString("withdrawalpubkey") != "":
		withdrawalPubKeyBytes, err := hex.DecodeString(strings.TrimPrefix(viper.GetString("withdrawalpubkey"), "0x"))
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode withdrawal public key")
		}
		if len(withdrawalPubKeyBytes) != 48 {
			return nil, errors.New("withdrawal public key must be exactly 48 bytes in length")
		}
		withdrawalPubKey, err := e2types.BLSPublicKeyFromBytes(withdrawalPubKeyBytes)
		if err != nil {
			return nil, errors.Wrap(err, "withdrawal public key is not valid")
		}
		data.withdrawalCredentials = util.SHA256(withdrawalPubKey.Marshal())
		// This is hard-coded, to allow deposit data to be generated without a connection to the beacon node.
		data.withdrawalCredentials[0] = byte(0) // BLS_WITHDRAWAL_PREFIX
	case viper.GetString("withdrawaladdress") != "":
		// TODO checksum.
		withdrawalAddressBytes, err := hex.DecodeString(strings.TrimPrefix(viper.GetString("withdrawaladdress"), "0x"))
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode withdrawal address")
		}
		if len(withdrawalAddressBytes) != 20 {
			return nil, errors.New("withdrawal address must be exactly 20 bytes in length")
		}
		data.withdrawalCredentials = make([]byte, 32)
		copy(data.withdrawalCredentials[12:32], withdrawalAddressBytes[:])
		// This is hard-coded, to allow deposit data to be generated without a connection to the beacon node.
		data.withdrawalCredentials[0] = byte(1) // ETH1_ADDRESS_WITHDRAWAL_PREFIX
	default:
		return nil, errors.New("withdrawal account, public key or address is required")
	}

	if viper.GetString("depositvalue") == "" {
		return nil, errors.New("deposit value is required")
	}
	amount, err := string2eth.StringToGWei(viper.GetString("depositvalue"))
	if err != nil {
		return nil, errors.Wrap(err, "deposit value is invalid")
	}
	data.amount = spec.Gwei(amount)
	// This is hard-coded, to allow deposit data to be generated without a connection to the beacon node.
	if data.amount < 1000000000 { // MIN_DEPOSIT_AMOUNT
		return nil, errors.New("deposit value must be at least 1 Ether")
	}

	data.forkVersion, err = inputForkVersion(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to obtain fork version")
	}

	copy(data.domain[:], e2types.Domain(e2types.DomainDeposit, data.forkVersion[:], e2types.ZeroGenesisValidatorsRoot))

	return data, nil
}

func inputForkVersion(ctx context.Context) (*spec.Version, error) {
	// Default to mainnet.
	forkVersion := &spec.Version{0x00, 0x00, 0x00, 0x00}

	// Override if supplied.
	if viper.GetString("forkversion") != "" {
		data, err := hex.DecodeString(strings.TrimPrefix(viper.GetString("forkversion"), "0x"))
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode fork version")
		}
		if len(forkVersion) != 4 {
			return nil, errors.New("fork version must be exactly 4 bytes in length")
		}

		copy(forkVersion[:], data)
	}
	return forkVersion, nil
}
