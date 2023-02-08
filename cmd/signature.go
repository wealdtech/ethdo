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
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// signatureCmd represents the signature command.
var signatureCmd = &cobra.Command{
	Use:     "signature",
	Aliases: []string{"sig"},
	Short:   "Manage signatures",
	Long:    `Sign data and verify signatures.`,
}

func init() {
	RootCmd.AddCommand(signatureCmd)
}

var (
	dataFlag   *pflag.Flag
	domainFlag *pflag.Flag
)

func signatureFlags(cmd *cobra.Command) {
	if dataFlag == nil {
		cmd.Flags().String("data", "", "the data, as a hex string")
		dataFlag = cmd.Flags().Lookup("data")
		if err := viper.BindPFlag("signature-data", dataFlag); err != nil {
			panic(err)
		}
		cmd.Flags().String("domain", "0x0000000000000000000000000000000000000000000000000000000000000000", "the BLS domain, as a hex string")
		domainFlag = cmd.Flags().Lookup("domain")
		if err := viper.BindPFlag("signature-domain", domainFlag); err != nil {
			panic(err)
		}
	} else {
		cmd.Flags().AddFlag(dataFlag)
		cmd.Flags().AddFlag(domainFlag)
	}
}
