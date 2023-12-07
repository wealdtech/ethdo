// Copyright © 2020 - 2023 Weald Technology Trading.
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
	"encoding/hex"

	eth2client "github.com/attestantio/go-eth2-client"
	"github.com/attestantio/go-eth2-client/api"
	"github.com/pkg/errors"
)

// networks is a map of deposit contract addresses to networks.
var networks = map[string]string{
	"00000000219ab540356cbb839cbe05303d7705fa": "Mainnet",
	"07b39f4fde4a38bace212b546dac87c58dfe3fdc": "Medalla",
	"8c5fecdc472e27bc447696f431e425d02dd46a8c": "Pyrmont",
	"ff50ed3d0ec03ac01d4c79aad74928bff48a7b2b": "Prater",
	"6f22ffbc56eff051aecf839396dd1ed9ad6bba9d": "Ropsten",
	"7f02c3e3c98b133055b8b348b2ac625669ed295d": "Sepolia",
	"4242424242424242424242424242424242424242": "Holesky",
}

// Network returns the name of the network., calculated from the deposit contract information.
// If not known, returns "Unknown".
func Network(ctx context.Context, eth2Client eth2client.Service) (string, error) {
	var address []byte

	if eth2Client == nil {
		return "", errors.New("no Ethereum 2 client supplied")
	}

	provider, isProvider := eth2Client.(eth2client.SpecProvider)
	if !isProvider {
		return "", errors.New("client does not provide deposit contract address")
	}
	specResponse, err := provider.Spec(ctx, &api.SpecOpts{})
	if err != nil {
		return "", errors.Wrap(err, "failed to obtain chain specification")
	}
	depositContractAddress, exists := specResponse.Data["DEPOSIT_CONTRACT_ADDRESS"]
	if exists {
		address = depositContractAddress.([]byte)
	}

	return network(address), nil
}

// network returns a network given an Ethereum 1 contract address.
func network(address []byte) string {
	if network, exists := networks[hex.EncodeToString(address)]; exists {
		return network
	}
	return "Unknown"
}
