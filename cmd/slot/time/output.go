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

package slottime

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
)

type dataOut struct {
	debug     bool
	quiet     bool
	verbose   bool
	startTime time.Time
	endTime   time.Time
}

func output(_ context.Context, data *dataOut) (string, error) {
	if data == nil {
		return "", errors.New("no data")
	}

	if data.quiet {
		return "", nil
	}
	if data.verbose {
		return fmt.Sprintf("%s - %s", data.startTime, data.endTime), nil
	}
	return data.startTime.String(), nil
}
