// Copyright © 2020 Weald Technology Trading
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
	"github.com/spf13/viper"
	attesterinclusion "github.com/wealdtech/ethdo/cmd/attester/inclusion"
)

var attesterInclusionCmd = &cobra.Command{
	Use:   "inclusion",
	Short: "Obtain information about attester inclusion",
	Long: `Obtain information about attester inclusion.  For example:

    ethdo attester inclusion --account=Validators/00001 --epoch=12345

In quiet mode this will return 0 if an attestation from the attester is found on the block of the given epoch, otherwise 1.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		res, err := attesterinclusion.Run(cmd)
		if err != nil {
			return err
		}
		if viper.GetBool("quiet") {
			return nil
		}
		if res != "" {
			fmt.Println(res)
		}
		return nil
	},
}

func init() {
	attesterCmd.AddCommand(attesterInclusionCmd)
	attesterFlags(attesterInclusionCmd)
	attesterInclusionCmd.Flags().Int64("epoch", -1, "the last complete epoch")
	attesterInclusionCmd.Flags().String("pubkey", "", "the public key of the attester")
	attesterInclusionCmd.Flags().Int64("index", -1, "the index of the attester")
}

func attesterInclusionBindings() {
	if err := viper.BindPFlag("epoch", attesterInclusionCmd.Flags().Lookup("epoch")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("pubkey", attesterInclusionCmd.Flags().Lookup("pubkey")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("index", attesterInclusionCmd.Flags().Lookup("index")); err != nil {
		panic(err)
	}
}
