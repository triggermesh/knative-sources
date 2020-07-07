# Copyright (c) 2020 TriggerMesh Inc.

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

#    http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM golang:1.14-stretch AS builder

ENV CGO_ENABLED 0
ENV GOOS linux
ENV GOARCH amd64

RUN apt update && \
    apt install -y ca-certificates && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /go/src/sources

COPY go.mod go.sum ./
RUN go mod download

# NOTE: try not to modify the instructions above to keep cached layers

COPY ./ .
RUN BIN_OUTPUT_DIR=/bin make -C zendesk controller && \
    rm -rf ${GOPATH} && \
    rm -rf ${HOME}/.cache

FROM registry.access.redhat.com/ubi8/ubi-minimal

LABEL name "Triggermesh Event Source Controller for Zendesk"
LABEL vendor "Triggermesh"
LABEL version "v0.1.0"
LABEL release "1"
LABEL summary "The Triggermesh Event Source Controller for Zendesk"
LABEL description "This is the Triggermesh Knative Event Source Controller for Zendesk"

COPY licenses/ /licenses/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /bin/zendesk-controller /zendesk-controller

ENTRYPOINT ["/zendesk-controller"]
