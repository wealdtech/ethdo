// Copyright © 2025 Weald Technology Trading.
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

package blocktrail

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	eth2client "github.com/attestantio/go-eth2-client"
	"github.com/attestantio/go-eth2-client/api"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
	standardchaintime "github.com/wealdtech/ethdo/services/chaintime/standard"
	"github.com/wealdtech/ethdo/util"
)

func (c *command) process(ctx context.Context) error {
	// Obtain information we need to process.
	if err := c.setup(ctx); err != nil {
		return err
	}

	untilRoot := phase0.Root{}
	var untilBlock phase0.Slot
	switch {
	case strings.ToLower(c.target) == "justified", strings.ToLower(c.target) == "finalized":
		// Nothing to do.
	case strings.HasPrefix(c.target, "0x"):
		// Assume a root.
		if err := json.Unmarshal([]byte(fmt.Sprintf("%q", c.target)), &untilRoot); err != nil {
			return err
		}
	default:
		// Assume a block number.
		tmp, err := strconv.ParseUint(c.target, 10, 64)
		if err != nil {
			return err
		}
		untilBlock = phase0.Slot(tmp)
	}

	blockID := c.blockID
	for range c.maxBlocks {
		step := &step{}

		blockResponse, err := c.blocksProvider.SignedBeaconBlock(ctx, &api.SignedBeaconBlockOpts{
			Block: blockID,
		})
		if err != nil {
			var apiError *api.Error
			if errors.As(err, &apiError) && apiError.StatusCode == http.StatusNotFound {
				return errors.New("empty beacon block")
			}
			return errors.Wrap(err, "failed to obtain beacon block")
		}
		block := blockResponse.Data

		step.Slot, err = block.Slot()
		if err != nil {
			return err
		}
		step.Root, err = block.Root()
		if err != nil {
			return err
		}
		step.ParentRoot, err = block.ParentRoot()
		if err != nil {
			return err
		}
		executionBlock, err := block.ExecutionBlockNumber()
		if err != nil {
			return err
		}
		step.ExecutionBlock = phase0.Slot(executionBlock)
		step.ExecutionHash, err = block.ExecutionBlockHash()
		if err != nil {
			return err
		}

		if c.debug {
			data, err := json.Marshal(step)
			if err == nil {
				fmt.Fprintf(os.Stderr, "Step is %s\n", string(data))
			}
		}

		c.steps = append(c.steps, step)

		blockID = step.ParentRoot.String()

		if c.target == "justified" && bytes.Equal(step.Root[:], c.justifiedCheckpoint.Root[:]) {
			c.found = true
			break
		}
		if c.target == "finalized" && bytes.Equal(step.Root[:], c.finalizedCheckpoint.Root[:]) {
			c.found = true
			break
		}
		if untilBlock > 0 && step.Slot == untilBlock {
			c.found = true
			break
		}
		if (!untilRoot.IsZero()) && bytes.Equal(step.Root[:], untilRoot[:]) {
			c.found = true
			break
		}
	}

	return nil
}

func (c *command) setup(ctx context.Context) error {
	var err error

	// Connect to the client.
	c.consensusClient, err = util.ConnectToBeaconNode(ctx, &util.ConnectOpts{
		Address:       c.connection,
		Timeout:       c.timeout,
		AllowInsecure: c.allowInsecureConnections,
		LogFallback:   !c.quiet,
	})
	if err != nil {
		return errors.Wrap(err, "failed to connect to beacon node")
	}

	c.chainTime, err = standardchaintime.New(ctx,
		standardchaintime.WithSpecProvider(c.consensusClient.(eth2client.SpecProvider)),
		standardchaintime.WithGenesisProvider(c.consensusClient.(eth2client.GenesisProvider)),
	)
	if err != nil {
		return errors.Wrap(err, "failed to set up chaintime service")
	}

	var isProvider bool
	c.blocksProvider, isProvider = c.consensusClient.(eth2client.SignedBeaconBlockProvider)
	if !isProvider {
		return errors.New("connection does not provide signed beacon block information")
	}
	c.blockHeadersProvider, isProvider = c.consensusClient.(eth2client.BeaconBlockHeadersProvider)
	if !isProvider {
		return errors.New("connection does not provide beacon block header information")
	}

	finalityProvider, isProvider := c.consensusClient.(eth2client.FinalityProvider)
	if !isProvider {
		return errors.New("connection does not provide finality information")
	}
	finalityResponse, err := finalityProvider.Finality(ctx, &api.FinalityOpts{
		State: "head",
	})
	if err != nil {
		return errors.Wrap(err, "failed to obtain finality")
	}
	finality := finalityResponse.Data
	c.justifiedCheckpoint = finality.Justified
	if c.debug {
		fmt.Fprintf(os.Stderr, "Justified checkpoint is %d / %#x\n", c.justifiedCheckpoint.Epoch, c.justifiedCheckpoint.Root)
	}
	c.finalizedCheckpoint = finality.Finalized
	if c.debug {
		fmt.Fprintf(os.Stderr, "Finalized checkpoint is %d / %#x\n", c.finalizedCheckpoint.Epoch, c.finalizedCheckpoint.Root)
	}

	return nil
}
