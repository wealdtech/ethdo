// Copyright © 2019, 2020 Weald Technology Trading
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

package nodeevents

import (
	"context"
	"os"
	"testing"

	"github.com/attestantio/go-eth2-client/auto"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestProcess(t *testing.T) {
	if os.Getenv("ETHDO_TEST_CONNECTION") == "" {
		t.Skip("ETHDO_TEST_CONNECTION not configured; cannot run tests")
	}
	os.Setenv("ETHDO_ALLOW_INSECURE_CONNECTIONS", "true")

	eth2Client, err := auto.New(context.Background(),
		auto.WithLogLevel(zerolog.Disabled),
		auto.WithAddress(os.Getenv("ETHDO_TEST_CONNECTION")),
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
			name: "TopicsNil",
			dataIn: &dataIn{
				eth2Client: eth2Client,
			},
			err: "failed to connect for events: no topics supplied",
		},
		{
			name: "TopicsUnknown",
			dataIn: &dataIn{
				eth2Client: eth2Client,
				topics:     []string{"foo"},
			},
			err: "failed to connect for events: unsupported event topic foo",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := process(context.Background(), test.dataIn)
			if test.err != "" {
				require.EqualError(t, err, test.err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
