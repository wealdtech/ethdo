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

package accountcreate

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/attestantio/dirk/testing/daemon"
	"github.com/attestantio/dirk/testing/resources"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	e2types "github.com/wealdtech/go-eth2-types/v2"
	dirk "github.com/wealdtech/go-eth2-wallet-dirk"
	keystorev4 "github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"
	hd "github.com/wealdtech/go-eth2-wallet-hd/v2"
	nd "github.com/wealdtech/go-eth2-wallet-nd/v2"
	scratch "github.com/wealdtech/go-eth2-wallet-store-scratch"
	"google.golang.org/grpc/credentials"
)

func TestProcess(t *testing.T) {
	require.NoError(t, e2types.InitBLS())

	testNDWallet, err := nd.CreateWallet(context.Background(),
		"Test",
		scratch.New(),
		keystorev4.New(),
	)
	require.NoError(t, err)
	testHDWallet, err := hd.CreateWallet(context.Background(),
		"Test",
		[]byte("pass"),
		scratch.New(),
		keystorev4.New(),
		[]byte{
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		},
	)
	require.NoError(t, err)

	// #nosec G404
	port1 := uint32(12000 + rand.Intn(4000))
	// #nosec G404
	port2 := uint32(12000 + rand.Intn(4000))
	// #nosec G404
	port3 := uint32(12000 + rand.Intn(4000))
	peers := map[uint64]string{
		1: fmt.Sprintf("signer-test01:%d", port1),
		2: fmt.Sprintf("signer-test02:%d", port2),
		3: fmt.Sprintf("signer-test03:%d", port3),
	}
	_, path, err := daemon.New(context.Background(), "", 1, port1, peers)
	require.NoError(t, err)
	defer os.RemoveAll(path)
	_, path, err = daemon.New(context.Background(), "", 2, port2, peers)
	require.NoError(t, err)
	defer os.RemoveAll(path)
	_, path, err = daemon.New(context.Background(), "", 3, port3, peers)
	require.NoError(t, err)
	defer os.RemoveAll(path)
	endpoints := []*dirk.Endpoint{
		dirk.NewEndpoint("signer-test01", port1),
		dirk.NewEndpoint("signer-test02", port2),
		dirk.NewEndpoint("signer-test03", port3),
	}
	credentials, err := credentialsFromCerts(context.Background(), resources.ClientTest01Crt, resources.ClientTest01Key, resources.CACrt)
	require.NoError(t, err)
	testDistributedWallet, err := dirk.OpenWallet(context.Background(), "Wallet 3", credentials, endpoints)
	require.NoError(t, err)

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
			name: "WalletPassphraseIncorrect",
			dataIn: &dataIn{
				timeout:          5 * time.Second,
				wallet:           testHDWallet,
				accountName:      "Good",
				passphrase:       "ce%NohGhah4ye5ra",
				walletPassphrase: "bad",
				participants:     1,
				signingThreshold: 1,
			},
			err: "failed to unlock wallet: incorrect passphrase",
		},
		{
			name: "PassphraseMissing",
			dataIn: &dataIn{
				timeout:          5 * time.Second,
				wallet:           testHDWallet,
				accountName:      "Good",
				passphrase:       "",
				walletPassphrase: "pass",
				participants:     1,
				signingThreshold: 1,
			},
			err: "passphrase is required",
		},
		{
			name: "PassphraseWeak",
			dataIn: &dataIn{
				timeout:          5 * time.Second,
				wallet:           testHDWallet,
				accountName:      "Good",
				passphrase:       "poor",
				walletPassphrase: "pass",
				participants:     1,
				signingThreshold: 1,
			},
			err: "supplied passphrase is weak; use a stronger one or run with the --allow-weak-passphrases flag",
		},
		{
			name: "Good",
			dataIn: &dataIn{
				timeout:          5 * time.Second,
				wallet:           testHDWallet,
				accountName:      "Good",
				passphrase:       "ce%NohGhah4ye5ra",
				walletPassphrase: "pass",
				participants:     1,
				signingThreshold: 1,
			},
		},
		{
			name: "PathMalformed",
			dataIn: &dataIn{
				timeout:          5 * time.Second,
				wallet:           testHDWallet,
				accountName:      "Pathed",
				passphrase:       "ce%NohGhah4ye5ra",
				walletPassphrase: "pass",
				participants:     1,
				signingThreshold: 1,
				path:             "n/12381/3600/1/2/3",
			},
			err: "path does not match expected format m/...",
		},
		{
			name: "PathPassphraseMissing",
			dataIn: &dataIn{
				timeout:          5 * time.Second,
				wallet:           testHDWallet,
				accountName:      "Pathed",
				passphrase:       "",
				walletPassphrase: "pass",
				participants:     1,
				signingThreshold: 1,
				path:             "m/12381/3600/1/2/3",
			},
			err: "passphrase is required",
		},
		{
			name: "PathNotSupported",
			dataIn: &dataIn{
				timeout:          5 * time.Second,
				wallet:           testNDWallet,
				accountName:      "Pathed",
				passphrase:       "ce%NohGhah4ye5ra",
				walletPassphrase: "pass",
				participants:     1,
				signingThreshold: 1,
				path:             "m/12381/3600/1/2/3",
			},
			err: "wallet does not support account creation with an explicit path",
		},
		{
			name: "GoodWithPath",
			dataIn: &dataIn{
				timeout:          5 * time.Second,
				wallet:           testHDWallet,
				accountName:      "Pathed",
				passphrase:       "ce%NohGhah4ye5ra",
				walletPassphrase: "pass",
				participants:     1,
				signingThreshold: 1,
				path:             "m/12381/3600/1/2/3",
			},
		},
		{
			name: "DistributedSigningThresholdZero",
			dataIn: &dataIn{
				timeout:          5 * time.Second,
				wallet:           testDistributedWallet,
				accountName:      "Remote",
				passphrase:       "ce%NohGhah4ye5ra",
				walletPassphrase: "pass",
				participants:     3,
				signingThreshold: 0,
			},
			err: "signing threshold required",
		},
		{
			name: "DistributedSigningThresholdNotHalf",
			dataIn: &dataIn{
				timeout:          5 * time.Second,
				wallet:           testDistributedWallet,
				accountName:      "Remote",
				passphrase:       "ce%NohGhah4ye5ra",
				walletPassphrase: "pass",
				participants:     3,
				signingThreshold: 1,
			},
			err: "signing threshold must be more than half the number of participants",
		},
		{
			name: "DistributedSigningThresholdTooHigh",
			dataIn: &dataIn{
				timeout:          5 * time.Second,
				wallet:           testDistributedWallet,
				accountName:      "Remote",
				passphrase:       "ce%NohGhah4ye5ra",
				walletPassphrase: "pass",
				participants:     3,
				signingThreshold: 4,
			},
			err: "signing threshold cannot be higher than the number of participants",
		},
		{
			name: "DistributedNotSupported",
			dataIn: &dataIn{
				timeout:          5 * time.Second,
				wallet:           testNDWallet,
				accountName:      "Remote",
				passphrase:       "ce%NohGhah4ye5ra",
				walletPassphrase: "pass",
				participants:     3,
				signingThreshold: 2,
			},
			err: "wallet does not support distributed account creation",
		},
		{
			name: "DistributedGood",
			dataIn: &dataIn{
				timeout:          5 * time.Second,
				wallet:           testDistributedWallet,
				accountName:      "Remote",
				passphrase:       "ce%NohGhah4ye5ra",
				walletPassphrase: "pass",
				participants:     3,
				signingThreshold: 2,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := process(context.Background(), test.dataIn)
			if test.err != "" {
				require.EqualError(t, err, test.err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.dataIn.accountName, res.account.Name())
			}
		})
	}
}

func TestNilData(t *testing.T) {
	_, err := processStandard(context.Background(), nil)
	require.EqualError(t, err, "no data")
	_, err = processPathed(context.Background(), nil)
	require.EqualError(t, err, "no data")
	_, err = processDistributed(context.Background(), nil)
	require.EqualError(t, err, "no data")
}

func credentialsFromCerts(ctx context.Context, clientCert []byte, clientKey []byte, caCert []byte) (credentials.TransportCredentials, error) {
	clientPair, err := tls.X509KeyPair(clientCert, clientKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load client keypair")
	}

	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{clientPair},
		MinVersion:   tls.VersionTLS13,
	}

	if caCert != nil {
		cp := x509.NewCertPool()
		if !cp.AppendCertsFromPEM(caCert) {
			return nil, errors.New("failed to add CA certificate")
		}
		tlsCfg.RootCAs = cp
	}

	return credentials.NewTLS(tlsCfg), nil
}
