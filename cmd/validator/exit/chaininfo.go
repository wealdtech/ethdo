// Copyright © 2023 Weald Technology Trading.
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

package validatorexit

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/wealdtech/ethdo/beacon"
)

// obtainChainInfo obtains the chain information required to create an exit operation.
func (c *command) obtainChainInfo(ctx context.Context) error {
	var err error
	// Use the offline preparation file if present (and we haven't been asked to recreate it).
	if !c.prepareOffline {
		if err = c.obtainChainInfoFromFile(ctx); err == nil {
			return nil
		}
	}

	if c.offline {
		// If we are here it means that we are offline without chain information, and cannot continue.
		return fmt.Errorf("failed to obtain offline preparation file: %w", err)
	}

	return c.obtainChainInfoFromNode(ctx)
}

// obtainChainInfoFromFile obtains chain information from a pre-generated file.
func (c *command) obtainChainInfoFromFile(_ context.Context) error {
	_, err := os.Stat(offlinePreparationFilename)
	if err != nil {
		if c.debug {
			fmt.Fprintf(os.Stderr, "Failed to read offline preparation file: %v\n", err)
		}
		return err
	}

	if c.debug {
		fmt.Fprintf(os.Stderr, "%s found; loading chain state\n", offlinePreparationFilename)
	}
	data, err := os.ReadFile(offlinePreparationFilename)
	if err != nil {
		if c.debug {
			fmt.Fprintf(os.Stderr, "failed to load offline preparation file: %v\n", err)
		}
		return err
	}
	c.chainInfo = &beacon.ChainInfo{}
	if err := json.Unmarshal(data, c.chainInfo); err != nil {
		if c.debug {
			fmt.Fprintf(os.Stderr, "offline preparation file invalid: %v\n", err)
		}
		return err
	}

	return nil
}

// obtainChainInfoFromNode obtains chain info from a beacon node.
func (c *command) obtainChainInfoFromNode(ctx context.Context) error {
	if c.debug {
		fmt.Fprintf(os.Stderr, "Populating chain info from beacon node\n")
	}

	var err error
	c.chainInfo, err = beacon.ObtainChainInfoFromNode(ctx, c.consensusClient, c.chainTime)
	if err != nil {
		return err
	}

	return nil
}

// writeChainInfoToFile prepares for an offline run of this command by dumping
// the chain information to a file.
func (c *command) writeChainInfoToFile(_ context.Context) error {
	data, err := json.Marshal(c.chainInfo)
	if err != nil {
		return errors.Wrap(err, "failed to generate chain info JSON")
	}
	if err := os.WriteFile(offlinePreparationFilename, data, 0o600); err != nil {
		return errors.Wrap(err, "failed write chain info JSON")
	}

	return nil
}
