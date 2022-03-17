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
func ParseEpoch(ctx context.Context, chainTime chaintime.Service, epochStr string) (phase0.Epoch, error) {
	switch epochStr {
	case "", "current":
		return chainTime.CurrentEpoch(), nil
	case "last":
		return chainTime.CurrentEpoch() - 1, nil
	default:
		val, err := strconv.ParseInt(epochStr, 10, 64)
		if err != nil {
			return 0, errors.Wrap(err, "failed to parse epoch")
		}
		if val >= 0 {
			return phase0.Epoch(val), nil
		}
		return chainTime.CurrentEpoch() + phase0.Epoch(val), nil
	}
}
