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

	"github.com/stretchr/testify/require"
	"github.com/wealdtech/ethdo/util"
	e2types "github.com/wealdtech/go-eth2-types/v2"
)

func TestBLSID(t *testing.T) {
	require.NoError(t, e2types.InitBLS())

	tests := []struct {
		name   string
		in     uint64
		strRes string
	}{
		{
			name:   "Zero",
			in:     0,
			strRes: "0",
		},
		{
			name:   "One",
			in:     1,
			strRes: "1",
		},
		{
			name:   "High",
			in:     0x7fffffff,
			strRes: "2147483647",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			blsID := util.BLSID(test.in)
			require.Equal(t, test.strRes, blsID.GetDecString())
		})
	}
}
