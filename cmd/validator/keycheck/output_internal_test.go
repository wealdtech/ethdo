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

package validatorkeycheck

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOutput(t *testing.T) {
	tests := []struct {
		name     string
		dataOut  *dataOut
		exitCode int
		expected []string
		err      string
	}{
		{
			name: "Nil",
			err:  "no data",
		},
		{
			name: "Not found",
			dataOut: &dataOut{
				match: false,
			},
			exitCode: 1,
			expected: []string{
				"Could not confirm withdrawal credentials with given information",
			},
		},
		{
			name: "Found",
			dataOut: &dataOut{
				match: true,
			},
			expected: []string{
				"Withdrawal credentials confirmed",
			},
		},
		{
			name: "FoundWithPath",
			dataOut: &dataOut{
				match: true,
				path:  "m/12381/3600/10/0",
			},
			expected: []string{
				"Withdrawal credentials confirmed at path m/12381/3600/10/0",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, exitCode, err := output(context.Background(), test.dataOut)
			if test.err != "" {
				require.EqualError(t, err, test.err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.exitCode, exitCode)
				for _, expected := range test.expected {
					require.True(t, strings.Contains(res, expected))
				}
			}
		})
	}
}
