// Copyright © 2021 Weald Technology Trading.
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

package standard

import (
	eth2client "github.com/attestantio/go-eth2-client"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type parameters struct {
	logLevel             zerolog.Level
	genesisTimeProvider  eth2client.GenesisTimeProvider
	specProvider         eth2client.SpecProvider
	forkScheduleProvider eth2client.ForkScheduleProvider
}

// Parameter is the interface for service parameters.
type Parameter interface {
	apply(*parameters)
}

type parameterFunc func(*parameters)

func (f parameterFunc) apply(p *parameters) {
	f(p)
}

// WithLogLevel sets the log level for the module.
func WithLogLevel(logLevel zerolog.Level) Parameter {
	return parameterFunc(func(p *parameters) {
		p.logLevel = logLevel
	})
}

// WithGenesisTimeProvider sets the genesis time provider.
func WithGenesisTimeProvider(provider eth2client.GenesisTimeProvider) Parameter {
	return parameterFunc(func(p *parameters) {
		p.genesisTimeProvider = provider
	})
}

// WithSpecProvider sets the spec provider.
func WithSpecProvider(provider eth2client.SpecProvider) Parameter {
	return parameterFunc(func(p *parameters) {
		p.specProvider = provider
	})
}

// WithForkScheduleProvider sets the fork schedule provider.
func WithForkScheduleProvider(provider eth2client.ForkScheduleProvider) Parameter {
	return parameterFunc(func(p *parameters) {
		p.forkScheduleProvider = provider
	})
}

// parseAndCheckParameters parses and checks parameters to ensure that mandatory parameters are present and correct.
func parseAndCheckParameters(params ...Parameter) (*parameters, error) {
	parameters := parameters{
		logLevel: zerolog.GlobalLevel(),
	}
	for _, p := range params {
		if params != nil {
			p.apply(&parameters)
		}
	}

	if parameters.specProvider == nil {
		return nil, errors.New("no spec provider specified")
	}
	if parameters.genesisTimeProvider == nil {
		return nil, errors.New("no genesis time provider specified")
	}
	if parameters.forkScheduleProvider == nil {
		return nil, errors.New("no fork schedule provider specified")
	}

	return &parameters, nil
}
