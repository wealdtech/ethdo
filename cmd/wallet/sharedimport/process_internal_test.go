// Copyright © 2021, 2022 Weald Technology Trading
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

package walletsharedimport

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	e2types "github.com/wealdtech/go-eth2-types/v2"
)

func TestProcess(t *testing.T) {
	require.NoError(t, e2types.InitBLS())

	export := []byte(`{"version":1,"participants":5,"threshold":3,"data":"0x0106951ed83407552b501d97a31ee7bf6655450723dfcb0b8448690ce85838b7ba563cf536edf58bbb04f22cab8baee062c602175768d6419965545da206062b40cefe9887d2e89250b96cf99de1fcb2cc462b3eeb6b60128df66d5540edb93cfbdc805d353bf0223ca3f5c1c223f19928af742a54f2c2a60491f6fdae4bc5abc621babb625b6ec3610c3ce7943826b79b0cf3b1a84dbbe6b09c7edc87628269775576d2a1047689f31035ac1847e5b6e2511a86e58948478bbf885b814059a3f1b7c72c312f4e9fd6d962847e2c38f3bdc8df5deacfb2b7fddf851e4324a2433ebbcd0598bee8b493c27a1951bab894f1963dcc262ca1b47bda15f620d2d8d5006e5f798071db64f40a980ac77c759ab3f116d66a160d7516c92afd7d38be2681cfdbd750e6133c4d50e5555d9a9b69d223f389da737e352338f8c0e4b96e413362afc3561975a397715ef2fcbbf270b1d8a5ef41fafa6fb7241c4664627b420a2b40d06a5706ebcb005a39a7ed066fa13a206e396a572bab94829de52550d912ddbe2ee85b8775bb5886eb783426e3c79c2129bbe87b6be777cb79d70294f2541fcccc9bad8f603774c843ee5c7cabcb2bd5b6d160bd7e871e5cf90d4aca4e1e521089fc6d131ba3f9c0a6c0bd942837d598a78c8fd7a1c45409fba388ba1d16433acd93122c964d930a7dc5c5018128f5243a752d3cb56e4d7e607508490818b0237777543c90e2048a4fedf20b453adc2fac7aa4824d6805ed258de66c0d51f9d37cc616f1f84e0873dbb9edff03de8ba5839b898b55eb549ee34f4e587f6dd5a2bc892b0f11caebc33b314239d9567cca1477318c708cedba6c301e9c8edf58f46a7b4a07883c2dff30fe54eaf243718ccc464f276bb4045e72081248238eb9855d8b5f993c2b1e6049c95e5622685857016c2e72a89b322a24399f4476f4a3f7c0e219e06f8e46939d29874bebd5fb24407ca260ef1db362a79403c46776e5b205f956771d14aec6b4c54340a655acf5396ef9487e7acd8a154ff4392a79d35377ece9c09fbb114a935ff0f18b4469b9f94436d7b1790920e2cfcad4b7e187d6ecb47dc23336366baa8a70b3536a7df2489bf12d92aade034a185e5cc0a349229431e37b7f587d1dedd6a41cbe3452b7186fa25f1f22d7d17ae5750b42640b973f4503cb129beaa07f7fbb08bf09292336c96a1666da36c481904df944f74a5bc003a5b9e41a47b8240a996991e23d60f83d96590a67a621c780840fd6a256627d1202550e2b7b8c10d7e43dad01a88ce9757effacb82494948c94dd6eadf4452e2d396fd135eee347672ac33a4d224d9b79ff9438c46073aae6a1104606ca5a44d52f2b2ac93b7fe60b4db61a738d4f5db87ab92d987bb176d374a6306b7d5f4c974ac17153fed99aa8826a579c6806b74b25f21d7b232098d8845dbfb2645849dc4daefb9d9cc1079062af37dc9b976a9915803ce96abe3786ab5bc3d7a62c7a698a5f75a6a65c4aaa6972800ce8dbbd43b682d3f8ecd2a3074d14082ed60d7f969de5d59c66e0f4f3812bbdc536a92947e1d027c63d8595737d58cb62887237eb4ef9c704345677e1faf9b9ef0c524a28e8703e2814e897fabdaae1b2cd71360d19ff6c35275ecccfc834682b9094b66f42877942fec0a1b620eea4d6c8ce7a128c2bf07d77b448330abbe4c2405f769fd790f67a6adaa678677dd2238e77e60a0c324c2ad73a9e499b8cd4282d8ac6337e291563b5df2507d4ec8fe2ca568ac5d6af10448233a2900d6833a8cf6cdc06142b95410b3b21976ce95bfca87512805a70f7d193dceb25d62d12b280863b3165f11ee3f411f39132bd9e618e00225fbf0f9e39f2af15c1ec6cddc3dfb81089a69a0a8db9befcb3f987cb5f5f288b3259369ecc904145cf6bc2977a49c2977058886601155cd974951b37e2dc828cf2edbf3a60c1a8ba5ebdc27ed83a95ce8af9fdc4e0b6213fdcc02f8576d05f9ffe387ada68a4f0e3e538eb6be8433bb90be816e1f9a34e3d6fd1c60c46380ec1307e18011befd9f6399ece3e82001a32c5991055c5363b544bf7ec66d01e6eb26da41c382fe7954817fdc7a2d067569758897277e88b20f4cd93d4f61a61a609757b08d3579677262db5aef082d0e79fab11f52c9d86c0768890df96957dbbb4d425d5d271c4b18394e2b0c4f7c89b9a"}`)
	shares := []string{
		"d04a162f3f648647acbfc5af0475041c3f64c3d72752ddc52ab53786802ed7dfea3929488dbefb3af582e713fe967a6ff24c86757186abe7d93afdcd81cdff4f8a",
		"f06533d9efae8b015a5b9c73d2b3652b5e0c80fa9a948fdcfbda3d4bd54ae31573c8649ffb0a8900dfe1cecb740b0c3a477938f3e01244cac39a068612beb72bbe",
		"682b8e6256ce6a4fde515060a326214f7a3789b79c11e2cb53e5b185d522d196ca0b76dea7a03d739ec87605ede429a9f214dfb06703dbb143d8d5b56413d7a0a7",
		"53ccad137def6fcbaac0ccfff0fdbb02ab3fa4ce075b221f15a80203a318f29f09cfc7a40b29c910675791f847e3e72dc6f80e74b80f517512c1fd6be14ff5b2ff",
		"ed2166659f7b5412a169ec83627386bc6ff1a31e67735d405b2bf7cb122ad7ced35c87e42c8e8f7ba90b5899a94be506687a9c5b353af2a216018d9f1bf61745a5",
	}

	dir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	datFile := filepath.Join(dir, "backup.dat")
	require.NoError(t, os.WriteFile(datFile, export, 0o600))
	defer os.RemoveAll(dir)

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
			name: "FileMissing",
			dataIn: &dataIn{
				timeout: 5 * time.Second,
			},
			err: "import file is required",
		},
		{
			name: "FileBad",
			dataIn: &dataIn{
				timeout: 5 * time.Second,
				file:    []byte("\001\002"),
				shares: []string{
					shares[0],
					shares[1],
					shares[2],
				},
			},
			err: "failed to unmarshal export: invalid character '\\x01' looking for beginning of value",
		},
		{
			name: "SharesTooLow",
			dataIn: &dataIn{
				timeout: 5 * time.Second,
				file:    export,
				shares: []string{
					shares[0],
					shares[1],
				},
			},
			err: "import requires 3 shares, 2 were provided",
		},
		{
			name: "SharesTooHigh",
			dataIn: &dataIn{
				timeout: 5 * time.Second,
				file:    export,
				shares: []string{
					shares[0],
					shares[1],
					shares[2],
					shares[3],
				},
			},
			err: "import requires 3 shares, 4 were provided",
		},
		{
			name: "ShareBad",
			dataIn: &dataIn{
				timeout: 5 * time.Second,
				file:    export,
				shares: []string{
					"xxx",
					shares[1],
					shares[2],
				},
			},
			err: "invalid share: encoding/hex: invalid byte: U+0078 'x'",
		},
		{
			name: "Good",
			dataIn: &dataIn{
				timeout: 5 * time.Second,
				file:    export,
				shares: []string{
					shares[0],
					shares[1],
					shares[2],
				},
			},
		},
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
