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

func TestAcceptablePassphrase(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		allowWeak bool
		expected  bool
	}{
		{
			name:     "Empty",
			input:    ``,
			expected: false,
		},
		{
			name:     "Simple",
			input:    `password`,
			expected: false,
		},
		{
			name:      "AllowedWeak",
			input:     `password`,
			allowWeak: true,
			expected:  true,
		},
		{
			name:     "Complex",
			input:    `Hu[J"yKH{z&-;[]'7T*Dm1:t`,
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			viper.Set("allow-weak-passphrases", test.allowWeak)
			require.Equal(t, test.expected, util.AcceptablePassphrase(test.input))
		})
	}
}
