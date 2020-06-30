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
	"strconv"
	"strings"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/triggermesh/knative-sources/zendesk/pkg/apis/sources/v1alpha1"
	"go.uber.org/zap"
)

// ZendeskAPIHandler listen for Zendesk API Events
type ZendeskAPIHandler interface {
	Start(stopCh <-chan struct{}) error
}

type zendeskAPIHandler struct {
	port  int
	token string

	ceClient cloudevents.Client
	srv      *http.Server

	logger *zap.SugaredLogger
}

// NewZendeskAPIHandler creates the default implementation of the Zendesk API Events handler
func NewZendeskAPIHandler(ceClient cloudevents.Client, port int, token string, logger *zap.SugaredLogger) ZendeskAPIHandler {
	return &zendeskAPIHandler{
		port:  port,
		token: token,

		ceClient: ceClient,
		logger:   logger,
	}
}

// Start the server for receiving Zendesk callbacks. Will block
// until the stop channel closes.
func (h *zendeskAPIHandler) Start(stopCh <-chan struct{}) error {
	h.logger.Info("Starting Zendesk event handler")

	m := http.NewServeMux()
	m.HandleFunc("/", h.handleAll)

	h.srv = &http.Server{
		Addr:    ":" + strconv.Itoa(h.port),
		Handler: m,
	}

	done := make(chan bool, 1)
	go h.gracefulShutdown(stopCh, done)

	h.logger.Infof("Zendesk Source is ready to handle requests at %s", h.srv.Addr)
	if err := h.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("could not listen on %s: %v", h.srv.Addr, err)
	}

	<-done
	h.logger.Infof("Server stopped")
	return nil
}

func (h *zendeskAPIHandler) authenticate(r *http.Request) (bool, error) {

	s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(s) != 2 {
		return false, errors.New("Not authorized")
	}

	b, err := base64.StdEncoding.DecodeString(s[1])
	if err != nil {
		return false, errors.New("Not authorized")
	}

	pair := strings.SplitN(string(b), ":", 2)
	if len(pair) != 2 {

		return false, errors.New("Not authorized")
	}

	if pair[0] != "username" || pair[1] != "password" {
		return false, errors.New("Not authorized")
	}

	return true, nil

}

// handleAll receives all Zendesk events at a single resource, it
// is up to this function to parse event wrapper and dispatch.
func (h *zendeskAPIHandler) handleAll(w http.ResponseWriter, r *http.Request) {

	if r.Body == nil {
		h.handleError(errors.New("request without body not supported"), w)
		return
	}

	authStatus, err := h.authenticate(r)
	if err != nil {
		h.handleError(err, w)
		return
	}

	if authStatus == false {
		h.handleError(errors.New("Authentication FAILED"), w)
		return
	}

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		h.handleError(err, w)
		return
	}

	event := &ZendeskEventWrapper{}
	err = json.Unmarshal(body, event)
	if err != nil {
		h.handleError(fmt.Errorf("could not unmarshall JSON request: %s", err.Error()), w)
		return
	}

	switch event.Type {
	case "event_callback":
		h.handleCallback(event, w)

	default:
		h.logger.Warnf("not supported content %q", event.Type)
	}
}

func (h *zendeskAPIHandler) gracefulShutdown(stopCh <-chan struct{}, done chan<- bool) {
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

func (h *zendeskAPIHandler) handleError(err error, w http.ResponseWriter) {
	h.logger.Error("An error ocurred", zap.Error(err))
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func (h *zendeskAPIHandler) handleCallback(wrapper *ZendeskEventWrapper, w http.ResponseWriter) {
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

func cloudEventFromEventWrapper(wrapper *ZendeskEventWrapper) (*cloudevents.Event, error) {
	event := cloudevents.NewEvent(cloudevents.VersionV1)

	event.SetID(wrapper.EventID)
	event.SetType(v1alpha1.ZendeskSourceEventType)
	event.SetSource(wrapper.TeamID)
	event.SetExtension("api_app_id", wrapper.APIAppID)
	event.SetTime(time.Unix(int64(wrapper.EventTime), 0))
	event.SetSubject(wrapper.Event.Type())
	if err := event.SetData(cloudevents.ApplicationJSON, wrapper.Event); err != nil {
		return nil, err
	}

	return &event, nil
}
