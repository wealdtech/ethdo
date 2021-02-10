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

// import (
// 	"context"
// 	"testing"
//
// 	api "github.com/attestantio/go-eth2-client/api/v1"
// 	"github.com/stretchr/testify/require"
// 	"github.com/wealdtech/ethdo/testutil"
// )
//
// func TestOutput(t *testing.T) {
// 	tests := []struct {
// 		name    string
// 		dataOut *dataOut
// 		res     string
// 		err     string
// 	}{
// 		{
// 			name: "Nil",
// 			err:  "no data",
// 		},
// 		{
// 			name:    "Empty",
// 			dataOut: &dataOut{},
// 			res:     "No duties found",
// 		},
// 		{
// 			name: "Present",
// 			dataOut: &dataOut{
// 				duty: &api.AttesterDuty{
// 					PubKey:                  testutil.HexToPubKey("0x933ad9491b62059dd065b560d256d8957a8c402cc6e8d8ee7290ae11e8f7329267a8811c397529dac52ae1342ba58c95"),
// 					Slot:                    1,
// 					ValidatorIndex:          2,
// 					CommitteeIndex:          3,
// 					CommitteeLength:         4,
// 					CommitteesAtSlot:        5,
// 					ValidatorCommitteeIndex: 6,
// 				},
// 			},
// 			res: "Validator attesting in slot 1 committee 3",
// 		},
// 		{
// 			name: "JSON",
// 			dataOut: &dataOut{
// 				json: true,
// 				duty: &api.AttesterDuty{
// 					PubKey:                  testutil.HexToPubKey("0x933ad9491b62059dd065b560d256d8957a8c402cc6e8d8ee7290ae11e8f7329267a8811c397529dac52ae1342ba58c95"),
// 					Slot:                    1,
// 					ValidatorIndex:          2,
// 					CommitteeIndex:          3,
// 					CommitteeLength:         4,
// 					CommitteesAtSlot:        5,
// 					ValidatorCommitteeIndex: 6,
// 				},
// 			},
// 			res: `{"pubkey":"0x933ad9491b62059dd065b560d256d8957a8c402cc6e8d8ee7290ae11e8f7329267a8811c397529dac52ae1342ba58c95","slot":"1","validator_index":"2","committee_index":"3","committee_length":"4","committees_at_slot":"5","validator_committee_index":"6"}`,
// 		},
// 	}
//
// 	for _, test := range tests {
// 		t.Run(test.name, func(t *testing.T) {
// 			res, err := output(context.Background(), test.dataOut)
// 			if test.err != "" {
// 				require.EqualError(t, err, test.err)
// 			} else {
// 				require.NoError(t, err)
// 				require.Equal(t, test.res, res)
// 			}
// 		})
// 	}
// }
