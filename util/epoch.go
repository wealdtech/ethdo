// Copyright © 2022 Weald Technology Trading
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

package util

import (
	"context"
	"strconv"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
	"github.com/wealdtech/ethdo/services/chaintime"
)

// ParseEpoch parses input to calculate the desired epoch.
func ParseEpoch(_ context.Context, chainTime chaintime.Service, epochStr string) (phase0.Epoch, error) {
	currentEpoch := chainTime.CurrentEpoch()
	switch epochStr {
	case "", "current", "head", "-0":
		return currentEpoch, nil
	case "last":
		if currentEpoch > 0 {
			currentEpoch--
		}
		return currentEpoch, nil
	default:
		val, err := strconv.ParseInt(epochStr, 10, 64)
		if err != nil {
			return 0, errors.Wrap(err, "failed to parse epoch")
		}
		if val >= 0 {
			return phase0.Epoch(val), nil
		}
		if phase0.Epoch(-val) > currentEpoch {
			return 0, nil
		}
		return currentEpoch + phase0.Epoch(val), nil
	}
}
