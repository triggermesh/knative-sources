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
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/triggermesh/knative-sources/slack/pkg/apis/sources/v1alpha1"
	"go.uber.org/zap"
)

// SlackEventAPIHandler listen for Slack API Events
type SlackEventAPIHandler interface {
	Start(stopCh <-chan struct{}) error
}

type slackEventAPIHandler struct {
	port  int
	token string
	appID string

	ceClient cloudevents.Client
	srv      *http.Server

	logger *zap.SugaredLogger
}

// NewSlackEventAPIHandler creates the default implementation of the Slack API Events handler
func NewSlackEventAPIHandler(ceClient cloudevents.Client, port int, token, appID string, logger *zap.SugaredLogger) SlackEventAPIHandler {
	return &slackEventAPIHandler{
		port:  port,
		token: token,
		appID: appID,

		ceClient: ceClient,
		logger:   logger,
	}
}

// Start the server for receiving Slack callbacks. Will block
// until the stop channel closes.
func (h *slackEventAPIHandler) Start(stopCh <-chan struct{}) error {
	h.logger.Info("Starting Slack event handler")

	m := http.NewServeMux()
	m.HandleFunc("/", h.handleAll)

	h.srv = &http.Server{
		Addr:    ":" + strconv.Itoa(h.port),
		Handler: m,
	}

	done := make(chan bool, 1)
	go h.gracefulShutdown(stopCh, done)

	h.logger.Infof("Server is ready to handle requests at %s", h.srv.Addr)
	if err := h.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("could not listen on %s: %v", h.srv.Addr, err)
	}

	<-done
	h.logger.Infof("Server stopped")
	return nil
}

// handleAll receives all Slack events at a single resource, it
// is up to this function to parse event wrapper and dispatch.
func (h *slackEventAPIHandler) handleAll(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		h.handleError(errors.New("request without body not supported"), w)
		return
	}

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		h.handleError(err, w)
		return
	}

	event := &SlackEventWrapper{}
	err = json.Unmarshal(body, event)
	if err != nil {
		h.handleError(fmt.Errorf("could not unmarshall JSON request: %s", err.Error()), w)
		return
	}

	// All paths that are not managed by this integration need
	// to return 2xx withing 3 seconds to Slack API, otherwise
	// the message will be retried.
	// Responses for those cases are returned in order to
	// achieve that, logs are written if relevant.

	if h.appID != "" && event.APIAppID != h.appID {
		// silently ignore, some other integration should take
		// care of this event.
		return
	}

	if h.token != "" && event.Token != h.token {
		h.logger.Error("Received wrong token for this integration")
		return
	}

	switch event.Type {
	case "event_callback":
		h.handleCallback(event, w)

	case "url_verification":
		h.handleChallenge(body, w)

	default:
		h.logger.Warnf("not supported content %q", event.Type)
	}
}

func (h *slackEventAPIHandler) gracefulShutdown(stopCh <-chan struct{}, done chan<- bool) {
	<-stopCh
	h.logger.Info("Server is shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	h.srv.SetKeepAlivesEnabled(false)
	if err := h.srv.Shutdown(ctx); err != nil {
		h.logger.Fatalf("Could not gracefully shutdown the server: %v", err)
	}
	close(done)
}

func (h *slackEventAPIHandler) handleError(err error, w http.ResponseWriter) {
	h.logger.Error("An error ocurred", zap.Error(err))
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func (h *slackEventAPIHandler) handleChallenge(body []byte, w http.ResponseWriter) {
	h.logger.Info("Challenge received")
	c := &SlackChallenge{}

	err := json.Unmarshal(body, c)
	if err != nil {
		h.handleError(err, w)
		return
	}

	cr := &SlackChallengeResponse{Challenge: c.Challenge}
	res, err := json.Marshal(cr)
	if err != nil {
		h.handleError(err, w)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, err = w.Write(res)
	if err != nil {
		h.handleError(err, w)
	}
}

func (h *slackEventAPIHandler) handleCallback(wrapper *SlackEventWrapper, w http.ResponseWriter) {
	h.logger.Info("callback received")

	event, err := cloudEventFromEventWrapper(wrapper)
	if err != nil {
		h.handleError(err, w)
		return
	}

	if result := h.ceClient.Send(context.Background(), *event); !cloudevents.IsACK(result) {
		h.handleError(err, w)
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
