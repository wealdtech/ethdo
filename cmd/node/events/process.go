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

package nodeevents

import (
	"context"
	"encoding/json"
	"fmt"

	eth2client "github.com/attestantio/go-eth2-client"
	api "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/pkg/errors"
)

func process(ctx context.Context, data *dataIn) error {
	if data == nil {
		return errors.New("no data")
	}

	err := data.eth2Client.(eth2client.EventsProvider).Events(ctx, data.topics, eventHandler)
	if err != nil {
		return errors.Wrap(err, "failed to connect for events")
	}

	<-ctx.Done()

	return nil
}

func eventHandler(event *api.Event) {
	if event.Data == nil {
		return
	}

	data, err := json.Marshal(event)
	if err == nil {
		fmt.Println(string(data))
	}
}
