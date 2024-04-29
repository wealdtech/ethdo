// Copyright © 2022, 2023 Weald Technology Trading.
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
	"strings"
	"time"

	consensusclient "github.com/attestantio/go-eth2-client"
	"github.com/attestantio/go-eth2-client/spec/bellatrix"
	capella "github.com/attestantio/go-eth2-client/spec/capella"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
	"github.com/prysmaticlabs/go-ssz"
	"github.com/wealdtech/ethdo/beacon"
	standardchaintime "github.com/wealdtech/ethdo/services/chaintime/standard"
	"github.com/wealdtech/ethdo/signing"
	"github.com/wealdtech/ethdo/util"
	e2types "github.com/wealdtech/go-eth2-types/v2"
	ethutil "github.com/wealdtech/go-eth2-util"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

// minTimeout is the minimum timeout for this command.
// It needs to be set here as we want timeouts to be low in general, but this can be pulling
// a lot of data for an unsophisticated audience so it's easier to set a higher timeout..
var minTimeout = 5 * time.Minute

// validatorPath is the regular expression that matches a validator  path.
var validatorPath = regexp.MustCompile("^m/12381/3600/[0-9]+/0/0$")

// numeric is the regular expression that matches a number.
var numeric = regexp.MustCompile(`^[0-9]+$`)

var (
	offlinePreparationFilename = "offline-preparation.json"
	changeOperationsFilename   = "change-operations.json"
)

func (c *command) process(ctx context.Context) error {
	if err := c.setup(ctx); err != nil {
		return err
	}

	if err := c.obtainChainInfo(ctx); err != nil {
		return err
	}

	if c.prepareOffline {
		return c.writeChainInfoToFile(ctx)
	}

	if err := c.generateDomain(ctx); err != nil {
		return err
	}

	if err := c.obtainOperations(ctx); err != nil {
		return err
	}

	if len(c.signedOperations) == 0 {
		return errors.New("no suitable validators found; no operations generated")
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

func (c *command) obtainOperations(ctx context.Context) error {
	if c.account == "" && c.mnemonic == "" && c.privateKey == "" && c.validator == "" {
		// No input information; fetch the operations from a file.
		err := c.obtainOperationsFromFileOrInput(ctx)
		if err == nil {
			// Success.
			return nil
		}
		if c.signedOperationsInput != "" {
			return errors.Wrap(err, "failed to obtain supplied signed operations")
		}
		return errors.Wrap(err, fmt.Sprintf("no account, mnemonic or private key specified, and no %s file loaded", changeOperationsFilename))
	}

	if c.mnemonic != "" {
		switch {
		case c.path != "":
			// Have a mnemonic and path.
			return c.generateOperationFromMnemonicAndPath(ctx)
		case c.validator != "":
			// Have a mnemonic and validator.
			return c.generateOperationFromMnemonicAndValidator(ctx)
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

	if c.privateKey != "" {
		// Have a private key.
		return c.generateOperationsFromPrivateKey(ctx)
	}

	return errors.New("unsupported combination of inputs; see help for details of supported combinations")
}

func (c *command) generateOperationFromMnemonicAndPath(ctx context.Context) error {
	seed, err := util.SeedFromMnemonic(c.mnemonic)
	if err != nil {
		return err
	}

	// Turn the validators in to a map for easy lookup.
	validators := make(map[string]*beacon.ValidatorInfo, 0)
	for _, validator := range c.chainInfo.Validators {
		validators[fmt.Sprintf("%#x", validator.Pubkey)] = validator
	}

	validatorKeyPath := c.path
	match := validatorPath.MatchString(c.path)
	if !match {
		return fmt.Errorf("path %s does not match EIP-2334 format for a validator", c.path)
	}

	found, err := c.generateOperationFromSeedAndPath(ctx, validators, seed, validatorKeyPath)
	if err != nil {
		return errors.Wrap(err, "failed to generate operation from seed and path")
	}
	// Function `c.generateOperationFromSeedAndPath()` will not return errors
	// in non-serious cases since it is called in a loop when searching a
	// mnemonic's key space without a specific path, so we need to check if a
	// validator was not found in our case (it should be found if a path is
	// provided) and return an error if not.
	if !found {
		return errors.New("no validator found with the provided path and mnemonic, please run with --debug to see more information")
	}

	return nil
}

func (c *command) generateOperationFromMnemonicAndValidator(ctx context.Context) error {
	seed, err := util.SeedFromMnemonic(c.mnemonic)
	if err != nil {
		return err
	}

	validatorInfo, err := c.chainInfo.FetchValidatorInfo(ctx, c.validator)
	if err != nil {
		return err
	}

	// Scan the keys from the seed to find the path.
	maxDistance := 1024
	if c.maxDistance > 0 {
		maxDistance = int(c.maxDistance)
	}
	// Start scanning the validator keys.
	var withdrawalAccount e2wtypes.Account
	for i := 0; ; i++ {
		if i == maxDistance {
			if c.debug {
				fmt.Fprintf(os.Stderr, "Gone %d indices without finding the validator, not scanning any further\n", maxDistance)
			}
			return fmt.Errorf("failed to find validator using the provided mnemonic, validator=%s, pubkey=%#x", c.validator, validatorInfo.Pubkey)
		}
		validatorKeyPath := fmt.Sprintf("m/12381/3600/%d/0/0", i)
		validatorPrivkey, err := ethutil.PrivateKeyFromSeedAndPath(seed, validatorKeyPath)
		if err != nil {
			return errors.Wrap(err, "failed to generate validator private key")
		}
		validatorPubkey := validatorPrivkey.PublicKey().Marshal()
		if bytes.Equal(validatorPubkey, validatorInfo.Pubkey[:]) {
			withdrawalKeyPath := strings.TrimSuffix(validatorKeyPath, "/0")
			withdrawalAccount, err = util.ParseAccount(ctx, c.mnemonic, []string{withdrawalKeyPath}, true)
			if err != nil {
				return errors.Wrap(err, "failed to create withdrawal account")
			}

			err = c.generateOperationFromAccount(ctx, validatorInfo, withdrawalAccount)
			if err != nil {
				return err
			}
			break
		}
	}

	return nil
}

func (c *command) generateOperationsFromMnemonicAndPrivateKey(ctx context.Context) error {
	// Functionally identical to a simple scan, so use that.
	return c.generateOperationsFromMnemonic(ctx)
}

func (c *command) generateOperationsFromMnemonic(ctx context.Context) error {
	seed, err := util.SeedFromMnemonic(c.mnemonic)
	if err != nil {
		return err
	}

	// Turn the validators in to a map for easy lookup.
	validators := make(map[string]*beacon.ValidatorInfo, 0)
	for _, validator := range c.chainInfo.Validators {
		validators[fmt.Sprintf("%#x", validator.Pubkey)] = validator
	}

	// Start scanning the validator keys.
	lastFoundIndex := 0
	foundValidatorCount := 0
	maxDistance := 1024
	if c.maxDistance > 0 {
		maxDistance = int(c.maxDistance)
	}
	for i := 0; ; i++ {
		// If no validators have been found in the last maxDistance indices, stop scanning.
		if i-lastFoundIndex > maxDistance {
			// If no validators were found at all, return an error.
			if foundValidatorCount == 0 {
				return fmt.Errorf("failed to find validators using the provided mnemonic: searched %d indices without finding a validator", maxDistance)
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
			foundValidatorCount++
		}
	}
	return nil
}

func (c *command) generateOperationsFromAccountAndWithdrawalAccount(ctx context.Context) error {
	validatorAccount, err := util.ParseAccount(ctx, c.account, nil, false)
	if err != nil {
		return errors.Wrap(err, "failed to obtain validator account")
	}

	withdrawalAccount, err := util.ParseAccount(ctx, c.withdrawalAccount, c.passphrases, true)
	if err != nil {
		return errors.Wrap(err, "failed to obtain withdrawal account")
	}

	validatorPubkey, err := util.BestPublicKey(validatorAccount)
	if err != nil {
		return err
	}
	validatorInfo, err := c.chainInfo.FetchValidatorInfo(ctx, fmt.Sprintf("%#x", validatorPubkey.Marshal()))
	if err != nil {
		return errors.Wrap(err, "failed to obtain validator info")
	}

	return c.generateOperationFromAccount(ctx, validatorInfo, withdrawalAccount)
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

	validatorPubkey, err := util.BestPublicKey(validatorAccount)
	if err != nil {
		return err
	}
	validatorInfo, err := c.chainInfo.FetchValidatorInfo(ctx, fmt.Sprintf("%#x", validatorPubkey.Marshal()))
	if err != nil {
		return errors.Wrap(err, "failed to obtain validator info")
	}

	return c.generateOperationFromAccount(ctx, validatorInfo, withdrawalAccount)
}

func (c *command) generateOperationsFromValidatorAndPrivateKey(ctx context.Context) error {
	validatorInfo, err := c.obtainValidatorInfoFromValidatorSpecifier(ctx)
	if err != nil {
		return err
	}

	withdrawalAccount, err := util.ParseAccount(ctx, c.privateKey, nil, true)
	if err != nil {
		return err
	}

	return c.generateOperationFromAccount(ctx, validatorInfo, withdrawalAccount)
}

func (c *command) obtainValidatorInfoFromValidatorSpecifier(ctx context.Context) (*beacon.ValidatorInfo, error) {
	if numeric.MatchString(c.validator) {
		// The validator specifier looks like an on-chain index.  Fetch directly from the
		// chain information.
		return c.chainInfo.FetchValidatorInfo(ctx, c.validator)
	}

	// The validator specifier Looks like some sort of account specifier.  Fetch the account first,
	// and then the validator information from its public key.
	validatorAccount, err := util.ParseAccount(ctx, c.validator, nil, false)
	if err != nil {
		return nil, err
	}

	validatorPubkey, err := util.BestPublicKey(validatorAccount)
	if err != nil {
		return nil, err
	}

	return c.chainInfo.FetchValidatorInfo(ctx, fmt.Sprintf("%#x", validatorPubkey.Marshal()))
}

func (c *command) generateOperationsFromPrivateKey(ctx context.Context) error {
	// Verify that the user provided a private key.
	if strings.HasPrefix(c.privateKey, "0x") {
		data, err := hex.DecodeString(strings.TrimPrefix(c.privateKey, "0x"))
		if err != nil {
			return errors.Wrap(err, "failed to parse account key")
		}
		if len(data) != 32 {
			return errors.New("account key must be 32 bytes")
		}
	} else {
		return errors.New("account key must be a hex string")
	}
	// Extract withdrawal account public key from supplied private key.
	withdrawalAccount, err := util.ParseAccount(ctx, c.privateKey, nil, true)
	if err != nil {
		return err
	}
	pubkey, err := util.BestPublicKey(withdrawalAccount)
	if err != nil {
		return err
	}
	withdrawalCredentials := ethutil.SHA256(pubkey.Marshal())
	withdrawalCredentials[0] = byte(0) // BLS_WITHDRAWAL_PREFIX

	found := false
	for _, validatorInfo := range c.chainInfo.Validators {
		// Skip validators which withdrawal key don't match with supplied withdrawal account public key.
		if !bytes.Equal(withdrawalCredentials, validatorInfo.WithdrawalCredentials) {
			continue
		}

		if err := c.generateOperationFromAccount(ctx, validatorInfo, withdrawalAccount); err != nil {
			return err
		}
		found = true
	}
	if !found {
		return fmt.Errorf("no validator found with withdrawal credentials %#x", withdrawalCredentials)
	}
	return nil
}

func (c *command) obtainOperationsFromFileOrInput(ctx context.Context) error {
	// Start off by attempting to use the provided signed operations.
	if c.signedOperationsInput != "" {
		return c.obtainOperationsFromInput(ctx)
	}
	// If not, read it from the file with the standard name.
	return c.obtainOperationsFromFile(ctx)
}

func (c *command) obtainOperationsFromFile(ctx context.Context) error {
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

func (c *command) obtainOperationsFromInput(ctx context.Context) error {
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

	for _, op := range c.signedOperations {
		if err := c.verifyOperation(ctx, op); err != nil {
			return err
		}
	}
	return nil
}

func (c *command) generateOperationFromSeedAndPath(ctx context.Context,
	validators map[string]*beacon.ValidatorInfo,
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
			fmt.Fprintf(os.Stderr, "no validator found with public key %s at path %s\n", validatorPubkey, path)
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
	validator *beacon.ValidatorInfo,
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
	validator *beacon.ValidatorInfo,
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
	if c.debug {
		fmt.Fprintf(os.Stderr, "Signing %#x with domain %#x by public key %#x\n", root, c.domain, withdrawalAccount.PublicKey().Marshal())
	}
	signature, err := signing.SignRoot(ctx, withdrawalAccount, nil, root, c.domain)
	if err != nil {
		return nil, errors.Wrap(err, "failed to sign credentials change operation")
	}

	return &capella.SignedBLSToExecutionChange{
		Message:   operation,
		Signature: signature,
	}, nil
}

func (c *command) parseWithdrawalAddress(_ context.Context) error {
	// Check that a withdrawal address has been provided.
	if c.withdrawalAddressStr == "" {
		return errors.New("no withdrawal address provided")
	}
	// Check that the withdrawal address contains a 0x prefix.
	if !strings.HasPrefix(c.withdrawalAddressStr, "0x") {
		return fmt.Errorf("withdrawal address %s does not contain a 0x prefix", c.withdrawalAddressStr)
	}
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
	validators := make(map[phase0.ValidatorIndex]*beacon.ValidatorInfo, 0)
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

func (c *command) verifyOperation(_ context.Context, op *capella.SignedBLSToExecutionChange) error {
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
		Domain:     c.domain,
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

func (c *command) validateOperation(_ context.Context,
	validators map[phase0.ValidatorIndex]*beacon.ValidatorInfo,
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

	// Ensure timeout is at least the minimum.
	if c.timeout < minTimeout {
		if c.debug {
			fmt.Fprintf(os.Stderr, "Increasing timeout to %v\n", minTimeout)
		}
		c.timeout = minTimeout
	}

	// Connect to the consensus node.
	var err error
	c.consensusClient, err = util.ConnectToBeaconNode(ctx, &util.ConnectOpts{
		Address:       c.connection,
		Timeout:       c.timeout,
		AllowInsecure: c.allowInsecureConnections,
		LogFallback:   !c.quiet,
	})
	if err != nil {
		return err
	}

	// Set up chaintime.
	c.chainTime, err = standardchaintime.New(ctx,
		standardchaintime.WithGenesisProvider(c.consensusClient.(consensusclient.GenesisProvider)),
		standardchaintime.WithSpecProvider(c.consensusClient.(consensusclient.SpecProvider)),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create chaintime service")
	}

	return nil
}

func (c *command) generateDomain(ctx context.Context) error {
	genesisValidatorsRoot, err := c.obtainGenesisValidatorsRoot(ctx)
	if err != nil {
		return err
	}
	forkVersion, err := c.obtainForkVersion(ctx)
	if err != nil {
		return err
	}

	root, err := (&phase0.ForkData{
		CurrentVersion:        forkVersion,
		GenesisValidatorsRoot: genesisValidatorsRoot,
	}).HashTreeRoot()
	if err != nil {
		return errors.Wrap(err, "failed to calculate signature domain")
	}

	copy(c.domain[:], c.chainInfo.BLSToExecutionChangeDomainType[:])
	copy(c.domain[4:], root[:])
	if c.debug {
		fmt.Fprintf(os.Stderr, "Domain is %#x\n", c.domain)
	}

	return nil
}

func (c *command) obtainGenesisValidatorsRoot(_ context.Context) (phase0.Root, error) {
	genesisValidatorsRoot := phase0.Root{}

	if c.genesisValidatorsRoot != "" {
		if c.debug {
			fmt.Fprintf(os.Stderr, "Genesis validators root supplied on the command line\n")
		}
		root, err := hex.DecodeString(strings.TrimPrefix(c.genesisValidatorsRoot, "0x"))
		if err != nil {
			return phase0.Root{}, errors.Wrap(err, "invalid genesis validators root supplied")
		}
		if len(root) != phase0.RootLength {
			return phase0.Root{}, errors.New("invalid length for genesis validators root")
		}
		copy(genesisValidatorsRoot[:], root)
	} else {
		if c.debug {
			fmt.Fprintf(os.Stderr, "Genesis validators root obtained from chain info\n")
		}
		copy(genesisValidatorsRoot[:], c.chainInfo.GenesisValidatorsRoot[:])
	}

	if c.debug {
		fmt.Fprintf(os.Stderr, "Using genesis validators root %#x\n", genesisValidatorsRoot)
	}
	return genesisValidatorsRoot, nil
}

func (c *command) obtainForkVersion(_ context.Context) (phase0.Version, error) {
	forkVersion := phase0.Version{}

	if c.forkVersion != "" {
		if c.debug {
			fmt.Fprintf(os.Stderr, "Fork version supplied on the command line\n")
		}
		version, err := hex.DecodeString(strings.TrimPrefix(c.forkVersion, "0x"))
		if err != nil {
			return phase0.Version{}, errors.Wrap(err, "invalid fork version supplied")
		}
		if len(version) != phase0.ForkVersionLength {
			return phase0.Version{}, errors.New("invalid length for fork version")
		}
		copy(forkVersion[:], version)
	} else {
		if c.debug {
			fmt.Fprintf(os.Stderr, "Fork version obtained from chain info\n")
		}
		// Use the genesis fork version for setting credentials as per the spec.
		copy(forkVersion[:], c.chainInfo.GenesisForkVersion[:])
	}

	if c.debug {
		fmt.Fprintf(os.Stderr, "Using fork version %#x\n", forkVersion)
	}
	return forkVersion, nil
}

// addressBytesToEIP55 converts a byte array in to an EIP-55 string format.
func addressBytesToEIP55(address []byte) string {
	bytes := []byte(hex.EncodeToString(address))
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
