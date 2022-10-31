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
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOutput(t *testing.T) {
	tests := []struct {
		name    string
		dataOut *dataOut
		res     string
		err     string
	}{
		{
			name: "Nil",
			err:  "no data",
		},
		{
			name: "Normal",
			dataOut: &dataOut{
				startTime: time.Unix(1606824023, 0),
			},
			res: "2020-12-01 12:00:23 +0000 UTC",
		},
		{
			name: "Verbose",
			dataOut: &dataOut{
				startTime: time.Unix(1606824023, 0),
				endTime:   time.Unix(1606824035, 0),
				verbose:   true,
			},
			res: "2020-12-01 12:00:23 +0000 UTC - 2020-12-01 12:00:35 +0000 UTC",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := output(context.Background(), test.dataOut)
			if test.err != "" {
				require.EqualError(t, err, test.err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.res, res)
			}
		})
	}
}
