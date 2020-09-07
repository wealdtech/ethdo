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
	"strconv"
	"strings"

	"github.com/gogo/protobuf/types"
	"github.com/pkg/errors"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/spf13/viper"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
	"google.golang.org/grpc"
)

// FetchChainConfig fetches the chain configuration from the beacon node.
// It tweaks the output to make it easier to work with by setting appropriate
// types.
func FetchChainConfig(conn *grpc.ClientConn) (map[string]interface{}, error) {
	if conn == nil {
		return nil, errors.New("no connection to beacon node")
	}
	beaconClient := ethpb.NewBeaconChainClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
	defer cancel()
	config, err := beaconClient.GetBeaconConfig(ctx, &types.Empty{})
	if err != nil {
		return nil, err
	}
	results := make(map[string]interface{})
	for k, v := range config.Config {
		// Handle integers
		if v == "0" {
			results[k] = uint64(0)
			continue
		}
		intVal, err := strconv.ParseUint(v, 10, 64)
		if err == nil && intVal != 0 {
			results[k] = intVal
			continue
		}

		// Handle byte arrays
		if strings.HasPrefix(v, "[") {
			vals := strings.Split(v[1:len(v)-1], " ")
			res := make([]byte, len(vals))
			for i, val := range vals {
				intVal, err := strconv.Atoi(val)
				if err != nil {
					return nil, errors.Wrapf(err, "failed to convert value %q for %s", v, k)
				}
				res[i] = byte(intVal)
			}
			results[k] = res
			continue
		}

		// String (or unhandled format)
		results[k] = v
	}
	return results, nil
}

func FetchLatestFilledSlot(conn *grpc.ClientConn) (uint64, error) {
	if conn == nil {
		return 0, errors.New("no connection to beacon node")
	}
	beaconClient := ethpb.NewBeaconChainClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
	defer cancel()

	chainHead, err := beaconClient.GetChainHead(ctx, &types.Empty{})
	if err != nil {
		return 0, errors.Wrap(err, "failed to obtain latest")
	}

	return chainHead.HeadSlot, nil
}

// FetchValidatorCommittees fetches the validator committees for a given epoch.
func FetchValidatorCommittees(conn *grpc.ClientConn, epoch uint64) (map[uint64][][]uint64, error) {
	if conn == nil {
		return nil, errors.New("no connection to beacon node")
	}
	beaconClient := ethpb.NewBeaconChainClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
	defer cancel()

	req := &ethpb.ListCommitteesRequest{
		QueryFilter: &ethpb.ListCommitteesRequest_Epoch{
			Epoch: epoch,
		},
	}
	resp, err := beaconClient.ListBeaconCommittees(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to obtain committees")
	}

	res := make(map[uint64][][]uint64)
	for slot, committees := range resp.Committees {
		res[slot] = make([][]uint64, len(resp.Committees))
		for i, committee := range committees.Committees {
			res[slot][uint64(i)] = make([]uint64, len(committee.ValidatorIndices))
			indices := make([]uint64, len(committee.ValidatorIndices))
			copy(indices, committee.ValidatorIndices)
			res[slot][uint64(i)] = indices
		}
	}

	return res, nil
}

// FetchValidator fetches the validator definition from the beacon node.
func FetchValidator(conn *grpc.ClientConn, account e2wtypes.Account) (*ethpb.Validator, error) {
	if conn == nil {
		return nil, errors.New("no connection to beacon node")
	}
	beaconClient := ethpb.NewBeaconChainClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
	defer cancel()

	var pubKey []byte
	if pubKeyProvider, ok := account.(e2wtypes.AccountCompositePublicKeyProvider); ok {
		pubKey = pubKeyProvider.CompositePublicKey().Marshal()
	} else if pubKeyProvider, ok := account.(e2wtypes.AccountPublicKeyProvider); ok {
		pubKey = pubKeyProvider.PublicKey().Marshal()
	} else {
		return nil, errors.New("Unable to obtain public key")
	}

	req := &ethpb.GetValidatorRequest{
		QueryFilter: &ethpb.GetValidatorRequest_PublicKey{
			PublicKey: pubKey,
		},
	}
	return beaconClient.GetValidator(ctx, req)
}

// FetchValidatorByIndex fetches the validator definition from the beacon node.
func FetchValidatorByIndex(conn *grpc.ClientConn, index uint64) (*ethpb.Validator, error) {
	if conn == nil {
		return nil, errors.New("no connection to beacon node")
	}
	beaconClient := ethpb.NewBeaconChainClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
	defer cancel()

	req := &ethpb.GetValidatorRequest{
		QueryFilter: &ethpb.GetValidatorRequest_Index{
			Index: index,
		},
	}
	return beaconClient.GetValidator(ctx, req)
}

// FetchValidatorBalance fetches the validator balance from the beacon node.
func FetchValidatorBalance(conn *grpc.ClientConn, account e2wtypes.Account) (uint64, error) {
	if conn == nil {
		return 0, errors.New("no connection to beacon node")
	}
	beaconClient := ethpb.NewBeaconChainClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
	defer cancel()

	var pubKey []byte
	if pubKeyProvider, ok := account.(e2wtypes.AccountCompositePublicKeyProvider); ok {
		pubKey = pubKeyProvider.CompositePublicKey().Marshal()
	} else if pubKeyProvider, ok := account.(e2wtypes.AccountPublicKeyProvider); ok {
		pubKey = pubKeyProvider.PublicKey().Marshal()
	} else {
		return 0, errors.New("Unable to obtain public key")
	}

	res, err := beaconClient.ListValidatorBalances(ctx, &ethpb.ListValidatorBalancesRequest{
		PublicKeys: [][]byte{pubKey},
	})
	if err != nil {
		return 0, err
	}
	if len(res.Balances) == 0 {
		return 0, errors.New("unknown validator")
	}
	return res.Balances[0].Balance, nil
}

// FetchValidatorPerformance fetches the validator performance from the beacon node.
func FetchValidatorPerformance(conn *grpc.ClientConn, account e2wtypes.Account) (bool, bool, bool, uint64, int64, error) {
	if conn == nil {
		return false, false, false, 0, 0, errors.New("no connection to beacon node")
	}
	beaconClient := ethpb.NewBeaconChainClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
	defer cancel()

	var pubKey []byte
	if pubKeyProvider, ok := account.(e2wtypes.AccountCompositePublicKeyProvider); ok {
		pubKey = pubKeyProvider.CompositePublicKey().Marshal()
	} else if pubKeyProvider, ok := account.(e2wtypes.AccountPublicKeyProvider); ok {
		pubKey = pubKeyProvider.PublicKey().Marshal()
	} else {
		return false, false, false, 0, 0, errors.New("Unable to obtain public key")
	}

	req := &ethpb.ValidatorPerformanceRequest{
		PublicKeys: [][]byte{pubKey},
	}
	res, err := beaconClient.GetValidatorPerformance(ctx, req)
	if err != nil {
		return false, false, false, 0, 0, err
	}
	if len(res.InclusionDistances) == 0 {
		return false, false, false, 0, 0, errors.New("unknown validator")
	}
	return res.CorrectlyVotedHead[0],
		res.CorrectlyVotedSource[0],
		res.CorrectlyVotedTarget[0],
		res.InclusionDistances[0],
		int64(res.BalancesAfterEpochTransition[0]) - int64(res.BalancesBeforeEpochTransition[0]),
		err
}

// FetchValidatorInfo fetches current validator info from the beacon node.
func FetchValidatorInfo(conn *grpc.ClientConn, account e2wtypes.Account) (*ethpb.ValidatorInfo, error) {
	if conn == nil {
		return nil, errors.New("no connection to beacon node")
	}
	beaconClient := ethpb.NewBeaconChainClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
	defer cancel()

	stream, err := beaconClient.StreamValidatorsInfo(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to contact beacon node")
	}

	var pubKey []byte
	if pubKeyProvider, ok := account.(e2wtypes.AccountCompositePublicKeyProvider); ok {
		pubKey = pubKeyProvider.CompositePublicKey().Marshal()
	} else if pubKeyProvider, ok := account.(e2wtypes.AccountPublicKeyProvider); ok {
		pubKey = pubKeyProvider.PublicKey().Marshal()
	} else {
		return nil, errors.New("Unable to obtain public key")
	}

	changeSet := &ethpb.ValidatorChangeSet{
		Action:     ethpb.SetAction_SET_VALIDATOR_KEYS,
		PublicKeys: [][]byte{pubKey},
	}
	err = stream.Send(changeSet)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send validator public key")
	}
	return stream.Recv()
}

// FetchChainInfo fetches current chain info from the beacon node.
func FetchChainInfo(conn *grpc.ClientConn) (*ethpb.ChainHead, error) {
	if conn == nil {
		return nil, errors.New("no connection to beacon node")
	}
	beaconClient := ethpb.NewBeaconChainClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
	defer cancel()

	return beaconClient.GetChainHead(ctx, &types.Empty{})
}

// FetchBlock fetches a block at a given slot from the beacon node.
func FetchBlock(conn *grpc.ClientConn, slot uint64) (*ethpb.SignedBeaconBlock, error) {
	if conn == nil {
		return nil, errors.New("no connection to beacon node")
	}
	beaconClient := ethpb.NewBeaconChainClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
	defer cancel()

	req := &ethpb.ListBlocksRequest{}
	if slot == 0 {
		req.QueryFilter = &ethpb.ListBlocksRequest_Genesis{Genesis: true}
	} else {
		req.QueryFilter = &ethpb.ListBlocksRequest_Slot{Slot: slot}
	}
	resp, err := beaconClient.ListBlocks(ctx, req)
	if err != nil {
		return nil, err
	}
	if len(resp.BlockContainers) == 0 {
		return nil, nil
	}
	return resp.BlockContainers[0].Block, nil
}

func StreamBlocks(conn *grpc.ClientConn) (ethpb.BeaconChain_StreamBlocksClient, error) {
	if conn == nil {
		return nil, errors.New("no connection to beacon node")
	}

	beaconClient := ethpb.NewBeaconChainClient(conn)
	stream, err := beaconClient.StreamBlocks(context.Background(), &types.Empty{})
	if err != nil {
		return nil, err
	}

	return stream, nil
}
