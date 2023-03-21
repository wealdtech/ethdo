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

package validatorduties

import (
	"context"
	"fmt"
	"strings"
	"time"

	api "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/pkg/errors"
)

type dataOut struct {
	debug                   bool
	quiet                   bool
	verbose                 bool
	genesisTime             time.Time
	slotDuration            time.Duration
	slotsPerEpoch           uint64
	thisEpochAttesterDuty   *api.AttesterDuty
	thisEpochProposerDuties []*api.ProposerDuty
	nextEpochAttesterDuty   *api.AttesterDuty
}

func output(_ context.Context, data *dataOut) (string, error) {
	if data == nil {
		return "", errors.New("no data")
	}

	if data.quiet {
		return "", nil
	}

	builder := strings.Builder{}

	now := time.Now()
	builder.WriteString("Current time: ")
	builder.WriteString(now.Format("15:04:05\n"))

	if data.thisEpochAttesterDuty != nil {
		thisEpochAttesterSlot := data.thisEpochAttesterDuty.Slot
		thisSlotStart := data.genesisTime.Add(time.Duration(thisEpochAttesterSlot) * data.slotDuration)
		thisSlotEnd := thisSlotStart.Add(data.slotDuration)
		if thisSlotEnd.After(now) {
			builder.WriteString("Upcoming attestation slot this epoch: ")
			builder.WriteString(thisSlotStart.Format("15:04:05"))
			builder.WriteString(" - ")
			builder.WriteString(thisSlotEnd.Format("15:04:05 ("))
			until := thisSlotStart.Sub(now)
			if until > 0 {
				builder.WriteString(fmt.Sprintf("%ds until start of slot)\n", int(until.Seconds())))
			} else {
				builder.WriteString("\n")
			}
		}
	}

	for _, proposerDuty := range data.thisEpochProposerDuties {
		proposerSlot := proposerDuty.Slot
		proposerSlotStart := data.genesisTime.Add(time.Duration(proposerSlot) * data.slotDuration)
		proposerSlotEnd := proposerSlotStart.Add(data.slotDuration)
		builder.WriteString("Upcoming proposer slot this epoch: ")
		builder.WriteString(proposerSlotStart.Format("15:04:05"))
		builder.WriteString(" - ")
		builder.WriteString(proposerSlotEnd.Format("15:04:05 ("))
		until := proposerSlotStart.Sub(now)
		if until > 0 {
			builder.WriteString(fmt.Sprintf("%ds until start of slot)\n", int(until.Seconds())))
		} else {
			builder.WriteString("\n")
		}
	}

	if data.nextEpochAttesterDuty != nil {
		nextEpochAttesterSlot := data.nextEpochAttesterDuty.Slot
		nextSlotStart := data.genesisTime.Add(time.Duration(nextEpochAttesterSlot) * data.slotDuration)
		nextSlotEnd := nextSlotStart.Add(data.slotDuration)
		builder.WriteString("Upcoming attestation slot next epoch: ")
		builder.WriteString(nextSlotStart.Format("15:04:05"))
		builder.WriteString(" - ")
		builder.WriteString(nextSlotEnd.Format("15:04:05 ("))
		until := nextSlotStart.Sub(now)
		builder.WriteString(fmt.Sprintf("%ds until start of slot)\n", int(until.Seconds())))

		nextEpoch := uint64(data.nextEpochAttesterDuty.Slot) / data.slotsPerEpoch
		nextEpochStart := data.genesisTime.Add(time.Duration(nextEpoch*data.slotsPerEpoch) * data.slotDuration)
		builder.WriteString("Next epoch starts ")
		builder.WriteString(nextEpochStart.Format("15:04:05 ("))
		until = nextEpochStart.Sub(now)
		if until > 0 {
			builder.WriteString(fmt.Sprintf("%ds until start of epoch)\n", int(until.Seconds())))
		} else {
			builder.WriteString("\n")
		}
	}

	return builder.String(), nil
}
