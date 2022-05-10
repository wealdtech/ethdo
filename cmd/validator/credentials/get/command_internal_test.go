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

package validatorcredentialsget

import (
	"context"
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestInput(t *testing.T) {
	if os.Getenv("ETHDO_TEST_CONNECTION") == "" {
		t.Skip("ETHDO_TEST_CONNECTION not configured; cannot run tests")
	}

	tests := []struct {
		name string
		vars map[string]interface{}
		err  string
	}{
		{
			name: "TimeoutMissing",
			vars: map[string]interface{}{},
			err:  "timeout is required",
		},
		{
			name: "ConnectionMissing",
			vars: map[string]interface{}{
				"timeout": "5s",
				"index":   "1",
			},
			err: "connection is required",
		},
		{
			name: "NoValidatorInfo",
			vars: map[string]interface{}{
				"timeout":    "5s",
				"connection": os.Getenv("ETHDO_TEST_CONNECTION"),
			},
			err: "one of account, index or pubkey required",
		},
		{
			name: "MultipleValidatorInfo",
			vars: map[string]interface{}{
				"timeout":    "5s",
				"connection": os.Getenv("ETHDO_TEST_CONNECTION"),
				"index":      "1",
				"pubkey":     "0x000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
			},
			err: "only one of account, index and pubkey allowed",
		},
		{
			name: "Good",
			vars: map[string]interface{}{
				"timeout":    "5s",
				"connection": os.Getenv("ETHDO_TEST_CONNECTION"),
				"index":      "1",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			viper.Reset()

			for k, v := range test.vars {
				viper.Set(k, v)
			}
			_, err := newCommand(context.Background())
			if test.err != "" {
				require.EqualError(t, err, test.err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
