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

package members

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
)

type dataOut struct {
	debug      bool
	quiet      bool
	verbose    bool
	json       bool
	validators []phase0.ValidatorIndex
}

func output(_ context.Context, data *dataOut) (string, error) {
	if data == nil {
		return "", errors.New("no data")
	}

	if data.quiet {
		return "", nil
	}

	if data.validators == nil {
		return "No sync committee validators found", nil
	}

	if data.json {
		bytes, err := json.Marshal(data.validators)
		if err != nil {
			return "", errors.Wrap(err, "failed to marshal JSON")
		}
		return string(bytes), nil
	}

	validators := make([]string, len(data.validators))
	for i := range data.validators {
		validators[i] = fmt.Sprintf("%d", data.validators[i])
	}

	return strings.Join(validators, ","), nil
}
