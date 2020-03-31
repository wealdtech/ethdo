// Copyright © 2017-2019 Weald Technology Trading
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
	types "github.com/wealdtech/go-eth2-wallet-types"
)

// accountKeyCmd represents the account key command
var accountKeyCmd = &cobra.Command{
	Use:   "key",
	Short: "Obtain the private key of an account.",
	Long: `Obtain the private key of an account.  For example:

    ethdo account key --account="Personal wallet/Operations" --passphrase="my account passphrase"

In quiet mode this will return 0 if the key can be obtained, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		assert(!remote, "account keys not available with remote wallets")
		assert(rootAccount != "", "--account is required")

		account, err := accountFromPath(rootAccount)
		errCheck(err, "Failed to access account")

		_, ok := account.(types.AccountPrivateKeyProvider)
		assert(ok, fmt.Sprintf("account %q does not provide its private key", rootAccount))

		assert(rootAccountPassphrase != "", "--passphrase is required")
		err = account.Unlock([]byte(rootAccountPassphrase))
		errCheck(err, "Failed to unlock account to obtain private key")
		defer account.Lock()
		privateKey, err := account.(types.AccountPrivateKeyProvider).PrivateKey()
		errCheck(err, "Failed to obtain private key")
		account.Lock()

		outputIf(!quiet, fmt.Sprintf("%#064x", privateKey.Marshal()))
		os.Exit(_exitSuccess)
	},
}

func init() {
	accountCmd.AddCommand(accountKeyCmd)
	accountFlags(accountKeyCmd)
}
