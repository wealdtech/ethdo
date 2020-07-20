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
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	types "github.com/wealdtech/go-eth2-wallet-types/v2"
)

var walletExportPassphrase string

var walletExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export a wallet",
	Long: `Export a wallet for backup of transfer.  For example:

    ethdo wallet export --wallet=primary --exportpassphrase="my export secret"

In quiet mode this will return 0 if the wallet is able to be exported, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
		defer cancel()

		assert(viper.GetString("remote") == "", "wallet export not available with remote wallets")
		assert(viper.GetString("wallet") != "", "--wallet is required")
		assert(walletExportPassphrase != "", "--exportpassphrase is required")

		wallet, err := walletFromPath(viper.GetString("wallet"))
		errCheck(err, "Failed to access wallet")

		_, ok := wallet.(types.WalletExporter)
		assert(ok, fmt.Sprintf("wallets of type %q do not allow exporting accounts", wallet.Type()))

		exportData, err := wallet.(types.WalletExporter).Export(ctx, []byte(walletExportPassphrase))
		errCheck(err, "Failed to export wallet")

		outputIf(!quiet, fmt.Sprintf("0x%x", exportData))
		os.Exit(_exitSuccess)
	},
}

func init() {
	walletCmd.AddCommand(walletExportCmd)
	walletFlags(walletExportCmd)
	walletExportCmd.Flags().StringVar(&walletExportPassphrase, "exportpassphrase", "", "Passphrase to protect the export")
}
