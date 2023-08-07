// Copyright © 2023 Weald Technology Trading.
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

package walletbatch

import (
	"context"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Run runs the command.
func Run(cmd *cobra.Command) (string, error) {
	ctx := context.Background()

	c, err := newCommand(ctx)
	if err != nil {
		return "", errors.Wrap(err, "failed to set up command")
	}

	// Further errors do not need a usage report.
	cmd.SilenceUsage = true

	if err := c.process(ctx); err != nil {
		return "", errors.Wrap(err, "failed to process")
	}

	if viper.GetBool("quiet") {
		return "", nil
	}

	results, err := c.output(ctx)
	if err != nil {
		return "", errors.Wrap(err, "failed to obtain output")
	}

	return results, nil
}
