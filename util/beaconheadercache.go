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
	"context"
	"errors"
	"fmt"
	"net/http"

	eth2client "github.com/attestantio/go-eth2-client"
	"github.com/attestantio/go-eth2-client/api"
	apiv1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec/phase0"
)

// BeaconBlockHeaderCache is a cache of beacon block headers.
type BeaconBlockHeaderCache struct {
	beaconBlockHeadersProvider eth2client.BeaconBlockHeadersProvider
	entries                    map[phase0.Slot]*beaconBlockHeaderEntry
}

// NewBeaconBlockHeaderCache makes a new beacon block header cache.
func NewBeaconBlockHeaderCache(provider eth2client.BeaconBlockHeadersProvider) *BeaconBlockHeaderCache {
	return &BeaconBlockHeaderCache{
		beaconBlockHeadersProvider: provider,
		entries:                    make(map[phase0.Slot]*beaconBlockHeaderEntry),
	}
}

type beaconBlockHeaderEntry struct {
	present bool
	value   *apiv1.BeaconBlockHeader
}

// Fetch the beacon block header for the given slot.
func (b *BeaconBlockHeaderCache) Fetch(ctx context.Context,
	slot phase0.Slot,
) (
	*apiv1.BeaconBlockHeader,
	error,
) {
	entry, exists := b.entries[slot]
	if !exists {
		response, err := b.beaconBlockHeadersProvider.BeaconBlockHeader(ctx, &api.BeaconBlockHeaderOpts{Block: fmt.Sprintf("%d", slot)})
		if err != nil {
			var apiErr *api.Error
			if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
				entry = &beaconBlockHeaderEntry{
					present: false,
				}
			} else {
				return nil, err
			}
		} else {
			entry = &beaconBlockHeaderEntry{
				present: true,
				value:   response.Data,
			}
		}

		b.entries[slot] = entry
	}
	return entry.value, nil
}
