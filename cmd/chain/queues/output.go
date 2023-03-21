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
	"encoding/json"
	"fmt"
	"strings"
)

type jsonOutput struct {
	ActivationQueue int `json:"activation_queue"`
	ExitQueue       int `json:"exit_queue"`
}

func (c *command) output(ctx context.Context) (string, error) {
	if c.quiet {
		return "", nil
	}

	if c.json {
		return c.outputJSON(ctx)
	}
	return c.outputText(ctx)
}

func (c *command) outputJSON(_ context.Context) (string, error) {
	output := &jsonOutput{
		ActivationQueue: c.activationQueue,
		ExitQueue:       c.exitQueue,
	}
	data, err := json.Marshal(output)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (c *command) outputText(_ context.Context) (string, error) {
	builder := strings.Builder{}

	if c.activationQueue > 0 {
		builder.WriteString(fmt.Sprintf("Activation queue: %d\n", c.activationQueue))
	}
	if c.exitQueue > 0 {
		builder.WriteString(fmt.Sprintf("Exit queue: %d\n", c.exitQueue))
	}

	return strings.TrimSuffix(builder.String(), "\n"), nil
}
