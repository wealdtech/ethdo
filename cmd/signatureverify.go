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
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wealdtech/ethdo/util"
	"github.com/wealdtech/go-bytesutil"
	e2types "github.com/wealdtech/go-eth2-types/v2"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

var signatureVerifySignature string
var signatureVerifySigner string

// signatureVerifyCmd represents the signature verify command
var signatureVerifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify signed data",
	Long: `Verify signed data.  For example:

    ethdo signature verify --data=0x5f24e819400c6a8ee2bfc014343cd971b7eb707320025a7bcd83e621e26c35b7 --signature=0x8888... --account="Personal wallet/Operations"

In quiet mode this will return 0 if the data can be signed, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
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

		account, err := signatureVerifyAccount()
		errCheck(err, "Failed to obtain account")
		outputIf(debug, fmt.Sprintf("Public key is %#x", account.PublicKey().Marshal()))

		var root [32]byte
		copy(root[:], data)
		verified, err := verifyRoot(account, root, domain, signature)
		errCheck(err, "Failed to verify data")
		assert(verified, "Failed to verify")

		outputIf(verbose, "Verified")
		os.Exit(_exitSuccess)
	},
}

// signatureVerifyAccount obtains the account for the signature verify command.
func signatureVerifyAccount() (e2wtypes.Account, error) {
	var account e2wtypes.Account
	var err error
	if viper.GetString("account") != "" {
		ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
		defer cancel()
		_, account, err = walletAndAccountFromPath(ctx, viper.GetString("account"))
		if err != nil {
			return nil, errors.Wrap(err, "failed to obtain account")
		}
	} else {
		pubKeyBytes, err := hex.DecodeString(strings.TrimPrefix(signatureVerifySigner, "0x"))
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("failed to decode public key %s", signatureVerifySigner))
		}
		account, err = util.NewScratchAccount(nil, pubKeyBytes)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("invalid public key %s", signatureVerifySigner))
		}
	}
	return account, nil
}
func init() {
	signatureCmd.AddCommand(signatureVerifyCmd)
	signatureFlags(signatureVerifyCmd)
	signatureVerifyCmd.Flags().StringVar(&signatureVerifySignature, "signature", "", "the signature to verify")
	signatureVerifyCmd.Flags().StringVar(&signatureVerifySigner, "signer", "", "the public key of the signer (only if --account is not supplied)")
}
