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

package validatorcredentialsget

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	eth2client "github.com/attestantio/go-eth2-client"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
	"github.com/wealdtech/ethdo/util"
)

func (c *command) process(ctx context.Context) error {
	// Obtain information we need to process.
	if err := c.setup(ctx); err != nil {
		return err
	}

	// Work out which validator we are dealing with.
	if err := c.fetchValidator(ctx); err != nil {
		return err
	}

	if c.debug {
		data, err := json.Marshal(c.validator)
		if err == nil {
			fmt.Println(string(data))
		}
	}

	return nil
}

func (c *command) setup(ctx context.Context) error {
	var err error

	// Connect to the consensus node.
	c.consensusClient, err = util.ConnectToBeaconNode(ctx, c.connection, c.timeout, c.allowInsecureConnections)
	if err != nil {
		return errors.Wrap(err, "failed to connect to consensus node")
	}

	// Obtain the validators provider.
	var isProvider bool
	c.validatorsProvider, isProvider = c.consensusClient.(eth2client.ValidatorsProvider)
	if !isProvider {
		return errors.New("consensu node does not provide validator information")
	}

	return nil
}

func (c *command) fetchValidator(ctx context.Context) error {
	if c.account != "" {
		_, account, err := util.WalletAndAccountFromInput(ctx)
		if err != nil {
			return errors.Wrap(err, "unable to obtain account")
		}

		accPubKey, err := util.BestPublicKey(account)
		if err != nil {
			return errors.Wrap(err, "unable to obtain public key for account")
		}
		pubKey := phase0.BLSPubKey{}
		copy(pubKey[:], accPubKey.Marshal())
		validators, err := c.validatorsProvider.ValidatorsByPubKey(ctx,
			"head",
			[]phase0.BLSPubKey{pubKey},
		)
		if err != nil {
			return errors.Wrap(err, "failed to obtain validator information")
		}
		if len(validators) == 0 {
			return errors.New("unknown validator")
		}
		for _, validator := range validators {
			c.validator = validator
		}
	}
	if c.index != "" {
		tmp, err := strconv.ParseUint(c.index, 10, 64)
		if err != nil {
			return errors.Wrap(err, "invalid validator index")
		}
		index := phase0.ValidatorIndex(tmp)
		validators, err := c.validatorsProvider.Validators(ctx,
			"head",
			[]phase0.ValidatorIndex{index},
		)
		if err != nil {
			return errors.Wrap(err, "failed to obtain validator information")
		}
		if _, exists := validators[index]; !exists {
			return errors.New("unknown validator")
		}
		c.validator = validators[index]
	}
	if c.pubKey != "" {
		bytes, err := hex.DecodeString(strings.TrimPrefix(c.pubKey, "0x"))
		if err != nil {
			return errors.Wrap(err, "invalid validator public key")
		}
		pubKey := phase0.BLSPubKey{}
		copy(pubKey[:], bytes)

		validators, err := c.validatorsProvider.ValidatorsByPubKey(ctx,
			"head",
			[]phase0.BLSPubKey{pubKey},
		)
		if err != nil {
			return errors.Wrap(err, "failed to obtain validator information")
		}
		if len(validators) == 0 {
			return errors.New("unknown validator")
		}
		for _, validator := range validators {
			c.validator = validator
		}
	}

	return nil
}
