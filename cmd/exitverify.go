// Copyright © 2020, 2022 Weald Technology Trading
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
	"os"
	"strings"

	consensusclient "github.com/attestantio/go-eth2-client"
	"github.com/attestantio/go-eth2-client/api"
	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wealdtech/ethdo/util"
	e2types "github.com/wealdtech/go-eth2-types/v2"
)

var exitVerifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify exit data is valid",
	Long: `Verify that exit data generated by "ethdo validator exit" is correct for a given account.  For example:

    ethdo exit verify --signed-operation=exitdata.json --validator=primary/current

In quiet mode this will return 0 if the exit is verified correctly, otherwise 1.`,
	Run: func(_ *cobra.Command, _ []string) {
		ctx := context.Background()

		assert(viper.GetString("signed-operation") != "", "signed-operation is required")
		signedOp, err := obtainSignedOperation(viper.GetString("signed-operation"))
		errCheck(err, "Failed to obtain signed operation")

		eth2Client, err := util.ConnectToBeaconNode(ctx, &util.ConnectOpts{
			Address:       viper.GetString("connection"),
			Timeout:       viper.GetDuration("timeout"),
			AllowInsecure: viper.GetBool("allow-insecure-connections"),
			LogFallback:   !viper.GetBool("quiet"),
		})
		errCheck(err, "Failed to connect to Ethereum 2 beacon node")

		validator, err := util.ParseValidator(ctx, eth2Client.(consensusclient.ValidatorsProvider), fmt.Sprintf("%d", signedOp.Message.ValidatorIndex), "head")
		errCheck(err, "failed to obtain validator")
		pubkey, err := validator.PubKey(ctx)
		errCheck(err, "failed to obtain validator public key")
		account, err := util.ParseAccount(ctx, pubkey.String(), nil, false)
		errCheck(err, "failed to obtain account")

		// Ensure the validator is in a suitable state.
		assert(validator.Status == v1.ValidatorStateActiveOngoing, "validator not in a suitable state to exit")

		// Obtain the hash tree root of the message to check the signature.
		opRoot, err := signedOp.Message.HashTreeRoot()
		errCheck(err, "Failed to obtain exit hash tree root")

		genesisResponse, err := eth2Client.(consensusclient.GenesisProvider).Genesis(ctx, &api.GenesisOpts{})
		errCheck(err, "Failed to obtain beacon chain genesis")
		genesis := genesisResponse.Data

		response, err := eth2Client.(consensusclient.SpecProvider).Spec(ctx, &api.SpecOpts{})
		errCheck(err, "Failed to obtain spec information")

		// Check against Capella fork version (EIP-7044)
		signatureBytes := make([]byte, 96)
		copy(signatureBytes, signedOp.Signature[:])
		sig, err := e2types.BLSSignatureFromBytes(signatureBytes)
		errCheck(err, "Invalid signature")

		domain := phase0.Domain{}
		forkRaw, ok := response.Data["CAPELLA_FORK_VERSION"]
		if !ok {
			err = errors.New("failed to obtain Capella fork version")
		}
		errCheck(err, "Failed to obtain fork version")

		fork, ok := forkRaw.(phase0.Version)
		if !ok {
			err = errors.New("fork version is not of a valid type")
		}
		errCheck(err, "Failed to obtain fork version")

		exitDomain, err := e2types.ComputeDomain(e2types.DomainVoluntaryExit, fork[:], genesis.GenesisValidatorsRoot[:])
		errCheck(err, "Failed to compute domain")

		copy(domain[:], exitDomain)
		verified, err := util.VerifyRoot(account, opRoot, domain, sig)
		errCheck(err, "Failed to verify voluntary exit")

		assert(verified, "Voluntary exit failed to verify against current and previous fork versions")

		outputIf(viper.GetBool("verbose"), "Verified")
		os.Exit(_exitSuccess)
	},
}

// obtainSignedOperation obtains exit data from an input, could be JSON itself or a path to JSON.
func obtainSignedOperation(input string) (*phase0.SignedVoluntaryExit, error) {
	var err error
	var data []byte
	// Input could be JSON or a path to JSON
	if strings.HasPrefix(input, "{") {
		// Looks like JSON
		data = []byte(input)
	} else {
		// Assume it's a path to JSON
		data, err = os.ReadFile(input)
		if err != nil {
			return nil, errors.Wrap(err, "failed to find deposit data file")
		}
	}
	signedOp := &phase0.SignedVoluntaryExit{}
	err = json.Unmarshal(data, signedOp)
	if err != nil {
		return nil, errors.Wrap(err, "data is not valid JSON")
	}

	return signedOp, nil
}

func init() {
	exitCmd.AddCommand(exitVerifyCmd)
	exitFlags(exitVerifyCmd)
	exitVerifyCmd.Flags().String("signed-operation", "", "JSON data, or path to JSON data")
}

func exitVerifyBindings(cmd *cobra.Command) {
	if err := viper.BindPFlag("signed-operation", cmd.Flags().Lookup("signed-operation")); err != nil {
		panic(err)
	}
}
