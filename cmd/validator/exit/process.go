// Copyright © 2019, 2020 Weald Technology Trading
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

	eth2client "github.com/attestantio/go-eth2-client"
	api "github.com/attestantio/go-eth2-client/api/v1"
	spec "github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
	"github.com/wealdtech/ethdo/core"
	"github.com/wealdtech/ethdo/signing"
)

// maxFutureEpochs is the farthest in the future for which an exit will be created.
var maxFutureEpochs = spec.Epoch(1024)

func process(ctx context.Context, data *dataIn) (*dataOut, error) {
	if data == nil {
		return nil, errors.New("no data")
	}

	if data.epoch > data.currentEpoch {
		if data.epoch-data.currentEpoch > maxFutureEpochs {
			return nil, errors.New("not generating exit for an epoch in the far future")
		}
	}
	results := &dataOut{
		forkVersion: data.fork.CurrentVersion,
		jsonOutput:  data.jsonOutput,
	}

	validator, err := fetchValidator(ctx, data)
	if err != nil {
		return nil, err
	}

	exit, err := generateExit(ctx, data, validator)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate voluntary exit")
	}
	root, err := exit.HashTreeRoot()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate root for voluntary exit")
	}

	if data.account != nil {
		signature, err := signing.SignRoot(ctx, data.account, data.passphrases, root, data.domain)
		if err != nil {
			return nil, errors.Wrap(err, "failed to sign voluntary exit")
		}

		results.signedVoluntaryExit = &spec.SignedVoluntaryExit{
			Message:   exit,
			Signature: signature,
		}
	} else {
		results.signedVoluntaryExit = data.signedVoluntaryExit
	}

	if !data.jsonOutput {
		if err := broadcastExit(ctx, data, results); err != nil {
			return nil, errors.Wrap(err, "failed to broadcast voluntary exit")
		}
	}

	return results, nil
}

func generateExit(ctx context.Context, data *dataIn, validator *api.Validator) (*spec.VoluntaryExit, error) {
	if data == nil {
		return nil, errors.New("no data")
	}

	if data.signedVoluntaryExit != nil {
		return data.signedVoluntaryExit.Message, nil
	}

	if validator == nil {
		return nil, errors.New("no validator")
	}

	exit := &spec.VoluntaryExit{
		Epoch:          data.epoch,
		ValidatorIndex: validator.Index,
	}
	return exit, nil
}

func broadcastExit(ctx context.Context, data *dataIn, results *dataOut) error {
	return data.eth2Client.(eth2client.VoluntaryExitSubmitter).SubmitVoluntaryExit(ctx, results.signedVoluntaryExit)
}

func fetchValidator(ctx context.Context, data *dataIn) (*api.Validator, error) {
	// Validator.
	if data.account == nil {
		return nil, nil
	}

	var validator *api.Validator
	validatorPubKeys := make([]spec.BLSPubKey, 1)
	pubKey, err := core.BestPublicKey(data.account)
	if err != nil {
		return nil, errors.Wrap(err, "failed to obtain public key for account")
	}
	copy(validatorPubKeys[0][:], pubKey.Marshal())
	validators, err := data.eth2Client.(eth2client.ValidatorsProvider).ValidatorsByPubKey(ctx, "head", validatorPubKeys)
	if err != nil {
		return nil, errors.Wrap(err, "failed to obtain validator from beacon node")
	}
	if len(validators) == 0 {
		return nil, errors.New("validator not known by beacon node")
	}
	for _, v := range validators {
		validator = v
	}
	if validator.Status != api.ValidatorStateActiveOngoing {
		return nil, errors.New("validator is not active; cannot exit")
	}
	return validator, nil
}
