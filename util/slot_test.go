// Copyright © 2023 Weald Technology Trading.
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

package util_test

import (
	"context"
	"testing"
	"time"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	standardchaintime "github.com/wealdtech/ethdo/services/chaintime/standard"
	"github.com/wealdtech/ethdo/testing/mock"
	"github.com/wealdtech/ethdo/util"
)

func TestParseSlot(t *testing.T) {
	ctx := context.Background()

	// genesis is 1 day ago.
	genesisTime := time.Now().AddDate(0, 0, -1)
	slotDuration := 12 * time.Second
	slotsPerSlot := uint64(32)
	epochsPerSyncCommitteePeriod := uint64(256)
	mockGenesisProvider := mock.NewGenesisProvider(genesisTime)
	mockSpecProvider := mock.NewSpecProvider(slotDuration, slotsPerSlot, epochsPerSyncCommitteePeriod)
	chainTime, err := standardchaintime.New(context.Background(),
		standardchaintime.WithLogLevel(zerolog.Disabled),
		standardchaintime.WithGenesisProvider(mockGenesisProvider),
		standardchaintime.WithSpecProvider(mockSpecProvider),
	)
	require.NoError(t, err)

	tests := []struct {
		name     string
		input    string
		err      string
		expected phase0.Slot
	}{
		{
			name:     "Genesis",
			input:    "0",
			expected: 0,
		},
		{
			name:  "Invalid",
			input: "invalid",
			err:   `failed to parse slot: strconv.ParseInt: parsing "invalid": invalid syntax`,
		},
		{
			name:     "Absolute",
			input:    "15",
			expected: 15,
		},
		{
			name:     "Current",
			input:    "current",
			expected: 7200,
		},
		{
			name:     "Last",
			input:    "last",
			expected: 7199,
		},
		{
			name:     "RelativeZero",
			input:    "-0",
			expected: 7200,
		},
		{
			name:     "Relative",
			input:    "-5",
			expected: 7195,
		},
		{
			name:     "RelativeFar",
			input:    "-50000",
			expected: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			slot, err := util.ParseSlot(ctx, chainTime, test.input)
			if test.err != "" {
				require.EqualError(t, err, test.err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expected, slot)
			}
		})
	}
}
