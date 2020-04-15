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

package util

import (
	"errors"

	"github.com/google/uuid"
	types "github.com/wealdtech/go-eth2-types/v2"
)

// ScratchAccount is an account that exists temporarily.
type ScratchAccount struct {
	id     uuid.UUID
	pubKey types.PublicKey
}

// NewScratchAccount creates a new local account.
func NewScratchAccount(pubKey []byte) (*ScratchAccount, error) {
	key, err := types.BLSPublicKeyFromBytes(pubKey)
	if err != nil {
		return nil, err
	}

	return &ScratchAccount{
		id:     uuid.New(),
		pubKey: key,
	}, nil
}

func (a *ScratchAccount) ID() uuid.UUID {
	return a.id
}

func (a *ScratchAccount) Name() string {
	return "scratch"
}

func (a *ScratchAccount) PublicKey() types.PublicKey {
	return a.pubKey
}

func (a *ScratchAccount) Path() string {
	return ""
}

func (a *ScratchAccount) Lock() {
}

func (a *ScratchAccount) Unlock([]byte) error {
	return nil
}

func (a *ScratchAccount) IsUnlocked() bool {
	return false
}

func (a *ScratchAccount) Sign(data []byte) (types.Signature, error) {
	return nil, errors.New("Not implemented")
}
