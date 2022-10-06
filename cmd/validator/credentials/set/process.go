// Copyright © 2022 Weald Technology Trading.
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

package validatorcredentialsset

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	consensusclient "github.com/attestantio/go-eth2-client"
	"github.com/attestantio/go-eth2-client/spec/bellatrix"
	capella "github.com/attestantio/go-eth2-client/spec/capella"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
	standardchaintime "github.com/wealdtech/ethdo/services/chaintime/standard"
	"github.com/wealdtech/ethdo/signing"
	"github.com/wealdtech/ethdo/util"
	ethutil "github.com/wealdtech/go-eth2-util"
)

func (c *command) process(ctx context.Context) error {
	if err := c.setup(ctx); err != nil {
		return err
	}

	if err := c.obtainOp(ctx); err != nil {
		return err
	}

	if c.json || c.offline {
		// Want JSON output, or cannot broadcast.
		return nil
	}

	if validated, reason := c.validateOp(ctx); !validated {
		return fmt.Errorf("operation failed validation: %s", reason)
	}

	return c.broadcastOp(ctx)
}

func (c *command) obtainOp(ctx context.Context) error {
	// See if we have been given an op.
	if c.signedOperation != "" {
		// Input could be JSON or a path to JSON.
		switch {
		case strings.HasPrefix(c.signedOperation, "{"):
			// Looks like JSON, nothing to do.
		default:
			// Assume it's a path to JSON
			data, err := os.ReadFile(c.signedOperation)
			if err != nil {
				return errors.Wrap(err, "failed to read signed operation file")
			}
			c.signedOperation = string(data)
		}
		// Unmarshal it to confirm it is valid.
		signedOp := &capella.SignedBLSToExecutionChange{}
		if err := json.Unmarshal([]byte(c.signedOperation), signedOp); err != nil {
			return err
		}
		return nil
	}

	// Need to create a new op.
	if err := c.fetchAccount(ctx); err != nil {
		return err
	}
	pubkey, err := util.BestPublicKey(c.withdrawalAccount)
	if err != nil {
		return err
	}
	blsPubkey := phase0.BLSPubKey{}
	copy(blsPubkey[:], pubkey.Marshal())

	withdrawalAddressBytes, err := hex.DecodeString(strings.TrimPrefix(c.withdrawalAddress, "0x"))
	if err != nil {
		return errors.Wrap(err, "failed to obtain execution address")
	}
	if len(withdrawalAddressBytes) != bellatrix.ExecutionAddressLength {
		return errors.New("withdrawal address must be exactly 20 bytes in length")
	}
	// Ensure the address is properly checksummed.
	checksummedAddress := addressBytesToEIP55(withdrawalAddressBytes)
	if checksummedAddress != c.withdrawalAddress {
		return fmt.Errorf("withdrawal address checksum does not match (expected %s)", checksummedAddress)
	}
	withdrawalAddress := bellatrix.ExecutionAddress{}
	copy(withdrawalAddress[:], withdrawalAddressBytes)

	if c.offline {
		err = c.obtainOpOffline(ctx, blsPubkey, withdrawalAddress)
	} else {
		err = c.obtainOpOnline(ctx, blsPubkey, withdrawalAddress)
	}
	if err != nil {
		return errors.Wrap(err, "failed to obtain operation")
	}

	root, err := c.op.HashTreeRoot()
	if err != nil {
		return errors.Wrap(err, "failed to generate root for credentials change operation")
	}

	// Sign the operation.
	signature, err := signing.SignRoot(ctx, c.withdrawalAccount, nil, root, c.domain)
	if err != nil {
		return errors.Wrap(err, "failed to sign credentials change operation")
	}

	c.signedOp = &capella.SignedBLSToExecutionChange{
		Message:   c.op,
		Signature: signature,
	}

	return nil
}

func (c *command) obtainOpOffline(ctx context.Context,
	pubkey phase0.BLSPubKey,
	withdrawalAddress bellatrix.ExecutionAddress,
) error {
	if c.validator == "" {
		return errors.New("validator index must be supplied when offline")
	}
	validatorIndex, err := strconv.ParseUint(c.validator, 10, 64)
	if err != nil {
		return errors.Wrap(err, "validator must be an index when offline")
	}

	if c.forkVersion == "" {
		return errors.New("fork version must be supplied when offline")
	}
	forkVersionBytes, err := hex.DecodeString(strings.TrimPrefix(c.forkVersion, "0x"))
	if err != nil {
		return errors.Wrap(err, "fork version invalid")
	}
	if len(forkVersionBytes) != phase0.ForkVersionLength {
		return errors.New("fork version incorrect length")
	}
	forkVersion := phase0.Version{}
	copy(forkVersion[:], forkVersionBytes)

	if c.genesisValidatorsRoot == "" {
		return errors.New("genesis validators root must be supplied when offline")
	}
	genesisValidatorsRootBytes, err := hex.DecodeString(strings.TrimPrefix(c.genesisValidatorsRoot, "0x"))
	if err != nil {
		return errors.Wrap(err, "genesis validators root invalid")
	}
	if len(genesisValidatorsRootBytes) != phase0.RootLength {
		return errors.New("genesis validators root incorrect length")
	}
	genesisValidatorsRoot := phase0.Root{}
	copy(genesisValidatorsRoot[:], genesisValidatorsRootBytes)

	// Generate the domain.
	forkData := &phase0.ForkData{
		CurrentVersion:        forkVersion,
		GenesisValidatorsRoot: genesisValidatorsRoot,
	}
	root, err := forkData.HashTreeRoot()
	if err != nil {
		return errors.Wrap(err, "failed to calculate signature domain")
	}
	c.domain = phase0.Domain{}
	copy(c.domain[:], []byte{0x0a, 0x00, 0x00, 0x00}) // DOMAIN_BLS_TO_EXECUTION_CHANGE.
	copy(c.domain[4:], root[:])

	// Generate the change operation.
	c.op = &capella.BLSToExecutionChange{
		ValidatorIndex:     phase0.ValidatorIndex(validatorIndex),
		FromBLSPubkey:      pubkey,
		ToExecutionAddress: withdrawalAddress,
	}

	return nil
}

func (c *command) obtainOpOnline(ctx context.Context,
	pubkey phase0.BLSPubKey,
	withdrawalAddress bellatrix.ExecutionAddress,
) error {
	// Ensure the validator is correct and suitable.
	if err := c.fetchChainInfo(ctx); err != nil {
		return err
	}
	// TODO Move to broadcast.
	if c.validatorInfo.Validator.WithdrawalCredentials[0] != 0x00 {
		return errors.New("validator withdrawal credentials are not using BLS; cannot change")
	}
	{
		// TODO remove.
		x, _ := json.Marshal(c.validatorInfo)
		fmt.Printf("%s\n", string(x))
	}

	// Generate the change operation.
	c.op = &capella.BLSToExecutionChange{
		ValidatorIndex:     c.validatorInfo.Index,
		FromBLSPubkey:      pubkey,
		ToExecutionAddress: withdrawalAddress,
	}

	return nil
}

func (c *command) validateOp(ctx context.Context,
) (
	bool,
	string,
) {
	// Confirm that the public key hashes to the existing withdrawal credentials (if available).
	if c.validatorInfo != nil {
		pubkey, err := util.BestPublicKey(c.withdrawalAccount)
		if err != nil {
			return false, "failed to obtain a public key for the withdrawal account"
		}
		blsHash := ethutil.Keccak256(pubkey.Marshal())
		// TODO remove.
		fmt.Printf("BLS pub key is %#x, hash is %#x\n", pubkey, blsHash)
		if !bytes.Equal(blsHash[1:], c.validatorInfo.Validator.WithdrawalCredentials[:]) {
			return false, "validator withdrawal credentials do not match current withdrawal credentials"
		}
	}

	return true, ""
}

func (c *command) broadcastOp(ctx context.Context) error {
	// Broadcast the operation.
	return c.consensusClient.(consensusclient.BLSToExecutionChangeSubmitter).SubmitBLSToExecutionChange(ctx, c.signedOp)
}

func (c *command) setup(ctx context.Context) error {
	if c.offline {
		return nil
	}

	var err error

	// Connect to the consensus node.
	c.consensusClient, err = util.ConnectToBeaconNode(ctx, c.connection, c.timeout, c.allowInsecureConnections)
	if err != nil {
		return errors.Wrap(err, "failed to connect to consensus node")
	}

	// Set up chaintime.
	c.chainTime, err = standardchaintime.New(ctx,
		standardchaintime.WithGenesisTimeProvider(c.consensusClient.(consensusclient.GenesisTimeProvider)),
		standardchaintime.WithForkScheduleProvider(c.consensusClient.(consensusclient.ForkScheduleProvider)),
		standardchaintime.WithSpecProvider(c.consensusClient.(consensusclient.SpecProvider)),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create chaintime service")
	}

	return nil
}

func (c *command) fetchChainInfo(ctx context.Context) error {
	var err error

	// Obtain the validators provider.
	validatorsProvider, isProvider := c.consensusClient.(consensusclient.ValidatorsProvider)
	if !isProvider {
		return errors.New("consensus node does not provide validator information")
	}

	c.validatorInfo, err = util.ParseValidator(ctx, validatorsProvider, c.validator, "head")
	if err != nil {
		return errors.Wrap(err, "failed to obtain validator")
	}

	epoch := c.chainTime.CurrentEpoch()

	// Obtain the domain type.
	spec, err := c.consensusClient.(consensusclient.SpecProvider).Spec(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to obtain spec")
	}
	domainType, exists := spec["DOMAIN_BLS_TO_EXECUTION_CHANGE"].(phase0.DomainType)
	if !exists {
		return errors.New("failed to obtain DOMAIN_BLS_TO_EXECUTION_CHANGE")
	}

	domainProvider, isProvider := c.consensusClient.(consensusclient.DomainProvider)
	if !isProvider {
		return errors.New("consensus node does not provide domain information")
	}
	c.domain, err = domainProvider.Domain(ctx, domainType, epoch)
	if err != nil {
		return errors.Wrap(err, "failed to obtain domain")
	}

	return nil
}

func (c *command) fetchAccount(ctx context.Context) error {
	var err error

	switch {
	case c.account != "":
		c.withdrawalAccount, err = util.ParseAccount(ctx, c.account, c.passphrases, true)
	case c.mnemonic != "":
		c.withdrawalAccount, err = util.ParseAccount(ctx, c.mnemonic, []string{c.path}, true)
	case c.privateKey != "":
		c.withdrawalAccount, err = util.ParseAccount(ctx, c.privateKey, nil, true)
	default:
		err = errors.New("account, mnemonic or private key must be supplied")
	}

	return err
}

// addressBytesToEIP55 converts a byte array in to an EIP-55 string format.
func addressBytesToEIP55(address []byte) string {
	bytes := []byte(fmt.Sprintf("%x", address))
	hash := ethutil.Keccak256(bytes)
	for i := 0; i < len(bytes); i++ {
		hashByte := hash[i/2]
		if i%2 == 0 {
			hashByte >>= 4
		} else {
			hashByte &= 0xf
		}
		if bytes[i] > '9' && hashByte > 7 {
			bytes[i] -= 32
		}
	}

	return fmt.Sprintf("0x%s", string(bytes))
}
