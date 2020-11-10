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

package validatorexit

import (
	"context"
	"os"
	"testing"
	"time"

	api "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/auto"
	spec "github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"github.com/wealdtech/ethdo/testutil"
	e2types "github.com/wealdtech/go-eth2-types/v2"
	e2wallet "github.com/wealdtech/go-eth2-wallet"
	keystorev4 "github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"
	nd "github.com/wealdtech/go-eth2-wallet-nd/v2"
	scratch "github.com/wealdtech/go-eth2-wallet-store-scratch"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

func TestProcess(t *testing.T) {
	if os.Getenv("ETHDO_TEST_CONNECTION") == "" {
		t.Skip("ETHDO_TEST_CONNECTION not configured; cannot run tests")
	}

	require.NoError(t, e2types.InitBLS())

	eth2Client, err := auto.New(context.Background(),
		auto.WithAddress(os.Getenv("ETHDO_TEST_CONNECTION")),
	)
	require.NoError(t, err)
	store := scratch.New()
	require.NoError(t, e2wallet.UseStore(store))
	testWallet, err := nd.CreateWallet(context.Background(), "Test wallet", store, keystorev4.New())
	require.NoError(t, err)
	require.NoError(t, testWallet.(e2wtypes.WalletLocker).Unlock(context.Background(), nil))
	viper.Set("passphrase", "pass")
	interop0, err := testWallet.(e2wtypes.WalletAccountImporter).ImportAccount(context.Background(),
		"Interop 0",
		testutil.HexToBytes("0x25295f0d1d592a90b333e26e85149708208e9f8e8bc18f6c77bd62f8ad7a6866"),
		[]byte("pass"),
	)
	require.NoError(t, err)

	//		activeValidator := &api.Validator{
	//			Index:   123,
	//			Balance: 32123456789,
	//			Status:  api.ValidatorStateActiveOngoing,
	//			Validator: &spec.Validator{
	//				PublicKey:                  testutil.HexToPubKey("0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c"),
	//				WithdrawalCredentials:      nil,
	//				EffectiveBalance:           32000000000,
	//				Slashed:                    false,
	//				ActivationEligibilityEpoch: 0,
	//				ActivationEpoch:            0,
	//				ExitEpoch:                  0,
	//				WithdrawableEpoch:          0,
	//			},
	//		}

	epochFork := &spec.Fork{
		PreviousVersion: spec.Version{0x00, 0x00, 0x00, 0x00},
		CurrentVersion:  spec.Version{0x00, 0x00, 0x00, 0x00},
		Epoch:           0,
	}

	tests := []struct {
		name   string
		dataIn *dataIn
		err    string
	}{
		{
			name: "Nil",
			err:  "no data",
		},
		{
			name: "EpochTooLate",
			dataIn: &dataIn{
				timeout:      5 * time.Second,
				eth2Client:   eth2Client,
				fork:         epochFork,
				currentEpoch: 10,
				account:      interop0,
				passphrases:  []string{"pass"},
				epoch:        9999999,
				domain:       spec.Domain{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f},
			},
			err: "not generating exit for an epoch in the far future",
		},
		{
			name: "AccountUnknown",
			dataIn: &dataIn{
				timeout:      5 * time.Second,
				eth2Client:   eth2Client,
				fork:         epochFork,
				currentEpoch: 10,
				account:      interop0,
				passphrases:  []string{"pass"},
				epoch:        10,
				domain:       spec.Domain{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f},
			},
			err: "validator not known by beacon node",
		},
		// {
		// 	name: "Good",
		// 	dataIn: &dataIn{
		// 		timeout:      5 * time.Second,
		// 		eth2Client:   eth2Client,
		// 		fork:         epochFork,
		// 		currentEpoch: 10,
		// 		account:      interop0,
		// 		passphrases:  []string{"pass"},
		// 		epoch:        10,
		// 		domain:       spec.Domain{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f},
		// 	},
		// },
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := process(context.Background(), test.dataIn)
			if test.err != "" {
				require.EqualError(t, err, test.err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGenerateExit(t *testing.T) {
	activeValidator := &api.Validator{
		Index:   123,
		Balance: 32123456789,
		Status:  api.ValidatorStateActiveOngoing,
		Validator: &spec.Validator{
			PublicKey:                  testutil.HexToPubKey("0xa99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c"),
			WithdrawalCredentials:      nil,
			EffectiveBalance:           32000000000,
			Slashed:                    false,
			ActivationEligibilityEpoch: 0,
			ActivationEpoch:            0,
			ExitEpoch:                  0,
			WithdrawableEpoch:          0,
		},
	}

	tests := []struct {
		name      string
		validator *api.Validator
		dataIn    *dataIn
		err       string
	}{
		{
			name: "Nil",
			err:  "no data",
		},
		{
			name: "SignedVoluntaryExitGood",
			dataIn: &dataIn{
				signedVoluntaryExit: &spec.SignedVoluntaryExit{
					Message: &spec.VoluntaryExit{
						Epoch:          spec.Epoch(123),
						ValidatorIndex: spec.ValidatorIndex(456),
					},
					Signature: spec.BLSSignature{
						0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
						0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f,
						0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x2f,
						0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x3a, 0x3b, 0x3c, 0x3d, 0x3e, 0x3f,
						0x40, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48, 0x49, 0x4a, 0x4b, 0x4c, 0x4d, 0x4e, 0x4f,
						0x50, 0x51, 0x52, 0x53, 0x54, 0x55, 0x56, 0x57, 0x58, 0x59, 0x5a, 0x5b, 0x5c, 0x5d, 0x5e, 0x5f,
					},
				},
			},
		},
		{
			name:   "ValidatorMissing",
			dataIn: &dataIn{},
			err:    "no validator",
		},
		{
			name:      "ValidatorGood",
			dataIn:    &dataIn{},
			validator: activeValidator,
		},
		{
			name: "Good",
			dataIn: &dataIn{
				signedVoluntaryExit: &spec.SignedVoluntaryExit{
					Message: &spec.VoluntaryExit{
						Epoch:          spec.Epoch(123),
						ValidatorIndex: spec.ValidatorIndex(456),
					},
					Signature: spec.BLSSignature{
						0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
						0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f,
						0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x2f,
						0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x3a, 0x3b, 0x3c, 0x3d, 0x3e, 0x3f,
						0x40, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48, 0x49, 0x4a, 0x4b, 0x4c, 0x4d, 0x4e, 0x4f,
						0x50, 0x51, 0x52, 0x53, 0x54, 0x55, 0x56, 0x57, 0x58, 0x59, 0x5a, 0x5b, 0x5c, 0x5d, 0x5e, 0x5f,
					},
				},
			},
			validator: activeValidator,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := generateExit(context.Background(), test.dataIn, test.validator)
			if test.err != "" {
				require.EqualError(t, err, test.err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
