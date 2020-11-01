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

package grpc

import (
	"context"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

// Connect connects to an Ethereum 2 endpoint.
func Connect() (*grpc.ClientConn, error) {
	connection := ""
	if viper.GetString("connection") != "" {
		connection = viper.GetString("connection")
	}

	if connection == "" {
		return nil, errors.New("no connection")
	}
	// outputIf(debug, fmt.Sprintf("Connecting to %s", connection))

	opts := []grpc.DialOption{grpc.WithInsecure()}

	ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
	defer cancel()
	return grpc.DialContext(ctx, connection, opts...)
}
