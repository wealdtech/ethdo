// Copyright © 2020 Weald Technology Trading
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

package util

import (
	"context"
	"fmt"
	"net/url"
	"time"

	eth2client "github.com/attestantio/go-eth2-client"
	"github.com/attestantio/go-eth2-client/auto"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// ConnectToBeaconNode connects to a beacon node at the given address.
func ConnectToBeaconNode(ctx context.Context, address string, timeout time.Duration, allowInsecure bool) (eth2client.Service, error) {
	if timeout == 0 {
		return nil, errors.New("no timeout specified")
	}

	if !allowInsecure {
		// Ensure the connection is either secure or local.
		connectionURL, err := url.Parse(address)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse connection")
		}
		if connectionURL.Scheme == "http" &&
			connectionURL.Host != "localhost" &&
			connectionURL.Host != "127.0.0.1" {
			fmt.Println("Connections to remote beacon nodes should be secure.  This warning can be silenced with --allow-insecure-connections")
		}
	}
	eth2Client, err := auto.New(ctx,
		auto.WithLogLevel(zerolog.Disabled),
		auto.WithAddress(address),
		auto.WithTimeout(timeout),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to beacon node")
	}

	return eth2Client, nil
}
