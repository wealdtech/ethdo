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

package attesterinclusion

import (
	"context"
	"testing"

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
			name:    "Empty",
			dataOut: &dataOut{},
			res:     "Attestation not found",
		},
		{
			name: "Found",
			dataOut: &dataOut{
				found:            true,
				slot:             123,
				attestationIndex: 456,
				inclusionDelay:   7,
			},
			res: `Attestation included in block 123, index 456`,
		},
		{
			name: "Verbose",
			dataOut: &dataOut{
				verbose:          true,
				found:            true,
				slot:             123,
				attestationIndex: 456,
				inclusionDelay:   7,
				headCorrect:      true,
				headTimely:       false,
				sourceTimely:     false,
				targetCorrect:    true,
				targetTimely:     true,
			},
			res: `Attestation included in block 123, index 456
Inclusion delay: 7
Head correct: ✓
Head timely: ✕
Source timely: ✕
Target correct: ✓
Target timely: ✓`,
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
