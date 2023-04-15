// Copyright © 2021 Weald Technology Trading
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

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	walletsharedexport "github.com/wealdtech/ethdo/cmd/wallet/sharedexport"
)

var walletSharedExportCmd = &cobra.Command{
	Use:   "sharedexport",
	Short: "Export a wallet using Shamir secret sharing",
	Long: `Export a wallet for backup of transfer using Shamir secret sharing.  For example:

    ethdo wallet sharedexport --wallet=primary --participants=5 --threshold=3 --file=backup.dat`,
	RunE: func(cmd *cobra.Command, args []string) error {
		res, err := walletsharedexport.Run(cmd)
		if err != nil {
			return err
		}
		if res != "" {
			fmt.Println(res)
		}
		return nil
	},
}

func init() {
	walletCmd.AddCommand(walletSharedExportCmd)
	walletFlags(walletSharedExportCmd)
	walletSharedExportCmd.Flags().Uint32("participants", 0, "Number of participants in sharing scheme")
	walletSharedExportCmd.Flags().Uint32("threshold", 0, "Number of participants required to recover the export")
	walletSharedExportCmd.Flags().String("file", "", "Name of the file that stores the export")
}

func walletSharedExportBindings(cmd *cobra.Command) {
	if err := viper.BindPFlag("participants", cmd.Flags().Lookup("participants")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("threshold", cmd.Flags().Lookup("threshold")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("file", cmd.Flags().Lookup("file")); err != nil {
		panic(err)
	}
}
