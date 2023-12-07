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
	"strconv"
	"time"

	eth2client "github.com/attestantio/go-eth2-client"
	"github.com/attestantio/go-eth2-client/api"
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

		specResponse, err := eth2Client.(eth2client.SpecProvider).Spec(ctx, &api.SpecOpts{})
		errCheck(err, "Failed to obtain chain specification")

		if viper.GetBool("quiet") {
			return
		}

		// Tweak the spec for output.
		for k, v := range specResponse.Data {
			switch t := v.(type) {
			case phase0.Version:
				specResponse.Data[k] = fmt.Sprintf("%#x", t)
			case phase0.DomainType:
				specResponse.Data[k] = fmt.Sprintf("%#x", t)
			case time.Time:
				specResponse.Data[k] = strconv.FormatInt(t.Unix(), 10)
			case time.Duration:
				specResponse.Data[k] = strconv.FormatUint(uint64(t.Seconds()), 10)
			case []byte:
				specResponse.Data[k] = fmt.Sprintf("%#x", t)
			case uint64:
				specResponse.Data[k] = strconv.FormatUint(t, 10)
			}
		}

		if viper.GetBool("json") {
			data, err := json.Marshal(specResponse.Data)
			errCheck(err, "Failed to marshal JSON")
			fmt.Printf("%s\n", string(data))
		} else {
			keys := make([]string, 0, len(specResponse.Data))
			for k := range specResponse.Data {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, key := range keys {
				fmt.Printf("%s: %v\n", key, specResponse.Data[key])
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
