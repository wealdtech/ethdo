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
	"github.com/wealdtech/go-bytesutil"
	types "github.com/wealdtech/go-eth2-types"
)

// signatureSignCmd represents the signature sign command
var signatureSignCmd = &cobra.Command{
	Use:   "sign",
	Short: "Sign data",
	Long: `Sign presented data.  For example:

    ethereal signature sign --data="0x5FfC014343cd971B7eb70732021E26C35B744cc4" --account="Personal wallet/Operations" --passphrase="my account passphrase"

In quiet mode this will return 0 if the data can be signed, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		assert(signatureData != "", "--data is required")
		data, err := bytesutil.FromHexString(signatureData)
		errCheck(err, "Failed to parse data")

		domain := types.Domain([]byte{0, 0, 0, 0}, []byte{0, 0, 0, 0})
		if signatureDomain != "" {
			domainBytes, err := bytesutil.FromHexString(signatureDomain)
			errCheck(err, "Failed to parse domain")
			assert(len(domainBytes) == 8, "Domain data invalid")
		}

		account, err := accountFromPath(rootAccount)
		errCheck(err, "Failed to access account for signing")

		err = account.Unlock([]byte(rootAccountPassphrase))
		errCheck(err, "Failed to unlock account for signing")
		defer account.Lock()

		signature, err := account.Sign(data, domain)
		errCheck(err, "Failed to sign data")

		outputIf(!quiet, fmt.Sprintf("0x%096x", signature.Marshal()))
		os.Exit(_exit_success)
	},
}

func init() {
	signatureCmd.AddCommand(signatureSignCmd)
	signatureFlags(signatureSignCmd)
}
