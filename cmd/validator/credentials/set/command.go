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
	"github.com/attestantio/go-eth2-client/spec/bellatrix"
	capella "github.com/attestantio/go-eth2-client/spec/capella"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/wealdtech/ethdo/services/chaintime"
	"github.com/wealdtech/ethdo/util"
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
	withdrawalAddressStr  string
	forkVersion           string
	genesisValidatorsRoot string
	prepareOffline        bool

	// Beacon node connection.
	timeout                  time.Duration
	connection               string
	allowInsecureConnections bool

	// Information required to generate the operations.
	withdrawalAddress bellatrix.ExecutionAddress
	chainInfo         *chainInfo

	// Processing.
	consensusClient consensusclient.Service
	chainTime       chaintime.Service

	// Output.
	signedOperations []*capella.SignedBLSToExecutionChange
}

func newCommand(ctx context.Context) (*command, error) {
	c := &command{
		quiet:                    viper.GetBool("quiet"),
		verbose:                  viper.GetBool("verbose"),
		debug:                    viper.GetBool("debug"),
		offline:                  viper.GetBool("offline"),
		json:                     viper.GetBool("json"),
		timeout:                  viper.GetDuration("timeout"),
		connection:               viper.GetString("connection"),
		allowInsecureConnections: viper.GetBool("allow-insecure-connections"),
		prepareOffline:           viper.GetBool("prepare-offline"),
		account:                  viper.GetString("account"),
		passphrases:              util.GetPassphrases(),
		mnemonic:                 viper.GetString("mnemonic"),
		path:                     viper.GetString("path"),
		privateKey:               viper.GetString("private-key"),

		validator:             viper.GetString("validator"),
		withdrawalAddressStr:  viper.GetString("withdrawal-address"),
		forkVersion:           viper.GetString("fork-version"),
		genesisValidatorsRoot: viper.GetString("genesis-validators-root"),
	}

	// Timeout is required.
	if c.timeout == 0 {
		return nil, errors.New("timeout is required")
	}

	// We are generating information for offline use, we don't need any information
	// related to the accounts or signing.
	if c.prepareOffline {
		return c, nil
	}

	if c.account != "" && len(c.passphrases) == 0 {
		return nil, errors.New("passphrase required with account")
	}

	return c, nil
}
