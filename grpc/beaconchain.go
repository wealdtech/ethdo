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

package grpc

import (
	"context"

	"github.com/spf13/viper"
	"google.golang.org/grpc"

	"github.com/golang/protobuf/ptypes/empty"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	wtypes "github.com/wealdtech/go-eth2-wallet-types"
)

// FetchChainConfig fetches the chain configuration from the beacon node.
func FetchChainConfig(conn *grpc.ClientConn) (*ethpb.BeaconConfig, error) {
	beaconClient := ethpb.NewBeaconChainClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
	defer cancel()
	return beaconClient.GetBeaconConfig(ctx, &empty.Empty{})
}

// FetchValidator fetches validator information from the beacon node.
func FetchValidator(conn *grpc.ClientConn, account wtypes.Account) (*ethpb.Validator, error) {
	beaconClient := ethpb.NewBeaconChainClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
	defer cancel()

	req := &ethpb.GetValidatorRequest{
		QueryFilter: &ethpb.GetValidatorRequest_PublicKey{
			PublicKey: account.PublicKey().Marshal(),
		},
	}
	return beaconClient.GetValidator(ctx, req)
}
