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

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type dataIn struct {
	quiet bool
	// Derivation information.
	mnemonic string
	path     string
	// Output options.
	showKey bool
}

func input(ctx context.Context) (*dataIn, error) {
	data := &dataIn{}

	// Quiet.
	data.quiet = viper.GetBool("quiet")

	// Mnemonic.
	if viper.GetString("mnemonic") == "" {
		return nil, errors.New("mnemonic is required")
	}
	data.mnemonic = viper.GetString("mnemonic")

	// Path.
	if viper.GetString("path") == "" {
		return nil, errors.New("path is required")
	}
	data.path = viper.GetString("path")

	// Show key.
	data.showKey = viper.GetBool("show-key")

	return data, nil
}
