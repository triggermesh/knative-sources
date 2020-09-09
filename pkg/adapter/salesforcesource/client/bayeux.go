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
	"sync"
	"time"

	"golang.org/x/net/publicsuffix"
)

// Bayeux client according to CometD 3.1.14
// and the subset needed for Salesforce Streaming API.
// See: https://docs.cometd.org/current3/reference/
type Bayeux interface {
	Handshake() error
	Connect() error
}

type bayeux struct {
	creds *Credentials

	url      string
	client   *http.Client
	clientID string

	messages chan *ConnectResponse

	interval time.Duration
	ctx      context.Context
	mutex    sync.RWMutex
}

// NewBayeux creates a Bayeux client for Salesforce Streaming API consumption.
func NewBayeux(ctx context.Context, credentials *Credentials, apiVersion string, client *http.Client) Bayeux {
	return &bayeux{
		creds: credentials,

		url:    credentials.InstanceURL + "/cometd/" + apiVersion,
		client: client,

		ctx:      ctx,
		interval: 100 * time.Second,
	}
}

func (b *bayeux) Handshake() error {
	payload := `{"channel": "/meta/handshake", "supportedConnectionTypes": ["long-polling"], "version": "1.0"}`

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

func (b *bayeux) connectLoop() error {
	for {
		select {
		case msg := <-b.messages:
			fmt.Printf("received message %v\n", msg)
		case <-b.ctx.Done():
			fmt.Printf("context cancelled\n")
		}
	}
}

// Connect will start to receive events.
// This is a blocking function
func (b *bayeux) Connect() error {
	payload := `{"channel": "/meta/connect", "connectionType": "long-polling", "clientId": "` + b.clientID + `"}`
	res, err := b.doPost(payload)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	var by []byte
	if res.Body != nil {
		by, _ = ioutil.ReadAll(res.Body)
	}
	// Restore the io.ReadCloser to its original state
	res.Body = ioutil.NopCloser(bytes.NewBuffer(by))
	// Use the content
	s := string(by)
	fmt.Printf("\nResponse Body: %s\n\n", s)

	c := []ConnectResponse{}
	err = json.NewDecoder(res.Body).Decode(&c)
	if err != nil {
		return fmt.Errorf("could not decode connect response: %+w", err)
	}

	if len(c) == 0 {
		return errors.New("empty connect response")
	}

	fmt.Printf("\n\nthis is the connect response\n\n%v\n", c)
	fmt.Printf("\n\nthis is the connect response, el0\n\n%v\n", c[0])

	return nil

}

func (b *bayeux) Disconnect() {

}

func (b *bayeux) Subscribe() {

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

// HandshakeResponse for Bayeux protocol
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
		SObject json.RawMessage `json:"sobject"`
	} `json:"data,omitempty"`
	Advice struct {
		Reconnect string `json:"reconnect,omitempty"`
		Timeout   int    `json:"timeout,omitempty"`
		Interval  int    `json:"interval,omitempty"`
	} `json:"advice,omitempty"`
}
