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
		return true, nil
	}

	return false, nil

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
	h.logger.Info("got event:")
	h.logger.Info(event)

	// Getting `runtime error: invalid memory address or nil pointer dereference` herre
	// See the full error message at the bottom of this file

	cEvent, err := h.cloudEventFromEventWrapper(event)
	if err != nil {
		h.logger.Info("Error Creating CloudEvent")
		h.handleError(err, w)
	}
	if result := h.ceClient.Send(context.Background(), *cEvent); !cloudevents.IsACK(result) {
		h.logger.Info("Error Sending CloudEvent")
		h.handleError(result, w)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// fix this
	res, err := json.Marshal(`200:Big Pog! Thanks Zendesk!`)
	if err != nil {
		h.handleError(err, w)
	}
	_, err = w.Write(res)
	if err != nil {
		h.logger.Info("Error Writing HTTP response")
		h.handleError(err, w)
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

// fix this
func (h *zendeskAPIHandler) cloudEventFromEventWrapper(wrapper *ZendeskEventWrapper) (*cloudevents.Event, error) {
	h.logger.Info("Proccesing Zendesk event")
	data, err := json.Marshal(wrapper)
	if err != nil {
		return nil, err
	}
	event := cloudevents.NewEvent(cloudevents.VersionV1)

	event.SetID("wrapper.EventID")
	event.SetType("functions.zendessk.sources.triggermesh.io")
	event.SetSource("https://github.com/cloudevents/spec/pull")
	//event.SetExtension("api_app_id", "wrapper.APIAppID")
	//event.SetTime(time.Unix(int64(120), 0))
	event.SetSubject("New Zendesk Ticket")
	if err := event.SetData(cloudevents.ApplicationJSON, data); err != nil {
		return nil, fmt.Errorf("failed to set event data: %w", err)
	}

	return &event, nil
}

// (*ServeMux).ServeHTTP\n\tnet/http/server.go:2387\nnet/http.serverHandler.ServeHTTP\n\tnet/http/server.go:2807\nnet/http.(*conn).serve\n\tnet/http/server.go:1895"}
// 2020/07/01 07:04:29 http: panic serving 10.1.8.1:49486: runtime error: invalid memory address or nil pointer dereference
// goroutine 76 [running]:
// net/http.(*conn).serve.func1(0xc000676780)
// 	net/http/server.go:1772 +0x139
// panic(0x159ca80, 0x2698a20)
// 	runtime/panic.go:975 +0x3e3
// github.com/triggermesh/knative-sources/zendesk/pkg/adapter.(*zendeskAPIHandler).handleError(0xc00017a980, 0x0, 0x0, 0x1ad6920, 0xc0000d2000)
// 	github.com/triggermesh/knative-sources/zendesk/pkg/adapter/zendesk.go:186 +0x1fc
// github.com/triggermesh/knative-sources/zendesk/pkg/adapter.(*zendeskAPIHandler).handleAll(0xc00017a980, 0x1ad6920, 0xc0000d2000, 0xc0000c2c00)
// 	github.com/triggermesh/knative-sources/zendesk/pkg/adapter/zendesk.go:154 +0x623
// net/http.HandlerFunc.ServeHTTP(0xc0004fe490, 0x1ad6920, 0xc0000d2000, 0xc0000c2c00)
// 	net/http/server.go:2012 +0x44
// net/http.(*ServeMux).ServeHTTP(0xc00017aa00, 0x1ad6920, 0xc0000d2000, 0xc0000c2c00)
// 	net/http/server.go:2387 +0x1a5
// net/http.serverHandler.ServeHTTP(0xc000424380, 0x1ad6920, 0xc0000d2000, 0xc0000c2c00)
// 	net/http/server.go:2807 +0xa3
// net/http.(*conn).serve(0xc000676780, 0x1ada0e0, 0xc0006566c0)
// 	net/http/server.go:1895 +0x86c
// created by net/http.(*Server).Serve
// 	net/http/server.go:2933 +0x35c
