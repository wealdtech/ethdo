// Copyright © 2017-2020 Weald Technology Trading
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
	"os"

	"github.com/spf13/cobra"
	pb "github.com/wealdtech/eth2-signer-api/pb/v1"
	"github.com/wealdtech/go-bytesutil"
	types "github.com/wealdtech/go-eth2-types/v2"
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

		//		domain := types.Domain([]byte{0, 0, 0, 0}, []byte{0, 0, 0, 0})
		if signatureDomain != "" {
			domainBytes, err := bytesutil.FromHexString(signatureDomain)
			errCheck(err, "Failed to parse domain")
			assert(len(domainBytes) == 8, "Domain data invalid")
		}

		var pubKey types.PublicKey
		assert(signatureVerifyPubKey == "" || rootAccount == "", "Either --pubkey or --account should be supplied")
		if rootAccount != "" {
			if remote {
				listerClient := pb.NewListerClient(remoteGRPCConn)
				listAccountsReq := &pb.ListAccountsRequest{
					Paths: []string{
						rootAccount,
					},
				}
				resp, err := listerClient.ListAccounts(context.Background(), listAccountsReq)
				errCheck(err, "Failed to access account")
				assert(resp.State == pb.ResponseState_SUCCEEDED, "Failed to obtain account")
				assert(len(resp.Accounts) == 1, "No such account")
				pubKey, err = types.BLSPublicKeyFromBytes(resp.Accounts[0].PublicKey)
				errCheck(err, "Invalid public key provided for account")
			} else {
				account, err := accountFromPath(rootAccount)
				errCheck(err, "Unknown account")
				pubKey = account.PublicKey()
			}
		} else {
			pubKeyBytes, err := bytesutil.FromHexString(signatureVerifyPubKey)
			errCheck(err, "Invalid public key")
			pubKey, err = types.BLSPublicKeyFromBytes(pubKeyBytes)
			errCheck(err, "Invalid public key")
		}
		// TODO data + domain -> root
		verified := signature.Verify(data, pubKey)
		if !verified {
			outputIf(!quiet, "Not verified")
			os.Exit(_exitFailure)
		}
		outputIf(!quiet, "Verified")
		os.Exit(_exitSuccess)
	},
}

func init() {
	signatureCmd.AddCommand(signatureVerifyCmd)
	signatureFlags(signatureVerifyCmd)
	signatureVerifyCmd.Flags().StringVar(&signatureVerifySignature, "signature", "", "the signature to verify")
	signatureVerifyCmd.Flags().StringVar(&signatureVerifyPubKey, "signer", "", "the public key of the signer (only if --account is not supplied)")
}
