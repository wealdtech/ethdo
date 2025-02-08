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

	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
	"github.com/wealdtech/ethdo/services/chaintime"
)

// AttestationHead returns the head for which the attestation should have voted.
func AttestationHead(ctx context.Context,
	headersCache *BeaconBlockHeaderCache,
	attestation *spec.VersionedAttestation,
) (
	phase0.Root,
	error,
) {
	attestationData, err := attestation.Data()
	if err != nil {
		return phase0.Root{}, errors.Wrap(err, "failed to obtain attestation data")
	}

	slot := attestationData.Slot
	for {
		header, err := headersCache.Fetch(ctx, slot)
		if err != nil {
			return phase0.Root{}, err
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

		return header.Root, nil
	}
}

// AttestationHeadCorrect returns true if the given attestation had the correct head.
func AttestationHeadCorrect(ctx context.Context,
	headersCache *BeaconBlockHeaderCache,
	attestation *spec.VersionedAttestation,
) (
	bool,
	error,
) {
	attestationData, err := attestation.Data()
	if err != nil {
		return false, errors.Wrap(err, "failed to obtain attestation data")
	}

	slot := attestationData.Slot
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

		return bytes.Equal(header.Root[:], attestationData.BeaconBlockRoot[:]), nil
	}
}

// AttestationTarget returns the target for which the attestation should have voted.
func AttestationTarget(ctx context.Context,
	headersCache *BeaconBlockHeaderCache,
	chainTime chaintime.Service,
	attestation *spec.VersionedAttestation,
) (
	phase0.Root,
	error,
) {
	attestationData, err := attestation.Data()
	if err != nil {
		return phase0.Root{}, errors.Wrap(err, "failed to obtain attestation data")
	}

	// Start with first slot of the target epoch.
	slot := chainTime.FirstSlotOfEpoch(attestationData.Target.Epoch)
	for {
		header, err := headersCache.Fetch(ctx, slot)
		if err != nil {
			return phase0.Root{}, err
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

		return header.Root, nil
	}
}

// AttestationTargetCorrect returns true if the given attestation had the correct target.
func AttestationTargetCorrect(ctx context.Context,
	headersCache *BeaconBlockHeaderCache,
	chainTime chaintime.Service,
	attestation *spec.VersionedAttestation,
) (
	bool,
	error,
) {
	attestationData, err := attestation.Data()
	if err != nil {
		return false, errors.Wrap(err, "failed to obtain attestation data")
	}

	// Start with first slot of the target epoch.
	slot := chainTime.FirstSlotOfEpoch(attestationData.Target.Epoch)
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

		return bytes.Equal(header.Root[:], attestationData.Target.Root[:]), nil
	}
}
