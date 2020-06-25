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

package adapter

import (
	"context"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"go.uber.org/zap"

	"knative.dev/eventing/pkg/adapter/v2"
	"knative.dev/pkg/logging"
)

const defaultListenPort = 8080

// New adapter implementation
func New(ctx context.Context, aEnv adapter.EnvConfigAccessor, ceClient cloudevents.Client) adapter.Adapter {
	env := aEnv.(*envAccessor)
	logger := logging.FromContext(ctx)

	return &slackAdapter{
		handler: NewSlackEventAPIHandler(ceClient, defaultListenPort, env.Token, env.AppID, logger.Named("handler")),
		logger:  logger,
	}
}

var _ adapter.Adapter = (*slackAdapter)(nil)

type slackAdapter struct {
	handler SlackEventAPIHandler
	logger  *zap.SugaredLogger
}

// Start runs the Slack handler.
func (a *slackAdapter) Start(stopCh <-chan struct{}) error {
	return a.handler.Start(stopCh)
}
