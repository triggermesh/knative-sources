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
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/triggermesh/knative-sources/pkg/apis/sources/v1alpha1"
	"go.uber.org/zap"
)

// ZendeskAPIHandler listen for Zendesk API Events
type ZendeskAPIHandler interface {
	Start(ctx context.Context) error
}

const (
	serverPort                = "8080"
	serverShutdownGracePeriod = time.Second * 10
)

const (
	// auth header prefix, it is important that the blank
	// space is present at the end for string manipulation
	// at auth parsing function.
	authPrefix = "Basic "
)

type zendeskAPIHandler struct {
	username string
	password string

	ceClient    cloudevents.Client
	srv         *http.Server
	eventsource string

	logger *zap.SugaredLogger
}

// NewZendeskAPIHandler creates the default implementation of the Zendesk API Events handler
func NewZendeskAPIHandler(ceClient cloudevents.Client, username, password, eventsource string, logger *zap.SugaredLogger) ZendeskAPIHandler {
	return &zendeskAPIHandler{
		username:    username,
		password:    password,
		eventsource: eventsource,
		ceClient:    ceClient,
		logger:      logger,
	}
}

// Start the server for receiving Zendesk events. Will block until the stop channel closes.
func (h *zendeskAPIHandler) Start(ctx context.Context) error {
	h.logger.Info("Starting Zendesk event handler...")
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// handle stop signals
	go func() {
		<-ctx.Done()
		h.logger.Info("Shutdown signal received. Terminating")
		h.srv.SetKeepAlivesEnabled(false)
		cancel()
	}()

	m := http.NewServeMux()
	m.HandleFunc("/", h.handleAll)
	http.HandleFunc("/health", healthCheckHandler)

	h.srv = &http.Server{
		Addr:    ":" + serverPort,
		Handler: m,
	}

	serverErrCh := make(chan error)
	defer close(serverErrCh)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		h.logger.Infof("Zendesk Source is ready to handle requests at %s", h.srv.Addr)
		serverErrCh <- h.srv.ListenAndServe()
		wg.Done()
	}()

	var err error

	select {
	case serverErr := <-serverErrCh:
		if serverErr != nil {
			err = fmt.Errorf("failure during runtime of Zendesk event handler: %w", serverErr)
		}
		cancel()

	case <-ctx.Done():
		h.logger.Info("Shutting server down")

		shutdownCtx, cancelTimeout := context.WithTimeout(ctx, serverShutdownGracePeriod)
		defer cancelTimeout()
		if shutdownErr := h.srv.Shutdown(shutdownCtx); shutdownErr != nil {
			err = fmt.Errorf("error during server shutdown: %w", shutdownErr)
		}

		// unblock server goroutine
		<-serverErrCh
	}

	wg.Wait()
	return err
}

func (h *zendeskAPIHandler) validateAuthHeader(r *http.Request) error {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, authPrefix) {
		return errors.New("incorrect auth header")
	}

	content, err := base64.StdEncoding.DecodeString(auth[len(authPrefix):])
	if err != nil {
		return errors.New("could not decode the auth header")
	}

	pair := strings.SplitN(string(content), ":", 2)
	if len(pair) != 2 {
		return errors.New("misformated credentials at auth header")
	}

	if pair[0] != h.username || pair[1] != h.password {
		return fmt.Errorf("credentials received for user %q are not valid", pair[0])
	}

	return nil
}

// handleAll receives all Zendesk events at a single resource, it
// is up to this function to parse event wrapper and dispatch.
func (h *zendeskAPIHandler) handleAll(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		h.handleError(errors.New("request without body not supported"), w)
		return
	}

	if err := h.validateAuthHeader(r); err != nil {
		h.handleError(err, w)
		return
	}

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		h.handleError(err, w)
		return
	}

	event := &ZendeskEvent{}
	err = json.Unmarshal(body, event)
	if err != nil {
		h.handleError(fmt.Errorf("could not unmarshall JSON request: %w", err), w)
		return
	}

	cEvent, err := h.cloudEventFromWrapper(event)
	if err != nil {
		h.handleError(fmt.Errorf("could not create Cloud Event: %w", err), w)
	}

	if result := h.ceClient.Send(context.Background(), *cEvent); !cloudevents.IsACK(result) {
		h.handleError(fmt.Errorf("could not send Cloud Event: %w", result), w)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

func (h *zendeskAPIHandler) handleError(err error, w http.ResponseWriter) {
	h.logger.Error("An error ocurred", zap.Error(err))
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func (h *zendeskAPIHandler) cloudEventFromWrapper(ze *ZendeskEvent) (*cloudevents.Event, error) {
	data, err := json.Marshal(ze)
	if err != nil {
		return nil, err
	}
	event := cloudevents.NewEvent(cloudevents.VersionV1)

	if ticketType := ze.Type(); ticketType != "" {
		event.SetExtension("ticket_type", ticketType)
	}
	event.SetID(ze.ID())
	event.SetType(v1alpha1.ZendeskSourceEventType)
	event.SetSource(h.eventsource)
	event.SetSubject(ze.Title())

	if err := event.SetData(cloudevents.ApplicationJSON, data); err != nil {
		return nil, fmt.Errorf("failed to set event data: %w", err)
	}

	return &event, nil
}

func healthCheckHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}
