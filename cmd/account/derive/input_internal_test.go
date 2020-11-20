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

package accountderive

import (
	"context"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	e2types "github.com/wealdtech/go-eth2-types/v2"
)

func TestInput(t *testing.T) {
	require.NoError(t, e2types.InitBLS())

	tests := []struct {
		name string
		vars map[string]interface{}
		res  *dataIn
		err  string
	}{
		{
			name: "MnemonicMissing",
			vars: map[string]interface{}{
				"path": "m/12381/3600/0/0",
			},
			err: "mnemonic is required",
		},
		{
			name: "PathMissing",
			vars: map[string]interface{}{
				"mnemonic": "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art",
			},
			err: "path is required",
		},
		{
			name: "Good",
			vars: map[string]interface{}{
				"mnemonic": "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art",
				"path":     "m/12381/3600/0/0",
			},
			res: &dataIn{
				mnemonic: "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art",
				path:     "m/12381/3600/0/0",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			viper.Reset()
			for k, v := range test.vars {
				viper.Set(k, v)
			}
			res, err := input(context.Background())
			if test.err != "" {
				require.EqualError(t, err, test.err)
			} else {
				require.NoError(t, err)
				// Cannot compare accounts directly, so need to check each element individually.
				require.Equal(t, test.res.mnemonic, res.mnemonic)
				require.Equal(t, test.res.path, res.path)
			}
		})
	}
}
