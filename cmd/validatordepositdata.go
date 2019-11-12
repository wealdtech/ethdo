// Copyright © 2019 Weald Technology Trading
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

	"github.com/prysmaticlabs/go-ssz"
	"github.com/spf13/cobra"
	types "github.com/wealdtech/go-eth2-types"
	util "github.com/wealdtech/go-eth2-util"
	string2eth "github.com/wealdtech/go-string2eth"
)

var validatorDepositDataValidatorAccount string
var validatorDepositDataWithdrawalAccount string
var validatorDepositDataDepositValue string

var validatorDepositDataCmd = &cobra.Command{
	Use:   "depositdata",
	Short: "Generate deposit data for a validator",
	Long: `Generate data for a deposit to the Ethereum 1 validator contract.  For example:

    ethdo validator depositdata --validatoraccount=primary/validator --withdrawalaccount=primary/current --value="32 Ether"

In quiet mode this will return 0 if the the data can be generated correctly, otherwise 1.

The information generated can be passed to ethereal to create a deposit from the Ethereum 1 chain.

In quiet mode this will return 0 if the the data can be generated correctly, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		assert(validatorDepositDataValidatorAccount != "", "--validatoraccount is required")
		validatorAccount, err := accountFromPath(validatorDepositDataValidatorAccount)
		errCheck(err, "Failed to obtain validator account")
		outputIf(debug, fmt.Sprintf("Validator public key is 0x048%x", validatorAccount.PublicKey().Marshal()))

		assert(validatorDepositDataWithdrawalAccount != "", "--withdrawalaccount is required")
		withdrawalAccount, err := accountFromPath(validatorDepositDataWithdrawalAccount)
		errCheck(err, "Failed to obtain withdrawal account")
		outputIf(debug, fmt.Sprintf("Withdrawal public key is 0x048%x", withdrawalAccount.PublicKey().Marshal()))

		withdrawalCredentials := util.SHA256(withdrawalAccount.PublicKey().Marshal())
		errCheck(err, "Failed to hash withdrawal credentials")
		withdrawalCredentials[0] = byte(0) // BLSWithdrawalPrefix
		outputIf(debug, fmt.Sprintf("Withdrawal credentials are 0x%032x", withdrawalCredentials))

		assert(validatorDepositDataDepositValue != "", "--depositvalue is required")
		val, err := string2eth.StringToGWei(validatorDepositDataDepositValue)
		errCheck(err, "Invalid value")
		assert(val >= 1000000000, "deposit value must be at least 1 Ether")

		depositData := struct {
			PubKey                []byte `ssz-size:"48"`
			WithdrawalCredentials []byte `ssz-size:"32"`
			Value                 uint64
			Signature             []byte `ssz-size:"96"`
		}{
			PubKey:                validatorAccount.PublicKey().Marshal(),
			WithdrawalCredentials: withdrawalCredentials,
			Value:                 val,
		}

		signingRoot, err := ssz.SigningRoot(depositData)
		errCheck(err, "Failed to generate deposit data signing root")
		outputIf(debug, fmt.Sprintf("Signing root is %x", signingRoot))
		domain := types.Domain(types.DomainDeposit, []byte{0, 0, 0, 0})
		signature, err := sign(validatorDepositDataValidatorAccount, signingRoot[:], domain)
		errCheck(err, "Failed to sign deposit data signing root")
		depositData.Signature = signature.Marshal()
		outputIf(debug, fmt.Sprintf("Deposit data signature is %x", depositData.Signature))

		depositDataRoot, err := ssz.HashTreeRoot(depositData)
		errCheck(err, "Failed to generate deposit data root")
		outputIf(debug, fmt.Sprintf("Deposit data root is %x", depositDataRoot))

		outputIf(!quiet, fmt.Sprintf(`{"pubkey":"%048x","withdrawal_credentials":"%032x","signature":"%096x","value":%d,"deposit_data_root":"%032x"}`, depositData.PubKey, depositData.WithdrawalCredentials, depositData.Signature, val, depositDataRoot))
		os.Exit(0)
	},
}

func init() {
	validatorCmd.AddCommand(validatorDepositDataCmd)
	validatorFlags(validatorDepositDataCmd)
	validatorDepositDataCmd.Flags().StringVar(&validatorDepositDataValidatorAccount, "validatoraccount", "", "Account of the account carrying out the validation")
	validatorDepositDataCmd.Flags().StringVar(&validatorDepositDataWithdrawalAccount, "withdrawalaccount", "", "Account of the account to which the validator funds will be withdrawn")
	validatorDepositDataCmd.Flags().StringVar(&validatorDepositDataDepositValue, "depositvalue", "", "Value of the amount to be deposited")
}
