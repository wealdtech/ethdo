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

package slottime

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/attestantio/go-eth2-client/auto"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestProcess(t *testing.T) {
	if os.Getenv("ETHDO_TEST_CONNECTION") == "" {
		t.Skip("ETHDO_TEST_CONNECTION not configured; cannot run tests")
	}
	eth2Client, err := auto.New(context.Background(),
		auto.WithLogLevel(zerolog.Disabled),
		auto.WithAddress(os.Getenv("ETHDO_TEST_CONNECTION")),
	)
	require.NoError(t, err)

	tests := []struct {
		name     string
		dataIn   *dataIn
		expected time.Time
		err      string
	}{
		{
			name: "Nil",
			err:  "no data",
		},
		{
			name: "Slot0",
			dataIn: &dataIn{
				eth2Client: eth2Client,
				slot:       "0",
			},
			expected: time.Unix(1606824023, 0),
		},
		{
			name: "Slot1",
			dataIn: &dataIn{
				eth2Client: eth2Client,
				slot:       "1",
			},
			expected: time.Unix(1606824035, 0),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := process(context.Background(), test.dataIn)
			if test.err != "" {
				require.EqualError(t, err, test.err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expected, res.startTime)
			}
		})
	}
}
