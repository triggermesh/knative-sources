module github.com/triggermesh/knative-sources

go 1.14

// Top-level module control over the exact version used for important direct dependencies.
// https://github.com/golang/go/wiki/Modules#when-should-i-use-the-replace-directive
replace (
	github.com/cloudevents/sdk-go/v2 => github.com/cloudevents/sdk-go/v2 v2.0.0-RC3
	k8s.io/apimachinery => k8s.io/apimachinery v0.16.8
	k8s.io/client-go => k8s.io/client-go v0.16.8
	k8s.io/code-generator => k8s.io/code-generator v0.16.8
)

require (
	cloud.google.com/go v0.47.0 // indirect
	contrib.go.opencensus.io/exporter/stackdriver v0.13.1 // indirect
	contrib.go.opencensus.io/exporter/zipkin v0.1.1 // indirect
	github.com/aws/aws-sdk-go v1.30.23 // indirect
	github.com/blang/semver v3.5.1+incompatible // indirect
	github.com/cloudevents/sdk-go/v2 v2.0.0-preview8
	github.com/golang/protobuf v1.3.3 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/googleapis/gnostic v0.3.1 // indirect
	github.com/gorilla/mux v1.7.3 // indirect
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/json-iterator/go v1.1.8 // indirect
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/onsi/ginkgo v1.11.0 // indirect
	github.com/onsi/gomega v1.8.1 // indirect
	github.com/openzipkin/zipkin-go v0.2.0 // indirect
	github.com/prometheus/client_golang v1.1.0 // indirect
	github.com/prometheus/common v0.7.0 // indirect
	github.com/robfig/cron v1.2.0 // indirect
	github.com/slack-go/slack v0.6.4
	go.opencensus.io v0.22.3 // indirect
	go.uber.org/zap v1.14.1
	golang.org/x/lint v0.0.0-20200130185559-910be7a94367 // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/sys v0.0.0-20200223170610-d5e6a3e2c0ae // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	golang.org/x/tools v0.0.0-20200214144324-88be01311a71 // indirect
	gomodules.xyz/jsonpatch/v2 v2.1.0 // indirect
	google.golang.org/appengine v1.6.5 // indirect
	k8s.io/api v0.17.4
	k8s.io/apiextensions-apiserver v0.17.3 // indirect
	k8s.io/apimachinery v0.17.4
	k8s.io/client-go v0.17.4
	knative.dev/eventing v0.14.0
	knative.dev/pkg v0.0.0-20200414233146-0eed424fa4ee
)