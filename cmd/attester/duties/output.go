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
	"encoding/json"
	"fmt"

	api "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/pkg/errors"
)

type dataOut struct {
	debug   bool
	quiet   bool
	verbose bool
	json    bool
	duty    *api.AttesterDuty
}

func output(_ context.Context, data *dataOut) (string, error) {
	if data == nil {
		return "", errors.New("no data")
	}

	if data.quiet {
		return "", nil
	}

	if data.duty == nil {
		return "No duties found", nil
	}

	if data.json {
		bytes, err := json.Marshal(data.duty)
		if err != nil {
			return "", errors.Wrap(err, "failed to marshalJSON")
		}
		return string(bytes), nil
	}

	return fmt.Sprintf("Validator attesting in slot %d committee %d", data.duty.Slot, data.duty.CommitteeIndex), nil
}
