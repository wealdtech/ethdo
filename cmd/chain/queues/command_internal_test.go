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

package chainqueues

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
			name: "Good",
			vars: map[string]interface{}{
				"timeout":    "5s",
				"connection": os.Getenv("ETHDO_TEST_CONNECTION"),
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
