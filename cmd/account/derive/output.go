// Copyright © 2020, 2023 Weald Technology Trading
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

package accountderive

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/wealdtech/ethdo/util"
	e2types "github.com/wealdtech/go-eth2-types/v2"
	ethutil "github.com/wealdtech/go-eth2-util"
	keystorev4 "github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"
)

type dataOut struct {
	showPrivateKey            bool
	showWithdrawalCredentials bool
	generateKeystore          bool
	key                       *e2types.BLSPrivateKey
	path                      string
}

func output(ctx context.Context, data *dataOut) (string, error) {
	if data == nil {
		return "", errors.New("no data")
	}
	if data.key == nil {
		return "", errors.New("no key")
	}

	if data.generateKeystore {
		return outputKeystore(ctx, data)
	}

	builder := strings.Builder{}

	if data.showPrivateKey {
		builder.WriteString(fmt.Sprintf("Private key: %#x\n", data.key.Marshal()))
	}
	if data.showWithdrawalCredentials {
		withdrawalCredentials := ethutil.SHA256(data.key.PublicKey().Marshal())
		withdrawalCredentials[0] = byte(0) // BLS_WITHDRAWAL_PREFIX
		builder.WriteString(fmt.Sprintf("Withdrawal credentials: %#x\n", withdrawalCredentials))
	}
	if !(data.showPrivateKey || data.showWithdrawalCredentials) {
		builder.WriteString(fmt.Sprintf("Public key: %#x\n", data.key.PublicKey().Marshal()))
	}

	return builder.String(), nil
}

func outputKeystore(_ context.Context, data *dataOut) (string, error) {
	passphrase, err := util.GetPassphrase()
	if err != nil {
		return "", errors.New("no passphrase supplied")
	}

	encryptor := keystorev4.New()
	crypto, err := encryptor.Encrypt(data.key.Marshal(), passphrase)
	if err != nil {
		return "", errors.New("failed to encrypt private key")
	}

	uuid, err := uuid.NewRandom()
	if err != nil {
		return "", errors.New("failed to generate UUID")
	}
	ks := make(map[string]interface{})
	ks["uuid"] = uuid.String()
	ks["pubkey"] = hex.EncodeToString(data.key.PublicKey().Marshal())
	ks["version"] = 4
	ks["path"] = data.path
	ks["crypto"] = crypto
	out, err := json.Marshal(ks)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal keystore JSON")
	}

	keystoreFilename := fmt.Sprintf("keystore-%s-%d.json", strings.ReplaceAll(data.path, "/", "_"), time.Now().Unix())

	if err := os.WriteFile(keystoreFilename, out, 0o600); err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("failed to write %s", keystoreFilename))
	}
	return "", nil
}
