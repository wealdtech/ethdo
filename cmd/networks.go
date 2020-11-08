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

package cmd

import (
	"fmt"

	"github.com/wealdtech/ethdo/grpc"
)

// networks is a map of deposit contract addresses to networks.
var networks = map[string]string{
	"16e82d77882a663454ef92806b7deca1d394810f": "Altona",
	"0f0f0fc0530007361933eab5db97d09acdd6c1c8": "Onyx",
	"07b39f4fde4a38bace212b546dac87c58dfe3fdc": "Medalla",
}

// network returns the name of the network, if known.
func network() string {
	if err := connect(); err != nil {
		return "Unknown"
	}

	depositContractAddress, err := grpc.FetchDepositContractAddress(eth2GRPCConn)
	if err != nil {
		return "Unknown"
	}
	outputIf(debug, fmt.Sprintf("Deposit contract is %x", depositContractAddress))

	depositContract := fmt.Sprintf("%x", depositContractAddress)
	if network, exists := networks[depositContract]; exists {
		return network
	}
	return "Unknown"
}
