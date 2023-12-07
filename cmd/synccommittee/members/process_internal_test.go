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

package members

import (
	"context"
	"os"
	"testing"

	eth2client "github.com/attestantio/go-eth2-client"
	"github.com/attestantio/go-eth2-client/auto"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	standardchaintime "github.com/wealdtech/ethdo/services/chaintime/standard"
)

func TestProcess(t *testing.T) {
	if os.Getenv("ETHDO_TEST_CONNECTION") == "" {
		t.Skip("ETHDO_TEST_CONNECTION not configured; cannot run tests")
	}
	eth2Client, err := auto.New(context.Background(),
		auto.WithLogLevel(zerolog.Disabled),
		auto.WithAddress(os.Getenv("ETHDO_TEST_CONNECTION")),
	)
	require.NoError(t, err)

	chainTime, err := standardchaintime.New(context.Background(),
		standardchaintime.WithGenesisProvider(eth2Client.(eth2client.GenesisProvider)),
		standardchaintime.WithSpecProvider(eth2Client.(eth2client.SpecProvider)),
	)
	require.NoError(t, err)

	tests := []struct {
		name   string
		dataIn *dataIn
		err    string
	}{
		{
			name: "Nil",
			err:  "no data",
		},
		{
			name: "Good",
			dataIn: &dataIn{
				eth2Client: eth2Client,
				chainTime:  chainTime,
				epoch:      "-1",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := process(context.Background(), test.dataIn)
			if test.err != "" {
				require.EqualError(t, err, test.err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
