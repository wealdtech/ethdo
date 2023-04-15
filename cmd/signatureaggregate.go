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
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/herumi/bls-eth-go-binary/bls"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wealdtech/ethdo/util"
)

var signatureAggregateSignatures []string

// signatureAggregateCmd represents the signature aggregate command.
var signatureAggregateCmd = &cobra.Command{
	Use:   "aggregate",
	Short: "Aggregate signatures",
	Long: `Aggregate signatures, either threshold or absolute.  For example:

    ethdo signature aggregate --signatures=0x5f24e819400c6a8ee2bfc014343cd971b7eb707320025a7bcd83e621e26c35b7,

Signatures are specified as "signature" for simple aggregation, and as "id:signature" for threshold aggregation.

In quiet mode this will return 0 if the signatures can be aggregated, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		assert(len(signatureAggregateSignatures) > 1, "multiple signatures required to aggregate")
		var signature *bls.Sign
		var err error
		if strings.Contains(signatureAggregateSignatures[0], ":") {
			signature, err = generateThresholdSignature()
		} else {
			signature, err = generateAggregateSignature()
		}
		errCheck(err, "Failed to aggregate signature")

		outputIf(!viper.GetBool("quiet"), fmt.Sprintf("%#x", signature.Serialize()))
		os.Exit(_exitSuccess)
	},
}

func generateThresholdSignature() (*bls.Sign, error) {
	ids := make([]bls.ID, len(signatureAggregateSignatures))
	sigs := make([]bls.Sign, len(signatureAggregateSignatures))

	for i := range signatureAggregateSignatures {
		parts := strings.Split(signatureAggregateSignatures[i], ":")
		if len(parts) != 2 {
			return nil, errors.New("invalid threshold signature format")
		}
		id, err := strconv.ParseUint(parts[0], 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, "invalid threshold signature ID")
		}
		ids[i] = *util.BLSID(id)
		sigBytes, err := hex.DecodeString(strings.TrimPrefix(parts[1], "0x"))
		if err != nil {
			return nil, errors.Wrap(err, "invalid threshold signature ID")
		}
		if err := sigs[i].Deserialize(sigBytes); err != nil {
			return nil, errors.Wrap(err, "invalid signature")
		}
	}

	var compositeSig bls.Sign
	if err := compositeSig.Recover(sigs, ids); err != nil {
		return nil, err
	}

	return &compositeSig, nil
}

func generateAggregateSignature() (*bls.Sign, error) {
	sigs := make([]bls.Sign, len(signatureAggregateSignatures))
	for i := range signatureAggregateSignatures {
		sigBytes, err := hex.DecodeString(strings.TrimPrefix(signatureAggregateSignatures[i], "0x"))
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode signature")
		}
		if err := sigs[i].Deserialize(sigBytes); err != nil {
			return nil, errors.Wrap(err, "invalid signature")
		}
	}
	var aggregateSig bls.Sign
	aggregateSig.Aggregate(sigs)

	return &aggregateSig, nil
}

func init() {
	signatureCmd.AddCommand(signatureAggregateCmd)
	signatureAggregateCmd.Flags().StringArrayVar(&signatureAggregateSignatures, "signature", nil, "a signature to aggregate (supply once for each signature)")
	signatureFlags(signatureAggregateCmd)
}
