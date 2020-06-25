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

// import (
// 	cloudevents "github.com/cloudevents/sdk-go/v2"
// )

package adapter

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	cloudeventst "github.com/cloudevents/sdk-go/v2/client/test"
	"github.com/stretchr/testify/assert"
	zapt "go.uber.org/zap/zaptest"
)

func TestSlackChallenge(t *testing.T) {

	logger := zapt.NewLogger(t).Sugar()
	ceClient, chEvent := cloudeventst.NewMockSenderClient(t, 1)

	tc := map[string]struct {
		body  io.Reader
		appID string
		token string

		expectedCode     int
		expectedContains string

		expectedEventID   string
		expectedEventData string
	}{
		"nil body": {
			body: nil,

			expectedCode:     http.StatusInternalServerError,
			expectedContains: "request without body not supported",
		},
		"not a JSON message": {
			body: read("this is not an expected message"),

			expectedCode:     http.StatusInternalServerError,
			expectedContains: "could not unmarshall JSON request:",
		},

		"not an expected message": {
			body: read(`{"hello":"world"}`),

			expectedCode: http.StatusOK,
		},

		"url verification": {
			body: read(`
			{
				"token": "Jhj5dZrVaK7ZwHHjRyZWjbDl",
				"challenge": "3eZbrw1aBm2rZgRNFdxV2595E9CY3gmdALWMmHkvFXO7tYXAYM8P",
				"type": "url_verification"
			}`),

			expectedCode:     http.StatusOK,
			expectedContains: `{"challenge":"3eZbrw1aBm2rZgRNFdxV2595E9CY3gmdALWMmHkvFXO7tYXAYM8P"}`,
		},

		"wrong App ID": {
			body: read(`
			{
        "token": "XXYYZZ",
        "team_id": "TXXXXXXXX",
        "api_app_id": "AXXXXXXXXX",
        "type": "event_callback",
        "event_id": "Ev08MFMKH6",
        "event_time": 1234567890
			}`),
			appID: "ZYYYYYYYYYY",

			expectedCode: http.StatusOK,
		},

		"wrong Token": {
			body: read(`
			{
        "token": "XXYYZZ",
        "team_id": "TXXXXXXXX",
        "api_app_id": "AXXXXXXXXX",
        "type": "event_callback",
        "event_id": "Ev08MFMKH6",
        "event_time": 1234567890
			}`),
			token: "AABBCC",

			expectedCode: http.StatusOK,
		},

		"handle callback": {
			body: read(`
			{
        "token": "XXYYZZ",
        "team_id": "TXXXXXXXX",
        "api_app_id": "AXXXXXXXXX",
        "event": {
                "type": "name_of_event",
                "event_ts": "1234567890.123456",
                "user": "UXXXXXXX1"
        },
        "type": "event_callback",
        "authed_users": [
                "UXXXXXXX1",
                "UXXXXXXX2"
        ],
        "event_id": "Ev08MFMKH6",
        "event_time": 1234567890
			}`),

			expectedCode:      http.StatusOK,
			expectedEventID:   "Ev08MFMKH6",
			expectedEventData: `{"event_ts":"1234567890.123456","type":"name_of_event","user":"UXXXXXXX1"}`,
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			handler := &slackEventAPIHandler{
				appID:    c.appID,
				token:    c.token,
				ceClient: ceClient,
				logger:   logger,
			}

			req, _ := http.NewRequest("GET", "/", c.body)

			rr := httptest.NewRecorder()
			th := http.HandlerFunc(handler.handleAll)

			th.ServeHTTP(rr, req)

			assert.Equal(t, c.expectedCode, rr.Code, "unexpected response code")
			assert.Contains(t, rr.Body.String(), c.expectedContains, "could not find expected response")

			if c.expectedEventID != "" {
				select {
				case event := <-chEvent:
					assert.Equal(t, c.expectedEventID, event.ID(), "event ID does not match")
					assert.Equal(t, c.expectedEventData, string(event.Data()), "event Data does not match")

				case <-time.After(1 * time.Second):
					assert.Fail(t, "expected cloud event by ID %q was not sent", c.expectedEventID)
				}
			}
		})
	}
}

func read(s string) io.Reader {
	return strings.NewReader(s)
}
