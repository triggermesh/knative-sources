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

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/net/publicsuffix"
)

const connectChannel = "/meta/connect"
const subscribeChannel = "/meta/subscribe"
const handshakeChannel = "/meta/handshake"

// Bayeux client according to CometD 3.1.14
// and the subset needed for Salesforce Streaming API.
// See: https://docs.cometd.org/current3/reference/
type Bayeux interface {
	Handshake() error
	Connect() ([]ConnectResponse, error)
	Start() error
	Subscribe(topic string, replayID int) ([]SubscriptionResponse, error)
}

type bayeux struct {
	creds *Credentials

	url           string
	client        *http.Client
	clientID      string
	subscriptions []string

	errCh  chan error
	msgCh  chan *ConnectResponse
	stopCh chan struct{}

	logger *zap.SugaredLogger
	ctx    context.Context
	mutex  sync.RWMutex
}

// NewBayeux creates a Bayeux client for Salesforce Streaming API consumption.
func NewBayeux(ctx context.Context, credentials *Credentials, apiVersion string, subscriptions []string, client *http.Client, logger *zap.SugaredLogger) Bayeux {
	return &bayeux{
		creds: credentials,

		subscriptions: subscriptions,
		url:           credentials.InstanceURL + "/cometd/" + apiVersion,
		client:        client,

		msgCh:  make(chan *ConnectResponse),
		errCh:  make(chan error),
		stopCh: make(chan struct{}),

		logger: logger,
		ctx:    ctx,
	}
}

func (b *bayeux) Start() error {
	if err := b.Handshake(); err != nil {
		return err
	}

	for _, s := range b.subscriptions {
		sbs, err := b.Subscribe(s, -2)
		if err != nil {
			return err
		}
		for _, sb := range sbs {
			if !sb.Successful {
				return fmt.Errorf("could not subscribe to %s: %s", sb.Subscription, sb.Error)
			}
			b.logger.Infof("subscribed to %s", sb.Subscription)
		}
	}

	// Connect loop will run until context is done
	go func() {
		for {
			select {
			case <-b.ctx.Done():
				close(b.stopCh)
				return
			default:
				crs, err := b.Connect()
				if err != nil {
					b.errCh <- err
					continue
				}

				for i := range crs {
					if strings.HasPrefix(crs[i].Channel, "/meta") {
						b.manageMeta(&crs[i])
						continue
					}

					b.msgCh <- &crs[i]
				}
			}
		}
	}()

	// Worker loop will run until the connect loop is stopped
	for {
		select {
		case msg := <-b.msgCh:
			if msg.Channel == connectChannel || msg.Channel == subscribeChannel {
				b.logger.Infof("", msg.Channel, msg.Successful)
			}

			fmt.Printf("received message: %v\n", msg)
		case msg := <-b.errCh:
			fmt.Printf("received error: %v\n", msg)
		case <-b.stopCh:
			return nil
		}
	}

	return nil
}

func (b *bayeux) manageMeta(cr *ConnectResponse) {
	if cr.Successful {
		b.logger.Infof("meta channel (channel: %s client: %s) ok", cr.Channel, cr.ClientID)
		return
	}

	b.logger.Warnf("meta channel (channel: %s client: %s) was not successful: %+v", cr.Channel, cr.ClientID, *cr)

	// Note: this case is based on Bayeux theory and very untested.
	if cr.Advice.Reconnect == "handshake" {
		b.logger.Infof("executing handshake as advised by channel response")
		if err := b.Handshake(); err != nil {
			b.errCh <- err
		}
	}
}

func (b *bayeux) Handshake() error {
	payload := `{"channel": "` + handshakeChannel + `", "supportedConnectionTypes": ["long-polling"], "version": "1.0"}`

	b.mutex.Lock()
	defer b.mutex.Unlock()

	jOpts := cookiejar.Options{publicsuffix.List}
	jar, err := cookiejar.New(&jOpts)
	if err != nil {
		return fmt.Errorf("could not setup cookiejar: %+w", err)
	}

	b.client.Jar = jar
	res, err := b.doPost(payload)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	h := []HandshakeResponse{}
	err = json.NewDecoder(res.Body).Decode(&h)
	if err != nil {
		return fmt.Errorf("could not decode handshake response: %+w", err)
	}

	if len(h) == 0 {
		return errors.New("empty handshake response")
	}

	b.clientID = h[0].ClientID

	return nil
}

// func (b *bayeux) connectLoop() error {
// 	for {
// 		select {
// 		case msg := <-b.msgCh:
// 			fmt.Printf("received message %v\n", msg)
// 		case err := <-b.errCh:
// 			fmt.Printf("received error %v\n", err)
// 		case <-b.ctx.Done():
// 			fmt.Printf("context cancelled\n")
// 		}
// 	}
// }

// Connect will start to receive events.
// This is a blocking function
func (b *bayeux) Connect() ([]ConnectResponse, error) {

	payload := `{"channel": "` + connectChannel + `", "connectionType": "long-polling", "clientId": "` + b.clientID + `"}`

	res, err := b.doPost(payload)
	// TODO check if context is done
	if err != nil {
		//b.errCh <- err
		return nil, fmt.Errorf("error sending connect request: %+w", err)
	}
	defer res.Body.Close()

	// var by []byte
	// if res.Body != nil {
	// 	by, _ = ioutil.ReadAll(res.Body)
	// }
	// // Restore the io.ReadCloser to its original state
	// res.Body = ioutil.NopCloser(bytes.NewBuffer(by))
	// // Use the content
	// s := string(by)
	// fmt.Printf("\nConnect  Body: %s\n\n", s)

	c := []ConnectResponse{}
	err = json.NewDecoder(res.Body).Decode(&c)
	if err != nil {
		return nil, fmt.Errorf("could not decode connect response: %+w", err)
	}

	if len(c) == 0 {
		return nil, errors.New("empty connect response")
	}

	return c, nil
}

func (b *bayeux) Disconnect() {

}

func (b *bayeux) Subscribe(topic string, replayID int) ([]SubscriptionResponse, error) {
	payload := `{"channel": "` + subscribeChannel + `", "subscription": "` + topic + `", "clientId": "` + b.clientID + `","ext":{"replay": {"` + topic + `": "` + strconv.Itoa(replayID) + `"}}}`
	res, err := b.doPost(payload)

	if err != nil {
		return nil, fmt.Errorf("error sending subscription request: %+w", err)
	}
	defer res.Body.Close()

	s := []SubscriptionResponse{}
	err = json.NewDecoder(res.Body).Decode(&s)

	if err != nil {
		return nil, fmt.Errorf("could not decode subscription response: %+w", err)
	}

	if len(s) == 0 {
		return nil, errors.New("empty subscription response")
	}

	return s, nil
}

func (b *bayeux) Unsubscribe() {

}

func (b *bayeux) doPost(payload string) (*http.Response, error) {
	req, err := http.NewRequest("POST", b.url, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		return nil, fmt.Errorf("could not build request: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+b.creds.Token)

	req = req.WithContext(b.ctx)
	res, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not execute request: %w", err)
	}

	if res.StatusCode >= 300 {
		msg := fmt.Sprintf("received unexpected status code %d", res.StatusCode)
		resb, err := ioutil.ReadAll(res.Body)
		if err != nil {
			msg += ": " + string(resb)
		}
		return nil, errors.New(msg)
	}

	return res, nil
}

// CommonResponse for all structures at Bayeux protocol.
type CommonResponse struct {
	Channel    string `json:"channel"`
	ClientID   string `json:"clientId"`
	Successful bool   `json:"successful"`
	Error      string `json:"error,omitempty"`
}

// HandshakeResponse for Bayeux protocol.
type HandshakeResponse struct {
	Channel         string          `json:"channel"`
	Version         string          `json:"version"`
	ConnectionTypes []string        `json:"supportedConnectionTypes"`
	ClientID        string          `json:"clientId"`
	Successful      bool            `json:"successful"`
	Extension       json.RawMessage `json:"ext,omitempty"`
}

// ConnectResponse for Bayeux protocol
type ConnectResponse struct {
	Channel    string `json:"channel"`
	ClientID   string `json:"clientId"`
	Successful bool   `json:"successful"`
	Error      string `json:"error,omitempty"`
	Data       struct {
		Event struct {
			CreatedDate time.Time `json:"createdDate"`
			ReplayID    int       `json:"replayId"`
			Type        string    `json:"type"`
		} `json:"event,omitempty"`
		Schema  string          `json:"schema"`
		SObject json.RawMessage `json:"sobject"`
		Payload json.RawMessage `json:"payload"`
	} `json:"data,omitempty"`
	Advice struct {
		Reconnect string `json:"reconnect,omitempty"`
		Timeout   int    `json:"timeout,omitempty"`
		Interval  int    `json:"interval,omitempty"`
	} `json:"advice,omitempty"`
}

// SubscriptionResponse for Bayeux protocol
type SubscriptionResponse struct {
	CommonResponse `json:",inline"`
	Subscription   string `json:"subscription,omitempty"`
}
