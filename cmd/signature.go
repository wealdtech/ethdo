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
)

var signatureData string
var signatureDomain string

// signatureCmd represents the signature command
var signatureCmd = &cobra.Command{
	Use:     "signature",
	Aliases: []string{"sig"},
	Short:   "Manage signatures",
	Long:    `Sign data and verify signatures.`,
}

func init() {
	RootCmd.AddCommand(signatureCmd)
}

func signatureFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&signatureData, "data", "", "the hex string of data")
	cmd.Flags().StringVar(&signatureDomain, "domain", "", "the hex string of the BLS domain (defaults to 0x0000000000000000)")
}
