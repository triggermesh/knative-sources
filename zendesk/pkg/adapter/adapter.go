package adapter

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"go.uber.org/zap"

	"knative.dev/eventing/pkg/adapter/v2"
	"knative.dev/pkg/logging"
)

// New adapter implementation
func New(ctx context.Context, aEnv adapter.EnvConfigAccessor, ceClient cloudevents.Client) adapter.Adapter {
	//env := aEnv.(*envAccessor)
	logger := logging.FromContext(ctx)

	return &zendeskAdapter{
		ceClient: ceClient,

		logger: logger,
	}
}

var _ adapter.Adapter = (*zendeskAdapter)(nil)

type zendeskAdapter struct {
	ceClient cloudevents.Client

	threadiness int
	logger      *zap.SugaredLogger
}

const (
	serverPort                = "8080"
	serverShutdownGracePeriod = time.Second * 10
	subscriptionRecheckPeriod = time.Second * 10
)

// Start runs the adapter.
// Returns if stopCh is closed or Send() returns an error.
func (a *zendeskAdapter) Start(stopCh <-chan struct{}) error {
	// ctx gets canceled to stop goroutines
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// handle stop signals
	go func() {
		<-stopCh
		a.logger.Info("Shutdown signal received. Terminating")
		cancel()
	}()

	http.HandleFunc("/", a.handler)
	//http.HandleFunc("/health", healthCheckHandler)

	server := &http.Server{Addr: ":" + serverPort}
	serverErrCh := make(chan error)
	defer close(serverErrCh)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		a.logger.Info("Serving on port " + serverPort)
		serverErrCh <- server.ListenAndServe()
		wg.Done()
	}()

	// /* TODO(antoineco): we should delete the subscription when the source
	//    is deleted by can't do it from the adapter because a) it should
	//    scale to zero b) it shouldn't have access to the Kubernetes API to
	//    read the event source object.
	//    Ref. https://github.com/triggermesh/aws-event-sources/issues/157
	// */
	// wg.Add(1)
	// go func() {
	// 	a.runSubscriptionReconciler(ctx, subscriptionRecheckPeriod)
	// 	wg.Done()
	// }()

	var err error

	select {
	case serverErr := <-serverErrCh:
		if serverErr != nil {
			err = fmt.Errorf("failure during runtime of SNS notification handler: %w", serverErr)
		}
		cancel()

	case <-ctx.Done():
		a.logger.Info("Shutting server down")

		shutdownCtx, cancelTimeout := context.WithTimeout(ctx, serverShutdownGracePeriod)
		defer cancelTimeout()
		if shutdownErr := server.Shutdown(shutdownCtx); shutdownErr != nil {
			err = fmt.Errorf("error during server shutdown: %w", shutdownErr)
		}

		// unblock server goroutine
		<-serverErrCh
	}

	wg.Wait()
	return err
}

func (a *zendeskAdapter) sendCloudEvent(ceCh <-chan cloudevents.Event, stopCh <-chan struct{}) {
	for {
		select {
		case ce := <-ceCh:
			a.logger.Infof("received CloudEvent: %+v", ce)
			if err := a.ceClient.Send(context.Background(), ce); err != nil {
				a.logger.Errorw("failed to send event", zap.String("event", ce.String()), zap.Error(err))
			}
		case <-stopCh:
			a.logger.Infof("received stop signal")
			return
		}
	}
}

func (a *zendeskAdapter) handler(w http.ResponseWriter, r *http.Request) {

	fmt.Println("-------------------HEADER--------------------------")

	hdr := r.Header

	for key, element := range hdr {
		fmt.Println("Key:", key, "=>", "Element:", element)
	}
	fmt.Println("------------------BODY--------------------------")

	r.ParseForm()

	for key, value := range r.Form {
		fmt.Printf("%s = %s\n", key, value)
	}

	fmt.Println("------------------AUTH--------------------------")

	s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(s) != 2 {
		http.Error(w, "Not authorized", 401)
		return
	}

	b, err := base64.StdEncoding.DecodeString(s[1])
	if err != nil {
		http.Error(w, err.Error(), 401)
		return
	}

	pair := strings.SplitN(string(b), ":", 2)
	if len(pair) != 2 {
		http.Error(w, "Not authorized", 401)
		return
	}

	if pair[0] != "username" || pair[1] != "password" {
		http.Error(w, "Not authorized", 401)
		return
	}

	fmt.Println("authenticated")

	// do a back flip
	// if result := a.ceClient.Send(context.Background(), *event); !cloudevents.IsACK(result) {
	// 	h.handleError(err, w)
	// }

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}

// func sendCloudEvent() error {
// 	event := cloudevents.NewEvent(cloudevents.VersionV1)

// 	event.SetID(wrapper.EventID)
// 	event.SetType(v1alpha1.SlackSourceEventType)
// 	event.SetSource(wrapper.TeamID)
// 	event.SetExtension("api_app_id", wrapper.APIAppID)
// 	event.SetTime(time.Unix(int64(wrapper.EventTime), 0))
// 	event.SetSubject(wrapper.Event.Type())
// 	if err := event.SetData(cloudevents.ApplicationJSON, wrapper.Event); err != nil {
// 		return nil, err
// 	}

// 	return nil
// }
