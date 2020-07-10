module github.com/triggermesh/knative-sources

go 1.14

// Top-level module control over the exact version used for important direct dependencies.
// https://github.com/golang/go/wiki/Modules#when-should-i-use-the-replace-directive
replace (
	k8s.io/apimachinery => k8s.io/apimachinery v0.16.8
	k8s.io/client-go => k8s.io/client-go v0.16.8
	k8s.io/code-generator => k8s.io/code-generator v0.16.8
)

require (
	contrib.go.opencensus.io/exporter/ocagent v0.7.0 // indirect
	contrib.go.opencensus.io/exporter/prometheus v0.2.0 // indirect
	contrib.go.opencensus.io/exporter/stackdriver v0.13.2 // indirect
	github.com/aws/aws-sdk-go v1.33.5 // indirect
	github.com/cloudevents/sdk-go/v2 v2.0.1-0.20200630063327-b91da81265fe
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/nukosuke/go-zendesk v0.7.5
	github.com/prometheus/client_golang v1.7.1 // indirect
	github.com/prometheus/statsd_exporter v0.17.0 // indirect
	github.com/stretchr/testify v1.6.1
	go.uber.org/zap v1.15.0
	golang.org/x/crypto v0.0.0-20200709230013-948cd5f35899 // indirect
	golang.org/x/net v0.0.0-20200707034311-ab3426394381 // indirect
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e // indirect
	google.golang.org/api v0.29.0 // indirect
	google.golang.org/genproto v0.0.0-20200709005830-7a2ca40e9dc3 // indirect
	k8s.io/api v0.18.5
	k8s.io/apiextensions-apiserver v0.17.8 // indirect
	k8s.io/apimachinery v0.18.5
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	k8s.io/klog/v2 v2.3.0 // indirect
	k8s.io/utils v0.0.0-20200619165400-6e3d28b6ed19 // indirect
	knative.dev/eventing v0.16.0
	knative.dev/pkg v0.0.0-20200710003319-43f4f824e3a3
	knative.dev/serving v0.16.0
)
