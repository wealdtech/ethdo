// Copyright © 2017-2023 Weald Technology Trading
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

	spec "github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wealdtech/ethdo/util"
	"github.com/wealdtech/go-bytesutil"
	e2types "github.com/wealdtech/go-eth2-types/v2"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

var (
	signatureVerifySignature string
	signatureVerifySigner    string
)

// signatureVerifyCmd represents the signature verify command.
var signatureVerifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify signed data",
	Long: `Verify signed data.  For example:

    ethdo signature verify --data=0x5f24e819400c6a8ee2bfc014343cd971b7eb707320025a7bcd83e621e26c35b7 --signature=0x8888... --account="Personal wallet/Operations"

In quiet mode this will return 0 if the data can be signed, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
		defer cancel()

		assert(viper.GetString("signature-data") != "", "--data is required")
		data, err := bytesutil.FromHexString(viper.GetString("signature-data"))
		errCheck(err, "Failed to parse data")
		assert(len(data) == 32, "data to verify must be 32 bytes")

		assert(signatureVerifySignature != "", "--signature is required")
		signatureBytes, err := bytesutil.FromHexString(signatureVerifySignature)
		errCheck(err, "Failed to parse signature")
		signature, err := e2types.BLSSignatureFromBytes(signatureBytes)
		errCheck(err, "Invalid signature")

		domain := e2types.Domain(e2types.DomainType([4]byte{0, 0, 0, 0}), e2types.ZeroForkVersion, e2types.ZeroGenesisValidatorsRoot)
		if viper.GetString("signature-domain") != "" {
			domain, err = bytesutil.FromHexString(viper.GetString("signature-domain"))
			errCheck(err, "Failed to parse domain")
			assert(len(domain) == 32, "Domain data invalid")
		}

		var account e2wtypes.Account
		switch {
		case viper.GetString("account") != "":
			account, err = util.ParseAccount(ctx, viper.GetString("account"), nil, false)
		case viper.GetString("private-key") != "":
			account, err = util.ParseAccount(ctx, viper.GetString("private-key"), nil, false)
		case viper.GetString("public-key") != "":
			account, err = util.ParseAccount(ctx, viper.GetString("public-key"), nil, false)
		}
		errCheck(err, "Failed to obtain account")
		outputIf(viper.GetBool("debug"), fmt.Sprintf("Public key is %#x", account.PublicKey().Marshal()))

		var specDomain spec.Domain
		copy(specDomain[:], domain)
		var root [32]byte
		copy(root[:], data)
		verified, err := util.VerifyRoot(account, root, specDomain, signature)
		errCheck(err, "Failed to verify data")
		assert(verified, "Failed to verify")

		outputIf(viper.GetBool("verbose"), "Verified")
		os.Exit(_exitSuccess)
	},
}

func init() {
	signatureCmd.AddCommand(signatureVerifyCmd)
	signatureFlags(signatureVerifyCmd)
	signatureVerifyCmd.Flags().StringVar(&signatureVerifySignature, "signature", "", "the signature to verify")
	signatureVerifyCmd.Flags().StringVar(&signatureVerifySigner, "signer", "", "the public key of the signer (only if --account is not supplied)")
}
