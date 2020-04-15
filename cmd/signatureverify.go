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

	"github.com/prysmaticlabs/go-ssz"
	"github.com/spf13/cobra"
	pb "github.com/wealdtech/eth2-signer-api/pb/v1"
	"github.com/wealdtech/go-bytesutil"
	e2types "github.com/wealdtech/go-eth2-types/v2"
)

var signatureVerifySignature string
var signatureVerifyPubKey string

// signatureVerifyCmd represents the signature verify command
var signatureVerifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify signed data",
	Long: `Verify signed data.  For example:

    ethereal signature verify --data=0x5f24e819400c6a8ee2bfc014343cd971b7eb707320025a7bcd83e621e26c35b7 --signature=0x8888... --account="Personal wallet/Operations"

In quiet mode this will return 0 if the data can be signed, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		assert(signatureData != "", "--data is required")
		data, err := bytesutil.FromHexString(signatureData)
		errCheck(err, "Failed to parse data")
		assert(len(data) == 32, "data to verify must be 32 bytes")

		assert(signatureVerifySignature != "", "--signature is required")
		signatureBytes, err := bytesutil.FromHexString(signatureVerifySignature)
		errCheck(err, "Failed to parse signature")
		signature, err := e2types.BLSSignatureFromBytes(signatureBytes)
		errCheck(err, "Invalid signature")

		domain := e2types.Domain(e2types.DomainType([4]byte{0, 0, 0, 0}), e2types.ZeroForkVersion, e2types.ZeroGenesisValidatorsRoot)
		if signatureDomain != "" {
			domainBytes, err := bytesutil.FromHexString(signatureDomain)
			errCheck(err, "Failed to parse domain")
			assert(len(domainBytes) == 32, "Domain data invalid")
		}

		var pubKey e2types.PublicKey
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
				pubKey, err = e2types.BLSPublicKeyFromBytes(resp.Accounts[0].PublicKey)
				errCheck(err, "Invalid public key provided for account")
			} else {
				account, err := accountFromPath(rootAccount)
				errCheck(err, "Unknown account")
				pubKey = account.PublicKey()
			}
		} else {
			pubKeyBytes, err := bytesutil.FromHexString(signatureVerifyPubKey)
			errCheck(err, "Invalid public key")
			pubKey, err = e2types.BLSPublicKeyFromBytes(pubKeyBytes)
			errCheck(err, "Invalid public key")
		}
		container := &SigningContainer{
			Root:   data,
			Domain: domain,
		}
		root, err := ssz.HashTreeRoot(container)
		errCheck(err, "Failed to create signing root")

		verified := signature.Verify(root[:], pubKey)
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
