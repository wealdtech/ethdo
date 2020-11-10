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
	"encoding/json"

	spec "github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
	"github.com/wealdtech/ethdo/util"
)

type dataOut struct {
	jsonOutput          bool
	forkVersion         spec.Version
	signedVoluntaryExit *spec.SignedVoluntaryExit
}

func output(ctx context.Context, data *dataOut) (string, error) {
	if data == nil {
		return "", errors.New("no data")
	}

	if data.signedVoluntaryExit == nil {
		return "", errors.New("no signed voluntary exit")
	}

	if data.jsonOutput {
		return outputJSON(ctx, data)
	}

	return "", nil
}

func outputJSON(ctx context.Context, data *dataOut) (string, error) {
	validatorExitData := &util.ValidatorExitData{
		Data:        data.signedVoluntaryExit,
		ForkVersion: data.forkVersion,
	}
	bytes, err := json.Marshal(validatorExitData)
	if err != nil {
		return "", errors.Wrap(err, "failed to generate JSON")
	}
	return string(bytes), nil
}
