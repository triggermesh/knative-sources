FROM golang:1.14-stretch AS builder

ENV CGO_ENABLED 0
ENV GOOS linux
ENV GOARCH amd64

WORKDIR /go/src/zendesksource-adapter

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN BIN_OUTPUT_DIR=/bin make zendesksource-adapter && \
    mkdir /kodata && \
    mv .git/* /kodata/ && \
    rm -rf ${GOPATH} && \
    rm -rf ${HOME}/.cache

FROM registry.access.redhat.com/ubi8/ubi-minimal

LABEL name "Triggermesh event source adapter for Zendesk"
LABEL vendor "Triggermesh"
LABEL version "v0.1.0"
LABEL release "1"
LABEL summary "The Triggermesh event source adapter for Zendesk"
LABEL description "This is the Triggermesh Zendesk event source adapter for Knative"

# Emulate ko builds
# https://github.com/google/ko/blob/v0.5.0/README.md#including-static-assets
ENV KO_DATA_PATH /kodata

COPY --from=builder /kodata/ ${KO_DATA_PATH}/
COPY --from=builder /bin/zendesksource-adapter /
COPY licenses/ /licenses/

ENTRYPOINT ["/zendesksource-adapter"]
