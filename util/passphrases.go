// Copyright © 2019 Weald Technology Trading
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
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// GetStorePassphrases() fetches the store passphrase supplied by the user.
func GetStorePassphrase() string {
	return viper.GetString("store-passphrase")
}

// GetWalletPassphrases() fetches the wallet passphrase supplied by the user.
func GetWalletPassphrase() string {
	return viper.GetString("wallet-passphrase")
}

// GetPassphrases() fetches the passphrases supplied by the user.
func GetPassphrases() []string {
	return viper.GetStringSlice("passphrase")
}

// getPassphrase fetches the passphrase supplied by the user.
func GetPassphrase() (string, error) {
	passphrases := GetPassphrases()
	if len(passphrases) == 0 {
		return "", errors.New("passphrase is required")
	}
	if len(passphrases) > 1 {
		return "", errors.New("multiple passphrases supplied")
	}
	return passphrases[0], nil
}

// GetOptionalPassphrase fetches the passphrase if supplied by the user.
func GetOptionalPassphrase() (string, error) {
	passphrases := GetPassphrases()
	if len(passphrases) == 0 {
		return "", nil
	}
	if len(passphrases) > 1 {
		return "", errors.New("multiple passphrases supplied")
	}
	return passphrases[0], nil
}
