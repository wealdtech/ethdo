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

package accountimport

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

type dataOut struct {
	account e2wtypes.Account
}

func output(_ context.Context, data *dataOut) (string, error) {
	if data == nil {
		return "", errors.New("no data")
	}
	if data.account == nil {
		return "", errors.New("no account")
	}

	if pubKeyProvider, ok := data.account.(e2wtypes.AccountCompositePublicKeyProvider); ok {
		return fmt.Sprintf("%#x", pubKeyProvider.CompositePublicKey().Marshal()), nil
	}

	if pubKeyProvider, ok := data.account.(e2wtypes.AccountPublicKeyProvider); ok {
		return fmt.Sprintf("%#x", pubKeyProvider.PublicKey().Marshal()), nil
	}

	return "", errors.New("no public key available")
}
