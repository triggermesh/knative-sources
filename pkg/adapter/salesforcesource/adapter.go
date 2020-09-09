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
	"errors"
	"fmt"
	"net/http"

	cloudevents "github.com/cloudevents/sdk-go/v2"
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

	creds, err := sfclient.AuthenticateCredentialsJWT(env.CertKey, env.ClientID, env.User, env.AuthServer, http.DefaultClient)
	client := sfclient.NewBayeux(ctx, creds, env.Version, http.DefaultClient)

	if err != nil {
		logger.Panic(err)
	}

	return &salesforceAdapter{
		client: client,

		ceClient: ceClient,
		logger:   logging.FromContext(ctx),
	}
}

// Start runs the handler.
func (a *salesforceAdapter) Start(ctx context.Context) error {
	if err := a.client.Handshake(); err != nil {
		return fmt.Errorf("error handshaking Salesforce: %w", err)
	}
	if err := a.client.Connect(); err != nil {
		return fmt.Errorf("error connecting to Salesforce: %w", err)
	}

	return errors.New("not implemented")
}
