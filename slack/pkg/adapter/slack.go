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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/triggermesh/knative-sources/slack/pkg/apis/sources/v1alpha1"
	"go.uber.org/zap"
)

const serverShutdownGracePeriod = time.Second * 10

// SlackEventAPIHandler listen for Slack API Events
type SlackEventAPIHandler interface {
	Start(ctx context.Context) error
}

type slackEventAPIHandler struct {
	port          int
	signingSecret string
	appID         string

	ceClient cloudevents.Client
	srv      *http.Server

	time   timeWrap
	logger *zap.SugaredLogger
}

// NewSlackEventAPIHandler creates the default implementation of the Slack API Events handler
func NewSlackEventAPIHandler(ceClient cloudevents.Client, port int, signingSecret, appID string, tw timeWrap, logger *zap.SugaredLogger) SlackEventAPIHandler {
	return &slackEventAPIHandler{
		port:          port,
		signingSecret: signingSecret,
		appID:         appID,

		ceClient: ceClient,
		time:     tw,
		logger:   logger,
	}
}

// Start the server for receiving Slack callbacks. Will block
// until the stop channel closes.
func (h *slackEventAPIHandler) Start(ctx context.Context) error {
	h.logger.Info("Starting Slack event handler")

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// handle stop signals
	go func() {
		<-ctx.Done()
		h.logger.Info("Shutdown signal received. Terminating")
		cancel()
	}()

	m := http.NewServeMux()
	m.HandleFunc("/", h.handleAll)

	h.srv = &http.Server{
		Addr:    ":" + strconv.Itoa(h.port),
		Handler: m,
	}

	serverErrCh := make(chan error)
	defer close(serverErrCh)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		h.logger.Infof("Server is ready to handle requests at %s", h.srv.Addr)
		serverErrCh <- h.srv.ListenAndServe()
		wg.Done()
	}()

	var err error
	select {
	case serverErr := <-serverErrCh:
		if serverErr != nil {
			err = fmt.Errorf("Shutting server down %w", serverErr)
		}
		cancel()

	case <-ctx.Done():

		shutdownCtx, cancelTimeout := context.WithTimeout(ctx, serverShutdownGracePeriod)
		defer cancelTimeout()

		if shutdownErr := h.srv.Shutdown(shutdownCtx); shutdownErr != nil {
			err = fmt.Errorf("error during server shutdown: %w", shutdownErr)
		}

		// unblock server goroutine
		<-serverErrCh
	}

	wg.Wait()
	h.logger.Infof("Server stopped")
	return err
}

// handleAll receives all Slack events at a single resource, it
// is up to this function to parse event wrapper and dispatch.
func (h *slackEventAPIHandler) handleAll(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		h.handleError(errors.New("request without body not supported"), http.StatusBadRequest, w)
		return
	}

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		h.handleError(err, http.StatusInternalServerError, w)
		return
	}

	if h.signingSecret != "" {
		err = h.verifySigning(r.Header, body)
		if err != nil {
			h.handleError(err, http.StatusUnauthorized, w)
			return
		}
	}

	event := &SlackEventWrapper{}
	err = json.Unmarshal(body, event)
	if err != nil {
		h.handleError(fmt.Errorf("could not unmarshall JSON request: %w", err), http.StatusBadRequest, w)
		return
	}

	// All paths that are not managed by this integration and are
	// not errors need to return 2xx withing 3 seconds to Slack API.
	// Otherwise the message will be retried.
	// See: https://api.slack.com/events-api#receiving_events (Responding to Events)

	if h.appID != "" && event.APIAppID != h.appID {
		// silently ignore, some other integration should take
		// care of this event.
		return
	}

	// There are only 2 documented types to be received from the Events API
	// - `event_callback`, See: https://api.slack.com/events-api#receiving_events
	// - `event_callback`, See: https://api.slack.com/events-api#subscriptions
	switch event.Type {
	case "event_callback":
		h.handleCallback(event, w)

	case "url_verification":
		h.handleChallenge(body, w)

	default:
		h.logger.Warnf("not supported content %q", event.Type)
	}
}

func (h *slackEventAPIHandler) handleError(err error, code int, w http.ResponseWriter) {
	h.logger.Error("An error ocurred", zap.Error(err))
	http.Error(w, err.Error(), code)
}

func (h *slackEventAPIHandler) handleChallenge(body []byte, w http.ResponseWriter) {
	h.logger.Info("Challenge received")
	c := &SlackChallenge{}

	err := json.Unmarshal(body, c)
	if err != nil {
		h.handleError(err, http.StatusBadRequest, w)
		return
	}

	cr := &SlackChallengeResponse{Challenge: c.Challenge}
	res, err := json.Marshal(cr)
	if err != nil {
		h.handleError(err, http.StatusBadRequest, w)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, err = w.Write(res)
	if err != nil {
		h.handleError(err, http.StatusInternalServerError, w)
	}
}

func (h *slackEventAPIHandler) handleCallback(wrapper *SlackEventWrapper, w http.ResponseWriter) {
	h.logger.Info("callback received")

	event, err := cloudEventFromEventWrapper(wrapper)
	if err != nil {
		h.handleError(err, http.StatusBadRequest, w)
		return
	}

	if result := h.ceClient.Send(context.Background(), *event); !cloudevents.IsACK(result) {
		h.handleError(err, http.StatusInternalServerError, w)
	}
}

func cloudEventFromEventWrapper(wrapper *SlackEventWrapper) (*cloudevents.Event, error) {
	event := cloudevents.NewEvent(cloudevents.VersionV1)

	event.SetID(wrapper.EventID)
	event.SetType(v1alpha1.SlackSourceEventType)
	event.SetSource(wrapper.TeamID)
	event.SetExtension("api_app_id", wrapper.APIAppID)
	event.SetTime(time.Unix(int64(wrapper.EventTime), 0))
	event.SetSubject(wrapper.Event.Type())
	if err := event.SetData(cloudevents.ApplicationJSON, wrapper.Event); err != nil {
		return nil, err
	}

	return &event, nil
}
