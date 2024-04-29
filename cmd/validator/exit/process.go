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

package validatorexit

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
	"time"

	consensusclient "github.com/attestantio/go-eth2-client"
	apiv1 "github.com/attestantio/go-eth2-client/api/v1"
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

var (
	offlinePreparationFilename = "offline-preparation.json"
	exitOperationsFilename     = "exit-operations.json"
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
		return fmt.Errorf("operations failed validation: %s", reason)
	}

	if c.json || c.offline {
		if c.debug {
			fmt.Fprintf(os.Stderr, "Not broadcasting exit operations\n")
		}
		// Want JSON output, or cannot broadcast.
		return nil
	}

	return c.broadcastOperations(ctx)
}

func (c *command) obtainOperations(ctx context.Context) error {
	if c.mnemonic == "" && c.privateKey == "" && c.validator == "" {
		// No input information; fetch the operation from a file.
		err := c.obtainOperationsFromFileOrInput(ctx)
		if err == nil {
			// Success.
			return nil
		}
		if c.signedOperationsInput != "" {
			return errors.Wrap(err, "failed to obtain supplied signed operation")
		}
		return errors.Wrap(err, fmt.Sprintf("no account, mnemonic or private key specified, and no %s file loaded", exitOperationsFilename))
	}

	if c.mnemonic != "" {
		switch {
		case c.path != "":
			// Have a mnemonic and path.
			return c.generateOperationFromMnemonicAndPath(ctx)
		case c.validator != "":
			// Have a mnemonic and validator.
			return c.generateOperationFromMnemonicAndValidator(ctx)
		default:
			// Have a mnemonic only.
			return c.generateOperationsFromMnemonic(ctx)
		}
	}

	if c.privateKey != "" {
		return c.generateOperationFromPrivateKey(ctx)
	}

	if c.validator != "" {
		return c.generateOperationFromValidator(ctx)
	}

	return errors.New("unsupported combination of inputs; see help for details of supported combinations")
}

func (c *command) generateOperationFromMnemonicAndPath(ctx context.Context) error {
	// Turn the validators in to a map for easy lookup.
	validators := make(map[string]*beacon.ValidatorInfo, 0)
	for _, validator := range c.chainInfo.Validators {
		validators[fmt.Sprintf("%#x", validator.Pubkey)] = validator
	}

	seed, err := util.SeedFromMnemonic(c.mnemonic)
	if err != nil {
		return err
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

	if c.debug {
		fmt.Fprintf(os.Stderr, "Searching for validator with index %d and public key %s\n", validatorInfo.Index, validatorInfo.Pubkey.String())
	}

	// Scan the keys from the seed to find the path.
	maxDistance := 1024
	if c.maxDistance > 0 {
		maxDistance = int(c.maxDistance)
	}
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
		if bytes.Equal(validatorPubkey, validatorInfo.Pubkey[:]) {
			validatorAccount, err := util.ParseAccount(ctx, c.mnemonic, []string{validatorKeyPath}, true)
			if err != nil {
				return errors.Wrap(err, "failed to create withdrawal account")
			}

			if err := c.generateOperationFromAccount(ctx, validatorAccount); err != nil {
				return err
			}
			break
		}
	}

	return nil
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

	// Scan the keys from the seed to find the path.
	maxDistance := 1024
	if c.maxDistance > 0 {
		maxDistance = int(c.maxDistance)
	}
	// Start scanning the validator keys.
	lastFoundIndex := 0
	foundValidatorCount := 0
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
			// We log errors but keep going.
			if c.debug {
				fmt.Fprintf(os.Stderr, "Failed to generate for path %s: %v\n", validatorKeyPath, err.Error())
			}
		}
		if found {
			lastFoundIndex = i
			foundValidatorCount++
		}
	}

	return nil
}

func (c *command) generateOperationFromPrivateKey(ctx context.Context) error {
	validatorAccount, err := util.ParseAccount(ctx, c.privateKey, nil, true)
	if err != nil {
		return errors.Wrap(err, "failed to parse validator account")
	}

	return c.generateOperationFromAccount(ctx, validatorAccount)
}

func (c *command) generateOperationFromValidator(ctx context.Context) error {
	validatorAccount, err := util.ParseAccount(ctx, c.validator, c.passphrases, true)
	if err != nil {
		return errors.Wrap(err, "failed to parse validator account")
	}

	return c.generateOperationFromAccount(ctx, validatorAccount)
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
	_, err := os.Stat(exitOperationsFilename)
	if err != nil {
		return errors.Wrap(err, "failed to read exit operations file")
	}
	if c.debug {
		fmt.Fprintf(os.Stderr, "%s found; loading operations\n", exitOperationsFilename)
	}
	data, err := os.ReadFile(exitOperationsFilename)
	if err != nil {
		return errors.Wrap(err, "failed to read exit operations file")
	}
	if err := json.Unmarshal(data, &c.signedOperations); err != nil {
		return errors.Wrap(err, "failed to parse exit operations file")
	}

	return c.verifySignedOperations(ctx)
}

func (c *command) obtainOperationsFromInput(ctx context.Context) error {
	if !strings.HasPrefix(c.signedOperationsInput, "{") &&
		!strings.HasPrefix(c.signedOperationsInput, "[") {
		// This looks like a file; read it in.
		data, err := os.ReadFile(c.signedOperationsInput)
		if err != nil {
			return errors.Wrap(err, "failed to read input file")
		}
		c.signedOperationsInput = string(data)
	}

	if strings.HasPrefix(c.signedOperationsInput, "{") {
		// Single operation; put it in an array.
		c.signedOperationsInput = fmt.Sprintf("[%s]", c.signedOperationsInput)
	}

	if err := json.Unmarshal([]byte(c.signedOperationsInput), &c.signedOperations); err != nil {
		return errors.Wrap(err, "failed to parse exit operation input")
	}

	return c.verifySignedOperations(ctx)
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

	privateKey := fmt.Sprintf("%#x", validatorPrivkey.Marshal())

	validatorInfo, exists := validators[fmt.Sprintf("%#x", validatorPrivkey.PublicKey().Marshal())]
	if !exists {
		return false, errors.New("unknown validator")
	}
	if validatorInfo.State == apiv1.ValidatorStateActiveExiting ||
		validatorInfo.State == apiv1.ValidatorStateActiveSlashed ||
		validatorInfo.State == apiv1.ValidatorStateExitedUnslashed ||
		validatorInfo.State == apiv1.ValidatorStateExitedSlashed ||
		validatorInfo.State == apiv1.ValidatorStateWithdrawalPossible ||
		validatorInfo.State == apiv1.ValidatorStateWithdrawalDone {
		return false, fmt.Errorf("validator is in state %v, not suitable to generate an exit", validatorInfo.State)
	}

	validatorAccount, err := util.ParseAccount(ctx, privateKey, nil, true)
	if err != nil {
		if c.debug {
			fmt.Fprintf(os.Stderr, "no validator found at path %s: %v\n", path, err)
		}
		return false, nil
	}
	if validatorAccount == nil {
		return false, nil
	}

	if err := c.generateOperationFromAccount(ctx, validatorAccount); err != nil {
		if c.debug {
			fmt.Fprintf(os.Stderr, "failed to generate operation at path %s: %v\n", path, err)
		}
		return false, nil
	}

	return true, nil
}

func (c *command) generateOperationFromAccount(ctx context.Context,
	account e2wtypes.Account,
) error {
	pubKey, err := util.BestPublicKey(account)
	if err != nil {
		return err
	}

	info, err := c.chainInfo.FetchValidatorInfo(ctx, fmt.Sprintf("%#x", pubKey.Marshal()))
	if err != nil {
		return err
	}

	epoch, err := c.selectEpoch()
	if err != nil {
		return err
	}
	if c.debug {
		fmt.Fprintf(os.Stderr, "Using %d for epoch\n", epoch)
	}

	signedOperation, err := c.createSignedOperation(ctx, info, account, epoch)
	if err != nil {
		return err
	}
	c.signedOperations = append(c.signedOperations, signedOperation)

	return nil
}

func (c *command) selectEpoch() (phase0.Epoch, error) {
	if c.epoch == "" {
		// No user-supplied epoch; use the one from chain info.
		return c.chainInfo.Epoch, nil
	}

	epoch, err := strconv.ParseInt(c.epoch, 10, 64)
	if err != nil {
		return 0, errors.New("epoch invalid")
	}

	if epoch < 0 {
		// Relative epoch.
		return c.chainInfo.Epoch - phase0.Epoch(-epoch), nil
	}

	return phase0.Epoch(epoch), nil
}

func (c *command) createSignedOperation(ctx context.Context,
	validator *beacon.ValidatorInfo,
	account e2wtypes.Account,
	epoch phase0.Epoch,
) (
	*phase0.SignedVoluntaryExit,
	error,
) {
	pubkey, err := util.BestPublicKey(account)
	if err != nil {
		return nil, err
	}
	if c.debug {
		fmt.Fprintf(os.Stderr, "Using %#x as best public key for %s\n", pubkey.Marshal(), account.Name())
	}
	blsPubkey := phase0.BLSPubKey{}
	copy(blsPubkey[:], pubkey.Marshal())

	operation := &phase0.VoluntaryExit{
		Epoch:          epoch,
		ValidatorIndex: validator.Index,
	}
	root, err := operation.HashTreeRoot()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate root for exit operation")
	}

	// Sign the operation.
	if c.debug {
		fmt.Fprintf(os.Stderr, "Signing %#x with domain %#x by public key %#x\n", root, c.domain, account.PublicKey().Marshal())
	}
	signature, err := signing.SignRoot(ctx, account, nil, root, c.domain)
	if err != nil {
		return nil, errors.Wrap(err, "failed to sign exit operation")
	}

	return &phase0.SignedVoluntaryExit{
		Message:   operation,
		Signature: signature,
	}, nil
}

func (c *command) verifySignedOperations(ctx context.Context) error {
	for _, op := range c.signedOperations {
		if err := c.verifySignedOperation(ctx, op); err != nil {
			return err
		}
	}
	return nil
}

func (c *command) verifySignedOperation(ctx context.Context, op *phase0.SignedVoluntaryExit) error {
	root, err := op.Message.HashTreeRoot()
	if err != nil {
		return errors.Wrap(err, "failed to generate message root")
	}

	sigBytes := make([]byte, len(op.Signature))
	copy(sigBytes, op.Signature[:])
	sig, err := e2types.BLSSignatureFromBytes(sigBytes)
	if err != nil {
		if c.verbose {
			fmt.Fprintf(os.Stderr, "Invalid signature: %v\n", err.Error())
		}
		return errors.New("invalid signature")
	}

	container := &phase0.SigningData{
		ObjectRoot: root,
		Domain:     c.domain,
	}
	signingRoot, err := ssz.HashTreeRoot(container)
	if err != nil {
		return errors.Wrap(err, "failed to generate signing root")
	}

	validatorInfo, err := c.chainInfo.FetchValidatorInfo(ctx, fmt.Sprintf("%d", op.Message.ValidatorIndex))
	if err != nil {
		return err
	}

	pubkeyBytes := make([]byte, len(validatorInfo.Pubkey[:]))
	copy(pubkeyBytes, validatorInfo.Pubkey[:])
	pubkey, err := e2types.BLSPublicKeyFromBytes(pubkeyBytes)
	if err != nil {
		return errors.Wrap(err, "invalid public key")
	}

	if !sig.Verify(signingRoot[:], pubkey) {
		return errors.New("signature does not verify")
	}

	return nil
}

func (c *command) validateOperations(ctx context.Context) (bool, string) {
	for _, op := range c.signedOperations {
		valid, issue := c.validateOperation(ctx, op)
		if !valid {
			return valid, issue
		}
	}

	return true, ""
}

func (c *command) validateOperation(_ context.Context,
	op *phase0.SignedVoluntaryExit,
) (
	bool,
	string,
) {
	var validatorInfo *beacon.ValidatorInfo
	for _, chainValidatorInfo := range c.chainInfo.Validators {
		if chainValidatorInfo.Index == op.Message.ValidatorIndex {
			validatorInfo = chainValidatorInfo
			break
		}
	}
	if validatorInfo == nil {
		return false, "validator not known on chain"
	}
	if c.debug {
		fmt.Fprintf(os.Stderr, "Validator exit operation: %v", op)
		fmt.Fprintf(os.Stderr, "On-chain validator info: %v\n", validatorInfo)
	}

	if validatorInfo.State == apiv1.ValidatorStateActiveExiting ||
		validatorInfo.State == apiv1.ValidatorStateActiveSlashed ||
		validatorInfo.State == apiv1.ValidatorStateExitedUnslashed ||
		validatorInfo.State == apiv1.ValidatorStateExitedSlashed ||
		validatorInfo.State == apiv1.ValidatorStateWithdrawalPossible ||
		validatorInfo.State == apiv1.ValidatorStateWithdrawalDone {
		return false, fmt.Sprintf("validator is in state %v, not suitable to generate an exit", validatorInfo.State)
	}

	return true, ""
}

func (c *command) broadcastOperations(ctx context.Context) error {
	for _, op := range c.signedOperations {
		if c.debug {
			data, err := json.Marshal(op)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Broadcasting %s\n", string(data))
			}
		}
		if err := c.consensusClient.(consensusclient.VoluntaryExitSubmitter).SubmitVoluntaryExit(ctx, op); err != nil {
			return err
		}
	}

	return nil
}

func (c *command) setup(ctx context.Context) error {
	if c.offline {
		return nil
	}

	// Ensure timeout is at least the minimum.
	if c.timeout < minTimeout {
		if c.debug {
			fmt.Fprintf(os.Stderr, "Increasing timeout to %v\n", minTimeout)
			c.timeout = minTimeout
		}
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
	forkVersion, err := c.obtainExitForkVersion(ctx)
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

	copy(c.domain[:], c.chainInfo.VoluntaryExitDomainType[:])
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

func (c *command) obtainExitForkVersion(_ context.Context) (phase0.Version, error) {
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
		// Use the Capella fork version for generating an exit as per the spec.
		copy(forkVersion[:], c.chainInfo.ExitForkVersion[:])
	}

	if c.debug {
		fmt.Fprintf(os.Stderr, "Using fork version %#x\n", forkVersion)
	}

	return forkVersion, nil
}
