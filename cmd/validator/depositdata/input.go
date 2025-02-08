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
	"fmt"
	"os"
	"strings"
	"time"

	spec "github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	ethdoutil "github.com/wealdtech/ethdo/util"
	e2types "github.com/wealdtech/go-eth2-types/v2"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
	string2eth "github.com/wealdtech/go-string2eth"
)

type dataIn struct {
	debug             bool
	format            string
	timeout           time.Duration
	withdrawalAccount string
	withdrawalPubKey  string
	withdrawalAddress string
	amount            spec.Gwei
	validatorAccounts []e2wtypes.Account
	forkVersion       *spec.Version
	domain            *spec.Domain
	passphrases       []string
	compounding       bool
}

func input() (*dataIn, error) {
	var err error
	data := &dataIn{
		debug:       viper.GetBool("debug"),
		forkVersion: &spec.Version{},
		domain:      &spec.Domain{},
		compounding: viper.GetBool("compounding"),
	}

	if viper.GetString("validatoraccount") == "" {
		return nil, errors.New("validator account is required")
	}

	if viper.GetDuration("timeout") == 0 {
		return nil, errors.New("timeout is required")
	}
	data.timeout = viper.GetDuration("timeout")

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

	data.withdrawalAccount = viper.GetString("withdrawalaccount")
	data.withdrawalPubKey = viper.GetString("withdrawalpubkey")
	data.withdrawalAddress = viper.GetString("withdrawaladdress")
	withdrawalDetailsPresent := 0
	if data.withdrawalAccount != "" {
		withdrawalDetailsPresent++
	}
	if data.withdrawalPubKey != "" {
		withdrawalDetailsPresent++
	}
	if data.withdrawalAddress != "" {
		withdrawalDetailsPresent++
	}
	if withdrawalDetailsPresent == 0 {
		return nil, errors.New("withdrawal account, public key or address is required")
	}
	if withdrawalDetailsPresent > 1 {
		return nil, errors.New("only one of withdrawal account, public key or address is allowed")
	}
	if data.compounding && data.withdrawalAddress == "" {
		return nil, errors.New("a compounding validator must be created with a withdrawal address")
	}

	if viper.GetString("depositvalue") == "" {
		return nil, errors.New("deposit value is required")
	}
	amount, err := string2eth.StringToGWei(viper.GetString("depositvalue"))
	if err != nil {
		return nil, errors.Wrap(err, "deposit value is invalid")
	}
	data.amount = spec.Gwei(amount)

	data.forkVersion, err = inputForkVersion(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to obtain fork version")
	}

	copy(data.domain[:], e2types.Domain(e2types.DomainDeposit, data.forkVersion[:], e2types.ZeroGenesisValidatorsRoot))
	if data.debug {
		fmt.Fprintf(os.Stderr, "Deposit domain is %#x\n", e2types.DomainDeposit)
		fmt.Fprintf(os.Stderr, "Fork version is %#x\n", *data.forkVersion)
		fmt.Fprintf(os.Stderr, "Genesis validators root is %#x\n", e2types.ZeroGenesisValidatorsRoot)
		fmt.Fprintf(os.Stderr, "Signature domain is %#x\n", *data.domain)
	}

	return data, nil
}

func inputForkVersion(_ context.Context) (*spec.Version, error) {
	// Default to mainnet.
	forkVersion := &spec.Version{0x00, 0x00, 0x00, 0x00}

	// Override if supplied.
	if viper.GetString("forkversion") != "" {
		data, err := hex.DecodeString(strings.TrimPrefix(viper.GetString("forkversion"), "0x"))
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode fork version")
		}
		if len(data) != 4 {
			return nil, errors.New("fork version must be exactly 4 bytes in length")
		}

		copy(forkVersion[:], data)
	}
	return forkVersion, nil
}
