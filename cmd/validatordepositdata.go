// Copyright © 2019, 2020 Weald Technology Trading
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
	"fmt"
	"os"
	"strings"

	"github.com/prysmaticlabs/go-ssz"
	"github.com/spf13/cobra"
	e2types "github.com/wealdtech/go-eth2-types/v2"
	util "github.com/wealdtech/go-eth2-util"
	string2eth "github.com/wealdtech/go-string2eth"
)

var validatorDepositDataValidatorAccount string
var validatorDepositDataWithdrawalAccount string
var validatorDepositDataDepositValue string
var validatorDepositDataRaw bool

var validatorDepositDataCmd = &cobra.Command{
	Use:   "depositdata",
	Short: "Generate deposit data for one or more validators",
	Long: `Generate data for deposits to the Ethereum 1 validator contract.  For example:

    ethdo validator depositdata --validatoraccount=primary/validator --withdrawalaccount=primary/current --value="32 Ether"

If validatoraccount is provided with an account path it will generate deposit data for all matching accounts.

The information generated can be passed to ethereal to create a deposit from the Ethereum 1 chain.

In quiet mode this will return 0 if the the data can be generated correctly, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		assert(validatorDepositDataValidatorAccount != "", "--validatoraccount is required")
		validatorWallet, err := walletFromPath(validatorDepositDataValidatorAccount)
		errCheck(err, "Failed to obtain validator wallet")
		validatorAccounts, err := accountsFromPath(validatorDepositDataValidatorAccount)
		errCheck(err, "Failed to obtain validator account")
		assert(len(validatorAccounts) > 0, "Failed to obtain validator account")
		if len(validatorAccounts) == 1 {
			outputIf(debug, fmt.Sprintf("Validator public key is %048x", validatorAccounts[0].PublicKey().Marshal()))
		} else {
			for _, validatorAccount := range validatorAccounts {
				outputIf(verbose, fmt.Sprintf("Creating deposit for %s/%s", validatorWallet.Name(), validatorAccount.Name()))
				outputIf(debug, fmt.Sprintf("Validator public key is %048x", validatorAccount.PublicKey().Marshal()))
			}
		}

		assert(validatorDepositDataWithdrawalAccount != "", "--withdrawalaccount is required")
		withdrawalAccount, err := accountFromPath(validatorDepositDataWithdrawalAccount)
		errCheck(err, "Failed to obtain withdrawal account")
		outputIf(debug, fmt.Sprintf("Withdrawal public key is %048x", withdrawalAccount.PublicKey().Marshal()))

		withdrawalCredentials := util.SHA256(withdrawalAccount.PublicKey().Marshal())
		errCheck(err, "Failed to hash withdrawal credentials")
		// TODO fetch this from the node.
		withdrawalCredentials[0] = byte(0) // BLSWithdrawalPrefix
		outputIf(debug, fmt.Sprintf("Withdrawal credentials are %032x", withdrawalCredentials))

		assert(validatorDepositDataDepositValue != "", "--depositvalue is required")
		val, err := string2eth.StringToGWei(validatorDepositDataDepositValue)
		errCheck(err, "Invalid value")
		// TODO fetch this from the node.
		assert(val >= 1000000000, "deposit value must be at least 1 Ether")

		// For each key, generate deposit data
		outputs := make([]string, 0)
		for _, validatorAccount := range validatorAccounts {
			depositData := struct {
				PubKey                []byte `ssz-size:"48"`
				WithdrawalCredentials []byte `ssz-size:"32"`
				Value                 uint64
			}{
				PubKey:                validatorAccount.PublicKey().Marshal(),
				WithdrawalCredentials: withdrawalCredentials,
				Value:                 val,
			}
			outputIf(debug, fmt.Sprintf("Deposit data:\n\tPublic key: %x\n\tWithdrawal credentials: %x\n\tValue: %d", depositData.PubKey, depositData.WithdrawalCredentials, depositData.Value))
			domain := e2types.Domain(e2types.DomainDeposit, e2types.ZeroForkVersion, e2types.ZeroGenesisValidatorsRoot)
			outputIf(debug, fmt.Sprintf("Domain is %x", domain))
			err = validatorAccount.Unlock([]byte(rootAccountPassphrase))
			errCheck(err, "Failed to unlock validator account")
			signature, err := signStruct(validatorAccount, depositData, domain)
			validatorAccount.Lock()
			errCheck(err, "Failed to generate deposit data signature")

			signedDepositData := struct {
				PubKey                []byte `ssz-size:"48"`
				WithdrawalCredentials []byte `ssz-size:"32"`
				Value                 uint64
				Signature             []byte `ssz-size:"96"`
			}{
				PubKey:                validatorAccount.PublicKey().Marshal(),
				WithdrawalCredentials: withdrawalCredentials,
				Value:                 val,
				Signature:             signature.Marshal(),
			}
			outputIf(debug, fmt.Sprintf("Signed deposit data:\n\tPublic key: %x\n\tWithdrawal credentials: %x\n\tValue: %d\n\tSignature: %x", signedDepositData.PubKey, signedDepositData.WithdrawalCredentials, signedDepositData.Value, signedDepositData.Signature))

			depositDataRoot, err := ssz.HashTreeRoot(signedDepositData)
			errCheck(err, "Failed to generate deposit data root")
			outputIf(debug, fmt.Sprintf("Deposit data root is %x", depositDataRoot))

			if validatorDepositDataRaw {
				// Build a raw transaction by hand
				txData := []byte{0x22, 0x89, 0x51, 0x18}
				// Pointer to validator public key
				txData = append(txData, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80}...)
				// Pointer to withdrawal credentials
				txData = append(txData, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xe0}...)
				// Pointer to validator signature
				txData = append(txData, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x20}...)
				// Deposit data root
				txData = append(txData, depositDataRoot[:]...)
				// Validator public key (pad to 32-byte boundary)
				txData = append(txData, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x30}...)
				txData = append(txData, validatorAccount.PublicKey().Marshal()...)
				txData = append(txData, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}...)
				// Withdrawal credentials
				txData = append(txData, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x20}...)
				txData = append(txData, withdrawalCredentials...)
				// Deposit signature
				txData = append(txData, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x60}...)
				txData = append(txData, signedDepositData.Signature...)
				outputs = append(outputs, fmt.Sprintf("%#x", txData))
			} else {
				outputs = append(outputs, fmt.Sprintf(`{"account":"%s","pubkey":"%048x","withdrawal_credentials":"%032x","signature":"%096x","value":%d,"deposit_data_root":"%032x","version":1}`, fmt.Sprintf("%s/%s", validatorWallet.Name(), validatorAccount.Name()), signedDepositData.PubKey, signedDepositData.WithdrawalCredentials, signedDepositData.Signature, val, depositDataRoot))
			}
		}

		if quiet {
			os.Exit(0)
		}

		if len(outputs) == 1 {
			fmt.Printf("%s\n", outputs[0])
		} else {
			fmt.Printf("[")
			fmt.Print(strings.Join(outputs, ","))
			fmt.Println("]")
		}
	},
}

func init() {
	validatorCmd.AddCommand(validatorDepositDataCmd)
	validatorFlags(validatorDepositDataCmd)
	validatorDepositDataCmd.Flags().StringVar(&validatorDepositDataValidatorAccount, "validatoraccount", "", "Account of the account carrying out the validation")
	validatorDepositDataCmd.Flags().StringVar(&validatorDepositDataWithdrawalAccount, "withdrawalaccount", "", "Account of the account to which the validator funds will be withdrawn")
	validatorDepositDataCmd.Flags().StringVar(&validatorDepositDataDepositValue, "depositvalue", "", "Value of the amount to be deposited")
	validatorDepositDataCmd.Flags().BoolVar(&validatorDepositDataRaw, "raw", false, "Print raw deposit data transaction data")
}
