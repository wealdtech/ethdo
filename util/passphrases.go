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
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// GetStorePassphrase fetches the store passphrase supplied by the user.
func GetStorePassphrase(store string) string {
	// Try store-specific passphrase.
	storePassphrase := viper.GetString(fmt.Sprintf("stores.%s.passphrase", store))
	if storePassphrase == "" {
		// Try generic passphrase.
		storePassphrase = viper.GetString("store-passphrase")
	}
	if storePassphrase == "" {
		// Try deprecated name.
		storePassphrase = viper.GetString("storepassphrase")
	}
	return storePassphrase
}

// GetWalletPassphrase fetches the wallet passphrase supplied by the user.
func GetWalletPassphrase() string {
	walletPassphrase := viper.GetString("wallet-passphrase")
	if walletPassphrase == "" {
		// Try deprecated name.
		walletPassphrase = viper.GetString("walletpassphrase")
	}
	return walletPassphrase
}

// GetPassphrases fetches the passphrases supplied by the user.
func GetPassphrases() []string {
	return viper.GetStringSlice("passphrase")
}

// GetPassphrase fetches the passphrase supplied by the user.
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
