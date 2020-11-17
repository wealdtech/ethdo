// Copyright © 2020 Weald Technology Trading
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
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wealdtech/ethdo/testutil"
)

func TestNetworksInternal(t *testing.T) {
	tests := []struct {
		name     string
		address  []byte
		expected string
	}{
		{
			name:     "Empty",
			expected: "Unknown",
		},
		{
			name:     "Unknown",
			address:  testutil.HexToBytes("0000000000000000000000000000000000000000"),
			expected: "Unknown",
		},
		{
			name:     "Mainnet",
			address:  testutil.HexToBytes("00000000219ab540356cbb839cbe05303d7705fa"),
			expected: "Mainnet",
		},
		{
			name:     "Pyrmont",
			address:  testutil.HexToBytes("8c5fecdc472e27bc447696f431e425d02dd46a8c"),
			expected: "Pyrmont",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.expected, network(test.address))
		})
	}
}
