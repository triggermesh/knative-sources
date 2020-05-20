# Development

## Contents

- [Development](#development)
  - [Contents](#contents)
  - [Building](#building)
  - [Running the controller](#running-the-controller)
    - [Locally](#locally)
    - [Inside a cluster](#inside-a-cluster)
  - [Adding event sources](#adding-event-sources)

## Building

Triggermesh Knative Sources use:

- make
- go 1.14
- go modules

Makefile `help` flag shows all targets per Source

```sh
$ make help

Triggermesh Slack Event Source for Knative
Usage:
  make <source>
  help                 Display this help
  mod-download         Download go modules
  build                Build the binary
  release              Build release binaries
  test                 Run unit tests
  cover                Generate code coverage
  lint                 Lint source files
  vet                  Vet source files
  fmt                  Format source files
  fmt-test             Check source formatting
  image                Builds the container image
  cloudbuild-test      Test container image build with Google Cloud Build
  cloudbuild           Build and publish image to GCR
  clean                Clean build artifacts
  gen-all              Generate all
  gen-deepcopy         Generate deepcopy for API objects
  gen-client           Generate clientset for API objects
  gen-lister           Generate listers for API objects
  gen-informer         Generate informers for API objects
  gen-injection        Generate injection for API objects
  gen-crd              Generate CRD manifests for API objects
```


## Running the controller

### Locally

Providing that the local environment is configured with a valid kubeconfig (either in `~/.kube/config` or set via the
`KUBECONFIG` environment variable), running the controller locally from the current development branch is as simple as
executing

```sh
go run ./cmd/controller
```

> :information_source: The source controller requires a few environment variables to be exported in order to start.
>
> `SYSTEM_NAMESPACE`
>
> The namespace in which the controller sources its configuration from (logging, observability). This can potentially be
> set to any namespace, including `default`, since the controller falls back to a default configuration if the
> aforementioned ConfigMaps are missing.
>
> `METRICS_DOMAIN`
>
> The domain to use for surfacing metrics.
>
> **Optional**. Only required when a ConfigMap for observability (`config-observability` by default) actually exists in
> the namespace defined by `SYSTEM_NAMESPACE`.

### Inside a cluster

One can build/push container images and deploy all relevant Kubernetes objects to a running cluster in a single command
using [ko](https://github.com/google/ko).

```sh
$ ko apply --local -f ./config/
...
2020/04/07 13:44:00 Using base gcr.io/distroless/static:latest for github.com/triggermesh/knative-sources/slack/cmd/controller
2020/04/07 13:44:01 Building github.com/triggermesh/knative-sources/slack/cmd/controller
2020/04/07 13:44:05 Loading ko.local/slack-controller-0d0554a556
2020/04/07 13:44:06 Loaded ko.local/slack-source-controller-0d0554a556
2020/04/07 13:44:06 Adding tag latest
2020/04/07 13:44:06 Added tag latest
deployment.apps/slack-source-controller created
```

The controller will be deployed to the `triggermesh` namespace.

```console
$ kubectl -n triggermesh get deployment/slack-source-controller
NAME                           READY   UP-TO-DATE   AVAILABLE   AGE
slack-source-controller   1/1     1            1           1m
```

> :information_source: Although `ko` does not make use of the `docker` client, the `--local` flag assumes the
> development environment is properly configured to import images in a Docker daemon. On `minikube`, for example, it is
> assumed that the environment variables defined by `minikube docker-env` are exported.
>
> Please refer to the `ko` documentation for further usage instructions.

## Adding event sources

Refer to Knative's [sample source](https://github.com/knative/sample-source) on how to develop event sources