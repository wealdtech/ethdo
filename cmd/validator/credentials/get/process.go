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
	"encoding/json"
	"fmt"

	eth2client "github.com/attestantio/go-eth2-client"
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
	c.consensusClient, err = util.ConnectToBeaconNode(ctx, &util.ConnectOpts{
		Address:       c.connection,
		Timeout:       c.timeout,
		AllowInsecure: c.allowInsecureConnections,
		LogFallback:   !c.quiet,
	})
	if err != nil {
		return errors.Wrap(err, "failed to connect to consensus node")
	}

	// Obtain the validators provider.
	var isProvider bool
	c.validatorsProvider, isProvider = c.consensusClient.(eth2client.ValidatorsProvider)
	if !isProvider {
		return errors.New("consensus node does not provide validator information")
	}

	return nil
}

func (c *command) fetchValidator(ctx context.Context) error {
	var err error
	c.validatorInfo, err = util.ParseValidator(ctx, c.validatorsProvider, c.validator, "head")
	if err != nil {
		return errors.Wrap(err, "failed to obtain validator information")
	}

	return nil
}
