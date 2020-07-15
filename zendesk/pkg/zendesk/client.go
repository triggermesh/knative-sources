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

package zendesk

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// ListTargetsResponse wraps Target list
type ListTargetsResponse struct {
	Targets []Target `json:"targets"`
	Page
}

// TargetCreate wraps Target creation
type TargetCreate struct {
	Target Target `json:"target"`
}

// TriggerCreate wraps Trigger creation
type TriggerCreate struct {
	Trigger Trigger `json:"trigger"`
}

// ListTriggersResponse wraps Trigger list
type ListTriggersResponse struct {
	Triggers []Trigger `json:"triggers"`
	Page
}

// Client for the Zendesk API subset supported
type Client interface {
	CreateTarget(ctx context.Context, target *Target) (*Target, error)
	ListTargets(ctx context.Context) (*ListTargetsResponse, error)
	DeleteTarget(ctx context.Context, id string) error

	CreateTrigger(ctx context.Context, trigger *Trigger) (*Trigger, error)
	ListTriggers(ctx context.Context) (*ListTriggersResponse, error)
	DeleteTrigger(ctx context.Context, id string) error
}

// NewClient creates default Zendesk API client
func NewClient(email, token, subdomain string, httpClient *http.Client) Client {
	return &client{
		auth:    "Basic " + base64.StdEncoding.EncodeToString([]byte(email+"/token:"+token)),
		baseURL: "https://" + subdomain + ".zendesk.com/api/v2/",
		client:  httpClient,
	}
}

type client struct {
	auth    string
	baseURL string
	client  *http.Client
}

func (c *client) ListTriggers(ctx context.Context) (*ListTriggersResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"triggers.json", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", c.auth)

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode >= 400 {
		return nil, errorFromResponse(res)
	}

	defer res.Body.Close()

	body := &ListTriggersResponse{}
	err = json.NewDecoder(res.Body).Decode(body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (c *client) ListTargets(ctx context.Context) (*ListTargetsResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"targets.json", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", c.auth)

	if err != nil {
		return nil, err
	}

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode >= 400 {
		return nil, errorFromResponse(res)
	}

	defer res.Body.Close()

	body := &ListTargetsResponse{}
	err = json.NewDecoder(res.Body).Decode(body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (c *client) CreateTarget(ctx context.Context, target *Target) (*Target, error) {
	b, err := json.Marshal(&TargetCreate{Target: *target})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"targets.json", bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", c.auth)

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode >= 400 {
		return nil, errorFromResponse(res)
	}

	defer res.Body.Close()

	t := &TargetCreate{}
	err = json.NewDecoder(res.Body).Decode(t)
	if err != nil {
		return nil, err
	}

	return &t.Target, nil
}

func (c *client) DeleteTarget(ctx context.Context, id string) error {
	req, err := http.NewRequestWithContext(ctx, "DELETE", c.baseURL+"targets/"+id+".json", nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", c.auth)

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode >= 400 {
		return errorFromResponse(res)
	}

	return nil
}

func (c *client) CreateTrigger(ctx context.Context, trigger *Trigger) (*Trigger, error) {
	b, err := json.Marshal(&TriggerCreate{Trigger: *trigger})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"triggers.json", bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", c.auth)

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode >= 400 {
		return nil, errorFromResponse(res)
	}

	defer res.Body.Close()

	t := &TriggerCreate{}
	err = json.NewDecoder(res.Body).Decode(t)
	if err != nil {
		return nil, err
	}

	return &t.Trigger, nil
}

func (c *client) DeleteTrigger(ctx context.Context, id string) error {
	req, err := http.NewRequestWithContext(ctx, "DELETE", c.baseURL+"triggers/"+id+".json", nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", c.auth)

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode >= 400 {
		return errorFromResponse(res)
	}

	return nil
}

func errorFromResponse(res *http.Response) error {
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("Zendesk API returned error (%d)", res.StatusCode)
	}

	return fmt.Errorf("Zendesk API returned error (%d): %s", res.StatusCode, string(b))
}
