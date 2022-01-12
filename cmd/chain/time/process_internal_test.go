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

package chaintime

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestProcess(t *testing.T) {
	if os.Getenv("ETHDO_TEST_CONNECTION") == "" {
		t.Skip("ETHDO_TEST_CONNECTION not configured; cannot run tests")
	}

	tests := []struct {
		name     string
		dataIn   *dataIn
		expected *dataOut
		err      string
	}{
		{
			name: "Nil",
			err:  "no data",
		},
		{
			name: "Slot",
			dataIn: &dataIn{
				connection:               os.Getenv("ETHDO_TEST_CONNECTION"),
				timeout:                  10 * time.Second,
				allowInsecureConnections: true,
				slot:                     "1",
			},
			expected: &dataOut{
				epochStart:                    time.Unix(1606824023, 0),
				epochEnd:                      time.Unix(1606824407, 0),
				slot:                          1,
				slotStart:                     time.Unix(1606824035, 0),
				slotEnd:                       time.Unix(1606824047, 0),
				syncCommitteePeriod:           0,
				syncCommitteePeriodStart:      time.Unix(1606824023, 0),
				syncCommitteePeriodEnd:        time.Unix(1606921943, 0),
				syncCommitteePeriodEpochStart: 0,
				syncCommitteePeriodEpochEnd:   255,
			},
		},
		{
			name: "Epoch",
			dataIn: &dataIn{
				connection:               os.Getenv("ETHDO_TEST_CONNECTION"),
				timeout:                  10 * time.Second,
				allowInsecureConnections: true,
				epoch:                    "2",
			},
			expected: &dataOut{
				epoch:                         2,
				epochStart:                    time.Unix(1606824791, 0),
				epochEnd:                      time.Unix(1606825175, 0),
				slot:                          64,
				slotStart:                     time.Unix(1606824791, 0),
				slotEnd:                       time.Unix(1606824803, 0),
				syncCommitteePeriod:           0,
				syncCommitteePeriodStart:      time.Unix(1606824023, 0),
				syncCommitteePeriodEnd:        time.Unix(1606921943, 0),
				syncCommitteePeriodEpochStart: 0,
				syncCommitteePeriodEpochEnd:   255,
			},
		},
		{
			name: "Timestamp",
			dataIn: &dataIn{
				connection:               os.Getenv("ETHDO_TEST_CONNECTION"),
				timeout:                  10 * time.Second,
				allowInsecureConnections: true,
				timestamp:                "2021-01-01T00:00:00+0000",
			},
			expected: &dataOut{
				epoch:                         6862,
				epochStart:                    time.Unix(1609459031, 0),
				epochEnd:                      time.Unix(1609459415, 0),
				slot:                          219598,
				slotStart:                     time.Unix(1609459199, 0),
				slotEnd:                       time.Unix(1609459211, 0),
				syncCommitteePeriod:           26,
				syncCommitteePeriodStart:      time.Unix(1609379927, 0),
				syncCommitteePeriodEnd:        time.Unix(1609477847, 0),
				syncCommitteePeriodEpochStart: 6656,
				syncCommitteePeriodEpochEnd:   6911,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := process(context.Background(), test.dataIn)
			if test.err != "" {
				require.EqualError(t, err, test.err)
			} else {
				require.NoError(t, err)
				fmt.Printf("****** %d %d\n", res.syncCommitteePeriodStart.Unix(), res.syncCommitteePeriodEnd.Unix())
				require.Equal(t, test.expected, res)
			}
		})
	}
}
