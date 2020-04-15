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
	"fmt"
	"os"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/spf13/cobra"
	"github.com/wealdtech/ethdo/grpc"
	string2eth "github.com/wealdtech/go-string2eth"
)

var blockInfoSlot int64

var blockInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Obtain information about a block",
	Long: `Obtain information about a block.  For example:

    ethdo block info --slot=12345

In quiet mode this will return 0 if the block information is present and not skipped, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := connect()
		errCheck(err, "Failed to obtain connection to Ethereum 2 beacon chain block")

		assert(blockInfoSlot != 0, "--slot is required")

		var slot uint64
		if blockInfoSlot < 0 {
			// TODO latest block.
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
		block := signedBlock.Block
		body := block.Body

		// General info.
		outputIf(verbose, fmt.Sprintf("Parent root: %#x", block.ParentRoot))
		outputIf(verbose, fmt.Sprintf("State root: %#x", block.StateRoot))
		if utf8.Valid(body.Graffiti) {
			fmt.Printf("Graffiti: %s\n", string(body.Graffiti))
		} else {
			fmt.Printf("Graffiti: %#x\n", body.Graffiti)
		}

		// Eth1 data.
		eth1Data := body.Eth1Data
		fmt.Printf("Ethereum 1 deposit count: %d\n", eth1Data.DepositCount)
		outputIf(verbose, fmt.Sprintf("Ethereum 1 deposit root: %#x", eth1Data.DepositRoot))
		outputIf(verbose, fmt.Sprintf("Ethereum 1 block hash: %#x", eth1Data.BlockHash))

		// Attestations.
		fmt.Printf("Attestations: %d\n", len(body.Attestations))
		if verbose {
			for i, att := range body.Attestations {
				fmt.Printf("\t%d:\n", i)

				fmt.Printf("\t\tCommittee index: %d\n", att.Data.CommitteeIndex)
				fmt.Printf("\t\tAttesters: %d\n", countSetBits(att.AggregationBits))
				fmt.Printf("\t\tAggregation bits: %s\n", bitsToString(att.AggregationBits))
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
				fmt.Printf("\t%d:\n", i)

				// Say what was slashed.
				att1 := slashing.Attestation_1
				att2 := slashing.Attestation_2
				slashedIndices := intersection(att1.AttestingIndices, att2.AttestingIndices)
				if len(slashedIndices) == 0 {
					continue
				}
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
				} else {
					if att1.Data.Source.Epoch < att2.Data.Source.Epoch &&
						att1.Data.Target.Epoch > att2.Data.Target.Epoch {
						fmt.Printf("\t\tSurround voted:\n")
						fmt.Printf("\t\t\tAttestation 1 vote: %d->%d\n", att1.Data.Source.Epoch, att1.Data.Target.Epoch)
						fmt.Printf("\t\t\tAttestation 2 vote: %d->%d\n", att2.Data.Source.Epoch, att2.Data.Target.Epoch)
					}
				}
			}
		}

		// TODO Proposer slashings once proposer slashings exist.

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

		os.Exit(_exitSuccess)
	},
}

// intersection returns a list of items common between the two sets.
func intersection(set1 []uint64, set2 []uint64) []uint64 {
	sort.Slice(set1, func(i, j int) bool { return set1[i] < set1[j] })
	sort.Slice(set2, func(i, j int) bool { return set2[i] < set2[j] })
	res := make([]uint64, 0)
	if len(set1) < len(set2) {
		set1, set2 = set2, set1
	}

	set2Pos := 0
	set2LastIndex := len(set2) - 1
	for set1Pos := range set1 {
		for set1[set1Pos] == set2[set2Pos] {
			res = append(res, set1[set1Pos])
			if set2Pos == set2LastIndex {
				break
			}
			set2Pos++
		}
		for set1[set1Pos] > set2[set2Pos] {
			if set2Pos == set2LastIndex {
				break
			}
			set2Pos++
		}
	}

	return res
}

// countSetBits counts the number of bits that are set in the given byte array.
func countSetBits(input []byte) int {
	total := 0
	for _, x := range input {
		item := uint8(x)
		for item > 0 {
			if item&0x01 == 1 {
				total++
			}
			item >>= 1
		}
	}
	return total
}

func bitsToString(input []byte) string {
	elements := make([]string, len(input))
	for i, x := range input {
		item := uint8(x)
		mask := uint8(0x80)
		element := ""
		for mask > 0 {
			if item&mask != 0 {
				element = fmt.Sprintf("%s✓", element)
			} else {
				element = fmt.Sprintf("%s✕", element)
			}
			mask >>= 1
		}
		elements[i] = element
	}
	return strings.Join(elements, " ")
}

func init() {
	blockCmd.AddCommand(blockInfoCmd)
	blockFlags(blockInfoCmd)
	blockInfoCmd.Flags().Int64Var(&blockInfoSlot, "slot", -1, "the default slot")

}
