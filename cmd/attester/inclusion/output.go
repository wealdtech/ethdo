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

package attesterinclusion

import (
	"context"
	"fmt"

	spec "github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
)

type dataOut struct {
	debug            bool
	quiet            bool
	verbose          bool
	slot             spec.Slot
	attestationIndex uint64
	inclusionDelay   spec.Slot
	found            bool
}

func output(ctx context.Context, data *dataOut) (string, error) {
	if data == nil {
		return "", errors.New("no data")
	}

	if !data.quiet {
		if data.found {
			return fmt.Sprintf("Attestation included in block %d, attestation %d (inclusion delay %d)", data.slot, data.attestationIndex, data.inclusionDelay), nil
		}
		return "Attestation not found", nil
	}
	return "", nil
}
