// Copyright © 2019 - 2023 Weald Technology Trading.
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

package blockinfo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	eth2client "github.com/attestantio/go-eth2-client"
	"github.com/attestantio/go-eth2-client/api"
	apiv1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/altair"
	"github.com/attestantio/go-eth2-client/spec/bellatrix"
	"github.com/attestantio/go-eth2-client/spec/capella"
	"github.com/attestantio/go-eth2-client/spec/deneb"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
	standardchaintime "github.com/wealdtech/ethdo/services/chaintime/standard"
)

var (
	jsonOutput bool
	sszOutput  bool
	results    *dataOut
)

func process(ctx context.Context, data *dataIn) (*dataOut, error) {
	if data == nil {
		return nil, errors.New("no data")
	}
	if data.blockID == "" && data.blockTime == "" {
		return nil, errors.New("no block ID or block time")
	}

	results = &dataOut{
		debug:      data.debug,
		verbose:    data.verbose,
		eth2Client: data.eth2Client,
	}

	specResponse, err := results.eth2Client.(eth2client.SpecProvider).Spec(ctx, &api.SpecOpts{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to obtain configuration information")
	}
	genesisResponse, err := results.eth2Client.(eth2client.GenesisProvider).Genesis(ctx, &api.GenesisOpts{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to obtain genesis information")
	}
	genesis := genesisResponse.Data
	results.genesisTime = genesis.GenesisTime
	results.slotDuration = specResponse.Data["SECONDS_PER_SLOT"].(time.Duration)
	results.slotsPerEpoch = specResponse.Data["SLOTS_PER_EPOCH"].(uint64)

	if data.blockTime != "" {
		data.blockID, err = timeToBlockID(ctx, data.eth2Client, data.blockTime)
		if err != nil {
			return nil, err
		}
	}

	blockResponse, err := results.eth2Client.(eth2client.SignedBeaconBlockProvider).SignedBeaconBlock(ctx, &api.SignedBeaconBlockOpts{
		Block: data.blockID,
	})
	if err != nil {
		var apiErr *api.Error
		if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
			if data.quiet {
				os.Exit(1)
			}

			return nil, errors.New("empty beacon block")
		}

		return nil, errors.Wrap(err, "failed to obtain beacon block")
	}
	block := blockResponse.Data
	if data.quiet {
		os.Exit(0)
	}

	switch block.Version {
	case spec.DataVersionPhase0:
		if err := outputPhase0Block(ctx, data.jsonOutput, block.Phase0); err != nil {
			return nil, errors.Wrap(err, "failed to output block")
		}
	case spec.DataVersionAltair:
		if err := outputAltairBlock(ctx, data.jsonOutput, data.sszOutput, block.Altair); err != nil {
			return nil, errors.Wrap(err, "failed to output block")
		}
	case spec.DataVersionBellatrix:
		if err := outputBellatrixBlock(ctx, data.jsonOutput, data.sszOutput, block.Bellatrix); err != nil {
			return nil, errors.Wrap(err, "failed to output block")
		}
	case spec.DataVersionCapella:
		if err := outputCapellaBlock(ctx, data.jsonOutput, data.sszOutput, block.Capella); err != nil {
			return nil, errors.Wrap(err, "failed to output block")
		}
	case spec.DataVersionDeneb:
		blobSidecarsResponse, err := results.eth2Client.(eth2client.BlobSidecarsProvider).BlobSidecars(ctx, &api.BlobSidecarsOpts{
			Block: data.blockID,
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to obtain blob sidecars")
		}
		blobSidecars := blobSidecarsResponse.Data
		if err := outputDenebBlock(ctx, data.jsonOutput, data.sszOutput, block.Deneb, blobSidecars); err != nil {
			return nil, errors.Wrap(err, "failed to output block")
		}
	default:
		return nil, errors.New("unknown block version")
	}

	if data.stream {
		jsonOutput = data.jsonOutput
		sszOutput = data.sszOutput
		if !jsonOutput && !sszOutput {
			fmt.Println("")
		}
		err := data.eth2Client.(eth2client.EventsProvider).Events(ctx, []string{"head"}, headEventHandler)
		if err != nil {
			return nil, errors.Wrap(err, "failed to start block stream")
		}
		<-ctx.Done()
	}

	return &dataOut{}, nil
}

func headEventHandler(event *apiv1.Event) {
	ctx := context.Background()

	// Only interested in head events.
	if event.Topic != "head" {
		return
	}

	blockID := fmt.Sprintf("%#x", event.Data.(*apiv1.HeadEvent).Block[:])
	blockResponse, err := results.eth2Client.(eth2client.SignedBeaconBlockProvider).SignedBeaconBlock(ctx, &api.SignedBeaconBlockOpts{
		Block: blockID,
	})
	if err != nil {
		if !jsonOutput && !sszOutput {
			fmt.Printf("Failed to obtain block: %v\n", err)
		}
		return
	}
	block := blockResponse.Data
	if block == nil {
		if !jsonOutput && !sszOutput {
			fmt.Println("Empty beacon block")
		}
		return
	}

	switch block.Version {
	case spec.DataVersionPhase0:
		err = outputPhase0Block(ctx, jsonOutput, block.Phase0)
	case spec.DataVersionAltair:
		err = outputAltairBlock(ctx, jsonOutput, sszOutput, block.Altair)
	case spec.DataVersionBellatrix:
		err = outputBellatrixBlock(ctx, jsonOutput, sszOutput, block.Bellatrix)
	case spec.DataVersionCapella:
		err = outputCapellaBlock(ctx, jsonOutput, sszOutput, block.Capella)
	case spec.DataVersionDeneb:
		var blobSidecarsResponse *api.Response[[]*deneb.BlobSidecar]
		blobSidecarsResponse, err = results.eth2Client.(eth2client.BlobSidecarsProvider).BlobSidecars(ctx, &api.BlobSidecarsOpts{
			Block: blockID,
		})
		if err == nil {
			err = outputDenebBlock(context.Background(), jsonOutput, sszOutput, block.Deneb, blobSidecarsResponse.Data)
		}
	default:
		err = errors.New("unknown block version")
	}
	if err != nil && !jsonOutput && !sszOutput {
		fmt.Printf("Failed to output block: %v\n", err)
		return
	}

	if !jsonOutput && !sszOutput {
		fmt.Println("")
	}
}

func outputPhase0Block(ctx context.Context, jsonOutput bool, signedBlock *phase0.SignedBeaconBlock) error {
	switch {
	case jsonOutput:
		data, err := json.Marshal(signedBlock)
		if err != nil {
			return errors.Wrap(err, "failed to generate JSON")
		}
		fmt.Printf("%s\n", string(data))
	default:
		data, err := outputPhase0BlockText(ctx, results, signedBlock)
		if err != nil {
			return errors.Wrap(err, "failed to generate text")
		}
		fmt.Print(data)
	}
	return nil
}

func outputAltairBlock(ctx context.Context, jsonOutput bool, sszOutput bool, signedBlock *altair.SignedBeaconBlock) error {
	switch {
	case jsonOutput:
		data, err := json.Marshal(signedBlock)
		if err != nil {
			return errors.Wrap(err, "failed to generate JSON")
		}
		fmt.Printf("%s\n", string(data))
	case sszOutput:
		data, err := signedBlock.MarshalSSZ()
		if err != nil {
			return errors.Wrap(err, "failed to generate SSZ")
		}
		fmt.Printf("%x\n", data)
	default:
		data, err := outputAltairBlockText(ctx, results, signedBlock)
		if err != nil {
			return errors.Wrap(err, "failed to generate text")
		}
		fmt.Print(data)
	}
	return nil
}

func outputBellatrixBlock(ctx context.Context, jsonOutput bool, sszOutput bool, signedBlock *bellatrix.SignedBeaconBlock) error {
	switch {
	case jsonOutput:
		data, err := json.Marshal(signedBlock)
		if err != nil {
			return errors.Wrap(err, "failed to generate JSON")
		}
		fmt.Printf("%s\n", string(data))
	case sszOutput:
		data, err := signedBlock.MarshalSSZ()
		if err != nil {
			return errors.Wrap(err, "failed to generate SSZ")
		}
		fmt.Printf("%x\n", data)
	default:
		data, err := outputBellatrixBlockText(ctx, results, signedBlock)
		if err != nil {
			return errors.Wrap(err, "failed to generate text")
		}
		fmt.Print(data)
	}
	return nil
}

func outputCapellaBlock(ctx context.Context, jsonOutput bool, sszOutput bool, signedBlock *capella.SignedBeaconBlock) error {
	switch {
	case jsonOutput:
		data, err := json.Marshal(signedBlock)
		if err != nil {
			return errors.Wrap(err, "failed to generate JSON")
		}
		fmt.Printf("%s\n", string(data))
	case sszOutput:
		data, err := signedBlock.MarshalSSZ()
		if err != nil {
			return errors.Wrap(err, "failed to generate SSZ")
		}
		fmt.Printf("%x\n", data)
	default:
		data, err := outputCapellaBlockText(ctx, results, signedBlock)
		if err != nil {
			return errors.Wrap(err, "failed to generate text")
		}
		fmt.Print(data)
	}
	return nil
}

func outputDenebBlock(ctx context.Context,
	jsonOutput bool,
	sszOutput bool,
	signedBlock *deneb.SignedBeaconBlock,
	blobs []*deneb.BlobSidecar,
) error {
	switch {
	case jsonOutput:
		data, err := json.Marshal(signedBlock)
		if err != nil {
			return errors.Wrap(err, "failed to generate JSON")
		}
		fmt.Printf("%s\n", string(data))
	case sszOutput:
		data, err := signedBlock.MarshalSSZ()
		if err != nil {
			return errors.Wrap(err, "failed to generate SSZ")
		}
		fmt.Printf("%x\n", data)
	default:
		data, err := outputDenebBlockText(ctx, results, signedBlock, blobs)
		if err != nil {
			return errors.Wrap(err, "failed to generate text")
		}
		fmt.Print(data)
	}
	return nil
}

func timeToBlockID(ctx context.Context, eth2Client eth2client.Service, input string) (string, error) {
	var timestamp time.Time

	switch {
	case strings.HasPrefix(input, "0x"):
		// Hex string.
		hexTime, err := strconv.ParseInt(strings.TrimPrefix(input, "0x"), 16, 64)
		if err != nil {
			return "", errors.Wrap(err, "failed to parse block time as hex string")
		}
		timestamp = time.Unix(hexTime, 0)
	case !strings.Contains(input, ":"):
		// No colon, assume decimal string.
		decTime, err := strconv.ParseInt(input, 10, 64)
		if err != nil {
			return "", errors.Wrap(err, "failed to parse block time as decimal string")
		}
		timestamp = time.Unix(decTime, 0)
	default:
		dateTime, err := time.Parse("2006-01-02T15:04:05", input)
		if err != nil {
			return "", errors.Wrap(err, "failed to parse block time as datetime")
		}
		timestamp = dateTime
	}

	// Assume timestamp.
	chainTime, err := standardchaintime.New(ctx,
		standardchaintime.WithSpecProvider(eth2Client.(eth2client.SpecProvider)),
		standardchaintime.WithGenesisProvider(eth2Client.(eth2client.GenesisProvider)),
	)
	if err != nil {
		return "", errors.Wrap(err, "failed to set up chaintime service")
	}

	return fmt.Sprintf("%d", chainTime.TimestampToSlot(timestamp)), nil
}
