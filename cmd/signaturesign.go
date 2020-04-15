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
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	pb "github.com/wealdtech/eth2-signer-api/pb/v1"
	"github.com/wealdtech/go-bytesutil"
	e2types "github.com/wealdtech/go-eth2-types/v2"
)

// signatureSignCmd represents the signature sign command
var signatureSignCmd = &cobra.Command{
	Use:   "sign",
	Short: "Sign a 32-byte piece of data",
	Long: `Sign presented data.  For example:

    ethereal signature sign --data=0x5f24e819400c6a8ee2bfc014343cd971b7eb707320025a7bcd83e621e26c35b7 --account="Personal wallet/Operations" --passphrase="my account passphrase"

In quiet mode this will return 0 if the data can be signed, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		assert(signatureData != "", "--data is required")
		data, err := bytesutil.FromHexString(signatureData)
		errCheck(err, "Failed to parse data")
		assert(len(data) == 32, "data to sign must be 32 bytes")

		domain := e2types.Domain(e2types.DomainType([4]byte{0, 0, 0, 0}), e2types.ZeroForkVersion, e2types.ZeroGenesisValidatorsRoot)
		if signatureDomain != "" {
			domainBytes, err := bytesutil.FromHexString(signatureDomain)
			errCheck(err, "Failed to parse domain")
			assert(len(domainBytes) == 32, "Domain data invalid")
		}

		assert(rootAccount != "", "--account is required")

		var signature e2types.Signature
		if remote {
			signClient := pb.NewSignerClient(remoteGRPCConn)
			signReq := &pb.SignRequest{
				Id:     &pb.SignRequest_Account{Account: rootAccount},
				Data:   data,
				Domain: domain,
			}
			ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
			defer cancel()
			resp, err := signClient.Sign(ctx, signReq)
			errCheck(err, "Failed to sign")
			switch resp.State {
			case pb.ResponseState_DENIED:
				die("Signing request denied")
			case pb.ResponseState_FAILED:
				die("Signing request failed")
			case pb.ResponseState_SUCCEEDED:
				signature, err = e2types.BLSSignatureFromBytes(resp.Signature)
				errCheck(err, "Invalid signature")
			}
		} else {
			account, err := accountFromPath(rootAccount)
			errCheck(err, "Failed to access account for signing")
			err = account.Unlock([]byte(rootAccountPassphrase))
			errCheck(err, "Failed to unlock account for signing")
			var fixedSizeData [32]byte
			copy(fixedSizeData[:], data)
			defer account.Lock()
			signature, err = signRoot(account, fixedSizeData, domain)
			errCheck(err, "Failed to sign data")
		}

		outputIf(!quiet, fmt.Sprintf("0x%096x", signature.Marshal()))
		os.Exit(_exitSuccess)
	},
}

func init() {
	signatureCmd.AddCommand(signatureSignCmd)
	signatureFlags(signatureSignCmd)
}
