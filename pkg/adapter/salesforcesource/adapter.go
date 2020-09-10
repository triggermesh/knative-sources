/*
Copyright (c) 2020 TriggerMesh Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package salesforcesource

import (
	"context"
	"encoding/json"
	"net/http"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"knative.dev/eventing/pkg/adapter/v2"
	"knative.dev/pkg/logging"

	sfclient "github.com/triggermesh/knative-sources/pkg/adapter/salesforcesource/client"
)

type salesforceAdapter struct {
	client sfclient.Bayeux

	ceClient cloudevents.Client
	logger   *zap.SugaredLogger
}

var _ adapter.Adapter = (*salesforceAdapter)(nil)

// NewAdapter implementation
func NewAdapter(ctx context.Context, aEnv adapter.EnvConfigAccessor, ceClient cloudevents.Client) adapter.Adapter {
	env := aEnv.(*envAccessor)
	logger := logging.FromContext(ctx)

	adapter := &salesforceAdapter{
		ceClient: ceClient,
		logger:   logging.FromContext(ctx),
	}

	creds, err := sfclient.AuthenticateCredentialsJWT(env.CertKey, env.ClientID, env.User, env.AuthServer, http.DefaultClient)
	if err != nil {
		logger.Panic(err)
	}
	adapter.client = sfclient.NewBayeux(ctx, creds, env.Version, env.Subscriptions, adapter.handle, http.DefaultClient, logger.Named("bayeux"))

	return adapter
}

// Start runs the handler.
func (a *salesforceAdapter) Start(ctx context.Context) error {
	return a.client.Start()
}

func (a *salesforceAdapter) handle(msg *sfclient.ConnectResponse) {
	event := cloudevents.NewEvent(cloudevents.VersionV1)

	// TODO REPLACE THIS!!!
	event.SetType("salesforce.stream")
	event.SetSource("source name")
	event.SetID(uuid.New().String())
	if err := event.SetData(cloudevents.ApplicationJSON, msg.Data); err != nil {
		a.logger.Error("failed to set event data: %w", err)
		return
	}
	event.SetSubject(subjectNameFromConnectResponse(msg))

	if result := a.ceClient.Send(context.Background(), event); !cloudevents.IsACK(result) {
		a.logger.Error("could not send CloudEvent: %s", event.String())
	}
}

func subjectNameFromConnectResponse(msg *sfclient.ConnectResponse) string {

	// if ChangeDataCapture look for entity/operation
	cdc := &sfclient.ChangeDataCapturePayload{}
	if err := json.Unmarshal(msg.Data.Payload, cdc); err == nil {
		return cdc.ChangeEventHeader.EntityName + "/" + cdc.ChangeEventHeader.ChangeType
	}

	// if PushTopic look for object-name/event-operation
	ptso := &sfclient.PushTopicSObject{}
	if err := json.Unmarshal(msg.Data.Payload, ptso); err == nil {
		return ptso.Name + "/" + msg.Data.Event.Type
	}

	// default to channel
	return msg.Channel
}
