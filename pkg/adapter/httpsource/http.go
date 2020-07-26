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

package httpsource

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"k8s.io/apimachinery/pkg/util/uuid"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"go.uber.org/zap"
)

const (
	serverPort                = "8080"
	serverShutdownGracePeriod = time.Second * 10
)

type httpHandler struct {
	eventType   string
	eventSource string

	username string
	password string

	ceClient cloudevents.Client
	srv      *http.Server

	logger *zap.SugaredLogger
}

// Start the server for receiving Http events. Will block until the stop channel closes.
func (h *httpHandler) Start(ctx context.Context) error {
	h.logger.Info("Starting Http event handler...")

	m := http.NewServeMux()
	m.HandleFunc("/", h.handleAll)
	http.HandleFunc("/health", healthCheckHandler)

	h.srv = &http.Server{
		Addr:    ":" + serverPort,
		Handler: m,
	}

	done := make(chan bool, 1)
	go h.gracefulShutdown(ctx.Done(), done)

	h.logger.Infof("Http Source is ready to handle requests at %s", h.srv.Addr)
	if err := h.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		// if an error occurs listening we don't want a graceful shutdown, the
		// server is not serving requests. Return and let the done channel die.
		return fmt.Errorf("could not listen on %s: %w", h.srv.Addr, err)
	}

	<-done
	h.logger.Infof("Server stopped")
	return nil
}

// handleAll receives all Http events at a single resource, it
// is up to this function to parse event wrapper and dispatch.
func (h *httpHandler) handleAll(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		h.handleError(errors.New("request without body not supported"), http.StatusBadRequest, w)
		return
	}

	if h.username != "" && h.password != "" {
		us, ps, ok := r.BasicAuth()
		if !ok {
			h.handleError(errors.New("Wrong authentication header"), http.StatusBadRequest, w)
			return
		}
		if us != h.username || ps != h.password {
			h.handleError(errors.New("Credentials are not valid"), http.StatusUnauthorized, w)
			return
		}
	}

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		h.handleError(err, http.StatusInternalServerError, w)
		return
	}

	event := cloudevents.NewEvent(cloudevents.VersionV1)
	event.SetType(h.eventType)
	event.SetSource(h.eventSource)
	event.SetID(string(uuid.NewUUID()))

	if err := event.SetData(cloudevents.ApplicationJSON, body); err != nil {
		h.handleError(fmt.Errorf("failed to set event data: %w", err), http.StatusInternalServerError, w)
		return
	}

	if result := h.ceClient.Send(context.Background(), event); !cloudevents.IsACK(result) {
		h.handleError(fmt.Errorf("could not send Cloud Event: %w", result), http.StatusInternalServerError, w)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (h *httpHandler) handleError(err error, code int, w http.ResponseWriter) {
	h.logger.Error("An error ocurred", zap.Error(err))
	http.Error(w, err.Error(), code)
}

func healthCheckHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

func (h *httpHandler) gracefulShutdown(stopCh <-chan struct{}, done chan<- bool) {
	<-stopCh
	h.logger.Info("Server is shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), serverShutdownGracePeriod)
	defer cancel()

	h.srv.SetKeepAlivesEnabled(false)
	if err := h.srv.Shutdown(ctx); err != nil {
		h.logger.Fatalf("Could not gracefully shutdown the server: %v", err)
	}
	close(done)
}
