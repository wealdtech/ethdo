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

package util

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	eth2client "github.com/attestantio/go-eth2-client"
	"github.com/attestantio/go-eth2-client/http"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// defaultBeaconNodeAddresses are default REST endpoint addresses for beacon nodes.
var defaultBeaconNodeAddresses = []string{
	"localhost:5052", // Lighthouse, Nimbus
	"localhost:5051", // Teku
	"localhost:3500", // Prysm
}

// fallbackBeaconNode is used if no other connection is supplied.
var fallbackBeaconNode = "http://mainnet-consensus.attestant.io/"

type ConnectOpts struct {
	Address       string
	Timeout       time.Duration
	AllowInsecure bool
	LogFallback   bool
}

// ConnectToBeaconNode connects to a beacon node at the given address.
func ConnectToBeaconNode(ctx context.Context, opts *ConnectOpts) (eth2client.Service, error) {
	if opts == nil {
		return nil, errors.New("no options specified")
	}

	if opts.Timeout == 0 {
		return nil, errors.New("no timeout specified")
	}

	if opts.Address != "" {
		// We have an explicit address; use it.
		return connectToBeaconNode(ctx, opts.Address, opts.Timeout, opts.AllowInsecure)
	}

	// Try the defaults.
	for _, address := range defaultBeaconNodeAddresses {
		client, err := connectToBeaconNode(ctx, address, opts.Timeout, opts.AllowInsecure)
		if err == nil {
			return client, nil
		}
	}

	// The user did not provide a connection, so attempt to use the fallback node.
	if opts.LogFallback {
		fmt.Fprintf(os.Stderr, "No connection supplied with --connection parameter and no local beacon node found, attempting to use mainnet fallback\n")
	}
	client, err := connectToBeaconNode(ctx, fallbackBeaconNode, opts.Timeout, true)
	if err == nil {
		return client, nil
	}

	return nil, errors.New("failed to connect to any beacon node")
}

func connectToBeaconNode(ctx context.Context, address string, timeout time.Duration, allowInsecure bool) (eth2client.Service, error) {
	if !strings.HasPrefix(address, "http") {
		address = fmt.Sprintf("http://%s", address)
	}
	if !allowInsecure {
		// Ensure the connection is either secure or local.
		connectionURL, err := url.Parse(address)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse connection")
		}
		if connectionURL.Scheme == "http" &&
			connectionURL.Host != "localhost" &&
			!strings.HasPrefix(connectionURL.Host, "localhost:") &&
			connectionURL.Host != "127.0.0.1" &&
			!strings.HasPrefix(connectionURL.Host, "127.0.0.1:") {
			fmt.Println("Connections to remote beacon nodes should be secure.  This warning can be silenced with --allow-insecure-connections")
		}
	}
	eth2Client, err := http.New(ctx,
		http.WithLogLevel(zerolog.Disabled),
		http.WithAddress(address),
		http.WithTimeout(timeout),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to beacon node")
	}

	return eth2Client, nil
}
