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

package validatorduties

import (
	"context"
	"strings"
	"testing"
	"time"

	api "github.com/attestantio/go-eth2-client/api/v1"
	spec "github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/stretchr/testify/require"
)

func TestOutput(t *testing.T) {
	tests := []struct {
		name     string
		dataOut  *dataOut
		expected []string
		err      string
	}{
		{
			name: "Nil",
			err:  "no data",
		},
		{
			name:     "Empty",
			dataOut:  &dataOut{},
			expected: []string{"Current time"},
		},
		{
			name: "Found",
			dataOut: &dataOut{
				genesisTime:   time.Unix(16000000000, 0),
				slotDuration:  12 * time.Second,
				slotsPerEpoch: 32,
				thisEpochAttesterDuty: &api.AttesterDuty{
					Slot: spec.Slot(1),
				},
				thisEpochProposerDuties: []*api.ProposerDuty{
					{
						Slot: spec.Slot(2),
					},
				},
				nextEpochAttesterDuty: &api.AttesterDuty{
					Slot: spec.Slot(40),
				},
			},
			expected: []string{
				"Current time",
				"Upcoming attestation slot this epoch",
				"Upcoming proposer slot this epoch",
				"Upcoming attestation slot next epoch",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := output(context.Background(), test.dataOut)
			if test.err != "" {
				require.EqualError(t, err, test.err)
			} else {
				require.NoError(t, err)
				for _, expected := range test.expected {
					require.True(t, strings.Contains(res, expected))
				}
			}
		})
	}
}
