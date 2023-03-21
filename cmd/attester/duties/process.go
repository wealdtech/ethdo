// Copyright © 2021 Weald Technology Trading
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

package attesterduties

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	eth2client "github.com/attestantio/go-eth2-client"
	api "github.com/attestantio/go-eth2-client/api/v1"
	spec "github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
	"github.com/wealdtech/ethdo/util"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

func process(ctx context.Context, data *dataIn) (*dataOut, error) {
	if data == nil {
		return nil, errors.New("no data")
	}

	var account e2wtypes.Account
	var err error
	if data.account != "" {
		ctx, cancel := context.WithTimeout(ctx, data.timeout)
		defer cancel()
		_, account, err = util.WalletAndAccountFromPath(ctx, data.account)
		if err != nil {
			return nil, errors.Wrap(err, "failed to obtain account")
		}
	} else {
		pubKeyBytes, err := hex.DecodeString(strings.TrimPrefix(data.pubKey, "0x"))
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("failed to decode public key %s", data.pubKey))
		}
		account, err = util.NewScratchAccount(nil, pubKeyBytes)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("invalid public key %s", data.pubKey))
		}
	}

	// Fetch validator
	pubKeys := make([]spec.BLSPubKey, 1)
	pubKey, err := util.BestPublicKey(account)
	if err != nil {
		return nil, errors.Wrap(err, "failed to obtain public key for account")
	}
	copy(pubKeys[0][:], pubKey.Marshal())
	validators, err := data.eth2Client.(eth2client.ValidatorsProvider).ValidatorsByPubKey(ctx, fmt.Sprintf("%d", uint64(data.epoch)*data.slotsPerEpoch), pubKeys)
	if err != nil {
		return nil, errors.New("failed to obtain validator information")
	}
	if len(validators) == 0 {
		return nil, errors.New("validator is not known")
	}
	var validator *api.Validator
	for _, v := range validators {
		validator = v
	}

	results := &dataOut{
		debug:   data.debug,
		quiet:   data.quiet,
		verbose: data.verbose,
	}

	duty, err := duty(ctx, data.eth2Client, validator, data.epoch)
	if err != nil {
		return nil, errors.Wrap(err, "failed to obtain duty for validator")
	}

	results.duty = duty

	return results, nil
}

func duty(ctx context.Context, eth2Client eth2client.Service, validator *api.Validator, epoch spec.Epoch) (*api.AttesterDuty, error) {
	// Find the attesting slot for the given epoch.
	duties, err := eth2Client.(eth2client.AttesterDutiesProvider).AttesterDuties(ctx, epoch, []spec.ValidatorIndex{validator.Index})
	if err != nil {
		return nil, errors.Wrap(err, "failed to obtain attester duties")
	}

	if len(duties) == 0 {
		return nil, errors.New("validator does not have duty for that epoch")
	}

	return duties[0], nil
}
