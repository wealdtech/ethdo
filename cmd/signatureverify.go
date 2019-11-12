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
	"os"

	"github.com/spf13/cobra"
	"github.com/wealdtech/go-bytesutil"
	types "github.com/wealdtech/go-eth2-types"
)

var signatureVerifySignature string
var signatureVerifyPubKey string

// signatureVerifyCmd represents the signature verify command
var signatureVerifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify signed data",
	Long: `Verify signed data.  For example:

    ethereal signature verify --data="0x5FfC014343cd971B7eb70732021E26C35B744cc4" --signature="0x8888..." --account="Personal wallet/Operations"

In quiet mode this will return 0 if the data can be signed, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		assert(signatureData != "", "--data is required")
		data, err := bytesutil.FromHexString(signatureData)
		errCheck(err, "Failed to parse data")

		assert(signatureVerifySignature != "", "--signature is required")
		signatureBytes, err := bytesutil.FromHexString(signatureVerifySignature)
		errCheck(err, "Failed to parse signature")
		signature, err := types.BLSSignatureFromBytes(signatureBytes)
		errCheck(err, "Invalid signature")

		domain := types.Domain([]byte{0, 0, 0, 0}, []byte{0, 0, 0, 0})
		if signatureDomain != "" {
			domainBytes, err := bytesutil.FromHexString(signatureDomain)
			errCheck(err, "Failed to parse domain")
			assert(len(domainBytes) == 8, "Domain data invalid")
		}

		var pubKey types.PublicKey
		assert(signatureVerifyPubKey == "" || rootAccount == "", "Either --pubkey or --account should be supplied")
		if rootAccount != "" {
			account, err := accountFromPath(rootAccount)
			errCheck(err, "Unknown account")
			pubKey = account.PublicKey()
		} else {
			pubKeyBytes, err := bytesutil.FromHexString(signatureVerifyPubKey)
			errCheck(err, "Invalid public key")
			pubKey, err = types.BLSPublicKeyFromBytes(pubKeyBytes)
			errCheck(err, "Invalid public key")
		}
		verified := signature.Verify(data, pubKey, domain)
		if !verified {
			outputIf(!quiet, "Not verified")
			os.Exit(_exit_failure)
		}
		outputIf(!quiet, "Verified")
		os.Exit(_exit_success)
	},
}

func init() {
	signatureCmd.AddCommand(signatureVerifyCmd)
	signatureFlags(signatureVerifyCmd)
	signatureVerifyCmd.Flags().StringVar(&signatureVerifySignature, "signature", "", "the signature to verify")
	signatureVerifyCmd.Flags().StringVar(&signatureVerifyPubKey, "signer", "", "the public key of the signer (only if --account is not supplied)")
}
