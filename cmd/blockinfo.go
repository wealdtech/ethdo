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
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/prysmaticlabs/go-bitfield"
	"github.com/prysmaticlabs/go-ssz"
	"github.com/spf13/cobra"
	"github.com/wealdtech/ethdo/grpc"
	string2eth "github.com/wealdtech/go-string2eth"
)

var blockInfoSlot int64
var blockInfoStream bool

var blockInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Obtain information about a block",
	Long: `Obtain information about a block.  For example:

    ethdo block info --slot=12345

In quiet mode this will return 0 if the block information is present and not skipped, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := connect()
		errCheck(err, "Failed to obtain connection to Ethereum 2 beacon chain block")

		config, err := grpc.FetchChainConfig(eth2GRPCConn)
		errCheck(err, "Failed to obtain beacon chain configuration")
		slotsPerEpoch := config["SlotsPerEpoch"].(uint64)
		secondsPerSlot := config["SecondsPerSlot"].(uint64)

		genesisTime, err := grpc.FetchGenesisTime(eth2GRPCConn)
		errCheck(err, "Failed to obtain beacon chain genesis")

		assert(blockInfoStream || blockInfoSlot != 0, "--slot or --stream is required")
		assert(!blockInfoStream || blockInfoSlot == -1, "--slot and --stream are not supported together")

		var slot uint64
		if blockInfoSlot < 0 {
			slot, err = grpc.FetchLatestFilledSlot(eth2GRPCConn)
			errCheck(err, "Failed to obtain slot of latest block")
		} else {
			slot = uint64(blockInfoSlot)
		}
		assert(slot > 0, "slot must be greater than 0")

		signedBlock, err := grpc.FetchBlock(eth2GRPCConn, slot)
		errCheck(err, "Failed to obtain block")
		if signedBlock == nil {
			outputIf(!quiet, "No block at that slot")
			os.Exit(_exitFailure)
		}
		outputBlock(signedBlock, genesisTime, secondsPerSlot, slotsPerEpoch)

		if blockInfoStream {
			stream, err := grpc.StreamBlocks(eth2GRPCConn)
			errCheck(err, "Failed to obtain block stream")
			for {
				fmt.Println()
				signedBlock, err := stream.Recv()
				errCheck(err, "Failed to obtain block")
				if signedBlock != nil {
					outputBlock(signedBlock, genesisTime, secondsPerSlot, slotsPerEpoch)
				}
			}
		}

		os.Exit(_exitSuccess)
	},
}

func outputBlock(signedBlock *ethpb.SignedBeaconBlock, genesisTime time.Time, secondsPerSlot uint64, slotsPerEpoch uint64) {
	block := signedBlock.Block
	body := block.Body

	// General info.
	bodyRoot, err := ssz.HashTreeRoot(block)
	errCheck(err, "Failed to calculate block body root")
	fmt.Printf("Slot: %d\n", block.Slot)
	fmt.Printf("Epoch: %d\n", block.Slot/slotsPerEpoch)
	fmt.Printf("Timestamp: %v\n", time.Unix(genesisTime.Unix()+int64(block.Slot*secondsPerSlot), 0))
	fmt.Printf("Block root: %#x\n", bodyRoot)
	outputIf(verbose, fmt.Sprintf("Parent root: %#x", block.ParentRoot))
	outputIf(verbose, fmt.Sprintf("State root: %#x", block.StateRoot))
	if len(body.Graffiti) > 0 && hex.EncodeToString(body.Graffiti) != "0000000000000000000000000000000000000000000000000000000000000000" {
		if utf8.Valid(body.Graffiti) {
			fmt.Printf("Graffiti: %s\n", string(body.Graffiti))
		} else {
			fmt.Printf("Graffiti: %#x\n", body.Graffiti)
		}
	}

	// Eth1 data.
	eth1Data := body.Eth1Data
	outputIf(verbose, fmt.Sprintf("Ethereum 1 deposit count: %d", eth1Data.DepositCount))
	outputIf(verbose, fmt.Sprintf("Ethereum 1 deposit root: %#x", eth1Data.DepositRoot))
	outputIf(verbose, fmt.Sprintf("Ethereum 1 block hash: %#x", eth1Data.BlockHash))

	validatorCommittees := make(map[uint64][][]uint64)

	// Attestations.
	fmt.Printf("Attestations: %d\n", len(body.Attestations))
	if verbose {
		for i, att := range body.Attestations {
			fmt.Printf("\t%d:\n", i)

			// Fetch committees for this epoch if not already obtained.
			committees, exists := validatorCommittees[att.Data.Slot]
			if !exists {
				attestationEpoch := att.Data.Slot / slotsPerEpoch
				epochCommittees, err := grpc.FetchValidatorCommittees(eth2GRPCConn, attestationEpoch)
				errCheck(err, "Failed to obtain committees")
				for k, v := range epochCommittees {
					validatorCommittees[k] = v
				}
				committees = validatorCommittees[att.Data.Slot]
			}

			fmt.Printf("\t\tCommittee index: %d\n", att.Data.CommitteeIndex)
			fmt.Printf("\t\tAttesters: %d/%d\n", att.AggregationBits.Count(), att.AggregationBits.Len())
			fmt.Printf("\t\tAggregation bits: %s\n", bitsToString(att.AggregationBits))
			fmt.Printf("\t\tAttesting indices: %s\n", attestingIndices(att.AggregationBits, committees[att.Data.CommitteeIndex]))
			fmt.Printf("\t\tSlot: %d\n", att.Data.Slot)
			fmt.Printf("\t\tBeacon block root: %#x\n", att.Data.BeaconBlockRoot)
			fmt.Printf("\t\tSource epoch: %d\n", att.Data.Source.Epoch)
			fmt.Printf("\t\tSource root: %#x\n", att.Data.Source.Root)
			fmt.Printf("\t\tTarget epoch: %d\n", att.Data.Target.Epoch)
			fmt.Printf("\t\tTarget root: %#x\n", att.Data.Target.Root)
		}
	}

	// Attester slashings.
	fmt.Printf("Attester slashings: %d\n", len(body.AttesterSlashings))
	if verbose {
		for i, slashing := range body.AttesterSlashings {
			// Say what was slashed.
			att1 := slashing.Attestation_1
			outputIf(debug, fmt.Sprintf("Attestation 1 attesting indices are %v", att1.AttestingIndices))
			att2 := slashing.Attestation_2
			outputIf(debug, fmt.Sprintf("Attestation 2 attesting indices are %v", att2.AttestingIndices))
			slashedIndices := intersection(att1.AttestingIndices, att2.AttestingIndices)
			if len(slashedIndices) == 0 {
				continue
			}

			fmt.Printf("\t%d:\n", i)

			fmt.Println("\t\tSlashed validators:")
			for _, slashedIndex := range slashedIndices {
				validator, err := grpc.FetchValidatorByIndex(eth2GRPCConn, slashedIndex)
				errCheck(err, "Failed to obtain validator information")
				fmt.Printf("\t\t\t%#x (%d)\n", validator.PublicKey, slashedIndex)
			}

			// Say what caused the slashing.
			if att1.Data.Target.Epoch == att2.Data.Target.Epoch {
				fmt.Printf("\t\tDouble voted for same target epoch (%d):\n", att1.Data.Target.Epoch)
				if !bytes.Equal(att1.Data.Target.Root, att2.Data.Target.Root) {
					fmt.Printf("\t\t\tAttestation 1 target epoch root: %#x\n", att1.Data.Target.Root)
					fmt.Printf("\t\t\tAttestation 2target epoch root: %#x\n", att2.Data.Target.Root)
				}
				if !bytes.Equal(att1.Data.BeaconBlockRoot, att2.Data.BeaconBlockRoot) {
					fmt.Printf("\t\t\tAttestation 1 beacon block root: %#x\n", att1.Data.BeaconBlockRoot)
					fmt.Printf("\t\t\tAttestation 2 beacon block root: %#x\n", att2.Data.BeaconBlockRoot)
				}
			} else if att1.Data.Source.Epoch < att2.Data.Source.Epoch &&
				att1.Data.Target.Epoch > att2.Data.Target.Epoch {
				fmt.Printf("\t\tSurround voted:\n")
				fmt.Printf("\t\t\tAttestation 1 vote: %d->%d\n", att1.Data.Source.Epoch, att1.Data.Target.Epoch)
				fmt.Printf("\t\t\tAttestation 2 vote: %d->%d\n", att2.Data.Source.Epoch, att2.Data.Target.Epoch)
			}
		}
	}

	fmt.Printf("Proposer slashings: %d\n", len(body.ProposerSlashings))
	// TODO verbose proposer slashings.

	// Deposits.
	fmt.Printf("Deposits: %d\n", len(body.Deposits))
	if verbose {
		for i, deposit := range body.Deposits {
			data := deposit.Data
			fmt.Printf("\t%d:\n", i)
			fmt.Printf("\t\tPublic key: %#x\n", data.PublicKey)
			fmt.Printf("\t\tAmount: %s\n", string2eth.GWeiToString(data.Amount, true))
			fmt.Printf("\t\tWithdrawal credentials: %#x\n", data.WithdrawalCredentials)
			fmt.Printf("\t\tSignature: %#x\n", data.Signature)
		}
	}

	// Voluntary exits.
	fmt.Printf("Voluntary exits: %d\n", len(body.VoluntaryExits))
	if verbose {
		for i, voluntaryExit := range body.VoluntaryExits {
			fmt.Printf("\t%d:\n", i)
			validator, err := grpc.FetchValidatorByIndex(eth2GRPCConn, voluntaryExit.Exit.ValidatorIndex)
			errCheck(err, "Failed to obtain validator information")
			fmt.Printf("\t\tValidator: %#x (%d)\n", validator.PublicKey, voluntaryExit.Exit.ValidatorIndex)
			fmt.Printf("\t\tEpoch: %d\n", voluntaryExit.Exit.Epoch)
		}
	}
}

// intersection returns a list of items common between the two sets.
func intersection(set1 []uint64, set2 []uint64) []uint64 {
	sort.Slice(set1, func(i, j int) bool { return set1[i] < set1[j] })
	sort.Slice(set2, func(i, j int) bool { return set2[i] < set2[j] })
	res := make([]uint64, 0)

	set1Pos := 0
	set2Pos := 0
	for set1Pos < len(set1) && set2Pos < len(set2) {
		switch {
		case set1[set1Pos] < set2[set2Pos]:
			set1Pos++
		case set2[set2Pos] < set1[set1Pos]:
			set2Pos++
		default:
			res = append(res, set1[set1Pos])
			set1Pos++
			set2Pos++
		}
	}

	return res
}

func bitsToString(input bitfield.Bitlist) string {
	bits := int(input.Len())

	res := ""
	for i := 0; i < bits; i++ {
		if input.BitAt(uint64(i)) {
			res = fmt.Sprintf("%s✓", res)
		} else {
			res = fmt.Sprintf("%s✕", res)
		}
		if i%8 == 7 {
			res = fmt.Sprintf("%s ", res)
		}
	}
	return strings.TrimSpace(res)
}

func attestingIndices(input bitfield.Bitlist, indices []uint64) string {
	bits := int(input.Len())
	res := ""
	for i := 0; i < bits; i++ {
		if input.BitAt(uint64(i)) {
			res = fmt.Sprintf("%s%d ", res, indices[i])
		}
	}
	return strings.TrimSpace(res)
}

func init() {
	blockCmd.AddCommand(blockInfoCmd)
	blockFlags(blockInfoCmd)
	blockInfoCmd.Flags().Int64Var(&blockInfoSlot, "slot", -1, "the latest slot with a block")
	blockInfoCmd.Flags().BoolVar(&blockInfoStream, "stream", false, "continually stream blocks as they arrive")
}
