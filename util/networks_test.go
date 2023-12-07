// Copyright © 2020, 2021 Weald Technology Trading
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
	"github.com/attestantio/go-eth2-client/api"
	"github.com/stretchr/testify/require"
	"github.com/wealdtech/ethdo/testutil"
	"github.com/wealdtech/ethdo/util"
)

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
func (c *specETH2Client) Spec(ctx context.Context, _ *api.SpecOpts) (*api.Response[map[string]any], error) {
	return &api.Response[map[string]any]{
		Data: map[string]any{
			"DEPOSIT_CONTRACT_ADDRESS": c.address,
		},
		Metadata: make(map[string]any),
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
			name: "MainnetSpec",
			service: &specETH2Client{
				address: testutil.HexToBytes("0x00000000219ab540356cbb839cbe05303d7705fa"),
			},
			network: "Mainnet",
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
