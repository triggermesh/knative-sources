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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

const (
	// grant type for OAuth JWT.
	// See: https://tools.ietf.org/html/rfc7523#page-10
	grantJWT = "urn:ietf:params:oauth:grant-type:jwt-bearer"

	oauthTokenPath = "/services/oauth2/token"
)

type claims struct {
	jwt.StandardClaims
}

// AuthenticateCredentialsJWT connects to the remote authentication endpoint and processes
// the request
func AuthenticateCredentialsJWT(certKey, clientID, user, server string, client *http.Client) (*Credentials, error) {
	audience := strings.TrimSuffix(server, "/")

	signKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(certKey))
	if err != nil {
		return nil, fmt.Errorf("unable to parse PEM private key: %w", err)
	}

	claims := &claims{
		StandardClaims: jwt.StandardClaims{
			Issuer:   clientID,
			Subject:  user,
			Audience: audience,
			// expiry needs to be set to 3 minutes or less
			// See: https://help.salesforce.com/articleView?id=remoteaccess_oauth_jwt_flow.htm
			ExpiresAt: time.Now().Add(time.Minute * 3).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	tokenString, err := token.SignedString(signKey)
	if err != nil {
		return nil, fmt.Errorf("could not sign JWT token: %w", err)
	}

	form := url.Values{}
	form.Add("grant_type", grantJWT)
	form.Add("assertion", tokenString)

	authURL := audience + oauthTokenPath
	req, err := http.NewRequest("POST", authURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("could not build authentication request: %w", err)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not execute authentication request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		msg := fmt.Sprintf("received unexpected status code %d from authentication", res.StatusCode)
		resb, err := ioutil.ReadAll(res.Body)
		if err != nil {
			msg += ": " + string(resb)
		}
		return nil, errors.New(msg)
	}

	c := &Credentials{}
	err = json.NewDecoder(res.Body).Decode(c)
	if err != nil {
		return nil, fmt.Errorf("could not decode authentication response into Salesforce grant: %w", err)
	}

	return c, nil
}
