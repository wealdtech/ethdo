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

package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wealdtech/go-bytesutil"
	wallet "github.com/wealdtech/go-eth2-wallet"
)

var walletImportData string
var walletImportPassphrase string

var walletImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import a wallet",
	Long: `Import a wallet.  For example:

    ethdo wallet import --importdata=primary --importpassphrase="my export secret"

In quiet mode this will return 0 if the wallet is imported successfully, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		assert(!remote, "wallet import not available with remote wallets")
		assert(walletImportData != "", "--walletimportdata is required")
		assert(walletImportPassphrase != "", "--importpassphrase is required")

		if !strings.HasPrefix(walletImportData, "0x") {
			outputIf(debug, fmt.Sprintf("Reading wallet import from file %s", walletImportData))
			// Assume this is a path
			fileData, err := ioutil.ReadFile(walletImportData)
			errCheck(err, "Failed to read wallet import data")
			walletImportData = strings.TrimSpace(string(fileData))
		}
		outputIf(debug, fmt.Sprintf("Wallet import data is of length %d", len(walletImportData)))
		importData, err := bytesutil.FromHexString(walletImportData)
		errCheck(err, "Failed to decode wallet data")

		_, err = wallet.ImportWallet(importData, []byte(walletImportPassphrase))
		errCheck(err, "Failed to import wallet")

		os.Exit(_exitSuccess)
	},
}

func init() {
	walletCmd.AddCommand(walletImportCmd)
	walletFlags(walletImportCmd)
	walletImportCmd.Flags().StringVar(&walletImportData, "importdata", "", "The data to import, or the name of a file to read")
	walletImportCmd.Flags().StringVar(&walletImportPassphrase, "importpassphrase", "", "Passphrase protecting the data to import")
}
