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
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wealdtech/ethdo/grpc"
	"github.com/wealdtech/ethdo/util"
	e2types "github.com/wealdtech/go-eth2-types/v2"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

var exitVerifyPubKey string

var exitVerifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify deposit data matches requirements",
	Long: `Verify deposit data matches requirements.  For example:

    ethdo deposit verify --data=depositdata.json --withdrawalaccount=primary/current --value="32 Ether"

The information generated can be passed to ethereal to create a deposit from the Ethereum 1 chain.

In quiet mode this will return 0 if the the data can be generated correctly, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
		defer cancel()

		assert(viper.GetString("account") != "" || exitVerifyPubKey != "", "account or public key is required")
		account, err := exitVerifyAccount(ctx)
		errCheck(err, "Failed to obtain account")

		assert(viper.GetString("exit.data") != "", "exit data is required")
		data, err := obtainExitData(viper.GetString("exit.Data"))
		errCheck(err, "Failed to obtain exit data")

		// Confirm signature is good.
		err = connect()
		errCheck(err, "Failed to obtain connection to Ethereum 2 beacon chain node")
		genesisValidatorsRoot, err := grpc.FetchGenesisValidatorsRoot(eth2GRPCConn)
		outputIf(debug, fmt.Sprintf("Genesis validators root is %x", genesisValidatorsRoot))
		errCheck(err, "Failed to obtain genesis validators root")
		domain := e2types.Domain(e2types.DomainVoluntaryExit, data.ForkVersion, genesisValidatorsRoot)
		exit := &ethpb.VoluntaryExit{
			Epoch:          data.Epoch,
			ValidatorIndex: data.ValidatorIndex,
		}
		sig, err := e2types.BLSSignatureFromBytes(data.Signature)
		errCheck(err, "Invalid signature")
		verified, err := verifyStruct(account, exit, domain, sig)
		errCheck(err, "Failed to verify voluntary exit")
		assert(verified, "Voluntary exit failed to verify")

		// TODO confirm fork version is valid (once we have a way of obtaining the current fork version).

		outputIf(verbose, "Verified")
		os.Exit(_exitSuccess)
	},
}

// obtainExitData obtains exit data from an input, could be JSON itself or a path to JSON.
func obtainExitData(input string) (*validatorExitData, error) {
	var err error
	var data []byte
	// Input could be JSON or a path to JSON
	if strings.HasPrefix(input, "{") {
		// Looks like JSON
		data = []byte(input)
	} else {
		// Assume it's a path to JSON
		data, err = ioutil.ReadFile(input)
		if err != nil {
			return nil, errors.Wrap(err, "failed to find deposit data file")
		}
	}
	exitData := &validatorExitData{}
	err = json.Unmarshal([]byte(data), exitData)
	if err != nil {
		return nil, errors.Wrap(err, "data is not valid JSON")
	}

	return exitData, nil
}

// exitVerifyAccount obtains the account for the exitVerify command.
func exitVerifyAccount(ctx context.Context) (e2wtypes.Account, error) {
	var account e2wtypes.Account
	var err error
	if viper.GetString("account") != "" {
		_, account, err = walletAndAccountFromPath(ctx, viper.GetString("account"))
		if err != nil {
			return nil, errors.Wrap(err, "failed to obtain account")
		}
	} else {
		pubKeyBytes, err := hex.DecodeString(strings.TrimPrefix(exitVerifyPubKey, "0x"))
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("failed to decode public key %s", exitVerifyPubKey))
		}
		account, err = util.NewScratchAccount(nil, pubKeyBytes)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("invalid public key %s", exitVerifyPubKey))
		}
	}
	return account, nil
}

func init() {
	exitCmd.AddCommand(exitVerifyCmd)
	exitFlags(exitVerifyCmd)
	exitVerifyCmd.Flags().String("data", "", "JSON data, or path to JSON data")
	exitVerifyCmd.Flags().StringVar(&exitVerifyPubKey, "pubkey", "", "Public key for which to verify exit")
	if err := viper.BindPFlag("exit.data", exitVerifyCmd.Flags().Lookup("data")); err != nil {
		panic(err)
	}
}
