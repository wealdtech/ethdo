// Copyright © 2022 Weald Technology Trading.
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

package validatorcredentialsset

import (
	"context"
	"time"

	consensusclient "github.com/attestantio/go-eth2-client"
	apiv1 "github.com/attestantio/go-eth2-client/api/v1"
	capella "github.com/attestantio/go-eth2-client/spec/capella"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/wealdtech/ethdo/services/chaintime"
	"github.com/wealdtech/ethdo/util"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

type command struct {
	quiet   bool
	verbose bool
	debug   bool
	offline bool
	json    bool

	// Input.
	account               string
	passphrases           []string
	mnemonic              string
	path                  string
	privateKey            string
	validator             string
	withdrawalAddress     string
	signedOperation       string
	forkVersion           string
	genesisValidatorsRoot string

	// Beacon node connection.
	timeout                  time.Duration
	connection               string
	allowInsecureConnections bool

	// Processing.
	consensusClient   consensusclient.Service
	chainTime         chaintime.Service
	withdrawalAccount e2wtypes.Account
	validatorInfo     *apiv1.Validator
	domain            phase0.Domain
	op                *capella.BLSToExecutionChange

	// Output.
	signedOp *capella.SignedBLSToExecutionChange
}

func newCommand(ctx context.Context) (*command, error) {
	c := &command{
		quiet:   viper.GetBool("quiet"),
		verbose: viper.GetBool("verbose"),
		debug:   viper.GetBool("debug"),
		offline: viper.GetBool("offline"),
		json:    viper.GetBool("json"),
	}

	// Timeout.
	if viper.GetDuration("timeout") == 0 {
		return nil, errors.New("timeout is required")
	}
	c.timeout = viper.GetDuration("timeout")

	c.connection = viper.GetString("connection")
	c.allowInsecureConnections = viper.GetBool("allow-insecure-connections")

	c.account = viper.GetString("account")
	c.passphrases = util.GetPassphrases()
	c.mnemonic = viper.GetString("mnemonic")
	c.path = viper.GetString("path")
	c.privateKey = viper.GetString("private-key")

	if c.account == "" && c.mnemonic == "" && c.privateKey == "" {
		return nil, errors.New("one of account, mnemonic or private key required")
	}

	if c.account != "" && len(c.passphrases) == 0 {
		return nil, errors.New("passphrase required with account")
	}

	if c.mnemonic != "" && c.path == "" {
		return nil, errors.New("path required with mnemonic")
	}

	if viper.GetString("validator") == "" {
		return nil, errors.New("validator is required")
	}
	c.validator = viper.GetString("validator")

	c.withdrawalAddress = viper.GetString("withdrawal-address")

	c.signedOperation = viper.GetString("signed-operation")

	c.forkVersion = viper.GetString("fork-version")
	c.genesisValidatorsRoot = viper.GetString("genesis-validators-root")

	return c, nil
}
