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

package validatorwithdrawl

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

//nolint:unparam
func (c *command) output(_ context.Context) (string, error) {
	if c.quiet {
		return "", nil
	}

	if c.json {
		data, err := json.Marshal(c.res)
		if err != nil {
			return "", errors.Wrap(err, "failed to marshal results")
		}
		return string(data), nil
	}

	return fmt.Sprintf("Withdrawal expected at %s in block %d", c.res.Expected.Format("2006-01-02T15:04:05"), c.res.Block), nil
}
