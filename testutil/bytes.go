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

package testutil

import (
	"encoding/hex"
	"strings"

	spec "github.com/attestantio/go-eth2-client/spec/phase0"
)

// HexToBytes converts a hex string to a byte array.
// This should only be used for pre-defined test strings; it will panic if the input is invalid.
func HexToBytes(input string) []byte {
	res, err := hex.DecodeString(strings.TrimPrefix(input, "0x"))
	if err != nil {
		panic(err)
	}
	return res
}

// HexToPubKey converts a hex string to a spec public key.
// This should only be used for pre-defined test strings; it will panic if the input is invalid.
func HexToPubKey(input string) spec.BLSPubKey {
	data := HexToBytes(input)
	var res spec.BLSPubKey
	copy(res[:], data)
	return res
}

// HexToSignature converts a hex string to a spec signature.
// This should only be used for pre-defined test strings; it will panic if the input is invalid.
func HexToSignature(input string) spec.BLSSignature {
	data := HexToBytes(input)
	var res spec.BLSSignature
	copy(res[:], data)
	return res
}

// HexToDomainType converts a hex string to a spec domain type.
// This should only be used for pre-defined test strings; it will panic if the input is invalid.
func HexToDomainType(input string) spec.DomainType {
	data := HexToBytes(input)
	var res spec.DomainType
	copy(res[:], data)
	return res
}

// HexToDomain converts a hex string to a spec domain.
// This should only be used for pre-defined test strings; it will panic if the input is invalid.
func HexToDomain(input string) spec.Domain {
	data := HexToBytes(input)
	var res spec.Domain
	copy(res[:], data)
	return res
}

// HexToVersion converts a hex string to a spec version.
// This should only be used for pre-defined test strings; it will panic if the input is invalid.
func HexToVersion(input string) spec.Version {
	data := HexToBytes(input)
	var res spec.Version
	copy(res[:], data)
	return res
}

// HexToRoot converts a hex string to a spec root.
// This should only be used for pre-defined test strings; it will panic if the input is invalid.
func HexToRoot(input string) spec.Root {
	data := HexToBytes(input)
	var res spec.Root
	copy(res[:], data)
	return res
}
