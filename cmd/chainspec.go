// Copyright © 2023 Weald Technology Trading.
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
	"encoding/json"
	"fmt"
	"sort"
	"time"

	eth2client "github.com/attestantio/go-eth2-client"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wealdtech/ethdo/util"
)

var chainSpecCmd = &cobra.Command{
	Use:   "spec",
	Short: "Obtain specification for a chain",
	Long: `Obtain specification for a chain.  For example:

    ethdo chain spec

In quiet mode this will return 0 if the chain specification can be obtained, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		eth2Client, err := util.ConnectToBeaconNode(ctx, &util.ConnectOpts{
			Address:       viper.GetString("connection"),
			Timeout:       viper.GetDuration("timeout"),
			AllowInsecure: viper.GetBool("allow-insecure-connections"),
			LogFallback:   !viper.GetBool("quiet"),
		})
		errCheck(err, "Failed to connect to Ethereum consensus node")

		spec, err := eth2Client.(eth2client.SpecProvider).Spec(ctx)
		errCheck(err, "Failed to obtain chain specification")

		if viper.GetBool("quiet") {
			return
		}

		// Tweak the spec for output.
		for k, v := range spec {
			switch t := v.(type) {
			case phase0.Version:
				spec[k] = fmt.Sprintf("%#x", t)
			case phase0.DomainType:
				spec[k] = fmt.Sprintf("%#x", t)
			case time.Time:
				spec[k] = fmt.Sprintf("%d", t.Unix())
			case time.Duration:
				spec[k] = fmt.Sprintf("%d", uint64(t.Seconds()))
			case []byte:
				spec[k] = fmt.Sprintf("%#x", t)
			case uint64:
				spec[k] = fmt.Sprintf("%d", t)
			}
		}

		if viper.GetBool("json") {
			data, err := json.Marshal(spec)
			errCheck(err, "Failed to marshal JSON")
			fmt.Printf("%s\n", string(data))
		} else {
			keys := make([]string, 0, len(spec))
			for k := range spec {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, key := range keys {
				fmt.Printf("%s: %v\n", key, spec[key])
			}
		}
	},
}

func init() {
	chainCmd.AddCommand(chainSpecCmd)
	chainFlags(chainSpecCmd)
}

func chainSpecBindings(_ *cobra.Command) {
}
