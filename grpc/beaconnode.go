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

	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

// FetchValidatorIndex fetches the index of a validator.
func FetchValidatorIndex(conn *grpc.ClientConn, account wtypes.Account) (uint64, error) {
	validatorClient := ethpb.NewBeaconNodeValidatorClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
	defer cancel()

	// Fetch the account.
	req := &ethpb.ValidatorIndexRequest{
		PublicKey: account.PublicKey().Marshal(),
	}
	resp, err := validatorClient.ValidatorIndex(ctx, req)
	if err != nil {
		return 0, err
	}

	return resp.Index, nil
}

// FetchValidatorState fetches the state of a validator.
func FetchValidatorState(conn *grpc.ClientConn, account wtypes.Account) (ethpb.ValidatorStatus, error) {
	validatorClient := ethpb.NewBeaconNodeValidatorClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
	defer cancel()

	// Fetch the account.
	req := &ethpb.ValidatorStatusRequest{
		PublicKey: account.PublicKey().Marshal(),
	}
	resp, err := validatorClient.ValidatorStatus(ctx, req)
	if err != nil {
		return 0, err
	}

	return resp.Status, nil
}
