// Copyright © 2019 - 2022 Weald Technology Trading.
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

package attesterinclusion

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
)

type dataOut struct {
	debug            bool
	quiet            bool
	verbose          bool
	attestation      *phase0.Attestation
	slot             phase0.Slot
	attestationIndex uint64
	inclusionDelay   phase0.Slot
	found            bool
	headCorrect      bool
	headTimely       bool
	sourceTimely     bool
	targetCorrect    bool
	targetTimely     bool
}

func output(_ context.Context, data *dataOut) (string, error) {
	buf := strings.Builder{}
	if data == nil {
		return buf.String(), errors.New("no data")
	}

	if !data.quiet {
		if data.found {
			buf.WriteString("Attestation included in block ")
			buf.WriteString(fmt.Sprintf("%d", data.slot))
			buf.WriteString(", index ")
			buf.WriteString(strconv.FormatUint(data.attestationIndex, 10))
			if data.verbose {
				buf.WriteString("\nInclusion delay: ")
				buf.WriteString(fmt.Sprintf("%d", data.inclusionDelay))
				buf.WriteString("\nHead correct: ")
				if data.headCorrect {
					buf.WriteString("✓")
				} else {
					buf.WriteString("✕")
				}
				buf.WriteString("\nHead timely: ")
				if data.headTimely {
					buf.WriteString("✓")
				} else {
					buf.WriteString("✕")
				}
				buf.WriteString("\nSource timely: ")
				if data.sourceTimely {
					buf.WriteString("✓")
				} else {
					buf.WriteString("✕")
				}
				buf.WriteString("\nTarget correct: ")
				if data.targetCorrect {
					buf.WriteString("✓")
				} else {
					buf.WriteString("✕")
				}
				buf.WriteString("\nTarget timely: ")
				if data.targetTimely {
					buf.WriteString("✓")
				} else {
					buf.WriteString("✕")
				}
			}
		} else {
			buf.WriteString("Attestation not found")
		}
	}
	return buf.String(), nil
}
