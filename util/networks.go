// Copyright © 2020 Weald Technology Trading
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

package util

import (
	"context"
	"fmt"

	eth2client "github.com/attestantio/go-eth2-client"
	"github.com/pkg/errors"
)

// networks is a map of deposit contract addresses to networks.
var networks = map[string]string{
	"00000000219ab540356cbb839cbe05303d7705fa": "Mainnet",
	"07b39f4fde4a38bace212b546dac87c58dfe3fdc": "Medalla",
	"8c5fecdc472e27bc447696f431e425d02dd46a8c": "Pyrmont",
}

// Network returns the name of the network., calculated from the deposit contract information.
// If not known, returns "Unknown".
func Network(ctx context.Context, eth2Client eth2client.Service) (string, error) {
	var address []byte
	var err error

	if eth2Client == nil {
		return "", errors.New("no Ethereum 2 client supplied")
	}

	if provider, isProvider := eth2Client.(eth2client.DepositContractProvider); isProvider {
		address, err = provider.DepositContractAddress(ctx)
		if err != nil {
			return "", errors.Wrap(err, "failed to obtain deposit contract address")
		}
	} else if provider, isProvider := eth2Client.(eth2client.SpecProvider); isProvider {
		config, err := provider.Spec(ctx)
		if err != nil {
			return "", errors.Wrap(err, "failed to obtain chain specification")
		}
		if config == nil {
			return "", errors.New("failed to return chain specification")
		}
		depositContractAddress, exists := config["DEPOSIT_CONTRACT_ADDRESS"]
		if exists {
			address = depositContractAddress.([]byte)
		}
	}

	return network(address), nil
}

// network returns a network given an Ethereum 1 contract address.
func network(address []byte) string {
	if network, exists := networks[fmt.Sprintf("%x", address)]; exists {
		return network
	}
	return "Unknown"
}
