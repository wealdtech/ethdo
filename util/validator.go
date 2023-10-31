// Copyright © 2022 Weald Technology Trading
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

package util

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	consensusclient "github.com/attestantio/go-eth2-client"
	"github.com/attestantio/go-eth2-client/api"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

// ValidatorIndex obtains the index of a validator.
func ValidatorIndex(ctx context.Context, client consensusclient.Service, account string, pubKey string, index string) (phase0.ValidatorIndex, error) {
	switch {
	case account != "":
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, account, err := WalletAndAccountFromPath(ctx, account)
		if err != nil {
			return 0, errors.Wrap(err, "failed to obtain account")
		}
		return accountToIndex(ctx, account, client)
	case pubKey != "":
		pubKeyBytes, err := hex.DecodeString(strings.TrimPrefix(pubKey, "0x"))
		if err != nil {
			return 0, errors.Wrap(err, fmt.Sprintf("failed to decode public key %s", pubKey))
		}
		account, err := NewScratchAccount(nil, pubKeyBytes)
		if err != nil {
			return 0, errors.Wrap(err, fmt.Sprintf("invalid public key %s", pubKey))
		}
		return accountToIndex(ctx, account, client)
	case index != "":
		val, err := strconv.ParseUint(index, 10, 64)
		if err != nil {
			return 0, err
		}
		return phase0.ValidatorIndex(val), nil
	default:
		return 0, errors.New("no validator")
	}
}

func accountToIndex(ctx context.Context, account e2wtypes.Account, client consensusclient.Service) (phase0.ValidatorIndex, error) {
	pubKey, err := BestPublicKey(account)
	if err != nil {
		return 0, err
	}

	pubKeys := make([]phase0.BLSPubKey, 1)
	copy(pubKeys[0][:], pubKey.Marshal())
	validatorsResponse, err := client.(consensusclient.ValidatorsProvider).Validators(ctx, &api.ValidatorsOpts{
		State:   "head",
		PubKeys: pubKeys,
	})
	if err != nil {
		return 0, err
	}

	for index := range validatorsResponse.Data {
		return index, nil
	}
	return 0, errors.New("validator not found")
}
