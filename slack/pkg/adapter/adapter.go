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
	"sync"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/wait"

	"knative.dev/eventing/pkg/adapter/v2"
	"knative.dev/pkg/logging"
)

// New adapter implementation
func New(ctx context.Context, aEnv adapter.EnvConfigAccessor, ceClient cloudevents.Client) adapter.Adapter {
	env := aEnv.(*envAccessor)
	logger := logging.FromContext(ctx)

	return &slackAdapter{
		client:      ceClient,
		token:       env.Token,
		threadiness: env.Threadiness,
		logger:      logger,
	}
}

var _ adapter.Adapter = (*slackAdapter)(nil)

type slackAdapter struct {
	client cloudevents.Client

	token       string
	threadiness int
	logger      *zap.SugaredLogger
}

// Start runs the adapter.
// Returns if stopCh is closed or Send() returns an error.
func (a *slackAdapter) Start(stopCh <-chan struct{}) error {
	a.logger.Info("Starting slack adapter")
	ceCh := make(chan cloudevents.Event)

	wg := sync.WaitGroup{}
	for i := 1; i <= a.threadiness; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			a.sendCloudEvent(ceCh, stopCh)
		}()
	}

	wait.Until(func() {
		p := newProcessor(a.token, a.logger, ceCh)
		p.Run(stopCh)
	}, 2*time.Second, stopCh)

	wg.Wait()
	return nil
}

func (a *slackAdapter) sendCloudEvent(ceCh <-chan cloudevents.Event, stopCh <-chan struct{}) {
	for {
		select {
		case ce := <-ceCh:
			a.logger.Infof("received CloudEvent: %+v", ce)
			if err := a.client.Send(context.Background(), ce); err != nil {
				a.logger.Errorw("failed to send event", zap.String("event", ce.String()), zap.Error(err))
			}
		case <-stopCh:
			a.logger.Infof("received stop signal")
			return
		}
	}
}
