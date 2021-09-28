// Copyright © 2021 Weald Technology Trading
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

	"github.com/spf13/cobra"
	chainverifysignedcontributionandproof "github.com/wealdtech/ethdo/cmd/chain/verify/signedcontributionandproof"
)

var chainVerifySignedContributionAndProofCmd = &cobra.Command{
	Use:   "signedcontributionandproof",
	Short: "Verify a signed contribution and proof",
	Long: `Verify a signed contribution and proof.  For example:

    ethdo chain verify signedcontributionandproof --data=... --validator=...

validator can be an account, a public key or an index.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		res, err := chainverifysignedcontributionandproof.Run(cmd)
		if err != nil {
			return err
		}
		if res != "" {
			fmt.Print(res)
		}
		return nil
	},
}

func init() {
	chainVerifyCmd.AddCommand(chainVerifySignedContributionAndProofCmd)
	chainVerifyFlags(chainVerifySignedContributionAndProofCmd)
}

func chainVerifySignedContributionAndProofBindings(cmd *cobra.Command) {
	chainVerifyBindings(cmd)
}
