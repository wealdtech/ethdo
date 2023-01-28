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

package util_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wealdtech/ethdo/util"
)

func TestUnmarshal(t *testing.T) {
	tests := []struct {
		name string
		in   []byte
		err  string
	}{
		{
			name: "Nil",
			err:  "unexpected end of JSON input",
		},
		{
			name: "Invalid",
			in:   []byte(`invalid`),
			err:  "invalid character 'i' looking for beginning of value",
		},
		{
			name: "ExitMissing",
			in:   []byte(`{"fork_version":"0x00000001"}`),
			err:  "exit missing",
		},
		{
			name: "ExitInvalid",
			in:   []byte(`{"exit":{},"fork_version":"0x00000001"}`),
			err:  "failed to unmarshal JSON: message missing",
		},
		{
			name: "ForkVersionMissing",
			in:   []byte(`{"exit":{"message":{"epoch":"0","validator_index":"0"},"signature":"0xb74eade64ebf1e02cc57e5d29517032c6ca99132fb8e7fb7e6d58c68713e581ef0ef88e2a6c599a007d997782abdd50b0f9763500a93a971c89cb2275583fe755d7c0e64f459ff22fcef5cab3f80848f0356e67c142b9cf3ee65613f56283d6e"}}`),
			err:  "fork version missing",
		},
		{
			name: "ForkVersionInvalid",
			in:   []byte(`{"exit":{"message":{"epoch":"0","validator_index":"0"},"signature":"0xb74eade64ebf1e02cc57e5d29517032c6ca99132fb8e7fb7e6d58c68713e581ef0ef88e2a6c599a007d997782abdd50b0f9763500a93a971c89cb2275583fe755d7c0e64f459ff22fcef5cab3f80848f0356e67c142b9cf3ee65613f56283d6e"},"fork_version":"invalid"}`),
			err:  "fork version invalid: encoding/hex: invalid byte: U+0069 'i'",
		},
		{
			name: "Good",
			in:   []byte(`{"exit":{"message":{"epoch":"0","validator_index":"0"},"signature":"0xb74eade64ebf1e02cc57e5d29517032c6ca99132fb8e7fb7e6d58c68713e581ef0ef88e2a6c599a007d997782abdd50b0f9763500a93a971c89cb2275583fe755d7c0e64f459ff22fcef5cab3f80848f0356e67c142b9cf3ee65613f56283d6e"},"fork_version":"0x00000001"}`),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var res util.ValidatorExitData
			err := json.Unmarshal(test.in, &res)
			if test.err != "" {
				require.EqualError(t, err, test.err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
