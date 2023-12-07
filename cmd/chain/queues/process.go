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

package chainqueues

import (
	"context"
	"fmt"

	eth2client "github.com/attestantio/go-eth2-client"
	"github.com/attestantio/go-eth2-client/api"
	"github.com/pkg/errors"
	standardchaintime "github.com/wealdtech/ethdo/services/chaintime/standard"
	"github.com/wealdtech/ethdo/util"
)

func (c *command) process(ctx context.Context) error {
	// Obtain information we need to process.
	if err := c.setup(ctx); err != nil {
		return err
	}

	epoch, err := util.ParseEpoch(ctx, c.chainTime, c.epoch)
	if err != nil {
		return err
	}

	response, err := c.validatorsProvider.Validators(ctx, &api.ValidatorsOpts{
		State: fmt.Sprintf("%d", c.chainTime.FirstSlotOfEpoch(epoch)),
	})
	if err != nil {
		return errors.Wrap(err, "failed to obtain validators")
	}

	for _, validator := range response.Data {
		if validator.Validator == nil {
			continue
		}
		if validator.Validator.ActivationEligibilityEpoch <= epoch && validator.Validator.ActivationEpoch > epoch {
			c.activationQueue++
		}
		if validator.Validator.ExitEpoch != 0xffffffffffffffff && validator.Validator.ExitEpoch > epoch {
			c.exitQueue++
		}
	}

	return nil
}

func (c *command) setup(ctx context.Context) error {
	var err error

	// Connect to the client.
	c.eth2Client, err = util.ConnectToBeaconNode(ctx, &util.ConnectOpts{
		Address:       c.connection,
		Timeout:       c.timeout,
		AllowInsecure: c.allowInsecureConnections,
		LogFallback:   !c.quiet,
	})
	if err != nil {
		return errors.Wrap(err, "failed to connect to beacon node")
	}

	c.chainTime, err = standardchaintime.New(ctx,
		standardchaintime.WithSpecProvider(c.eth2Client.(eth2client.SpecProvider)),
		standardchaintime.WithGenesisProvider(c.eth2Client.(eth2client.GenesisProvider)),
	)
	if err != nil {
		return errors.Wrap(err, "failed to set up chaintime service")
	}

	var isProvider bool
	c.validatorsProvider, isProvider = c.eth2Client.(eth2client.ValidatorsProvider)
	if !isProvider {
		return errors.New("connection does not provide validator information")
	}

	return nil
}
