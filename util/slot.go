// Copyright © 2023 Weald Technology Trading
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

// ParseSlot parses input to calculate the desired slot.
func ParseSlot(_ context.Context, chainTime chaintime.Service, slotStr string) (phase0.Slot, error) {
	currentSlot := chainTime.CurrentSlot()
	switch slotStr {
	case "", "current", "head", "-0":
		return currentSlot, nil
	case "last":
		if currentSlot > 0 {
			currentSlot--
		}
		return currentSlot, nil
	default:
		val, err := strconv.ParseInt(slotStr, 10, 64)
		if err != nil {
			return 0, errors.Wrap(err, "failed to parse slot")
		}
		if val >= 0 {
			return phase0.Slot(val), nil
		}
		if phase0.Slot(-val) > currentSlot {
			return 0, nil
		}
		return currentSlot + phase0.Slot(val), nil
	}
}
