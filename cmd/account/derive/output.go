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
	"fmt"
	"strings"

	"github.com/pkg/errors"
	e2types "github.com/wealdtech/go-eth2-types/v2"
	util "github.com/wealdtech/go-eth2-util"
)

type dataOut struct {
	showPrivateKey            bool
	showWithdrawalCredentials bool
	key                       *e2types.BLSPrivateKey
}

func output(ctx context.Context, data *dataOut) (string, error) {
	if data == nil {
		return "", errors.New("no data")
	}
	if data.key == nil {
		return "", errors.New("no key")
	}

	builder := strings.Builder{}

	if data.showPrivateKey {
		builder.WriteString(fmt.Sprintf("Private key: %#x\n", data.key.Marshal()))
	}
	if data.showWithdrawalCredentials {
		withdrawalCredentials := util.SHA256(data.key.PublicKey().Marshal())
		withdrawalCredentials[0] = byte(0) // BLS_WITHDRAWAL_PREFIX
		builder.WriteString(fmt.Sprintf("Withdrawal credentials: %#x\n", withdrawalCredentials))
	}
	if !(data.showPrivateKey || data.showWithdrawalCredentials) {
		builder.WriteString(fmt.Sprintf("Public key: %#x\n", data.key.PublicKey().Marshal()))
	}

	return builder.String(), nil
}
