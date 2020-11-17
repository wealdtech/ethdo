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

package util_test

import (
	"context"
	"testing"

	eth2client "github.com/attestantio/go-eth2-client"
	"github.com/stretchr/testify/require"
	"github.com/wealdtech/ethdo/testutil"
	"github.com/wealdtech/ethdo/util"
)

// A mock Ethereum 2 client service that returns supplied deposit information.
type depositETH2Client struct {
	address   []byte
	chainID   uint64
	networkID uint64
}

// Name returns the name of the client implementation.
func (c *depositETH2Client) Name() string {
	return "deposit mock"
}

// Address returns the address of the client.
func (c *depositETH2Client) Address() string {
	return "mock"
}

// DepositContractAddress provides the Ethereum 1 address of the deposit contract.
func (c *depositETH2Client) DepositContractAddress(ctx context.Context) ([]byte, error) {
	return c.address, nil
}

// DepositContractChainID provides the Ethereum 1 chain ID of the deposit contract.
func (c *depositETH2Client) DepositContractChainID(ctx context.Context) (uint64, error) {
	return c.chainID, nil
}

// DepositContractNetworkID provides the Ethereum 1 network ID of the deposit contract.
func (c *depositETH2Client) DepositContractNetworkID(ctx context.Context) (uint64, error) {
	return c.networkID, nil
}

// A mock Ethereum 2 client service that returns spec information.
type specETH2Client struct {
	address []byte
}

// Name returns the name of the client implementation.
func (c *specETH2Client) Name() string {
	return "spec mock"
}

// Address returns the address of the client.
func (c *specETH2Client) Address() string {
	return "mock"
}

// Spec provides the spec information of the chain.
func (c *specETH2Client) Spec(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{
		"DEPOSIT_CONTRACT_ADDRESS": c.address,
	}, nil
}

func TestNetworks(t *testing.T) {
	tests := []struct {
		name    string
		service eth2client.Service
		err     string
		network string
	}{
		{
			name: "Nil",
			err:  "no Ethereum 2 client supplied",
		},
		{
			name: "MainnetDeposit",
			service: &depositETH2Client{
				address:   testutil.HexToBytes("0x00000000219ab540356cbb839cbe05303d7705fa"),
				chainID:   0,
				networkID: 0,
			},
			network: "Mainnet",
		},
		{
			name: "MainnetSpec",
			service: &specETH2Client{
				address: testutil.HexToBytes("0x00000000219ab540356cbb839cbe05303d7705fa"),
			},
			network: "Mainnet",
		},
		{
			name: "UnknownDeposit",
			service: &depositETH2Client{
				address:   testutil.HexToBytes("0x1111111111111111111111111111111111111111"),
				chainID:   0,
				networkID: 0,
			},
			network: "Unknown",
		},
		{
			name: "UnknownSpec",
			service: &specETH2Client{
				address: testutil.HexToBytes("0x1111111111111111111111111111111111111111"),
			},
			network: "Unknown",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			network, err := util.Network(context.Background(), test.service)
			if test.err != "" {
				require.EqualError(t, err, test.err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.network, network)
			}
		})
	}
}
