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

	return &zendeskAdapter{
		handler: NewZendeskAPIHandler(ceClient, defaultListenPort, env.Token, logger.Named("handler")),
		logger:  logger,
	}
}

var _ adapter.Adapter = (*zendeskAdapter)(nil)

type zendeskAdapter struct {
	handler ZendeskAPIHandler
	logger  *zap.SugaredLogger
}

// Start runs the Zendesk handler.
func (a *zendeskAdapter) Start(stopCh <-chan struct{}) error {
	return a.handler.Start(stopCh)
}

// package adapter

// import (
// 	"context"
// 	"encoding/base64"
// 	"fmt"
// 	"net/http"
// 	"strings"
// 	"sync"
// 	"time"

// 	cloudevents "github.com/cloudevents/sdk-go/v2"
// 	"go.uber.org/zap"

// 	"knative.dev/eventing/pkg/adapter/v2"
// 	"knative.dev/pkg/logging"
// )

// // New adapter implementation
// func New(ctx context.Context, aEnv adapter.EnvConfigAccessor, ceClient cloudevents.Client) adapter.Adapter {
// 	//env := aEnv.(*envAccessor)
// 	logger := logging.FromContext(ctx)

// 	return &zendeskAdapter{
// 		ceClient: ceClient,

// 		logger: logger,
// 	}
// }

// var _ adapter.Adapter = (*zendeskAdapter)(nil)

// type zendeskAdapter struct {
// 	ceClient cloudevents.Client
// 	logger   *zap.SugaredLogger
// }

// const (
// 	serverPort                = "8080"
// 	serverShutdownGracePeriod = time.Second * 10
// 	subscriptionRecheckPeriod = time.Second * 10
// )

// // Start runs the adapter.
// // Returns if stopCh is closed or Send() returns an error.
// func (a *zendeskAdapter) Start(stopCh <-chan struct{}) error {
// 	// ctx gets canceled to stop goroutines
// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel()

// 	// handle stop signals
// 	go func() {
// 		<-stopCh
// 		a.logger.Info("Shutdown signal received. Terminating")
// 		cancel()
// 	}()

// 	http.HandleFunc("/", a.handler)
// 	//http.HandleFunc("/health", healthCheckHandler)

// 	server := &http.Server{Addr: ":" + serverPort}
// 	serverErrCh := make(chan error)
// 	defer close(serverErrCh)

// 	var wg sync.WaitGroup

// 	wg.Add(1)
// 	go func() {
// 		a.logger.Info("Serving on port " + serverPort)
// 		serverErrCh <- server.ListenAndServe()
// 		wg.Done()
// 	}()

// 	var err error

// 	select {
// 	case serverErr := <-serverErrCh:
// 		if serverErr != nil {
// 			err = fmt.Errorf("failure during runtime of notification handler: %w", serverErr)
// 		}
// 		cancel()

// 	case <-ctx.Done():
// 		a.logger.Info("Shutting server down")

// 		shutdownCtx, cancelTimeout := context.WithTimeout(ctx, serverShutdownGracePeriod)
// 		defer cancelTimeout()
// 		if shutdownErr := server.Shutdown(shutdownCtx); shutdownErr != nil {
// 			err = fmt.Errorf("error during server shutdown: %w", shutdownErr)
// 		}

// 		// unblock server goroutine
// 		<-serverErrCh
// 	}

// 	wg.Wait()
// 	return err
// }

// func (a *zendeskAdapter) sendCloudEvent(ceCh <-chan cloudevents.Event, stopCh <-chan struct{}) {
// 	for {
// 		select {
// 		case ce := <-ceCh:
// 			a.logger.Infof("received CloudEvent: %+v", ce)
// 			if err := a.ceClient.Send(context.Background(), ce); err != nil {
// 				a.logger.Errorw("failed to send event", zap.String("event", ce.String()), zap.Error(err))
// 			}
// 		case <-stopCh:
// 			a.logger.Infof("received stop signal")
// 			return
// 		}
// 	}
// }

// func (a *zendeskAdapter) handler(w http.ResponseWriter, r *http.Request) {

// 	s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
// 	if len(s) != 2 {
// 		http.Error(w, "Not authorized", 401)
// 		return
// 	}

// 	b, err := base64.StdEncoding.DecodeString(s[1])
// 	if err != nil {
// 		http.Error(w, err.Error(), 401)
// 		return
// 	}

// 	pair := strings.SplitN(string(b), ":", 2)
// 	if len(pair) != 2 {
// 		http.Error(w, "Not authorized", 401)
// 		return
// 	}

// 	if pair[0] != "username" || pair[1] != "password" {
// 		http.Error(w, "Not authorized", 401)
// 		return
// 	}

// 	fmt.Println("authenticated")

// 	// do a back flip
// 	//create and return cloud event
// 	//a.sendCloudEvent()

// 	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
// 	w.WriteHeader(http.StatusOK)
// 	fmt.Fprintln(w, "OK")
// }
