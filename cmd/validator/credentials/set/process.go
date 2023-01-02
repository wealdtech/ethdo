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
	"regexp"
	"strconv"
	"strings"

	consensusclient "github.com/attestantio/go-eth2-client"
	"github.com/attestantio/go-eth2-client/spec/bellatrix"
	capella "github.com/attestantio/go-eth2-client/spec/capella"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
	"github.com/prysmaticlabs/go-ssz"
	standardchaintime "github.com/wealdtech/ethdo/services/chaintime/standard"
	"github.com/wealdtech/ethdo/signing"
	"github.com/wealdtech/ethdo/util"
	e2types "github.com/wealdtech/go-eth2-types/v2"
	ethutil "github.com/wealdtech/go-eth2-util"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

// validatorPath is the regular expression that matches a validator  path.
var validatorPath = regexp.MustCompile("^m/12381/3600/[0-9]+/0/0$")

var offlinePreparationFilename = "offline-preparation.json"
var changeOperationsFilename = "change-operations.json"

func (c *command) process(ctx context.Context) error {
	if err := c.setup(ctx); err != nil {
		return err
	}

	if err := c.obtainRequiredInformation(ctx); err != nil {
		return err
	}

	if c.prepareOffline {
		return c.dumpRequiredInformation(ctx)
	}

	if err := c.generateOperations(ctx); err != nil {
		return err
	}

	if validated, reason := c.validateOperations(ctx); !validated {
		return fmt.Errorf("operation failed validation: %s", reason)
	}

	if c.json || c.offline {
		if c.debug {
			fmt.Fprintf(os.Stderr, "Not broadcasting credentials change operations\n")
		}
		// Want JSON output, or cannot broadcast.
		return nil
	}

	return c.broadcastOperations(ctx)
}

// obtainRequiredInformation obtains the information required to create a
// withdrawal credentials change operation.
func (c *command) obtainRequiredInformation(ctx context.Context) error {
	c.chainInfo = &chainInfo{
		Validators: make([]*validatorInfo, 0),
	}

	// Use the offline preparation file if present (and we haven't been asked to recreate it).
	if !c.prepareOffline {
		err := c.loadChainInfo(ctx)
		if err == nil {
			return nil
		}
	}

	if c.offline {
		return fmt.Errorf("could not find the %s file; this is required to have been previously generated using --offline-preparation on an online machine and be readable in the directory in which this command is being run", offlinePreparationFilename)
	}

	if err := c.populateChainInfo(ctx); err != nil {
		return err
	}

	return nil
}

// populateChainInfo populates chain info structure from a beacon node.
func (c *command) populateChainInfo(ctx context.Context) error {
	if c.debug {
		fmt.Fprintf(os.Stderr, "Populating chain info from beacon node\n")
	}

	// Obtain validators.
	validators, err := c.consensusClient.(consensusclient.ValidatorsProvider).Validators(ctx, "head", nil)
	if err != nil {
		return errors.Wrap(err, "failed to obtain validators")
	}

	for _, validator := range validators {
		c.chainInfo.Validators = append(c.chainInfo.Validators, &validatorInfo{
			Index:                 validator.Index,
			Pubkey:                validator.Validator.PublicKey,
			WithdrawalCredentials: validator.Validator.WithdrawalCredentials,
		})
	}

	// Obtain genesis validators root.
	if c.genesisValidatorsRoot != "" {
		// Genesis validators root supplied manually.
		genesisValidatorsRoot, err := hex.DecodeString(strings.TrimPrefix(c.genesisValidatorsRoot, "0x"))
		if err != nil {
			return errors.Wrap(err, "invalid genesis validators root supplied")
		}
		if len(genesisValidatorsRoot) != phase0.RootLength {
			return errors.New("invalid length for genesis validators root")
		}
		copy(c.chainInfo.GenesisValidatorsRoot[:], genesisValidatorsRoot)
	} else {
		// Genesis validators root obtained from beacon node.
		genesis, err := c.consensusClient.(consensusclient.GenesisProvider).Genesis(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to obtain genesis information")
		}
		c.chainInfo.GenesisValidatorsRoot = genesis.GenesisValidatorsRoot
	}
	if c.debug {
		fmt.Fprintf(os.Stderr, "Genesis validators root is %#x\n", c.chainInfo.GenesisValidatorsRoot)
	}

	// Obtain epoch.
	c.chainInfo.Epoch = c.chainTime.CurrentEpoch()

	// Obtain fork version.
	if c.forkVersion != "" {
		if err := c.populateChainInfoForkVersionFromInput(ctx); err != nil {
			return err
		}
	} else {
		if err := c.populateChainInfoForkVersionFromChain(ctx); err != nil {
			return err
		}
	}
	if c.debug {
		fmt.Fprintf(os.Stderr, "Fork version is %#x\n", c.chainInfo.ForkVersion)
	}

	// Calculate domain.
	spec, err := c.consensusClient.(consensusclient.SpecProvider).Spec(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to obtain spec")
	}
	domainType, exists := spec["DOMAIN_BLS_TO_EXECUTION_CHANGE"].(phase0.DomainType)
	if !exists {
		return errors.New("failed to obtain DOMAIN_BLS_TO_EXECUTION_CHANGE")
	}
	if c.debug {
		fmt.Fprintf(os.Stderr, "Domain type is %#x\n", domainType)
	}
	copy(c.chainInfo.Domain[:], domainType[:])

	root, err := (&phase0.ForkData{
		CurrentVersion:        c.chainInfo.ForkVersion,
		GenesisValidatorsRoot: c.chainInfo.GenesisValidatorsRoot,
	}).HashTreeRoot()
	if err != nil {
		return errors.Wrap(err, "failed to calculate signature domain")
	}
	copy(c.chainInfo.Domain[4:], root[:])

	if c.debug {
		fmt.Fprintf(os.Stderr, "Domain is %#x\n", c.chainInfo.Domain)
	}

	return nil
}

func (c *command) populateChainInfoForkVersionFromInput(_ context.Context) error {
	// Fork version supplied manually.
	forkVersion, err := hex.DecodeString(strings.TrimPrefix(c.forkVersion, "0x"))
	if err != nil {
		return errors.Wrap(err, "invalid fork version supplied")
	}
	if len(forkVersion) != phase0.ForkVersionLength {
		return errors.New("invalid length for fork version")
	}
	copy(c.chainInfo.ForkVersion[:], forkVersion)

	return nil
}

func (c *command) populateChainInfoForkVersionFromChain(ctx context.Context) error {
	// Fetch the capella fork version from the specification.
	spec, err := c.consensusClient.(consensusclient.SpecProvider).Spec(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to obtain spec")
	}
	tmp, exists := spec["CAPELLA_FORK_VERSION"]
	if !exists {
		return errors.New("capella fork version not known by chain")
	}
	capellaForkVersion, isForkVersion := tmp.(phase0.Version)
	if !isForkVersion {
		//nolint:revive
		return errors.New("CAPELLA_FORK_VERSION is not a fork version!")
	}
	c.chainInfo.ForkVersion = capellaForkVersion

	// Work through the fork schedule to find the latest current fork post-Capella.
	forkSchedule, err := c.consensusClient.(consensusclient.ForkScheduleProvider).ForkSchedule(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to obtain fork schedule")
	}
	foundCapella := false
	for i := range forkSchedule {
		if foundCapella && forkSchedule[i].Epoch <= c.chainInfo.Epoch {
			c.chainInfo.ForkVersion = forkSchedule[i].CurrentVersion
		}
		if bytes.Equal(forkSchedule[i].CurrentVersion[:], capellaForkVersion[:]) {
			foundCapella = true
		}
	}

	return nil
}

// dumpRequiredInformation prepares for an offline run of this command by dumping
// the chain information to a file.
func (c *command) dumpRequiredInformation(_ context.Context) error {
	data, err := json.Marshal(c.chainInfo)
	if err != nil {
		return err
	}
	if err := os.WriteFile(offlinePreparationFilename, data, 0600); err != nil {
		return err
	}

	return nil
}

func (c *command) generateOperations(ctx context.Context) error {
	if c.account == "" && c.mnemonic == "" && c.privateKey == "" && c.validator == "" {
		// No input information; fetch the operations from a file.
		err := c.loadOperations(ctx)
		if err == nil {
			// Success.
			return nil
		}
		return fmt.Errorf("no account, mnemonic or private key specified and no %s file loaded: %v", changeOperationsFilename, err)
	}

	if c.mnemonic != "" {
		switch {
		case c.path != "":
			// Have a mnemonic and path.
			return c.generateOperationsFromMnemonicAndPath(ctx)
		case c.validator != "":
			// Have a mnemonic and validator.
			return c.generateOperationsFromMnemonicAndValidator(ctx)
		case c.privateKey != "":
			// Have a mnemonic and a private key for the withdrawal address.
			return c.generateOperationsFromMnemonicAndPrivateKey(ctx)
		default:
			// Have a mnemonic and nothing else; scan.
			return c.generateOperationsFromMnemonic(ctx)
		}
	}

	if c.account != "" {
		switch {
		case c.withdrawalAccount != "":
			// Have an account and a withdrawal account.
			return c.generateOperationsFromAccountAndWithdrawalAccount(ctx)
		case c.privateKey != "":
			// Have an account and a private key for the withdrawal address.
			return c.generateOperationsFromAccountAndPrivateKey(ctx)
		}
	}

	if c.validator != "" && c.privateKey != "" {
		// Have a validator and a private key for the withdrawal address.
		return c.generateOperationsFromValidatorAndPrivateKey(ctx)
	}

	return errors.New("unsupported combination of inputs; see help for details of supported combinations")
}

func (c *command) loadChainInfo(_ context.Context) error {
	_, err := os.Stat(offlinePreparationFilename)
	if err != nil {
		if c.debug {
			fmt.Fprintf(os.Stderr, "Failed to read offline preparation file: %v\n", err)
		}
		return errors.Wrap(err, fmt.Sprintf("cannot find %s", offlinePreparationFilename))
	}

	if c.debug {
		fmt.Fprintf(os.Stderr, "%s found; loading chain state\n", offlinePreparationFilename)
	}
	data, err := os.ReadFile(offlinePreparationFilename)
	if err != nil {
		return errors.Wrap(err, "failed to read offline preparation file")
	}
	if err := json.Unmarshal(data, c.chainInfo); err != nil {
		return errors.Wrap(err, "failed to parse offline preparation file")
	}

	return nil
}

func (c *command) loadOperations(ctx context.Context) error {
	// Start off by attempting to use the provided signed operations.
	if c.signedOperationsInput != "" {
		return c.loadOperationsFromInput(ctx)
	}

	// If not, read it from the file with the standard name.
	_, err := os.Stat(changeOperationsFilename)
	if err != nil {
		return errors.Wrap(err, "failed to read change operations file")
	}
	if c.debug {
		fmt.Fprintf(os.Stderr, "%s found; loading operations\n", changeOperationsFilename)
	}
	data, err := os.ReadFile(changeOperationsFilename)
	if err != nil {
		return errors.Wrap(err, "failed to read change operations file")
	}
	if err := json.Unmarshal(data, &c.signedOperations); err != nil {
		return errors.Wrap(err, "failed to parse change operations file")
	}

	for _, op := range c.signedOperations {
		if err := c.verifyOperation(ctx, op); err != nil {
			return err
		}
	}

	return nil
}

func (c *command) verifyOperation(ctx context.Context, op *capella.SignedBLSToExecutionChange) error {
	root, err := op.Message.HashTreeRoot()
	if err != nil {
		return errors.Wrap(err, "failed to generate message root")
	}

	sigBytes := make([]byte, len(op.Signature))
	copy(sigBytes, op.Signature[:])
	sig, err := e2types.BLSSignatureFromBytes(sigBytes)
	if err != nil {
		return errors.Wrap(err, "invalid signature")
	}

	container := &phase0.SigningData{
		ObjectRoot: root,
		Domain:     c.chainInfo.Domain,
	}
	signingRoot, err := ssz.HashTreeRoot(container)
	if err != nil {
		return errors.Wrap(err, "failed to generate signing root")
	}

	pubkeyBytes := make([]byte, len(op.Message.FromBLSPubkey))
	copy(pubkeyBytes, op.Message.FromBLSPubkey[:])
	pubkey, err := e2types.BLSPublicKeyFromBytes(pubkeyBytes)
	if err != nil {
		return errors.Wrap(err, "invalid public key")
	}
	if !sig.Verify(signingRoot[:], pubkey) {
		return errors.New("signature does not verify")
	}

	return nil
}

func (c *command) loadOperationsFromInput(_ context.Context) error {
	if strings.HasPrefix(c.signedOperationsInput, "{") {
		// This looks like a single entry; turn it in to an array.
		c.signedOperationsInput = fmt.Sprintf("[%s]", c.signedOperationsInput)
	}

	if !strings.HasPrefix(c.signedOperationsInput, "[") {
		// This looks like a file; read it in.
		data, err := os.ReadFile(c.signedOperationsInput)
		if err != nil {
			return errors.Wrap(err, "failed to read input file")
		}
		c.signedOperationsInput = string(data)
	}

	if err := json.Unmarshal([]byte(c.signedOperationsInput), &c.signedOperations); err != nil {
		return errors.Wrap(err, "failed to parse change operations input")
	}

	return nil
}

func (c *command) generateOperationsFromMnemonic(ctx context.Context) error {
	seed, err := util.SeedFromMnemonic(c.mnemonic)
	if err != nil {
		return err
	}

	// Turn the validators in to a map for easy lookup.
	validators := make(map[string]*validatorInfo, 0)
	for _, validator := range c.chainInfo.Validators {
		validators[fmt.Sprintf("%#x", validator.Pubkey)] = validator
	}

	maxDistance := 1024
	// Start scanning the validator keys.
	lastFoundIndex := 0
	for i := 0; ; i++ {
		if i-lastFoundIndex > maxDistance {
			if c.debug {
				fmt.Fprintf(os.Stderr, "Gone %d indices without finding a validator, not scanning any further\n", maxDistance)
			}
			break
		}
		validatorKeyPath := fmt.Sprintf("m/12381/3600/%d/0/0", i)

		found, err := c.generateOperationFromSeedAndPath(ctx, validators, seed, validatorKeyPath)
		if err != nil {
			return errors.Wrap(err, "failed to generate operation from seed and path")
		}
		if found {
			lastFoundIndex = i
		}
	}
	return nil
}

func (c *command) generateOperationsFromMnemonicAndPrivateKey(ctx context.Context) error {
	// Functionally identical to a simple scan, so use that.
	return c.generateOperationsFromMnemonic(ctx)
}

func (c *command) generateOperationsFromMnemonicAndValidator(ctx context.Context) error {
	seed, err := util.SeedFromMnemonic(c.mnemonic)
	if err != nil {
		return err
	}

	validator, err := c.fetchValidatorInfo(ctx)
	if err != nil {
		return err
	}

	// Scan the keys from the seed to find the path.
	maxDistance := 1024
	// Start scanning the validator keys.
	for i := 0; ; i++ {
		if i == maxDistance {
			if c.debug {
				fmt.Fprintf(os.Stderr, "Gone %d indices without finding the validator, not scanning any further\n", maxDistance)
			}
			break
		}
		validatorKeyPath := fmt.Sprintf("m/12381/3600/%d/0/0", i)
		validatorPrivkey, err := ethutil.PrivateKeyFromSeedAndPath(seed, validatorKeyPath)
		if err != nil {
			return errors.Wrap(err, "failed to generate validator private key")
		}
		validatorPubkey := validatorPrivkey.PublicKey().Marshal()
		if bytes.Equal(validatorPubkey, validator.Pubkey[:]) {
			// Recreate the withdrawal credentials to ensure a match.
			withdrawalKeyPath := strings.TrimSuffix(validatorKeyPath, "/0")
			withdrawalPrivkey, err := ethutil.PrivateKeyFromSeedAndPath(seed, withdrawalKeyPath)
			if err != nil {
				return errors.Wrap(err, "failed to generate withdrawal private key")
			}
			withdrawalPubkey := withdrawalPrivkey.PublicKey()
			withdrawalCredentials := ethutil.SHA256(withdrawalPubkey.Marshal())
			withdrawalCredentials[0] = byte(0) // BLS_WITHDRAWAL_PREFIX
			if !bytes.Equal(withdrawalCredentials, validator.WithdrawalCredentials) {
				return fmt.Errorf("validator %#x withdrawal credentials %#x do not match expected credentials, cannot update", validatorPubkey, validator.WithdrawalCredentials)
			}

			withdrawalAccount, err := util.ParseAccount(ctx, c.mnemonic, []string{withdrawalKeyPath}, true)
			if err != nil {
				return errors.Wrap(err, "failed to create withdrawal account")
			}

			err = c.generateOperationFromAccount(ctx, validator, withdrawalAccount)
			if err != nil {
				return err
			}
			break
		}
	}

	return nil
}

func (c *command) generateOperationFromSeedAndPath(ctx context.Context,
	validators map[string]*validatorInfo,
	seed []byte,
	path string,
) (
	bool,
	error,
) {
	validatorPrivkey, err := ethutil.PrivateKeyFromSeedAndPath(seed, path)
	if err != nil {
		return false, errors.Wrap(err, "failed to generate validator private key")
	}
	validatorPubkey := fmt.Sprintf("%#x", validatorPrivkey.PublicKey().Marshal())
	validator, exists := validators[validatorPubkey]
	if !exists {
		if c.debug {
			fmt.Fprintf(os.Stderr, "No validator found with public key %s at path %s\n", validatorPubkey, path)
		}
		return false, nil
	}

	if c.verbose {
		fmt.Fprintf(os.Stderr, "Validator %d found with public key %s at path %s\n", validator.Index, validatorPubkey, path)
	}

	if validator.WithdrawalCredentials[0] != byte(0) {
		if c.debug {
			fmt.Fprintf(os.Stderr, "Validator %s has non-BLS withdrawal credentials %#x\n", validatorPubkey, validator.WithdrawalCredentials)
		}
		return false, nil
	}

	var withdrawalPubkey []byte
	var withdrawalAccount e2wtypes.Account
	if c.privateKey == "" {
		// Recreate the withdrawal credentials to ensure a match.
		withdrawalKeyPath := strings.TrimSuffix(path, "/0")
		withdrawalPrivkey, err := ethutil.PrivateKeyFromSeedAndPath(seed, withdrawalKeyPath)
		if err != nil {
			return false, errors.Wrap(err, "failed to generate withdrawal private key")
		}
		withdrawalPubkey = withdrawalPrivkey.PublicKey().Marshal()
		withdrawalAccount, err = util.ParseAccount(ctx, c.mnemonic, []string{withdrawalKeyPath}, true)
		if err != nil {
			return false, errors.Wrap(err, "failed to create withdrawal account")
		}

	} else {
		// Need the withdrawal credentials from the private key.
		withdrawalAccount, err = util.ParseAccount(ctx, c.privateKey, nil, true)
		if err != nil {
			return false, err
		}
		withdrawalPubkey = withdrawalAccount.PublicKey().Marshal()
	}
	withdrawalCredentials := ethutil.SHA256(withdrawalPubkey)
	withdrawalCredentials[0] = byte(0) // BLS_WITHDRAWAL_PREFIX
	if !bytes.Equal(withdrawalCredentials, validator.WithdrawalCredentials) {
		if c.verbose && c.privateKey == "" {
			fmt.Fprintf(os.Stderr, "Validator %s withdrawal credentials %#x do not match expected credentials, cannot update\n", validatorPubkey, validator.WithdrawalCredentials)
		}
		return false, nil
	}

	if c.debug {
		fmt.Fprintf(os.Stderr, "Validator %s eligible for setting credentials\n", validatorPubkey)
	}

	err = c.generateOperationFromAccount(ctx, validator, withdrawalAccount)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (c *command) generateOperationFromAccount(ctx context.Context,
	validator *validatorInfo,
	withdrawalAccount e2wtypes.Account,
) error {
	signedOperation, err := c.createSignedOperation(ctx, validator, withdrawalAccount)
	if err != nil {
		return err
	}
	c.signedOperations = append(c.signedOperations, signedOperation)
	return nil
}

func (c *command) createSignedOperation(ctx context.Context,
	validator *validatorInfo,
	withdrawalAccount e2wtypes.Account,
) (
	*capella.SignedBLSToExecutionChange,
	error,
) {
	pubkey, err := util.BestPublicKey(withdrawalAccount)
	if err != nil {
		return nil, err
	}
	if c.debug {
		fmt.Fprintf(os.Stderr, "Using %#x as best public key for %s\n", pubkey.Marshal(), withdrawalAccount.Name())
	}
	blsPubkey := phase0.BLSPubKey{}
	copy(blsPubkey[:], pubkey.Marshal())

	if err := c.parseWithdrawalAddress(ctx); err != nil {
		return nil, errors.Wrap(err, "invalid withdrawal address")
	}

	operation := &capella.BLSToExecutionChange{
		ValidatorIndex:     validator.Index,
		FromBLSPubkey:      blsPubkey,
		ToExecutionAddress: c.withdrawalAddress,
	}
	root, err := operation.HashTreeRoot()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate root for credentials change operation")
	}

	// Sign the operation.
	signature, err := signing.SignRoot(ctx, withdrawalAccount, nil, root, c.chainInfo.Domain)
	if err != nil {
		return nil, errors.Wrap(err, "failed to sign credentials change operation")
	}

	return &capella.SignedBLSToExecutionChange{
		Message:   operation,
		Signature: signature,
	}, nil
}

func (c *command) parseWithdrawalAddress(_ context.Context) error {
	withdrawalAddressBytes, err := hex.DecodeString(strings.TrimPrefix(c.withdrawalAddressStr, "0x"))
	if err != nil {
		return errors.Wrap(err, "failed to obtain execution address")
	}
	if len(withdrawalAddressBytes) != bellatrix.ExecutionAddressLength {
		return errors.New("withdrawal address must be exactly 20 bytes in length")
	}
	// Ensure the address is properly checksummed.
	checksummedAddress := addressBytesToEIP55(withdrawalAddressBytes)
	if checksummedAddress != c.withdrawalAddressStr {
		return fmt.Errorf("withdrawal address checksum does not match (expected %s)", checksummedAddress)
	}
	copy(c.withdrawalAddress[:], withdrawalAddressBytes)

	return nil
}

func (c *command) validateOperations(ctx context.Context) (bool, string) {
	// Turn the validators in to a map for easy lookup.
	validators := make(map[phase0.ValidatorIndex]*validatorInfo, 0)
	for _, validator := range c.chainInfo.Validators {
		validators[validator.Index] = validator
	}

	for _, signedOperation := range c.signedOperations {
		if validated, reason := c.validateOperation(ctx, validators, signedOperation); !validated {
			return validated, reason
		}
	}
	return true, ""
}

func (c *command) validateOperation(_ context.Context,
	validators map[phase0.ValidatorIndex]*validatorInfo,
	signedOperation *capella.SignedBLSToExecutionChange,
) (
	bool,
	string,
) {
	validator, exists := validators[signedOperation.Message.ValidatorIndex]
	if !exists {
		return false, "validator not known on chain"
	}
	if c.debug {
		fmt.Fprintf(os.Stderr, "Credentials change operation: %v", signedOperation)
		fmt.Fprintf(os.Stderr, "On-chain validator info: %v\n", validator)
	}

	if validator.WithdrawalCredentials[0] != byte(0) {
		return false, "validator is not using BLS withdrawal credentials"
	}

	withdrawalCredentials := ethutil.SHA256(signedOperation.Message.FromBLSPubkey[:])
	withdrawalCredentials[0] = byte(0) // BLS_WITHDRAWAL_PREFIX
	if !bytes.Equal(withdrawalCredentials, validator.WithdrawalCredentials) {
		if c.debug {
			fmt.Fprintf(os.Stderr, "validator withdrawal credentials %#x do not match calculated operation withdrawal credentials %#x\n", validator.WithdrawalCredentials, withdrawalCredentials)
		}
		return false, "validator withdrawal credentials do not match those in the operation"
	}

	return true, ""
}

func (c *command) broadcastOperations(ctx context.Context) error {
	return c.consensusClient.(consensusclient.BLSToExecutionChangesSubmitter).SubmitBLSToExecutionChanges(ctx, c.signedOperations)
}

func (c *command) setup(ctx context.Context) error {
	if c.offline {
		return nil
	}

	// Connect to the consensus node.
	var err error
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

func (c *command) fetchValidatorInfo(ctx context.Context) (*validatorInfo, error) {
	var validatorInfo *validatorInfo
	switch {
	case c.validator == "":
		return nil, errors.New("no validator specified")
	case strings.HasPrefix(c.validator, "0x"):
		// A public key
		for _, validator := range c.chainInfo.Validators {
			if strings.EqualFold(c.validator, fmt.Sprintf("%#x", validator.Pubkey)) {
				validatorInfo = validator
				break
			}
		}
	case strings.Contains(c.validator, "/"):
		// An account.
		_, account, err := util.WalletAndAccountFromPath(ctx, c.validator)
		if err != nil {
			return nil, errors.Wrap(err, "unable to obtain account")
		}
		accPubKey, err := util.BestPublicKey(account)
		if err != nil {
			return nil, errors.Wrap(err, "unable to obtain public key for account")
		}
		pubkey := fmt.Sprintf("%#x", accPubKey.Marshal())
		for _, validator := range c.chainInfo.Validators {
			if strings.EqualFold(pubkey, fmt.Sprintf("%#x", validator.Pubkey)) {
				validatorInfo = validator
				break
			}
		}
	default:
		// An index.
		index, err := strconv.ParseUint(c.validator, 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse validator index")
		}
		validatorIndex := phase0.ValidatorIndex(index)
		for _, validator := range c.chainInfo.Validators {
			if validator.Index == validatorIndex {
				validatorInfo = validator
				break
			}
		}
	}

	if validatorInfo == nil {
		return nil, errors.New("unknown validator")
	}

	return validatorInfo, nil
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

func (c *command) generateOperationsFromMnemonicAndPath(ctx context.Context) error {
	seed, err := util.SeedFromMnemonic(c.mnemonic)
	if err != nil {
		return err
	}

	// Turn the validators in to a map for easy lookup.
	validators := make(map[string]*validatorInfo, 0)
	for _, validator := range c.chainInfo.Validators {
		validators[fmt.Sprintf("%#x", validator.Pubkey)] = validator
	}

	validatorKeyPath := c.path
	match := validatorPath.Match([]byte(c.path))
	if !match {
		return fmt.Errorf("path %s does not match EIP-2334 format", c.path)
	}

	if _, err := c.generateOperationFromSeedAndPath(ctx, validators, seed, validatorKeyPath); err != nil {
		return errors.Wrap(err, "failed to generate operation from seed and path")
	}

	return nil
}

func (c *command) generateOperationsFromAccountAndWithdrawalAccount(ctx context.Context) error {
	validatorAccount, err := util.ParseAccount(ctx, c.account, nil, true)
	if err != nil {
		return err
	}

	withdrawalAccount, err := util.ParseAccount(ctx, c.withdrawalAccount, c.passphrases, true)
	if err != nil {
		return err
	}

	// Find the validator info given its account information.
	validatorPubkey := validatorAccount.PublicKey().Marshal()
	var validatorInfo *validatorInfo
	for _, validator := range c.chainInfo.Validators {
		if bytes.Equal(validator.Pubkey[:], validatorPubkey) {
			// Found it.
			validatorInfo = validator
		}
	}
	if validatorInfo == nil {
		return errors.New("could not find information for that validator on the chain")
	}

	if err := c.generateOperationFromAccount(ctx, validatorInfo, withdrawalAccount); err != nil {
		return err
	}

	return nil
}

func (c *command) generateOperationsFromAccountAndPrivateKey(ctx context.Context) error {
	validatorAccount, err := util.ParseAccount(ctx, c.account, nil, true)
	if err != nil {
		return err
	}

	withdrawalAccount, err := util.ParseAccount(ctx, c.privateKey, nil, true)
	if err != nil {
		return err
	}

	// Find the validator info given its account information.
	validatorPubkey := validatorAccount.PublicKey().Marshal()
	var validatorInfo *validatorInfo
	for _, validator := range c.chainInfo.Validators {
		if bytes.Equal(validator.Pubkey[:], validatorPubkey) {
			// Found it.
			validatorInfo = validator
		}
	}
	if validatorInfo == nil {
		return errors.New("could not find information for that validator on the chain")
	}

	if err := c.generateOperationFromAccount(ctx, validatorInfo, withdrawalAccount); err != nil {
		return err
	}

	return nil
}

func (c *command) generateOperationsFromValidatorAndPrivateKey(ctx context.Context) error {
	validator, err := c.fetchValidatorInfo(ctx)
	if err != nil {
		return err
	}

	validatorAccount, err := util.ParseAccount(ctx, validator.Pubkey.String(), nil, false)
	if err != nil {
		return err
	}

	withdrawalAccount, err := util.ParseAccount(ctx, c.privateKey, nil, true)
	if err != nil {
		return err
	}

	// Find the validator info given its account information.
	validatorPubkey := validatorAccount.PublicKey().Marshal()
	var validatorInfo *validatorInfo
	for _, validator := range c.chainInfo.Validators {
		if bytes.Equal(validator.Pubkey[:], validatorPubkey) {
			// Found it.
			validatorInfo = validator
		}
	}
	if validatorInfo == nil {
		return errors.New("could not find information for that validator on the chain")
	}

	if err := c.generateOperationFromAccount(ctx, validatorInfo, withdrawalAccount); err != nil {
		return err
	}

	return nil
}
