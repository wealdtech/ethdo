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

package chaintime

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	spec "github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
)

type dataOut struct {
	debug   bool
	quiet   bool
	verbose bool

	epoch                         spec.Epoch
	epochStart                    time.Time
	epochEnd                      time.Time
	slot                          spec.Slot
	slotStart                     time.Time
	slotEnd                       time.Time
	hasSyncCommittees             bool
	syncCommitteePeriod           uint64
	syncCommitteePeriodStart      time.Time
	syncCommitteePeriodEpochStart spec.Epoch
	syncCommitteePeriodEnd        time.Time
	syncCommitteePeriodEpochEnd   spec.Epoch
}

func output(_ context.Context, data *dataOut) (string, error) {
	if data == nil {
		return "", errors.New("no data")
	}

	if data.quiet {
		return "", nil
	}

	builder := strings.Builder{}

	builder.WriteString("Epoch ")
	builder.WriteString(fmt.Sprintf("%d", data.epoch))
	builder.WriteString("\n  Epoch start ")
	builder.WriteString(data.epochStart.Format("2006-01-02 15:04:05"))
	if data.verbose {
		builder.WriteString(" (")
		builder.WriteString(strconv.FormatInt(data.epochStart.Unix(), 10))
		builder.WriteString(")")
	}
	builder.WriteString("\n  Epoch end ")
	builder.WriteString(data.epochEnd.Format("2006-01-02 15:04:05"))
	if data.verbose {
		builder.WriteString(" (")
		builder.WriteString(strconv.FormatInt(data.epochEnd.Unix(), 10))
		builder.WriteString(")")
	}

	builder.WriteString("\nSlot ")
	builder.WriteString(fmt.Sprintf("%d", data.slot))
	builder.WriteString("\n  Slot start ")
	builder.WriteString(data.slotStart.Format("2006-01-02 15:04:05"))
	if data.verbose {
		builder.WriteString(" (")
		builder.WriteString(strconv.FormatInt(data.slotStart.Unix(), 10))
		builder.WriteString(")")
	}
	builder.WriteString("\n  Slot end ")
	builder.WriteString(data.slotEnd.Format("2006-01-02 15:04:05"))
	if data.verbose {
		builder.WriteString(" (")
		builder.WriteString(strconv.FormatInt(data.slotEnd.Unix(), 10))
		builder.WriteString(")")
	}

	if data.hasSyncCommittees {
		builder.WriteString("\nSync committee period ")
		builder.WriteString(strconv.FormatUint(data.syncCommitteePeriod, 10))
		builder.WriteString("\n  Sync committee period start ")
		builder.WriteString(data.syncCommitteePeriodStart.Format("2006-01-02 15:04:05"))
		builder.WriteString(" (epoch ")
		builder.WriteString(fmt.Sprintf("%d", data.syncCommitteePeriodEpochStart))
		if data.verbose {
			builder.WriteString(", ")
			builder.WriteString(strconv.FormatInt(data.syncCommitteePeriodStart.Unix(), 10))
		}
		builder.WriteString(")\n  Sync committee period end ")
		builder.WriteString(data.syncCommitteePeriodEnd.Format("2006-01-02 15:04:05"))
		builder.WriteString(" (epoch ")
		builder.WriteString(fmt.Sprintf("%d", data.syncCommitteePeriodEpochEnd))
		if data.verbose {
			builder.WriteString(", ")
			builder.WriteString(strconv.FormatInt(data.syncCommitteePeriodEnd.Unix(), 10))
		}
		builder.WriteString(")")
	}

	builder.WriteString("\n")

	return builder.String(), nil
}
