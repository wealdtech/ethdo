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
	"os"

	"github.com/spf13/cobra"
)

var accountInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Information about an account",
	Long: `Obtain information about an account.  For example:

    ethdo account info --account="primary/my funds"

In quiet mode this will return 0 if the account exists, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		assert(rootAccount != "", "--account is required")

		account, err := accountFromPath(rootAccount)
		errCheck(err, "Failed to access wallet")

		outputIf(!quiet, fmt.Sprintf("Public key: 0x%048x", account.PublicKey().Marshal()))
		if verbose {
			outputIf(account.Path() != "", fmt.Sprintf("Path: %s", account.Path()))
		}
		os.Exit(_exit_success)
	},
}

func init() {
	accountCmd.AddCommand(accountInfoCmd)
	accountFlags(accountInfoCmd)
}
