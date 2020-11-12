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
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"github.com/wealdtech/ethdo/util"
)

func TestBaseDir(t *testing.T) {
	tests := []struct {
		name     string
		inputs   map[string]interface{}
		expected string
	}{
		{
			name: "Nil",
		},
		{
			name: "Current",
			inputs: map[string]interface{}{
				"base-dir": "/tmp",
			},
			expected: "/tmp",
		},
		{
			name: "Deprecated",
			inputs: map[string]interface{}{
				"basedir": "/tmp",
			},
			expected: "/tmp",
		},
		{
			name: "Override",
			inputs: map[string]interface{}{
				"basedir":  "/tmp",
				"base-dir": "/tmp2",
			},
			expected: "/tmp2",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			viper.Reset()
			for k, v := range test.inputs {
				viper.Set(k, v)
			}
			res := util.GetBaseDir()
			require.Equal(t, test.expected, res)
		})
	}
}
