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
	"github.com/wealdtech/go-bytesutil"
	e2types "github.com/wealdtech/go-eth2-types/v2"
)

// signatureSignCmd represents the signature sign command
var signatureSignCmd = &cobra.Command{
	Use:   "sign",
	Short: "Sign a 32-byte piece of data",
	Long: `Sign presented data.  For example:

    ethdo signature sign --data=0x5f24e819400c6a8ee2bfc014343cd971b7eb707320025a7bcd83e621e26c35b7 --account="Personal wallet/Operations" --passphrase="my account passphrase"

In quiet mode this will return 0 if the data can be signed, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
		defer cancel()

		assert(viper.GetString("signature-data") != "", "--data is required")
		data, err := bytesutil.FromHexString(viper.GetString("signature-data"))
		errCheck(err, "Failed to parse data")
		assert(len(data) == 32, "data to sign must be 32 bytes")

		domain := e2types.Domain(e2types.DomainType([4]byte{0, 0, 0, 0}), e2types.ZeroForkVersion, e2types.ZeroGenesisValidatorsRoot)
		if viper.GetString("signature-domain") != "" {
			domain, err = bytesutil.FromHexString(viper.GetString("signature-domain"))
			errCheck(err, "Failed to parse domain")
			assert(len(domain) == 32, "Domain data invalid")
		}
		outputIf(debug, fmt.Sprintf("Domain is %#x", domain))

		assert(viper.GetString("account") != "", "--account is required")
		_, account, err := walletAndAccountFromInput(ctx)
		errCheck(err, "Failed to obtain account")

		var fixedSizeData [32]byte
		copy(fixedSizeData[:], data)
		signature, err := signRoot(account, fixedSizeData, domain)
		errCheck(err, "Failed to sign")

		outputIf(!quiet, fmt.Sprintf("%#x", signature.Marshal()))
		os.Exit(_exitSuccess)
	},
}

func init() {
	signatureCmd.AddCommand(signatureSignCmd)
	signatureFlags(signatureSignCmd)
}
