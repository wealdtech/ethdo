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

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	e2types "github.com/wealdtech/go-eth2-types/v2"
)

func TestLogLevel(t *testing.T) {
	require.NoError(t, e2types.InitBLS())

	tests := []struct {
		name  string
		input string
		level zerolog.Level
	}{
		{
			name:  "Unknown",
			input: "unknown",
			level: zerolog.DebugLevel,
		},
		{
			name:  "Disabled",
			input: "None",
			level: zerolog.Disabled,
		},
		{
			name:  "Trace",
			input: "Trace",
			level: zerolog.TraceLevel,
		},
		{
			name:  "Debug",
			input: "Debug",
			level: zerolog.DebugLevel,
		},
		{
			name:  "Info",
			input: "Info",
			level: zerolog.InfoLevel,
		},
		{
			name:  "Warn",
			input: "Warn",
			level: zerolog.WarnLevel,
		},
		{
			name:  "Error",
			input: "Error",
			level: zerolog.ErrorLevel,
		},
		{
			name:  "Fatal",
			input: "Fatal",
			level: zerolog.FatalLevel,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			level := logLevel(test.input)
			require.Equal(t, test.level, level)
		})
	}
}
