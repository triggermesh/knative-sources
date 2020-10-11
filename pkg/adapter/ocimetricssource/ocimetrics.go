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

package ocimetricssource

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/monitoring"
	"github.com/triggermesh/knative-sources/pkg/apis/sources/v1alpha1"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"go.uber.org/zap"
	"knative.dev/eventing/pkg/adapter/v2"
)

type OCIMetricsAPIHandler interface {
	Start(ctx context.Context) error
}

const (
	serverPort                = "8080"
	serverShutdownGracePeriod = time.Second * 10
)

type ociMetricsAPIHandler struct {
	provider         common.ConfigurationProvider
	client           monitoring.MonitoringClient
	interval         int
	intervalType     time.Duration
	metricsQuery     string
	metricsNamespace string
	tenant           string
	ceClient         cloudevents.Client
	eventsource      string
	srv              *http.Server // This is used mainly for the health check
	context          context.Context

	logger *zap.SugaredLogger
}

func NewOciMetricsAPIHandler(ceClient cloudevents.Client, aEnv adapter.EnvConfigAccessor, eventsource string, logger *zap.SugaredLogger) OCIMetricsAPIHandler {
	var intervalType time.Duration
	env := aEnv.(*envAccessor)

	intervalStr := env.PollingFrequency
	it := string(intervalStr[len(intervalStr)-1])

	interval, err := strconv.Atoi(strings.TrimSuffix(intervalStr, it))
	if err != nil {
		logger.Panicw("cannot parse polling frequency", zap.Error(err))
	}

	// Parse the interval type
	switch it {
	case "m":
		intervalType = time.Minute
		if interval <= 0 {
			logger.Panic("interval minute is out of range")
		}
	case "h":
		intervalType = time.Hour
		if interval <= 0 {
			logger.Panic("interval hour is out of range")
		}
	case "d":
		intervalType = time.Hour * 24
		if interval <= 0 {
			logger.Panic("interval day is out of range")
		}
	}

	provider := common.NewRawConfigurationProvider(env.TenantOCID, env.UserOCID, env.OracleRegion, env.OracleApiKeyFingerprint, env.OracleApiKey, &env.OracleApiKeyPassphrase)

	monitoringClient, err := monitoring.NewMonitoringClientWithConfigurationProvider(provider)
	if err != nil {
		logger.Panicw("unable to create client", zap.Error(err))
	}

	return &ociMetricsAPIHandler{
		provider:         provider,
		ceClient:         ceClient,
		eventsource:      eventsource,
		logger:           logger,
		interval:         interval,
		intervalType:     intervalType,
		metricsQuery:     env.MetricsQuery,
		metricsNamespace: env.MetricsNamespace,
		tenant:           env.TenantOCID,
		client:           monitoringClient,
	}
}

func (o *ociMetricsAPIHandler) Start(ctx context.Context) error {
	o.logger.Info("Startinng OCI Metrics event handler with interval: ", o.interval)

	// Setup a timer for polling the metrics endpoint
	poll := time.NewTicker(time.Duration(o.interval) * o.intervalType)
	metricsCh := make(chan bool)
	defer poll.Stop()
	defer close(metricsCh)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	o.context = ctx

	// handle stop signals
	go func() {
		<-ctx.Done()
		o.logger.Info("Shutdown signal received. Terminating")
		o.srv.SetKeepAlivesEnabled(false)
		cancel()
	}()

	// fire the initial metrics request, and then start polling
	go func() {
		o.collectMetrics(time.Now())

		for {
			select {
			case <-metricsCh:
				return
			case t := <-poll.C:
				o.collectMetrics(t)
			}
		}
	}()

	http.HandleFunc("/health", healthCheckHandler)
	o.srv = &http.Server{
		Addr: ":" + serverPort,
	}

	serverErrCh := make(chan error)
	defer close(serverErrCh)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		o.logger.Infof("OCI Metrics Source is ready to handle requests at %s", o.srv.Addr)
		serverErrCh <- o.srv.ListenAndServe()
		wg.Done()
	}()

	var err error

	select {
	case serverErr := <-serverErrCh:
		if serverErr != nil {
			err = fmt.Errorf("failure during runtime of OCI Metrics event handler: %w", serverErr)
		}
		cancel()

	case <-ctx.Done():
		o.logger.Info("Shutting server down")

		// stop polling for metrics data
		metricsCh <- true

		shutdownCtx, cancelTimeout := context.WithTimeout(ctx, serverShutdownGracePeriod)
		defer cancelTimeout()
		if shutdownErr := o.srv.Shutdown(shutdownCtx); shutdownErr != nil {
			err = fmt.Errorf("error during server shutdown: %w", shutdownErr)
		}

		// unblock server goroutine
		<-serverErrCh
	}

	wg.Wait()
	return err
}

func (o *ociMetricsAPIHandler) collectMetrics(startTime time.Time) {
	o.logger.Debug("Firing metrics")

	reqDetails := monitoring.SummarizeMetricsDataDetails{
		Namespace: &o.metricsNamespace,
		Query:     &o.metricsQuery,
		StartTime: &common.SDKTime{Time: startTime.Add(o.intervalType * time.Duration(-o.interval))},
	}
	req := monitoring.SummarizeMetricsDataRequest{
		CompartmentId:               &o.tenant,
		SummarizeMetricsDataDetails: reqDetails,
	}

	response, err := o.client.SummarizeMetricsData(o.context, req)
	if err != nil {
		o.logger.Errorw("unable retrieving metrics", zap.Error(err))
	}

	event, err := o.cloudEventFromEventWrapper(&response)
	if err != nil {
		o.logger.Errorw("unable to package metrics", zap.Error(err))
	}

	if result := o.ceClient.Send(context.Background(), *event); !cloudevents.IsACK(result) {
		o.logger.Errorw("unable to send metrics", zap.Error(err))
	}
}

func (o *ociMetricsAPIHandler) cloudEventFromEventWrapper(response *monitoring.SummarizeMetricsDataResponse) (*cloudevents.Event, error) {
	event := cloudevents.NewEvent(cloudevents.VersionV1)

	event.SetID(*response.OpcRequestId)
	event.SetType(v1alpha1.OCIMetricsGenericEventType)
	event.SetSource(o.eventsource)
	if err := event.SetData(cloudevents.ApplicationJSON, response.Items); err != nil {
		return nil, err
	}

	return &event, nil
}

func healthCheckHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}
