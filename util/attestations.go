// Copyright © 2022 Weald Technology Trading.
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
	"bytes"
	"context"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/wealdtech/ethdo/services/chaintime"
)

// AttestationHeadCorrect returns true if the given attestation had the correct head.
func AttestationHeadCorrect(ctx context.Context,
	headersCache *BeaconBlockHeaderCache,
	attestation *phase0.Attestation,
) (
	bool,
	error,
) {
	slot := attestation.Data.Slot
	for {
		header, err := headersCache.Fetch(ctx, slot)
		if err != nil {
			return false, err
		}
		if header == nil {
			// No block.
			slot--
			continue
		}
		if !header.Canonical {
			// Not canonical.
			slot--
			continue
		}
		return bytes.Equal(header.Root[:], attestation.Data.BeaconBlockRoot[:]), nil
	}
}

// AttestationTargetCorrect returns true if the given attestation had the correct target.
func AttestationTargetCorrect(ctx context.Context,
	headersCache *BeaconBlockHeaderCache,
	chainTime chaintime.Service,
	attestation *phase0.Attestation,
) (
	bool,
	error,
) {
	// Start with first slot of the target epoch.
	slot := chainTime.FirstSlotOfEpoch(attestation.Data.Target.Epoch)
	for {
		header, err := headersCache.Fetch(ctx, slot)
		if err != nil {
			return false, err
		}
		if header == nil {
			// No block.
			slot--
			continue
		}
		if !header.Canonical {
			// Not canonical.
			slot--
			continue
		}
		return bytes.Equal(header.Root[:], attestation.Data.Target.Root[:]), nil
	}
}
